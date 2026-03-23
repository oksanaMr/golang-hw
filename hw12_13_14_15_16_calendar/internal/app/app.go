package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	storage storage.EventStorage
}

func New(storage storage.EventStorage) *App {
	return &App{
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id uuid.UUID, title string) (storage.Event, error) {
	return a.storage.Create(ctx, storage.Event{ID: id, Title: title})
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) (storage.Event, error) {
	return a.storage.Update(ctx, id, event)
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	return a.storage.Delete(ctx, id)
}

func (a *App) GetByID(ctx context.Context, id uuid.UUID) (storage.Event, error) {
	return a.storage.GetByID(ctx, id)
}

func (a *App) ListByDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListByDay(ctx, date)
}

func (a *App) ListByWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListByWeek(ctx, date)
}

func (a *App) ListByMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListByMonth(ctx, date)
}
