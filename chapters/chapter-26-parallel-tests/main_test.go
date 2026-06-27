package parallel_tests

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsafeCounter_ParallelRace(t *testing.T) {
	c := &UnsafeCounter{}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			c.Increment()
		})
	}
	wg.Wait()
	t.Logf("UnsafeCounter final value: %d (expected 1000)", c.Value())
}

func TestSafeCounter_Parallel(t *testing.T) {
	c := &SafeCounter{}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			c.Increment()
		})
	}
	wg.Wait()
	assert.Equal(t, 1000, c.Value())
}

func TestAtomicCounter_Parallel(t *testing.T) {
	c := &AtomicCounter{}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			c.Increment()
		})
	}
	wg.Wait()
	assert.Equal(t, int64(1000), c.Value())
}

func TestSliceWriter_Parallel(t *testing.T) {
	w := &SliceWriter{}
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w.Write(n)
		}(i)
	}
	wg.Wait()
	assert.Len(t, w.Read(), 100)
}

func TestParallelSummer(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	total := SumConcurrently(nums, 4)
	assert.Equal(t, 55, total)
}

func TestParallelSummer_Empty(t *testing.T) {
	assert.Equal(t, 0, SumConcurrently(nil, 4))
}

func TestCache_Parallel(t *testing.T) {
	c := NewCache()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set("key", "value")
			c.Get("key")
		}(i)
	}
	wg.Wait()
}

func TestCache_ReadWhileWrite(t *testing.T) {
	c := NewCache()
	c.Set("a", "1")
	c.Set("b", "2")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c.Set("c", "3")
	}()

	go func() {
		defer wg.Done()
		_, ok := c.Get("a")
		assert.True(t, ok)
	}()

	wg.Wait()
	assert.Equal(t, 3, c.Len())
}

func TestParallelSubtests(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}

	for _, v := range values {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Greater(t, v, 0)
		})
	}
}
