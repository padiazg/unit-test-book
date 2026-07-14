# Chapter 33: Event-Driven Run Loop Tests

## Description

Test goroutines whose `Run(ctx)` method returns an event channel and the system is driven by an external transport (sensor, queue, stream). These tests combine mock-based setup, external channel observation, and a structured `before`/`checks`/`after` lifecycle. Unlike injected-channel patterns where the test owns both sides, the test observes from the outside — reading from the returned channel with timeouts and verifying the channel closes on shutdown.

Real-world examples:
- `go-aqi/drivers/sps30/sps30_test.go` — `TestSPS30_Run` with transport mock, `select`-based event capture, and phase hooks
- `go-aqi/drivers/zh07/zh07i_test.go` — `TestZH07i_Run` with staged `mock.On("Read").Once()` expectations simulating byte-level protocol

## Code

```go
type TransportProvider interface {
	Read(buf []byte) (int, error)
	Write(data []byte) error
}

type Sensor struct {
	transport TransportProvider
	interval  time.Duration
	stop      chan struct{}
}

func (s *Sensor) Run(ctx context.Context) <-chan ReadingEvent {
	out := make(chan ReadingEvent)
	go func() {
		defer close(out)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			case <-ticker.C:
				s.collect(out)
			}
		}
	}()
	return out
}

func (s *Sensor) Stop() { close(s.stop) }

func (s *Sensor) readMeasurement() (float64, error) {
	if err := s.transport.Write([]byte{0x03, 0x00}); err != nil {
		return 0, fmt.Errorf("write command: %w", err)
	}
	buf := make([]byte, 6)
	n, err := s.transport.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}
	if n < 6 {
		return 0, fmt.Errorf("short read: got %d bytes", n)
	}
	return float64(buf[0])*256 + float64(buf[1]) + float64(buf[2])/100, nil
}
```

## Test

```go
type mockTransport struct{ mock.Mock }

func (m *mockTransport) Read(buf []byte) (int, error) {
	args := m.Called(buf)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockTransport) Write(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

type checkReadFn func(*testing.T, ReadingEvent)

var checkReadRun = func(fns ...checkReadFn) []checkReadFn { return fns }

func checkRunError(want string) checkReadFn {
	return func(t *testing.T, re ReadingEvent) {
		t.Helper()
		if want == "" {
			assert.NoError(t, re.Err)
			return
		}
		if assert.Error(t, re.Err) {
			assert.Contains(t, re.Err.Error(), want)
		}
	}
}

func checkReadValues(want float64) checkReadFn {
	return func(t *testing.T, re ReadingEvent) {
		t.Helper()
		assert.InDelta(t, want, re.PM2_5, 0.01)
	}
}

func TestSensor_Run(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		before   func(*mockTransport)
		after    func(*Sensor, context.CancelFunc)
		checks   []checkReadFn
	}{
		{
			name:     "emits valid readings on tick",
			interval: 10 * time.Millisecond,
			before: func(m *mockTransport) {
				m.On("Write", []byte{0x03, 0x00}).Return(nil).Maybe()
				m.On("Read", mock.MatchedBy(func(b []byte) bool {
					return len(b) == 6
				})).Run(func(args mock.Arguments) {
					buf := args.Get(0).([]byte)
					buf[0] = 0x01; buf[1] = 0x90; buf[2] = 50
				}).Return(6, nil).Maybe()
			},
			after: func(s *Sensor, _ context.CancelFunc) { s.Stop() },
			checks: checkReadRun(
				checkRunError(""),
				checkReadValues(400.5),
			),
		},
		{
			name:     "closes channel on context cancellation",
			interval: time.Hour,
			before: func(m *mockTransport) {
				m.On("Write", mock.Anything).Return(nil).Maybe()
				m.On("Read", mock.Anything).Return(6, nil).Maybe()
			},
			after: func(_ *Sensor, cancel context.CancelFunc) { cancel() },
			checks: checkReadRun(),
		},
		{
			name:     "emits error on transport failure",
			interval: 10 * time.Millisecond,
			before: func(m *mockTransport) {
				m.On("Write", []byte{0x03, 0x00}).
					Return(errors.New("write timeout")).Maybe()
			},
			after: func(s *Sensor, _ context.CancelFunc) { s.Stop() },
			checks: checkReadRun(
				checkRunError("write timeout"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			m := new(mockTransport)
			if tt.before != nil { tt.before(m) }

			s := NewSensor(m, tt.interval)
			ch := s.Run(ctx)

			time.Sleep(20 * time.Millisecond)

			if tt.after != nil { tt.after(s, cancel) }

			var got ReadingEvent
			select {
			case reading, ok := <-ch:
				if !ok {
					for _, fn := range tt.checks { fn(t, ReadingEvent{}) }
					return
				}
				got = reading
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for reading")
			}

			select {
			case _, ok := <-ch:
				if ok { for range ch {} }
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for channel to close")
			}

			for _, fn := range tt.checks { fn(t, got) }
		})
	}
}
```

## Testing Approach

Event-driven run loop tests combine three layers:

1. **`before`/`after` lifecycle** — `before` configures mock expectations for the specific scenario; `after` triggers shutdown (`s.Stop()` or `cancel()`). Each test case specifies only what differs — mock setup, interval, assertions, and teardown. The boilerplate of creating contexts, starting goroutines, and draining channels is shared in the loop body.

2. **Per-case interval** — short intervals (10ms) let the first tick fire before shutdown, producing a reading to capture. A long interval (`time.Hour`) for shutdown tests ensures no reading is produced before cancel — the channel closes purely via context cancellation.

3. **`.Maybe()` on mock expectations** — `Maybe()` prevents test failures from extra tick cycles that fire before `s.Stop()` takes effect. Since the test drains the channel after capture, any extra readings from racing ticks are discarded harmlessly.

4. **Two-phase channel observation** — a `20ms` sleep gives the goroutine time to produce a reading. Then `after()` triggers shutdown. The first `select` captures the reading (or detects early channel close for cancellation cases). The second `select` verifies the channel closes after shutdown.

5. **Error path coverage via `assert.Contains`** — `checkRunError` wraps the error check with `assert.Contains`, verifying the error is propagated with context (e.g. `"write timeout"` from `fmt.Errorf("write command: %w", err)`). This ensures the production code wraps errors properly, not just returns them bare.

6. **Related: channel-backed mocks for bidirectional protocols** — The `mockConn` pattern from `truco/truco-server/internal/adapters/primary/ws/handler_test.go` extends run-loop testing to bidirectional communication (WebSocket, gRPC streams). Unlike the external observation pattern in this chapter (reading from a returned channel), channel-backed mocks use buffered channels to *inject* incoming messages into the handler's read loop and capture outgoing writes. Each test case sets up mock expectations via `before`, pumps messages into `mockConn.inject()`, then observes responses through a `respCh` — combining the lifecycle patterns from chapters 10 and 33 with channel-based bidirectional simulation. See also chapter 34 for another infrastructure-level testing approach using embedded in-memory servers.
