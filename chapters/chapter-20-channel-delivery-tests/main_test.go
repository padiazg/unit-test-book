package channel_delivery_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type workerCheckFn func(*testing.T, [6]int)

var checkworker = func(fns ...workerCheckFn) []workerCheckFn { return fns }

func TestWorker(t *testing.T) {
	checkIndex := func(index int, want int) workerCheckFn {
		return func(t *testing.T, i [6]int) {
			t.Helper()
			assert.Equal(t, want, i[index])
		}
	}

	playersFrom := func(values ...[5]int) []Player {
		out := make([]Player, 0, len(values))
		for _, v := range values {
			out = append(out, Player{Numbers: v})
		}
		return out
	}

	tests := []struct {
		name   string
		values [][5]int
		draw   [5]int
		checks []workerCheckFn
	}{
		{
			name: "2:1, 3:2, 4:0, 5:2",
			values: [][5]int{
				{5, 2, 3, 4, 1},      // 5 coincidences
				{3, 2, 1, 10, 12},    // 3 coincidences
				{5, 4, 2, 9, 6},      // 3 coincidences
				{3, 2, 6, 7, 9},      // 2 coincidences
				{1, 2, 3, 4, 5},      // 5 coincidences
				{20, 21, 22, 23, 24}, // 0 coincidences
			},
			draw: [5]int{1, 2, 3, 4, 5},
			checks: checkworker(
				checkIndex(2, 1),
				checkIndex(3, 2),
				checkIndex(4, 0),
				checkIndex(5, 2),
			),
		},
		{
			name: "2:1, 3:1, 4:1, 5:2",
			values: [][5]int{
				{64, 90, 10, 81, 70}, // 5 coincidences
				{64, 2, 1, 11, 12},   // 1 coincidence
				{90, 81, 2, 9, 6},    // 2 coincidences
				{90, 81, 2, 10, 6},   // 3 coincidences
				{70, 10, 90, 64, 9},  // 4 coincidences
				{70, 81, 10, 90, 64}, // 5 coincidences
				{20, 21, 22, 23, 24}, // 0 coincidences
			},
			draw: [5]int{70, 81, 10, 90, 64},
			checks: checkworker(
				checkIndex(2, 1),
				checkIndex(3, 1),
				checkIndex(4, 1),
				checkIndex(5, 2),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			starts := make(chan job)
			results := make(chan [6]int)

			go Worker(playersFrom(tt.values...), starts, results)

			starts <- job{draw: Draw{Numbers: tt.draw}} // submit job

			var total [6]int
			partial := <-results // collect the single worker's result
			for k := 2; k <= 5; k++ {
				total[k] += partial[k]
			}

			for _, c := range tt.checks {
				c(t, total)
			}
		})
	}
}

type poolCountCheckFn func(*testing.T, int, int, int, int)

var checkPoolCount = func(fns ...poolCountCheckFn) []poolCountCheckFn { return fns }

func TestPool_Count(t *testing.T) {
	checkCounts := func(c2, c3, c4, c5 int) poolCountCheckFn {
		return func(t *testing.T, r, r2, r3, r4 int) {
			t.Helper()
			assert.Equal(t, [4]int{c2, c3, c4, c5}, [4]int{r, r2, r3, r4})
		}
	}

	playersFrom := func(values ...[5]int) []Player {
		out := make([]Player, 0, len(values))
		for _, v := range values {
			out = append(out, Player{Numbers: v})
		}
		return out
	}

	tests := []struct {
		name    string
		workers int
		draw    Draw
		players []Player
		checks  []poolCountCheckFn
	}{
		{
			name:    "2:1, 3:2, 4:0, 5:2",
			workers: 2,
			draw:    Draw{Numbers: [5]int{1, 2, 3, 4, 5}},
			players: playersFrom(
				[5]int{5, 2, 3, 4, 1},      // 5
				[5]int{3, 2, 1, 10, 12},    // 3
				[5]int{5, 4, 2, 9, 6},      // 3
				[5]int{3, 2, 6, 7, 9},      // 2
				[5]int{1, 2, 3, 4, 5},      // 5
				[5]int{20, 21, 22, 23, 24}, // 0
			),
			checks: checkPoolCount(
				checkCounts(1, 2, 0, 2),
			),
		},
		{
			name:    "more workers than players",
			workers: 8,
			draw:    Draw{Numbers: [5]int{70, 81, 10, 90, 64}},
			players: playersFrom(
				[5]int{64, 90, 10, 81, 70}, // 5
				[5]int{70, 81, 10, 90, 64}, // 5
			),
			checks: checkPoolCount(
				checkCounts(0, 0, 0, 2),
			),
		},
		{
			name:    "empty players",
			workers: 2,
			draw:    Draw{Numbers: [5]int{1, 2, 3, 4, 5}},
			players: nil,
			checks: checkPoolCount(
				checkCounts(0, 0, 0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool(tt.players, tt.workers)
			defer p.Close()

			c2, c3, c4, c5 := p.Count(tt.draw)
			for _, c := range tt.checks {
				c(t, c2, c3, c4, c5)
			}
		})
	}
}

func TestWorker_ExitsOnClosedChannel(t *testing.T) {
	starts := make(chan job)
	results := make(chan [6]int, 1)
	close(starts)

	done := make(chan struct{})
	go func() {
		defer close(done)
		Worker(nil, starts, results)
	}()

	select {
	case <-done: // worker exited cleanly
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after start channel closed")
	}
}
