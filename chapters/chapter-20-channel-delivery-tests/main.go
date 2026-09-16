package channel_delivery_tests

// Draw is a set of numbers to score players against.
type Draw struct {
	Numbers [5]int
}

// Player is a single ticket with five numbers.
type Player struct {
	Numbers [5]int
}

// job describes one draw to be scanned by a worker.
type job struct {
	draw Draw
}

// Worker scans its assigned players against each incoming draw and sends the
// coincidence counts (index = number of matches) back on results.
// It exits when the starts channel is closed.
func Worker(players []Player, starts <-chan job, results chan<- [6]int) {
	for j := range starts {
		var counts [6]int
		for _, p := range players {
			m := coincidences(p.Numbers, j.draw.Numbers)
			if m >= 2 {
				counts[m]++
			}
		}
		results <- counts
	}
}

// coincidences returns how many numbers a and b share.
func coincidences(a, b [5]int) int {
	seen := make(map[int]bool, len(a))
	for _, n := range a {
		seen[n] = true
	}
	c := 0
	for _, n := range b {
		if seen[n] {
			c++
		}
	}
	return c
}

// Pool holds a fixed set of workers, each pinned to a contiguous slice of the
// players and reused across every draw.
type Pool struct {
	results []chan [6]int
	starts  []chan job
	workers int
}

// NewPool splits players across workers and starts one Worker per worker.
func NewPool(players []Player, workers int) *Pool {
	if len(players) == 0 {
		return &Pool{}
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(players) {
		workers = len(players)
	}

	p := &Pool{
		workers: workers,
		starts:  make([]chan job, workers),
		results: make([]chan [6]int, workers),
	}

	chunk := (len(players) + workers - 1) / workers
	for w := range workers {
		start := w * chunk
		end := start + chunk
		if end > len(players) {
			end = len(players)
		}
		p.starts[w] = make(chan job)
		p.results[w] = make(chan [6]int)
		go Worker(players[start:end], p.starts[w], p.results[w])
	}
	return p
}

// Count scores draw against every player and returns the number of players
// with exactly 2, 3, 4, and 5 matches.
func (p *Pool) Count(draw Draw) (c2, c3, c4, c5 int) {
	// fan-out: send the draw to every worker
	for w := range p.workers {
		p.starts[w] <- job{draw: draw}
	}

	// fan-in: collect each worker's partial counts
	var total [6]int
	for w := range p.workers {
		partial := <-p.results[w]
		for k := 2; k <= 5; k++ {
			total[k] += partial[k]
		}
	}
	return total[2], total[3], total[4], total[5]
}

// Close stops all workers by closing their start channels.
func (p *Pool) Close() {
	for _, ch := range p.starts {
		close(ch)
	}
}
