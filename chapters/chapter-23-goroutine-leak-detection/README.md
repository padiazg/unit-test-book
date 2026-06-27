# Chapter 23: Goroutine Leak Detection

## Description

Use `go.uber.org/goleak` to detect goroutine leaks in tests. A goroutine leak occurs when a goroutine is started but never exits — it blocks forever on a channel, waits on a timer, or holds an abandoned mutex. `goleak.VerifyNone(t)` checks that no goroutines are running after a test completes.

Real-world example: `pantry/internal/core/services` — `goleak.VerifyTestMain(m)` catches leaks from all package tests.

## Code

```go
type LeakyProcessor struct {
	started bool
}

func (lp *LeakyProcessor) Start() {
	lp.started = true
	go func() { for { time.Sleep(time.Second) } }()
	// goroutine never exits — leak!
}

type SafeProcessor struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (sp *SafeProcessor) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	sp.cancel = cancel
	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		for {
			select {
			case <-ctx.Done(): return
			default: time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (sp *SafeProcessor) Stop() {
	sp.cancel()
	sp.wg.Wait()
}
```

## Test

```go
func verifyNoLeaks(t *testing.T) {
	t.Helper()
	goleak.VerifyNone(t)
}

func TestWorkerPool(t *testing.T) {
	defer verifyNoLeaks(t)
	jobs := make(chan int, 10)
	results := make(chan string, 100)
	wp := NewWorkerPool(3, jobs, results)
	for i := 0; i < 10; i++ { jobs <- i }
	close(jobs)
	wp.Stop()
	close(results)
	count := 0
	for range results { count++ }
	assert.Equal(t, 10, count)
}

func TestSafeProcessor(t *testing.T) {
	defer verifyNoLeaks(t)
	sp := NewSafeProcessor()
	sp.Start()
	sp.Stop()
}

func TestCachedService(t *testing.T) {
	defer verifyNoLeaks(t)
	cs := NewCachedService()
	cs.Set("k", "v")
	cs.Close()
}

func TestLeakyProcessor(t *testing.T) {
	t.Skip("demonstrates leak - skipped by goleak")
	lp := &LeakyProcessor{}
	lp.Start()
}
```

## Testing Approach

Goroutine leak detection:

1. **`verifyNoLeaks` helper** — wraps `goleak.VerifyNone(t)` with `t.Helper()`. Added via `defer` at the top of each clean test. If any goroutine is still running, the test fails with a dump of the leaked goroutine's stack.
2. **Skip leaky tests** — tests that intentionally demonstrate leaks (LeakyProcessor, LeakyCachedService) are skipped during goleak verification with `t.Skip`. The code remains as documentation but doesn't pollute the test suite.
3. **`wg.Wait()` for synchronization** — `SafeProcessor.Stop()` calls `wg.Wait()` to guarantee the goroutine has exited before the test continues. Without this, goleak might catch the goroutine in its final cleanup millisecond.
4. **`close(ch)` and `<-ch` coordination** — closing a channel as a broadcast signal (`close(done)`) lets goroutines exit cleanly. The test closes the done channel and waits (via WaitGroup or separate done channel).
