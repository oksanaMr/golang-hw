package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// WaitForKafka пытается подключиться к Kafka с ретраями
func WaitForKafka(ctx context.Context, brokers []string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		conn, err := kafka.Dial("tcp", brokers[0])
		if err == nil {
			conn.Close()
			return nil
		}

		// Экспоненциальный backoff
		waitTime := time.Duration(1<<uint(i)) * time.Second
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			fmt.Printf("Попытка %d: Kafka недоступна, ждем %v...\n", i+1, waitTime)
		}
	}
	return fmt.Errorf("не удалось подключиться к Kafka после %d попыток", maxRetries)
}
