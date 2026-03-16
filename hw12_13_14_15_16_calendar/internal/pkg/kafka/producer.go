package kafka

import (
	"context"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/segmentio/kafka-go"
)

type producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string, timeout time.Duration) (interfaces.Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: timeout,
		RequiredAcks: kafka.RequireAll,
	}

	return &producer{writer: writer}, nil
}

func (p *producer) Publish(ctx context.Context, messages ...interfaces.Message) error {
	kafkaMessages := make([]kafka.Message, len(messages))
	for i, msg := range messages {
		kafkaMessages[i] = kafka.Message{
			Key:   msg.Key,
			Value: msg.Value,
			Time:  time.Now(),
		}
	}

	return p.writer.WriteMessages(ctx, kafkaMessages...)
}

func (p *producer) Close() error {
	return p.writer.Close()
}
