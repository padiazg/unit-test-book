package channel_delivery_tests

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPool(t *testing.T) {
	jobs := []Job{
		{ID: 1, Value: 10},
		{ID: 2, Value: 20},
		{ID: 3, Value: 30},
	}

	results := RunPool(jobs, 2)
	require.Len(t, results, 3)

	got := map[int]int{}
	for _, r := range results {
		got[r.JobID] = r.Value
	}
	assert.Equal(t, 20, got[1])
	assert.Equal(t, 40, got[2])
	assert.Equal(t, 60, got[3])
}

func TestRunPool_NoJobs(t *testing.T) {
	results := RunPool(nil, 2)
	assert.Empty(t, results)
}

func TestRunPool_MoreWorkersThanJobs(t *testing.T) {
	jobs := []Job{{ID: 1, Value: 5}}
	results := RunPool(jobs, 10)
	require.Len(t, results, 1)
	assert.Equal(t, 10, results[0].Value)
}

func TestMerge(t *testing.T) {
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)

	for _, v := range []int{1, 2, 3} {
		ch1 <- v
	}
	close(ch1)

	for _, v := range []int{4, 5, 6} {
		ch2 <- v
	}
	close(ch2)

	out := Merge(ch1, ch2)

	got := make([]int, 0, 6)
	for v := range out {
		got = append(got, v)
	}
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6}, got)
}

func TestMerge_EmptyChannels(t *testing.T) {
	ch1 := make(chan int)
	close(ch1)
	ch2 := make(chan int)
	close(ch2)

	out := Merge(ch1, ch2)
	count := 0
	for range out {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestWorker_ReceivesClosedChannel(t *testing.T) {
	jobs := make(chan Job)
	results := make(chan Result, 1)
	close(jobs)

	done := make(chan struct{})
	go func() {
		defer close(done)
		Worker(jobs, results)
	}()

	select {
	case <-done:
		// worker exited cleanly
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after channel closed")
	}
}

func TestRunPool_ConcurrentSafe(t *testing.T) {
	jobs := make([]Job, 100)
	for i := range jobs {
		jobs[i] = Job{ID: i, Value: i}
	}

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			results := RunPool(jobs, 4)
			assert.Len(t, results, 100)
		})
	}
	wg.Wait()
}
