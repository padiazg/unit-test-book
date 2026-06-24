# Chapter 22: Goroutine Run Loops

## Description

Test goroutines with event loops that process messages, merge streams, fan-out work, and handle graceful shutdown via context cancellation or channel close. Run-loops are `select`-based goroutines that receive from one or more channels, process values, and send results.

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
