# Chapter 10: Before Hook Pattern

## Description

Extract test setup to a `before` hook function that returns a fresh fixture state. Each test case calls `before()` explicitly instead of a shared `setup()` mutated by every test. This prevents test pollution from shared mutable state (rate limiters, timestamps, sequence generators).

Real-world example: `hexago/internal/core/services/categories/categories_test.go:18` — `NewFixture()` returns a fresh service + repository for each test case.

## Code

```go
type RateLimiter struct {
	mu        sync.Mutex
	requests  int
	maxPerMin int
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.requests >= r.maxPerMin {
		return false
	}
	r.requests++
	return true
}
```

## Test

```go
func TestRateLimiter(t *testing.T) {
	type beforeReturns struct {
		limiter *RateLimiter
	}

	before := func(t *testing.T, max int) beforeReturns {
		t.Helper()
		return beforeReturns{
			limiter: NewRateLimiter(max),
		}
	}

	t.Run("allows within limit", func(t *testing.T) {
		f := before(t, 3)
		assert.True(t, f.limiter.Allow())
		assert.True(t, f.limiter.Allow())
		assert.True(t, f.limiter.Allow())
	})

	t.Run("blocks after limit", func(t *testing.T) {
		f := before(t, 2)
		assert.True(t, f.limiter.Allow())
		assert.True(t, f.limiter.Allow())
		assert.False(t, f.limiter.Allow()) // blocked
	})

	t.Run("separate limiters are isolated", func(t *testing.T) {
		a := before(t, 1)
		b := before(t, 1)
		assert.True(t, a.limiter.Allow())
		assert.True(t, b.limiter.Allow())
		assert.False(t, a.limiter.Allow())
		assert.False(t, b.limiter.Allow())
	})
}
```

## Testing Approach

The before hook pattern:

1. **Fresh state per case** — `before()` returns a new `RateLimiter` with the configured max. Each subtest gets its own instance. No risk of test A consuming requests and affecting test B.
2. **Typed return struct** — `beforeReturns` documents exactly what the test fixture provides. Adding a new dependency (e.g. a clock) doesn't require changing every test's setup, just the struct and the `before` function.
3. **Explicit over shared** — no `init()` or `TestMain` setup. Every subtest calls `before()` and sees the fixture it depends on. Makes the test self-documenting.
4. **`t.Helper()`** — the `before` hook calls `t.Helper()`, so failure line numbers point to the test assertion, not inside the setup function.
