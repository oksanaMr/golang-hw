package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" //nolint:revive
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

var _ storage.Storage = (*PostgresStorage)(nil)

type PostgresStorage struct {
	db *sql.DB
}

func New(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) Create(ctx context.Context, event model.Event) (model.Event, error) {
	query := `
	INSERT INTO events (
		id, title, event_time, duration, description, user_id, notify_before
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, title, event_time, duration, description, user_id, notify_before
	`

	// Генерируем новый UUID, если не задан
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	var created model.Event
	err := s.db.QueryRowContext(
		ctx, query,
		event.ID, event.Title, event.EventTime,
		event.Duration, event.Description, event.UserID,
		event.NotifyBefore,
	).Scan(
		&created.ID, &created.Title, &created.EventTime,
		&created.Duration, &created.Description, &created.UserID,
		&created.NotifyBefore,
	)
	if err != nil {
		// Проверяем на нарушение уникальности
		if isDuplicateError(err) {
			return model.Event{}, model.ErrEventAlreadyExists
		}
		return model.Event{}, fmt.Errorf("failed to create event: %w", err)
	}

	return created, nil
}

func (s *PostgresStorage) Update(ctx context.Context, id uuid.UUID, event model.Event) (model.Event, error) {
	query := `
	UPDATE events 
	SET title = $2, event_time = $3, duration = $4, 
		description = $5, user_id = $6, notify_before = $7,
		updated_at = NOW()
	WHERE id = $1
	RETURNING id, title, event_time, duration, description, user_id, notify_before
	`

	var updated model.Event
	err := s.db.QueryRowContext(
		ctx, query,
		id, event.Title, event.EventTime, event.Duration,
		event.Description, event.UserID, event.NotifyBefore,
	).Scan(
		&updated.ID, &updated.Title, &updated.EventTime,
		&updated.Duration, &updated.Description, &updated.UserID,
		&updated.NotifyBefore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Event{}, model.ErrEventNotFound
		}
		return model.Event{}, fmt.Errorf("failed to update event: %w", err)
	}

	return updated, nil
}

func (s *PostgresStorage) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return model.ErrEventNotFound
	}

	return nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, id uuid.UUID) (model.Event, error) {
	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE id = $1
	`

	var event model.Event
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.Title, &event.EventTime,
		&event.Duration, &event.Description, &event.UserID,
		&event.NotifyBefore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Event{}, model.ErrEventNotFound
		}
		return model.Event{}, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

func (s *PostgresStorage) ListByDay(ctx context.Context, date time.Time) ([]model.Event, error) {
	// Начало и конец указанного дня
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	//nolint:goconst
	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE event_time >= $1 AND event_time < $2
	ORDER BY event_time ASC
	`

	return s.queryEvents(ctx, query, startOfDay, endOfDay)
}

func (s *PostgresStorage) ListByWeek(ctx context.Context, date time.Time) ([]model.Event, error) {
	// Находим начало недели (понедельник)
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

	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE event_time >= $1 AND event_time < $2
	ORDER BY event_time ASC
	`

	return s.queryEvents(ctx, query, startOfWeek, endOfWeek)
}

func (s *PostgresStorage) ListByMonth(ctx context.Context, date time.Time) ([]model.Event, error) {
	// Начало месяца
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	// Начало следующего месяца
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE event_time >= $1 AND event_time < $2
	ORDER BY event_time ASC
	`

	return s.queryEvents(ctx, query, startOfMonth, endOfMonth)
}

func (s *PostgresStorage) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Event, error) {
	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE user_id = $1
	ORDER BY event_time DESC
	`

	return s.queryEvents(ctx, query, userID)
}

func (s *PostgresStorage) ListByUserAndPeriod(
	ctx context.Context, userID uuid.UUID, start, end time.Time,
) ([]model.Event, error) {
	query := `
	SELECT id, title, event_time, duration, description, user_id, notify_before
	FROM events
	WHERE user_id = $1 AND event_time >= $2 AND event_time < $3
	ORDER BY event_time ASC
	`

	return s.queryEvents(ctx, query, userID, start, end)
}

func (s *PostgresStorage) queryEvents(ctx context.Context, query string, args ...interface{}) ([]model.Event, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		err := rows.Scan(
			&event.ID, &event.Title, &event.EventTime,
			&event.Duration, &event.Description, &event.UserID,
			&event.NotifyBefore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

func isDuplicateError(err error) bool {
	// Для PostgreSQL код ошибки 23505 - unique_violation
	var pqErr interface {
		SQLState() string
	}
	if errors.As(err, &pqErr) {
		return pqErr.SQLState() == "23505"
	}
	return false
}

func (s *PostgresStorage) SaveNotification(ctx context.Context, notification *model.Notification) (model.Notification, error) {
	query := `
	INSERT INTO notifications (id, event_id, user_id, title, event_time) 
         VALUES ($1, $2, $3, $4, $5)
	RETURNING id, event_id, user_id, title, event_time
	`

	// Генерируем новый UUID, если не задан
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	var created model.Notification
	err := s.db.QueryRowContext(
		ctx, query,
		notification.ID, notification.EventID, notification.UserID,
		notification.Title, notification.EventTime,
	).Scan(
		&created.ID, &created.EventID, &created.UserID,
		&created.Title, &created.EventTime,
	)
	if err != nil {
		return model.Notification{}, fmt.Errorf("failed to create event: %w", err)
	}

	return created, nil
}

func (s *PostgresStorage) GetTotalNotifications(ctx context.Context) (int64, error) {
	query := `
	SELECT COUNT(*) 
	FROM notifications
	`
	var count int64
	err := s.db.QueryRowContext(ctx, query).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count total notifications: %w", err)
	}

	return count, nil
}

func (s *PostgresStorage) GetTodayNotifications(ctx context.Context, date time.Time) (int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	query := `
	SELECT COUNT(*) 
	FROM notifications
	WHERE event_time >= $1 AND event_time < $2
	`
	var count int64
	err := s.db.QueryRowContext(ctx, query, startOfDay, endOfDay).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return count, nil
}
