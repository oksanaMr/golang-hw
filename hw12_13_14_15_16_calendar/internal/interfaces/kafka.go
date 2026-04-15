package interfaces

import (
	"context"
)

type Message struct {
	Key   []byte
	Value []byte
}

type Producer interface {
	Publish(ctx context.Context, messages ...Message) error
	Close() error
}

type Consumer interface {
	Subscribe(ctx context.Context, handler MessageHandler) error
	Close() error
}

type MessageHandler func(ctx context.Context, msg Message)

type ConsumerGroup interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}
