package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

type Service struct {
	repo     storage.EventStorage
	producer interfaces.Producer
	logger   *logger.Logger
}

func NewService(repo storage.EventStorage, producer interfaces.Producer, logger *logger.Logger) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (s *Service) ProcessNotifications(ctx context.Context) {
	now := time.Now()
	events, err := s.repo.ListByDay(ctx, now)
	if err != nil {
		s.logger.Error("Ошибка получения событий", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	var messages []interfaces.Message

	for _, event := range events {
		if event.ShouldNotify() &&
			event.NotificationTime().Hour() == now.Hour() &&
			event.NotificationTime().Minute() == now.Minute() {

			// Создаем уведомление
			notification := &model.Notification{
				ID:        uuid.New(),
				EventID:   event.ID,
				UserID:    event.UserID,
				Title:     event.Title,
				EventTime: event.EventTime,
			}

			data, err := notification.ToJSON()
			if err != nil {
				s.logger.Error("Ошибка сериализации уведомления", "error", err)
				continue
			}

			messages = append(messages, interfaces.Message{
				Key:   []byte(event.ID.String()), // Используем ID события как ключ
				Value: data,
			})
			s.logger.Info("Отправка уведомления в Kafka", "notification", notification)
		}
	}

	if len(messages) == 0 {
		return
	}

	err = s.producer.Publish(ctx, messages...)
	if err != nil {
		s.logger.Error("Ошибка отправки в Kafka", "error", err)
		return
	}
}

func (s *Service) CleanupOldEvents(ctx context.Context) {
	yearAgo := time.Now().AddDate(-1, 0, 0)
	events, err := s.repo.ListByDay(ctx, yearAgo)
	if err != nil {
		s.logger.Error("Ошибка получения событий", "error", err)
		return
	}
	if len(events) == 0 {
		return
	}
	for _, event := range events {
		if err := s.repo.Delete(ctx, event.ID); err != nil {
			s.logger.Error("Ошибка удаления события", "eventID", event.ID, "error", err)
		}
	}
}
