# Chapter 20: Channel Delivery Tests

## Description

Test goroutine communication patterns using channels: fan-out (distribute jobs to workers), fan-in (merge multiple channels), buffered vs unbuffered channels, and graceful shutdown via close. Channel delivery tests verify that values flow correctly between goroutines without deadlocks or data loss.

## Code

```go
func Worker(jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		results <- Result{JobID: job.ID, Value: job.Value * 2}
	}
}

func RunPool(jobs []Job, workers int) []Result {
	jobCh := make(chan Job, len(jobs))
	resultCh := make(chan Result, len(jobs))
	for i := 0; i < workers; i++ {
		go Worker(jobCh, resultCh)
	}
	for _, j := range jobs { jobCh <- j }
	close(jobCh)
	var out []Result
	for i := 0; i < len(jobs); i++ {
		out = append(out, <-resultCh)
	}
	return out
}
```

## Test

```go
func TestRunPool(t *testing.T) {
	jobs := []Job{
		{ID: 1, Value: 10},
		{ID: 2, Value: 20},
		{ID: 3, Value: 30},
	}
	results := RunPool(jobs, 2)
	require.Len(t, results, 3)
	got := map[int]int{}
	for _, r := range results { got[r.JobID] = r.Value }
	assert.Equal(t, 20, got[1])
	assert.Equal(t, 40, got[2])
	assert.Equal(t, 60, got[3])
}

func TestRunPool_NoJobs(t *testing.T) {
	assert.Empty(t, RunPool(nil, 2))
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
	case <-done: // worker exited cleanly
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after channel closed")
	}
}

func TestMerge(t *testing.T) {
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)
	for _, v := range []int{1, 2, 3} { ch1 <- v }; close(ch1)
	for _, v := range []int{4, 5, 6} { ch2 <- v }; close(ch2)
	out := Merge(ch1, ch2)
	got := []int{}
	for v := range out { got = append(got, v) }
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6}, got)
}
```

## Testing Approach

Channel delivery tests:

1. **Buffered channels** — `jobCh := make(chan Job, len(jobs))` prevents senders from blocking. The buffer size matches the workload so all sends complete before any goroutine reads.
2. **`close(ch)` as shutdown signal** — `Worker` exits when `for job := range jobs` sees the closed+empty channel. The test verifies this with a `done` channel and timeout.
3. **Result collection** — the main goroutine collects exactly `len(jobs)` results from `resultCh`. Any mismatch means workers lost or duplicated work.
4. **`ElementsMatch` for merge** — merging two channels produces nondeterministic ordering. `assert.ElementsMatch` ignores order, focusing on value completeness.
