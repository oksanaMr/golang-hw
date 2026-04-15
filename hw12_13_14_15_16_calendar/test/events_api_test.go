// tests/integration/events_api_test.go
package test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Event соответствует структуре ответа
type Event struct {
	Description  *string   `json:"description,omitempty"`
	Duration     string    `json:"duration"`
	EventTime    time.Time `json:"eventTime"`
	Id           string    `json:"id"`
	NotifyBefore *string   `json:"notifyBefore,omitempty"`
	Title        string    `json:"title"`
	UserId       string    `json:"userId"`
}

// ErrorResponse структура ошибки
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

func TestEventsAPI(t *testing.T) {
	httpClient := NewTestHTTPClient()

	t.Run("create event - success with all fields", func(t *testing.T) {
		t.Log("Очистка таблиц")
		if err := truncateDB(); err != nil {
			t.Fatalf("Failed to trancate database: %v", err)
		}

		description := "Integration test event description"
		notifyBefore := "30m"

		event := CreateEventRequest{
			Title:        "Integration Test Event",
			Description:  &description,
			EventTime:    time.Now().Add(24 * time.Hour),
			Duration:     "1h30m",
			UserId:       uuid.New().String(),
			NotifyBefore: &notifyBefore,
		}

		resp, err := httpClient.Post("/events", event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
		}

		// Проверяем структуру ответа
		var result Event
		if err := httpClient.ParseResponse(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Проверяем все поля
		if result.Id == "" {
			t.Error("Expected event ID in response")
		}

		if _, err := uuid.Parse(result.Id); err != nil {
			t.Errorf("Event ID is not a valid UUID: %v", err)
		}

		if result.Title != event.Title {
			t.Errorf("Expected title %s, got %s", event.Title, result.Title)
		}

		if result.Duration == "" {
			t.Errorf("Expected duration %s, got %s", event.Duration, result.Duration)
		}

		if !result.EventTime.Truncate(time.Second).Equal(event.EventTime.Truncate(time.Second)) {
			t.Errorf("Expected event time %v, got %v", event.EventTime, result.EventTime)
		}

		if result.UserId != event.UserId {
			t.Errorf("Expected user ID %s, got %s", event.UserId, result.UserId)
		}

		if result.Description == nil || *result.Description != *event.Description {
			t.Errorf("Expected description %v, got %v", *event.Description, result.Description)
		}

		if result.NotifyBefore == nil {
			t.Errorf("Expected notifyBefore %v, got %v", *event.NotifyBefore, result.NotifyBefore)
		}

		t.Logf("Created event with ID: %s", result.Id)
	})

	t.Run("create event - validation error (missing required fields)", func(t *testing.T) {
		// Отправляем неполные данные
		invalidEvent := map[string]interface{}{
			"title": "Invalid Event",
			// Отсутствуют обязательные поля: eventTime, duration, userId
		}

		resp, err := httpClient.Post("/events", invalidEvent)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	t.Run("create event - validation error (invalid duration format)", func(t *testing.T) {
		event := CreateEventRequest{
			Title:     "Invalid Duration Event",
			EventTime: time.Now().Add(24 * time.Hour),
			Duration:  "invalid-duration",
			UserId:    uuid.New().String(),
		}

		resp, err := httpClient.Post("/events", event)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	t.Run("create event - validation error (invalid notifyBefore format)", func(t *testing.T) {
		invalidNotify := "invalid"

		event := CreateEventRequest{
			Title:        "Invalid Notify Event",
			EventTime:    time.Now().Add(24 * time.Hour),
			Duration:     "1h",
			UserId:       uuid.New().String(),
			NotifyBefore: &invalidNotify,
		}

		resp, err := httpClient.Post("/events", event)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})
}

func TestEventListingAPI(t *testing.T) {
	httpClient := NewTestHTTPClient()

	// Подготавливаем тестовые данные
	t.Run("setup test events", func(t *testing.T) {
		t.Log("Очистка таблиц")
		if err := truncateDB(); err != nil {
			t.Fatalf("Failed to trancate database: %v", err)
		}

		baseTime, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")

		events := []CreateEventRequest{
			{
				Title:     "Today Event",
				EventTime: baseTime,
				Duration:  "2h",
				UserId:    uuid.New().String(),
			},
			{
				Title:     "Tomorrow Event",
				EventTime: baseTime.AddDate(0, 0, 1),
				Duration:  "3h",
				UserId:    uuid.New().String(),
			},
			{
				Title:     "Yesterday Event",
				EventTime: baseTime.AddDate(0, 0, -1),
				Duration:  "1h30m",
				UserId:    uuid.New().String(),
			},
			{
				Title:     "Next Week Event",
				EventTime: baseTime.AddDate(0, 0, 10),
				Duration:  "2h",
				UserId:    uuid.New().String(),
			},
			{
				Title:     "Next Month Event",
				EventTime: baseTime.AddDate(0, 1, 0),
				Duration:  "1h",
				UserId:    uuid.New().String(),
			},
		}

		for _, event := range events {
			resp, err := httpClient.Post("/events", event)
			if err != nil {
				t.Fatalf("Failed to create event: %v", err)
			}

			var result Event
			if err := httpClient.ParseResponse(resp, &result); err != nil {
				t.Fatalf("Failed to parse error: %v", err)
			}
			t.Logf("Created event: %s - %s", result.Id, result.Title)
			resp.Body.Close()
		}
	})

	t.Run("get events for day", func(t *testing.T) {
		baseTime, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")
		dateStr := baseTime.Format("2006-01-02")

		resp, err := httpClient.Get("/events?date=" + dateStr)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var events []Event
		if err := httpClient.ParseResponse(resp, &events); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 event for today, got %d", len(events))
		}
	})

	t.Run("get events for week", func(t *testing.T) {
		baseTime, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")
		dateStr := baseTime.Format("2006-01-02")

		resp, err := httpClient.Get("/events/week/" + dateStr)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var events []Event
		if err := httpClient.ParseResponse(resp, &events); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(events) != 3 {
			t.Errorf("Expected 3 events for today, got %d", len(events))
		}
	})

	t.Run("get events for month", func(t *testing.T) {
		baseTime, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")
		dateStr := baseTime.Format("2006-01-02")

		resp, err := httpClient.Get("/events/month/" + dateStr)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var events []Event
		if err := httpClient.ParseResponse(resp, &events); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(events) != 4 {
			t.Errorf("Expected 4 events for today, got %d", len(events))
		}
	})
}
