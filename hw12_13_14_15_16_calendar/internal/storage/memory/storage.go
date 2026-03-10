package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

var _ storage.EventStorage = (*InMemoryStorage)(nil)

type InMemoryStorage struct {
	mu     sync.RWMutex
	events map[uuid.UUID]storage.Event
}

func New() *InMemoryStorage {
	return &InMemoryStorage{
		events: make(map[uuid.UUID]storage.Event),
	}
}

func (s *InMemoryStorage) Create(_ context.Context, event storage.Event) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; exists {
		return storage.Event{}, storage.ErrEventAlreadyExists
	}

	s.events[event.ID] = event
	return event, nil
}

func (s *InMemoryStorage) Update(_ context.Context, id uuid.UUID, event storage.Event) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.Event{}, storage.ErrEventNotFound
	}

	event.ID = id // гарантируем, что ID не изменился
	s.events[id] = event
	return event, nil
}

func (s *InMemoryStorage) Delete(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *InMemoryStorage) GetByID(_ context.Context, id uuid.UUID) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return storage.Event{}, storage.ErrEventNotFound
	}

	return event, nil
}

func (s *InMemoryStorage) ListByDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	year, month, day := date.Date()

	for _, event := range s.events {
		ey, em, ed := event.EventTime.Date()
		if ey == year && em == month && ed == day {
			result = append(result, event)
		}
	}

	return result, nil
}

func (s *InMemoryStorage) ListByWeek(_ context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Находим начало и конец недели
	weekday := date.Weekday()
	var startOfWeek time.Time
	switch weekday { //nolint:exhaustive
	case time.Sunday:
		startOfWeek = date.AddDate(0, 0, -6)
	default:
		startOfWeek = date.AddDate(0, 0, -int(weekday)+1)
	}

	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, date.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	var result []storage.Event
	for _, event := range s.events {
		if (event.EventTime.After(startOfWeek) || event.EventTime.Equal(startOfWeek)) &&
			event.EventTime.Before(endOfWeek) {
			result = append(result, event)
		}
	}

	return result, nil
}

func (s *InMemoryStorage) ListByMonth(_ context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	year, month, _ := date.Date()

	for _, event := range s.events {
		ey, em, _ := event.EventTime.Date()
		if ey == year && em == month {
			result = append(result, event)
		}
	}

	return result, nil
}
