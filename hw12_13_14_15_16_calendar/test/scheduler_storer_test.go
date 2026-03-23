package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DBEvent struct {
	ID           string
	Title        string
	EventTime    time.Time
	Duration     time.Duration
	Description  string
	UserID       string
	NotifyBefore time.Duration
}

type CreateEventRequest struct {
	Description  *string   `json:"description,omitempty"`
	Duration     string    `json:"duration"`
	EventTime    time.Time `json:"eventTime"`
	NotifyBefore *string   `json:"notifyBefore,omitempty"`
	Title        string    `json:"title"`
	UserId       string    `json:"userId"`
}

func TestSchedulerProcessesOnlyFutureEvents(t *testing.T) {
	t.Log("Очистка таблиц")
	if err := truncateDB(); err != nil {
		t.Fatalf("Failed to trancate database: %v", err)
	}

	now := time.Now()
	nextMinute := now.Truncate(time.Second).Add(time.Minute)

	t.Logf("Текущее время: %v", now.Format("15:04:05"))
	t.Logf("Следующая минута: %v", nextMinute.Format("15:04"))

	// Событие для следующей минуты (ДОЛЖНО обработаться)
	currentEvent := createEventForNotificationTime(t, nextMinute)

	// Событие для через 5 минут (НЕ должно обработаться)
	futureTime := nextMinute.Add(5 * time.Minute)
	futureEvent := createEventForNotificationTime(t, futureTime)

	// Событие для 5 минут назад (НЕ должно обработаться)
	pastTime := nextMinute.Add(-5 * time.Minute)
	pastEvent := createEventForNotificationTime(t, pastTime)

	insertDBEvent(t, currentEvent)
	insertDBEvent(t, futureEvent)
	insertDBEvent(t, pastEvent)

	t.Logf("Событие для текущей минуты: %v", currentEvent.ID)
	t.Logf("Событие для будущей минуты: %v (в %v)", futureEvent.ID, futureTime.Format("15:04"))
	t.Logf("Событие для прошлой минуты: %v (в %v)", pastEvent.ID, pastTime.Format("15:04"))

	time.Sleep(2 * time.Minute)
	time.Sleep(5 * time.Second) // Даем время на обработку

	// Проверяем уведомления
	var notificationCount int

	err := testDB.QueryRow(`
        SELECT COUNT(*) FROM notifications WHERE event_id = $1
    `, currentEvent.ID).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 1, notificationCount, "Должно быть уведомление для текущего события")

	err = testDB.QueryRow(`
        SELECT COUNT(*) FROM notifications WHERE event_id = $1
    `, futureEvent.ID).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount, "Не должно быть уведомления для будущего события")

	err = testDB.QueryRow(`
        SELECT COUNT(*) FROM notifications WHERE event_id = $1
    `, pastEvent.ID).Scan(&notificationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, notificationCount, "Не должно быть уведомления для прошлого события")
}

func TestAPIAndSchedulerIntegration(t *testing.T) {
	t.Log("Очистка таблиц")
	if err := truncateDB(); err != nil {
		t.Fatalf("Failed to trancate database: %v", err)
	}

	now := time.Now()
	nextMinute := now.Truncate(time.Second).Add(time.Minute)

	uniqueTitle := fmt.Sprintf("API Event %d", now.UnixNano())
	description := "Created via API test"
	notifyBefore := "15m"
	event := CreateEventRequest{
		Title:        uniqueTitle,
		EventTime:    nextMinute.Add(15 * time.Minute), // событие через 15 минут после уведомления
		Duration:     "1h",
		Description:  &description,
		UserId:       uuid.New().String(),
		NotifyBefore: &notifyBefore,
	}

	// Отправляем запрос к API
	body, err := json.Marshal(event)
	require.NoError(t, err)

	t.Logf("Отправка запроса к %s/events", apiBaseURL)
	t.Logf("Body: %s", string(body))

	resp, err := http.Post(
		fmt.Sprintf("%s/events", apiBaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	t.Logf("API response status: %d, body: %s", resp.StatusCode, string(responseBody))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	if resp.StatusCode != http.StatusCreated {
		t.FailNow()
	}

	time.Sleep(2 * time.Minute)
	time.Sleep(5 * time.Second)

	// ищем уведомление по заголовку события
	var notificationID string
	var notificationTitle string
	var notificationEventID string

	err = testDB.QueryRow(`
        SELECT n.id, n.title, n.event_id
        FROM notifications n
        WHERE n.title = $1
    `, uniqueTitle).Scan(&notificationID, &notificationTitle, &notificationEventID)

	require.NoError(t, err, "Должно быть найдено уведомление для созданного события")
	assert.NotEmpty(t, notificationID)
	assert.Equal(t, uniqueTitle, notificationTitle)
}

func createEventForNotificationTime(t *testing.T, targetTime time.Time) DBEvent {
	notifyBefore := 15 * time.Minute

	event := DBEvent{
		ID:           uuid.New().String(),
		Title:        fmt.Sprintf("Event for %s", targetTime.Format("15:04")),
		EventTime:    targetTime.Add(notifyBefore),
		Duration:     1 * time.Hour,
		Description:  "Test event",
		UserID:       uuid.New().String(),
		NotifyBefore: notifyBefore,
	}

	return event
}

func insertDBEvent(t *testing.T, event DBEvent) {
	_, err := testDB.Exec(`
        INSERT INTO events (id, title, event_time, duration, description, user_id, notify_before)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, event.ID, event.Title, event.EventTime, int64(event.Duration),
		event.Description, event.UserID, int64(event.NotifyBefore))
	require.NoError(t, err)

	// Для отладки выводим время уведомления
	notifyTime := event.EventTime.Add(-event.NotifyBefore)
	t.Logf("Событие %s: EventTime=%v, NotifyBefore=%v, NotificationTime=%v",
		event.ID[:8], event.EventTime.Format("15:04:05"),
		event.NotifyBefore, notifyTime.Format("15:04:05"))
}
