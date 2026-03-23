package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	// ID - уникальный идентификатор события (UUID)
	ID uuid.UUID `json:"id" db:"id"`

	// Title - заголовок события (короткий текст)
	Title string `json:"title" db:"title" validate:"required,min=1,max=255"`

	// EventTime - дата и время события
	EventTime time.Time `json:"eventTime" db:"event_time" validate:"required"`

	// Duration - длительность события
	Duration time.Duration `json:"duration" db:"duration"`

	// Description - описание события (длинный текст, опционально)
	Description string `json:"description,omitempty" db:"description"`

	// UserID - ID пользователя, владельца события
	UserID uuid.UUID `json:"userId" db:"user_id" validate:"required"`

	// NotifyBefore - за сколько времени высылать уведомление (опционально)
	// Значение 0 означает, что уведомление не требуется
	NotifyBefore time.Duration `json:"notifyBefore,omitempty" db:"notify_before"`
}

func NewEvent(title string, eventTime time.Time, duration time.Duration,
	description string, userID uuid.UUID, notifyBefore time.Duration,
) *Event {
	return &Event{
		ID:           uuid.New(),
		Title:        title,
		EventTime:    eventTime,
		Duration:     duration,
		Description:  description,
		UserID:       userID,
		NotifyBefore: notifyBefore,
	}
}

func (e *Event) EndTime() time.Time {
	return e.EventTime.Add(e.Duration)
}

func (e *Event) ShouldNotify() bool {
	return e.NotifyBefore > 0
}

func (e *Event) NotificationTime() time.Time {
	return e.EventTime.Add(-e.NotifyBefore)
}

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
)
