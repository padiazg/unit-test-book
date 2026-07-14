package goroutine_run_loops

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunLoop_StartStop(t *testing.T) {
	r := NewRunLoop()
	r.Start()
	r.Submit("hello")
	r.Submit("world")
	r.Stop()

	assert.Equal(t, []string{"hello", "world"}, r.Handled())
}

func TestRunLoop_Empty(t *testing.T) {
	r := NewRunLoop()
	r.Start()
	r.Stop()
	assert.Empty(t, r.Handled())
}

func TestRunLoop_ConcurrentSubmissions(t *testing.T) {
	r := NewRunLoop()
	r.Start()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Submit("msg")
		}(i)
	}
	wg.Wait()
	r.Stop()

	assert.Len(t, r.Handled(), 10)
}

func TestProcessor(t *testing.T) {
	t.Run("processes values until cancelled", func(t *testing.T) {
		input := make(chan int)
		output := make(chan int)
		p := NewProcessor(input, output)

		ctx, cancel := context.WithCancel(context.Background())
		go p.Run(ctx)

		input <- 5
		assert.Equal(t, 10, <-output)
		input <- 3
		assert.Equal(t, 6, <-output)

		cancel()
		p.Wait()
	})

	t.Run("exits on input close", func(t *testing.T) {
		input := make(chan int)
		output := make(chan int)
		p := NewProcessor(input, output)

		go p.Run(context.Background())

		input <- 7
		assert.Equal(t, 14, <-output)
		close(input)
		p.Wait()
	})
}

func TestMerger(t *testing.T) {
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)
	output := make(chan int, 6)

	m := NewMerger(output, ch1, ch2)
	go m.Run()
	<-m.started

	ch1 <- 1
	ch2 <- 100
	ch1 <- 2
	ch2 <- 200
	ch1 <- 3
	ch2 <- 300

	close(ch1)
	close(ch2)

	m.Wait()

	got := make([]int, 0, 6)
	for v := range output {
		got = append(got, v)
	}
	assert.ElementsMatch(t, []int{1, 2, 3, 100, 200, 300}, got)
}

func TestGenerator(t *testing.T) {
	t.Run("generates N values", func(t *testing.T) {
		output := make(chan int, 10)
		g := NewGenerator(output)

		go g.Run(context.Background(), 5)

		got := make([]int, 0)
		for v := range output {
			got = append(got, v)
		}
		assert.Equal(t, []int{0, 1, 2, 3, 4}, got)
	})

	t.Run("stops early", func(t *testing.T) {
		output := make(chan int, 100)
		g := NewGenerator(output)

		go g.Run(context.Background(), 100)
		g.Stop()

		got := make([]int, 0)
		for v := range output {
			got = append(got, v)
		}
		assert.Less(t, len(got), 100)
	})

	t.Run("cancellation", func(t *testing.T) {
		output := make(chan int, 100)
		g := NewGenerator(output)

		ctx, cancel := context.WithCancel(context.Background())
		go g.Run(ctx, 100)
		cancel()

		got := make([]int, 0)
		for v := range output {
			got = append(got, v)
		}
		assert.Less(t, len(got), 100)
	})
}

func TestFanOut(t *testing.T) {
	input := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		input <- i
	}
	close(input)

	outs := FanOut(input, 3)

	got := make([]int, 0)
	for _, ch := range outs {
		for v := range ch {
			got = append(got, v)
		}
	}
	assert.ElementsMatch(t, []int{10, 20, 30, 40, 50}, got)
}

func TestEventBus(t *testing.T) {
	t.Run("publish and receive", func(t *testing.T) {
		bus := NewEventBus()
		ch := make(chan Event, 1)
		bus.Subscribe("orders", ch)
		bus.Publish("orders", "order.created")
		select {
		case e := <-ch:
			assert.Equal(t, Event("order.created"), e)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for event")
		}
	})

	t.Run("different topic not received", func(t *testing.T) {
		bus := NewEventBus()
		ch := make(chan Event, 1)
		bus.Subscribe("orders", ch)
		bus.Publish("notifications", "ignored")
		select {
		case <-ch:
			t.Fatal("should not receive event on different topic")
		default:
		}
	})
}

func TestProcessMessages(t *testing.T) {
	r := NewRunLoop()
	msgs := []string{"a", "b", "c"}
	results := r.ProcessMessages(context.Background(), msgs)
	assert.Equal(t, []string{"processed: a", "processed: b", "processed: c"}, results)
}

func TestProcessMessages_Empty(t *testing.T) {
	r := NewRunLoop()
	results := r.ProcessMessages(context.Background(), nil)
	assert.Empty(t, results)
}

func TestProcessMessages_Cancellation(t *testing.T) {
	r := NewRunLoop()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := r.ProcessMessages(ctx, []string{"a", "b"})
	assert.Empty(t, results)
}

// Sensor tests — "Run() returns channel" pattern

func TestSensor_EmitsReadings(t *testing.T) {
	transport := func(buf []byte) (int, error) {
		buf[0] = 0x01
		buf[1] = 0x90
		return 2, nil
	}

	s := NewSensor(transport, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := s.Run(ctx)

	var got int
	select {
	case v, ok := <-ch:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		got = v
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for reading")
	}

	assert.Equal(t, 400, got) // 0x0190 = 400

	s.Stop()

	select {
	case _, ok := <-ch:
		if ok {
			for range ch {
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for channel to close")
	}
}

func TestSensor_ClosesOnContextCancel(t *testing.T) {
	transport := func(buf []byte) (int, error) {
		return 0, nil
	}

	s := NewSensor(transport, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	ch := s.Run(ctx)
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to close")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for channel to close")
	}
}

func TestSensor_ClosesOnStop(t *testing.T) {
	transport := func(buf []byte) (int, error) {
		return 0, nil
	}

	s := NewSensor(transport, time.Hour)
	ch := s.Run(context.Background())
	s.Stop()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to close")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for channel to close")
	}
}
