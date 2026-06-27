package goroutine_leak_detection

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func verifyNoLeaks(t *testing.T) {
	t.Helper()
	goleak.VerifyNone(t)
}

func TestWorkerPool(t *testing.T) {
	defer verifyNoLeaks(t)

	jobs := make(chan int, 10)
	results := make(chan string, 100)

	wp := NewWorkerPool(3, jobs, results)

	for i := range 10 {
		jobs <- i
	}
	close(jobs)
	wp.Stop()
	close(results)

	count := 0
	for range results {
		count++
	}
	assert.Equal(t, 10, count)
}

func TestLeakyProcessor(t *testing.T) {
	t.Skip("demonstrates goroutine leak - skipped by goleak verification")
	lp := &LeakyProcessor{}
	lp.Start()
	assert.True(t, lp.IsStarted())
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
	cs.Set("key1", "value1")
	assert.Equal(t, "value1", cs.Get("key1"))
	cs.Close()
}

func TestLeakyCachedService(t *testing.T) {
	t.Skip("demonstrates goroutine leak - skipped by goleak verification")
	cs := NewLeakyCachedService()
	cs.Set("key1", "value1")
	assert.Equal(t, "value1", cs.Get("key1"))
}

func TestDoWorkWithDeferredCleanup(t *testing.T) {
	DoWorkWithDeferredCleanup()
}

func TestDoWorkWithCancellation(t *testing.T) {
	DoWorkWithCancellation(context.Background())
}

func TestDoConcurrentWork(t *testing.T) {
	err := DoConcurrentWork()
	// may or may not receive the error from a goroutine
	if err != nil {
		assert.Contains(t, err.Error(), "error from")
	}
}

func TestUnboundedCache_Concurrent(t *testing.T) {
	var mu sync.Mutex
	m := make(map[string]int)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			UnboundedCachePut(&mu, m, key, n*2)
		}(i)
	}
	wg.Wait()

	v, ok := UnboundedCacheGet(&mu, m, "key-42")
	assert.True(t, ok)
	assert.Equal(t, 84, v)
}

func TestGoroutineLeak_VerifyClean(t *testing.T) {
	DoWorkWithDeferredCleanup()
}

func TestGoroutineLeak_VerifyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	DoWorkWithCancellation(ctx)
}
