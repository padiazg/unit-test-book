package goroutine_leak_detection

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WorkerPool struct {
	wg sync.WaitGroup
}

func NewWorkerPool(workers int, jobs <-chan int, results chan<- string) *WorkerPool {
	wp := &WorkerPool{}

	for i := range workers {
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()
			for j := range jobs {
				results <- fmt.Sprintf("worker-%d processed %d", id, j)
			}
		}(i)
	}
	return wp
}

func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
}

type LeakyProcessor struct {
	started bool
}

func (lp *LeakyProcessor) Start() {
	lp.started = true
	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
}

func (lp *LeakyProcessor) IsStarted() bool {
	return lp.started
}

type SafeProcessor struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewSafeProcessor() *SafeProcessor {
	return &SafeProcessor{}
}

func (sp *SafeProcessor) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	sp.cancel = cancel
	sp.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	})
}

func (sp *SafeProcessor) Stop() {
	if sp.cancel != nil {
		sp.cancel()
		sp.wg.Wait()
	}
}

type CachedService struct {
	mu    sync.Mutex
	cache map[string]string
	done  chan struct{}
	wg    sync.WaitGroup
}

func NewCachedService() *CachedService {
	cs := &CachedService{
		cache: make(map[string]string),
		done:  make(chan struct{}),
	}
	cs.wg.Add(1)
	go cs.evictLoop()
	return cs
}

func (cs *CachedService) evictLoop() {
	defer cs.wg.Done()
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cs.mu.Lock()
			cs.cache = make(map[string]string)
			cs.mu.Unlock()
		case <-cs.done:
			return
		}
	}
}

func (cs *CachedService) Set(key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cache[key] = value
}

func (cs *CachedService) Get(key string) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.cache[key]
}

func (cs *CachedService) Close() {
	close(cs.done)
	cs.wg.Wait()
}

type LeakyCachedService struct {
	mu    sync.Mutex
	cache map[string]string
}

func NewLeakyCachedService() *LeakyCachedService {
	cs := &LeakyCachedService{
		cache: make(map[string]string),
	}
	go cs.evictLoop()
	return cs
}

func (cs *LeakyCachedService) evictLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cs.mu.Lock()
		cs.cache = make(map[string]string)
		cs.mu.Unlock()
	}
}

func (cs *LeakyCachedService) Set(key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cache[key] = value
}

func (cs *LeakyCachedService) Get(key string) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.cache[key]
}

func DoWorkWithDeferredCleanup() {
	ch := make(chan int)
	go func() {
		for range ch {
		}
	}()
	ch <- 42
	close(ch)
}

func DoWorkWithLeakedGoroutine() {
	ch := make(chan int)
	go func() {
		for range ch {
		}
	}()
	ch <- 42
	// goroutine still blocks on ch even after function returns
}

func DoWorkWithCancellation(ctx context.Context) {
	ch := make(chan int, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}()
	ch <- 42
	close(ch)
}

func DoConcurrentWork() error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i := range 3 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n == 1 {
				errCh <- fmt.Errorf("error from %d", n)
			}
		}(i)
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func UnboundedCachePut(mu *sync.Mutex, m map[string]int, key string, value int) {
	mu.Lock()
	m[key] = value
	mu.Unlock()
}

func UnboundedCacheGet(mu *sync.Mutex, m map[string]int, key string) (int, bool) {
	mu.Lock()
	defer mu.Unlock()
	v, ok := m[key]
	return v, ok
}
