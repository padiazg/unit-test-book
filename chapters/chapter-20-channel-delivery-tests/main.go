package channel_delivery_tests

type Job struct {
	ID    int
	Value int
}

type Result struct {
	JobID int
	Value int
}

func Worker(jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		results <- Result{JobID: job.ID, Value: job.Value * 2}
	}
}

func RunPool(jobs []Job, workers int) []Result {
	jobCh := make(chan Job, len(jobs))
	resultCh := make(chan Result, len(jobs))

	for range workers {
		go Worker(jobCh, resultCh)
	}

	for _, j := range jobs {
		jobCh <- j
	}
	close(jobCh)

	var out []Result
	for range jobs {
		out = append(out, <-resultCh)
	}
	return out
}

func Merge(chs ...<-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for _, ch := range chs {
			for v := range ch {
				out <- v
			}
		}
		close(out)
	}()
	return out
}
