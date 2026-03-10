package storage //nolint:gci,gofmt,gofumpt

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type EventStorage interface {
	Create(ctx context.Context, event Event) (Event, error)

	Update(ctx context.Context, id uuid.UUID, event Event) (Event, error)

	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (Event, error)

	ListByDay(ctx context.Context, date time.Time) ([]Event, error)

	ListByWeek(ctx context.Context, date time.Time) ([]Event, error)

	ListByMonth(ctx context.Context, date time.Time) ([]Event, error)

	// ListByUser(ctx context.Context, userID uuid.UUID) ([]Event, error)
	// ListByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]Event, error)
}

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrInvalidEventData   = errors.New("invalid event data")
)
