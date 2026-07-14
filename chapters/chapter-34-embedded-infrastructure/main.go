package embedded_infrastructure

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type EventBus struct {
	client *redis.Client
}

func NewEventBus(client *redis.Client) *EventBus {
	return &EventBus{client: client}
}

func (b *EventBus) Publish(ctx context.Context, channel, message string) error {
	return b.client.Publish(ctx, channel, message).Err()
}

func (b *EventBus) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return b.client.Subscribe(ctx, channel)
}

func (b *EventBus) Close() error {
	return b.client.Close()
}
