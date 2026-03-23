package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
)

type EventStorage interface {
	Create(ctx context.Context, event model.Event) (model.Event, error)

	Update(ctx context.Context, id uuid.UUID, event model.Event) (model.Event, error)

	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (model.Event, error)

	ListByDay(ctx context.Context, date time.Time) ([]model.Event, error)

	ListByWeek(ctx context.Context, date time.Time) ([]model.Event, error)

	ListByMonth(ctx context.Context, date time.Time) ([]model.Event, error)

	// ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Event, error)
	// ListByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Event, error)
}
