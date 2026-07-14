package event_driven_run_loop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransport struct {
	mock.Mock
}

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
					buf[0] = 0x01
					buf[1] = 0x90
					buf[2] = 50
				}).Return(6, nil).Maybe()
			},
			after: func(s *Sensor, _ context.CancelFunc) {
				s.Stop()
			},
			checks: checkReadRun(
				checkRunError(""),
				checkReadValues(400.5),
			),
		},
		{
			name:     "closes channel on context cancellation",
			interval: time.Hour, // no tick fires — channel closes only via cancel
			before: func(m *mockTransport) {
				m.On("Write", mock.Anything).Return(nil).Maybe()
				m.On("Read", mock.Anything).Return(6, nil).Maybe()
			},
			after: func(_ *Sensor, cancel context.CancelFunc) {
				cancel()
			},
			checks: checkReadRun(),
		},
		{
			name:     "emits error on transport failure",
			interval: 10 * time.Millisecond,
			before: func(m *mockTransport) {
				m.On("Write", []byte{0x03, 0x00}).
					Return(errors.New("write timeout")).Maybe()
			},
			after: func(s *Sensor, _ context.CancelFunc) {
				s.Stop()
			},
			checks: checkReadRun(
				checkRunError("write timeout"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			m := new(mockTransport)

			if tt.before != nil {
				tt.before(m)
			}

			s := NewSensor(m, tt.interval)
			ch := s.Run(ctx)

			time.Sleep(20 * time.Millisecond)

			if tt.after != nil {
				tt.after(s, cancel)
			}

			var got ReadingEvent

			select {
			case reading, ok := <-ch:
				if !ok {
					for _, fn := range tt.checks {
						fn(t, ReadingEvent{})
					}
					return
				}
				got = reading
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for reading")
			}

			select {
			case _, ok := <-ch:
				if ok {
					for range ch {
					}
				}
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for channel to close")
			}

			for _, fn := range tt.checks {
				fn(t, got)
			}
		})
	}
}


