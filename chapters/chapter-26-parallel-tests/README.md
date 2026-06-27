# Chapter 26: Parallel Tests

## Description

Test concurrent access to shared data with `t.Parallel()`, `sync.Mutex`, and `sync/atomic`. Unsafe operations (plain `int` increment, slice append without lock) produce data races; safe variants using mutexes or atomics prevent them. Run tests with `go test -race` to detect unsynchronized access.

Real-world example: `ollama-tools/internal/registry` — parallel subtests with `-race` flag to verify thread-safe cache.

## Code

```go
type UnsafeCounter struct { value int }
func (c *UnsafeCounter) Increment() { c.value++ }

type SafeCounter struct {
	mu    sync.Mutex
	value int
}
func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

type AtomicCounter struct {
	value atomic.Int64
}
func (c *AtomicCounter) Increment() { c.value.Add(1) }
```

## Test

```go
func TestUnsafeCounter_ParallelRace(t *testing.T) {
	c := &UnsafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()
	// Run with -race: detects data race on c.value
	t.Logf("final value: %d (expected 1000)", c.Value())
}

func TestSafeCounter_Parallel(t *testing.T) {
	c := &SafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1000, c.Value())
}

func TestAtomicCounter_Parallel(t *testing.T) {
	c := &AtomicCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(1000), c.Value())
}

func TestParallelSubtests(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}
	for _, v := range values {
		v := v
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Greater(t, v, 0)
		})
	}
}
```

## Testing Approach

Parallel test patterns:

1. **Race detection** — `TestUnsafeCounter_ParallelRace` demonstrates a data race. Running with `go test -race` reports the race. The `SafeCounter` and `AtomicCounter` tests are race-free.
2. **`wg.Wait()` for goroutine completion** — all parallel tests use `sync.WaitGroup` to ensure goroutines finish before assertions. Without it, some goroutines may not have incremented yet.
3. **`t.Parallel()` for subtests** — `TestParallelSubtests` runs each subtest in parallel with `t.Parallel()`. The `v := v` copy captures the loop variable per-iteration, preventing the closure-iteration bug.
4. **`sync/atomic` for single-value contention** — `AtomicCounter` uses `atomic.Int64.Add(1)` which is lock-free and faster than a mutex for simple counters. Use `sync.Mutex` when protecting multi-field structs or compound operations.
