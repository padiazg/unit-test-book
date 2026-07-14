# Chapter 34: Embedded Infrastructure Testing

## Description

Test code that depends on external infrastructure (Redis, databases, message queues) using in-memory replacements rather than mocks at the application layer. This gives you realistic protocol coverage — serialization, pub/sub semantics, connection lifecycle — without needing external processes.

Real-world examples:
- `truco/truco-server/internal/adapters/secondary/external/notifier_redis_test.go` — `miniredis.RunT(t)` for Redis pub/sub testing with table-driven `before`/`checks` lifecycle
- The same pattern extends to `httptest.NewServer` for HTTP, `:memory:` SQLite for databases, and embedded message brokers

## Code

```go
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
```

## Test

```go
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
```

## Testing Approach

This test combines three patterns from earlier chapters into a single integrated test suite:

1. **Embedded infrastructure via `miniredis.RunT(t)`** — starts a real Redis-compatible server in the current process. The test connects with a standard `redis.Client`, exactly like production. `RunT` registers `t.Cleanup` for automatic shutdown — no `defer` boilerplate, no leaked servers on failure.

2. **`before`/`after` lifecycle** — each test case specifies only what differs: `before` publishes events via the miniredis instance (acting as the external producer), `after` triggers shutdown via `bus.Close()`. The boilerplate of creating miniredis, connecting clients, subscribing, and draining channels lives in the loop body. This is the same pattern from chapter 10 applied to infrastructure-heavy tests.

3. **Closure-check factories** — `checkMessage(want string)` and `checkNoMessage()` are factory closures (chapter 7) that capture the expected value and return a `checkEventFn`. The `checkEvent` collection builder (chapter 6) composes them into a slice. Test cases mix and match — one might need only `checkMessage("hello")`, another combines `checkNoMessage()`.

4. **Channel draining via goroutine** — a background goroutine uses `for msg := range ch` to drain the go-redis channel. When `bus.Close()` closes the client connection, the PubSub's internal goroutine detects the error and closes the channel, terminating the `range` loop. The last captured message (or zero-value if none arrived) becomes the test result.

5. **Subscription handshake timing** — Redis subscriptions are asynchronous: `SUBSCRIBE` is sent, the server acknowledges, and only then do messages flow. The `100ms` sleep gives this handshake time to complete before `before` publishes. The optional `50ms` sleep after publishing lets events arrive before `after` closes the connection.

6. **When to use embedded infra vs mocks** — Embedded infrastructure tests are more realistic but slower than interface mocks (chapters 11–15). Use them for protocol compliance, serialization, and connection lifecycle. Use mocks for fast, focused unit tests where you control every return path.
