package parallel_tests

import (
	"sync"
	"sync/atomic"
)

type UnsafeCounter struct {
	value int
}

func (c *UnsafeCounter) Increment() {
	c.value++
}

func (c *UnsafeCounter) Value() int {
	return c.value
}

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

type AtomicCounter struct {
	value atomic.Int64
}

func (c *AtomicCounter) Increment() {
	c.value.Add(1)
}

func (c *AtomicCounter) Value() int64 {
	return c.value.Load()
}

type SliceWriter struct {
	mu   sync.Mutex
	data []int
}

func (w *SliceWriter) Write(n int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data = append(w.data, n)
}

func (w *SliceWriter) Read() []int {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]int, len(w.data))
	copy(out, w.data)
	return out
}

type UnsafeSliceWriter struct {
	data []int
}

func (w *UnsafeSliceWriter) Write(n int) {
	w.data = append(w.data, n)
}

func (w *UnsafeSliceWriter) Read() []int {
	out := make([]int, len(w.data))
	copy(out, w.data)
	return out
}

type ParallelSummer struct {
	mu    sync.Mutex
	total int
}

func (s *ParallelSummer) Add(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total += n
}

func (s *ParallelSummer) Total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.total
}

func SumConcurrently(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	chunkSize := (len(nums) + workers - 1) / workers

	var (
		mu    sync.Mutex
		total int
		wg    sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if end > len(nums) {
			end = len(nums)
		}
		go func(s, e int) {
			defer wg.Done()
			sum := 0
			for _, v := range nums[s:e] {
				sum += v
			}
			mu.Lock()
			total += sum
			mu.Unlock()
		}(start, end)
	}
	wg.Wait()
	return total
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]string)}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
