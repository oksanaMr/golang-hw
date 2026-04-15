package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/segmentio/kafka-go"
)

type consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string, topic string, timeout time.Duration) (interfaces.Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		SessionTimeout: timeout,
	})

	return &consumer{reader: reader}, nil
}

func (c *consumer) Subscribe(ctx context.Context, handler interfaces.MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)

			if err != nil {
				// Проверяем, не завершен ли контекст
				if ctx.Err() != nil {
					return ctx.Err()
				}

				fmt.Printf("Ошибка чтения сообщения: %v", err)
				continue
			}

			handler(ctx, interfaces.Message{
				Key:   msg.Key,
				Value: msg.Value,
			})
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
