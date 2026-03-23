package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	storage storage.EventStorage
	logger  *logger.Logger
}

func New(logger *logger.Logger, storage storage.EventStorage) *App {
	return &App{
		storage: storage,
		logger:  logger,
	}
}

func (a *App) CreateEvent(ctx context.Context, duration string, description, notifyBefore *string, eventTime time.Time, title string, userId uuid.UUID) (model.Event, error) {
	event := model.Event{
		ID:        uuid.New(),
		Title:     title,
		EventTime: eventTime,
		Duration:  parseDuration(duration),
		UserID:    uuid.UUID(userId),
	}

	if description != nil {
		event.Description = *description
	}
	if notifyBefore != nil {
		event.NotifyBefore = parseDuration(*notifyBefore)
	}

	result, err := a.storage.Create(ctx, event)
	if err != nil {
		a.logger.Error("Error create event", "error", err)
	}

	return result, err
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID,
	description, duration, notifyBefore, title *string, eventTime *time.Time,
	userId *uuid.UUID,
) (model.Event, error) {
	// Получаем существующее событие
	existingEvent, err := a.storage.GetByID(ctx, id)
	if err != nil {
		return existingEvent, err
	}

	// Обновляем поля
	if title != nil {
		existingEvent.Title = *title
	}
	if eventTime != nil {
		existingEvent.EventTime = *eventTime
	}
	if duration != nil {
		existingEvent.Duration = parseDuration(*duration)
	}
	if description != nil {
		existingEvent.Description = *description
	}
	if userId != nil {
		existingEvent.UserID = *userId
	}
	if notifyBefore != nil {
		existingEvent.NotifyBefore = parseDuration(*notifyBefore)
	}

	result, err := a.storage.Update(ctx, id, existingEvent)
	if err != nil {
		a.logger.Error("Error update event", "error", err)
	}

	return result, err
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	err := a.storage.Delete(ctx, id)
	if err != nil {
		a.logger.Error("Error delete event", "error", err)
	}

	return err
}

func (a *App) GetByID(ctx context.Context, id uuid.UUID) (model.Event, error) {
	result, err := a.storage.GetByID(ctx, id)
	if err != nil {
		a.logger.Error("Error getting event", "error", err)
	}

	return result, err
}

func (a *App) ListByDay(ctx context.Context, date time.Time) ([]model.Event, error) {
	result, err := a.storage.ListByDay(ctx, date)
	if err != nil {
		a.logger.Error("Error getting events by day", "error", err)
	}

	return result, err
}

func (a *App) ListByWeek(ctx context.Context, date time.Time) ([]model.Event, error) {
	result, err := a.storage.ListByWeek(ctx, date)
	if err != nil {
		a.logger.Error("Error getting events by week", "error", err)
	}

	return result, err
}

func (a *App) ListByMonth(ctx context.Context, date time.Time) ([]model.Event, error) {
	result, err := a.storage.ListByMonth(ctx, date)
	if err != nil {
		a.logger.Error("Error getting events by month", "error", err)
	}

	return result, err
}

func parseDuration(dur string) time.Duration {
	if dur == "" {
		return 0
	}
	d, _ := time.ParseDuration(dur)
	return d
}
