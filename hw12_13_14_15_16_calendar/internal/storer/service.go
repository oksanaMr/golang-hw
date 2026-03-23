package storer

import (
	"context"
	"encoding/json"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
)

type Service struct {
	repo   storage.NotificationStorage
	logger *logger.Logger
}

func NewService(repo storage.NotificationStorage, logger *logger.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) HandleMessage(ctx context.Context, msg interfaces.Message) {
	var notification model.Notification

	if err := json.Unmarshal(msg.Value, &notification); err != nil {
		s.logger.Error("Ошибка десериализации уведомления", "error", err)
		return
	}

	s.logger.Info("Получено сообщение из Kafka", "notification", notification)

	notification, err := s.repo.SaveNotification(ctx, &notification)
	if err != nil {
		s.logger.Error("Ошибка сохранения уведомления", "error", err)
		return
	}
	s.logger.Info("Сохранено уведомление", "notification", notification)
}
