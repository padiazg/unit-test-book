# Chapter 22: Goroutine Run Loops

## Description

Test goroutines with event loops that process messages, merge streams, fan-out work, and handle graceful shutdown via context cancellation or channel close. Run-loops are `select`-based goroutines that receive from one or more channels, process values, and send results.

Real-world example: `jokes/internal/adapters/primary/queue/consumer.go` — run loop consuming messages with context cancellation.

## Code

```go
type RunLoop struct {
	ch   chan string
	done chan struct{}
}

func (r *RunLoop) Start() {
	go func() {
		for msg := range r.ch {
			r.mu.Lock()
			r.handled = append(r.handled, msg)
			r.mu.Unlock()
		}
		close(r.done)
	}()
}

func (r *RunLoop) Submit(msg string) { r.ch <- msg }

func (r *RunLoop) Stop() {
	close(r.ch)
	<-r.done
}

type Processor struct {
	input  <-chan int
	output chan<- int
	done   chan struct{}
}

func (p *Processor) Run(ctx context.Context) {
	defer close(p.done)
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-p.input:
			if !ok { return }
			p.output <- v * 2
		}
	}
}
```

## Test

```go
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

func TestProcessor(t *testing.T) {
	t.Run("processes until cancelled", func(t *testing.T) {
		input := make(chan int)
		output := make(chan int)
		p := NewProcessor(input, output)
		ctx, cancel := context.WithCancel(context.Background())
		go p.Run(ctx)
		input <- 5
		assert.Equal(t, 10, <-output)
		cancel()
		p.Wait()
	})
}

func TestFanOut(t *testing.T) {
	input := make(chan int, 5)
	for i := 1; i <= 5; i++ { input <- i }
	close(input)
	outs := FanOut(input, 3)
	got := []int{}
	for _, ch := range outs {
		for v := range ch { got = append(got, v) }
	}
	assert.ElementsMatch(t, []int{10, 20, 30, 40, 50}, got)
}

func TestMerger(t *testing.T) {
	ch1, ch2 := make(chan int, 3), make(chan int, 3)
	output := make(chan int, 6)
	m := NewMerger(output, ch1, ch2)
	go m.Run()
	<-m.started
	ch1 <- 1; ch2 <- 100; ch1 <- 2; ch2 <- 200; ch1 <- 3; ch2 <- 300
	close(ch1); close(ch2)
	m.Wait()
	got := []int{}
	for v := range output { got = append(got, v) }
	assert.ElementsMatch(t, []int{1, 2, 3, 100, 200, 300}, got)
}
```

## Testing Approach

Goroutine run loop tests:

1. **Graceful shutdown** — `Stop()` closes the input channel, the for-range loop exits naturally, and `<-r.done` waits for cleanup. This avoids `context.Background()` in simple cases.
2. **`<-m.started` synchronizer** — the Merger exposes a `started` channel to signal when internal goroutines have been created. Without it, `m.Wait()` could close output before workers start.
3. **Cancellation via context** — `Processor.Run(ctx)` checks `ctx.Done()` in the select. Tests cancel the context and verify workers exit via `< -p.Wait()`.
4. **Buffered channels for tests** — test inputs use buffered channels with small capacities so they don't block. The merger output has exactly enough capacity for all expected values.

### Run() Returns Channel

A variant appears when `Run()` *returns* the channel instead of receiving one. The method creates the channel internally, the caller observes from the outside — common when wrapping hardware or external sources that push events.

**Code:**

```go
type Sensor struct {
	transport func(buf []byte) (int, error)
	interval  time.Duration
	stop      chan struct{}
}

func (s *Sensor) Run(ctx context.Context) <-chan int {
	out := make(chan int)
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
				buf := make([]byte, 4)
				n, err := s.transport(buf)
				if err != nil || n < 2 {
					continue
				}
				out <- int(buf[0])<<8 | int(buf[1])
			}
		}
	}()
	return out
}

func (s *Sensor) Stop() { close(s.stop) }
```

**Test:**

```go
func TestSensor_EmitsReadings(t *testing.T) {
	transport := func(buf []byte) (int, error) {
		buf[0] = 0x01; buf[1] = 0x90
		return 2, nil
	}

	s := NewSensor(transport, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := s.Run(ctx)

	var got int
	select {
	case v, ok := <-ch:
		if !ok { t.Fatal("channel closed unexpectedly") }
		got = v
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for reading")
	}

	assert.Equal(t, 400, got) // 0x0190

	s.Stop()
	select {
	case _, ok := <-ch:
		if ok { for range ch {} }
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for channel to close")
	}
}

func TestSensor_ClosesOnCancel(t *testing.T) {
	s := NewSensor(func(b []byte) (int, error) { return 0, nil }, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	ch := s.Run(ctx)
	cancel()
	select {
	case _, ok := <-ch:
		if ok { t.Fatal("expected channel to close") }
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for channel to close")
	}
}
```

**Key differences from injected-channel run loops:**

1. **External observation** — the test doesn't own the channel. It must `select` with a timeout to capture events, then `select` again to verify cleanup.
2. **Two-phase shutdown** — first stop the producer (`s.Stop()` or `cancel()`), then verify the channel closes. A `for range` drain after `Stop()` absorbs events sent before the goroutine sees the stop signal.
3. **Timeout as safety net** — every channel read in the test has a `<-time.After` fallback. Without it, a stuck goroutine hangs the test forever.
4. **Separate close verification** — the second `select` after shutdown confirms the goroutine exited and the channel closed. This catches goroutine leaks that don't affect output but accumulate in CI.
