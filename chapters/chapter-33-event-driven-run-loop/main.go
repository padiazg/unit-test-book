package event_driven_run_loop

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errShortRead = errors.New("short read")

type ReadingEvent struct {
	PM2_5 float64
	Err   error
}

type TransportProvider interface {
	Read(buf []byte) (int, error)
	Write(data []byte) error
}

type Sensor struct {
	transport TransportProvider
	interval  time.Duration
	stop      chan struct{}
}

func NewSensor(transport TransportProvider, interval time.Duration) *Sensor {
	return &Sensor{
		transport: transport,
		interval:  interval,
		stop:      make(chan struct{}),
	}
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

func (s *Sensor) Stop() {
	close(s.stop)
}

func (s *Sensor) collect(out chan<- ReadingEvent) {
	reading, err := s.readMeasurement()
	if err != nil {
		out <- ReadingEvent{Err: err}
		return
	}
	out <- ReadingEvent{PM2_5: reading}
}

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
		return 0, fmt.Errorf("%w: got %d bytes", errShortRead, n)
	}

	pm25 := float64(buf[0])*256 + float64(buf[1]) + float64(buf[2])/100
	return pm25, nil
}
