package embedded_infrastructure

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type checkEventFn func(*testing.T, string)

var checkEvent = func(fns ...checkEventFn) []checkEventFn { return fns }

func checkMessage(want string) checkEventFn {
	return func(t *testing.T, got string) {
		t.Helper()
		require.Equal(t, want, got, "message payload mismatch")
	}
}

func checkNoMessage() checkEventFn {
	return func(t *testing.T, got string) {
		t.Helper()
		require.Empty(t, got, "expected no message, got %q", got)
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		before  func(*testing.T, *miniredis.Miniredis, string)
		after   func(*EventBus, context.CancelFunc)
		checks  []checkEventFn
	}{
		{
			name:    "receives single event",
			channel: "events",
			before: func(t *testing.T, ms *miniredis.Miniredis, ch string) {
				ms.Publish(ch, "hello")
			},
			after: func(bus *EventBus, _ context.CancelFunc) {
				bus.Close()
			},
			checks: checkEvent(
				checkMessage("hello"),
			),
		},
		{
			name:    "receives last of multiple events",
			channel: "events",
			before: func(t *testing.T, ms *miniredis.Miniredis, ch string) {
				ms.Publish(ch, "first")
				time.Sleep(50 * time.Millisecond)
				ms.Publish(ch, "second")
			},
			after: func(bus *EventBus, _ context.CancelFunc) {
				bus.Close()
			},
			checks: checkEvent(
				checkMessage("second"),
			),
		},
		{
			name:    "channel closes on Close",
			channel: "events",
			after: func(bus *EventBus, _ context.CancelFunc) {
				bus.Close()
			},
			checks: checkEvent(
				checkNoMessage(),
			),
		},
		{
			name:    "channel isolation",
			channel: "other",
			before: func(t *testing.T, ms *miniredis.Miniredis, ch string) {
				ms.Publish("events", "should-not-receive")
			},
			after: func(bus *EventBus, _ context.CancelFunc) {
				bus.Close()
			},
			checks: checkEvent(
				checkNoMessage(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := miniredis.RunT(t)
			client := redis.NewClient(&redis.Options{Addr: ms.Addr()})
			t.Cleanup(func() { _ = client.Close() })
			bus := NewEventBus(client)

			ctx, cancel := context.WithCancel(context.Background())
			pubsub := bus.Subscribe(ctx, tt.channel)
			ch := pubsub.Channel()
			time.Sleep(100 * time.Millisecond)

			if tt.before != nil {
				tt.before(t, ms, tt.channel)
			}

			var (
				wg  sync.WaitGroup
				got string
			)

			wg.Go(func() {
				for msg := range ch {
					got = msg.Payload
				}
			})

			time.Sleep(50 * time.Millisecond)

			if tt.after != nil {
				tt.after(bus, cancel)
			}

			wg.Wait()

			for _, fn := range tt.checks {
				fn(t, got)
			}
		})
	}
}
