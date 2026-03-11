package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_Create(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := model.Event{
		ID:        uuid.New(),
		Title:     "Test Event",
		EventTime: time.Now(),
	}

	// Тест успешного создания
	created, err := s.Create(ctx, event)
	assert.NoError(t, err)
	assert.Equal(t, event, created)

	// Тест создания события с существующим ID
	_, err = s.Create(ctx, event)
	assert.ErrorIs(t, err, model.ErrEventAlreadyExists)
}

func TestInMemoryStorage_GetByID(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := model.Event{
		ID:        uuid.New(),
		Title:     "Test Event",
		EventTime: time.Now(),
	}
	_, err := s.Create(ctx, event)
	assert.NoError(t, err)

	// Тест получения существующего события
	found, err := s.GetByID(ctx, event.ID)
	assert.NoError(t, err)
	assert.Equal(t, event, found)

	// Тест получения несуществующего события
	_, err = s.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, model.ErrEventNotFound)
}

func TestInMemoryStorage_Update(t *testing.T) {
	s := New()
	ctx := context.Background()

	originalEvent := model.Event{
		ID:        uuid.New(),
		Title:     "Original Title",
		EventTime: time.Now(),
	}
	_, err := s.Create(ctx, originalEvent)
	assert.NoError(t, err)

	updatedEvent := model.Event{
		Title:     "Updated Title",
		EventTime: originalEvent.EventTime.Add(time.Hour),
	}

	// Тест успешного обновления
	result, err := s.Update(ctx, originalEvent.ID, updatedEvent)
	assert.NoError(t, err)
	assert.Equal(t, originalEvent.ID, result.ID)
	assert.Equal(t, updatedEvent.Title, result.Title)
	assert.Equal(t, updatedEvent.EventTime, result.EventTime)

	// Проверяем, что событие действительно обновилось
	found, err := s.GetByID(ctx, originalEvent.ID)
	assert.NoError(t, err)
	assert.Equal(t, result, found)

	// Тест обновления несуществующего события
	_, err = s.Update(ctx, uuid.New(), updatedEvent)
	assert.ErrorIs(t, err, model.ErrEventNotFound)
}

func TestInMemoryStorage_Delete(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := model.Event{
		ID:        uuid.New(),
		Title:     "Test Event",
		EventTime: time.Now(),
	}
	_, err := s.Create(ctx, event)
	assert.NoError(t, err)

	// Тест успешного удаления
	err = s.Delete(ctx, event.ID)
	assert.NoError(t, err)

	// Проверяем, что событие удалено
	_, err = s.GetByID(ctx, event.ID)
	assert.ErrorIs(t, err, model.ErrEventNotFound)

	// Тест удаления несуществующего события
	err = s.Delete(ctx, uuid.New())
	assert.ErrorIs(t, err, model.ErrEventNotFound)
}

func TestInMemoryStorage_ListByDay(t *testing.T) {
	s := New()
	ctx := context.Background()
	now := time.Now()

	today := model.Event{
		ID:        uuid.New(),
		Title:     "Today Event",
		EventTime: now,
	}
	tomorrow := model.Event{
		ID:        uuid.New(),
		Title:     "Tomorrow Event",
		EventTime: now.AddDate(0, 0, 1),
	}
	yesterday := model.Event{
		ID:        uuid.New(),
		Title:     "Yesterday Event",
		EventTime: now.AddDate(0, 0, -1),
	}

	_, _ = s.Create(ctx, today)
	_, _ = s.Create(ctx, tomorrow)
	_, _ = s.Create(ctx, yesterday)

	// Тест получения событий за сегодня
	events, err := s.ListByDay(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, today.ID, events[0].ID)

	// Тест получения событий за завтра
	events, err = s.ListByDay(ctx, now.AddDate(0, 0, 1))
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, tomorrow.ID, events[0].ID)

	// Тест для даты без событий
	events, err = s.ListByDay(ctx, now.AddDate(0, 0, 5))
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestInMemoryStorage_ListByWeek(t *testing.T) {
	s := New()
	ctx := context.Background()
	now, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")

	event1 := model.Event{
		ID:        uuid.New(),
		Title:     "Event 1",
		EventTime: now,
	}
	event2 := model.Event{
		ID:        uuid.New(),
		Title:     "Event 2",
		EventTime: now.AddDate(0, 0, 2), // +2 дня
	}
	event3 := model.Event{
		ID:        uuid.New(),
		Title:     "Event 3",
		EventTime: now.AddDate(0, 0, -2), // -2 дня
	}
	nextWeek := model.Event{
		ID:        uuid.New(),
		Title:     "Next Week Event",
		EventTime: now.AddDate(0, 0, 10), // следующая неделя
	}

	_, _ = s.Create(ctx, event1)
	_, _ = s.Create(ctx, event2)
	_, _ = s.Create(ctx, event3)
	_, _ = s.Create(ctx, nextWeek)

	// Тест получения событий за текущую неделю
	events, err := s.ListByWeek(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, events, 3) // события на этой неделе
}

func TestInMemoryStorage_ListByMonth(t *testing.T) {
	s := New()
	ctx := context.Background()
	now, _ := time.Parse("2006-01-02 15:04:05", "2026-03-06 15:04:05")

	thisMonth := model.Event{
		ID:        uuid.New(),
		Title:     "This Month",
		EventTime: now,
	}
	thisMonth2 := model.Event{
		ID:        uuid.New(),
		Title:     "This Month 2",
		EventTime: now.AddDate(0, 0, 15),
	}
	nextMonth := model.Event{
		ID:        uuid.New(),
		Title:     "Next Month",
		EventTime: now.AddDate(0, 1, 0),
	}
	lastMonth := model.Event{
		ID:        uuid.New(),
		Title:     "Last Month",
		EventTime: now.AddDate(0, -1, 0),
	}

	_, _ = s.Create(ctx, thisMonth)
	_, _ = s.Create(ctx, thisMonth2)
	_, _ = s.Create(ctx, nextMonth)
	_, _ = s.Create(ctx, lastMonth)

	// Тест получения событий за текущий месяц
	events, err := s.ListByMonth(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, events, 2)

	// Тест получения событий за следующий месяц
	events, err = s.ListByMonth(ctx, now.AddDate(0, 1, 0))
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, nextMonth.ID, events[0].ID)

	// Тест для месяца без событий
	events, err = s.ListByMonth(ctx, now.AddDate(0, 2, 0))
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	s := New()
	ctx := context.Background()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			event := model.Event{
				ID:        uuid.New(),
				Title:     "Concurrent Event",
				EventTime: time.Now(),
			}
			_, err := s.Create(ctx, event)
			assert.NoError(t, err)

			_, err = s.GetByID(ctx, event.ID)
			assert.NoError(t, err)

			err = s.Delete(ctx, event.ID)
			assert.NoError(t, err)

			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	events, err := s.ListByMonth(ctx, time.Now())
	assert.NoError(t, err)
	assert.Empty(t, events)
}
