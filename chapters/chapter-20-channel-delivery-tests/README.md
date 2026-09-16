---
section: concurrency
---

# Chapter 20: Channel Delivery Tests

## Description

Test goroutine communication patterns using channels: fan-out (distribute a job to a worker pool), fan-in (collect each worker's partial result), and graceful shutdown via `close`. Channel delivery tests verify that values flow correctly between goroutines without deadlocks or lost work.

The subject is a fixed worker pool: each worker is pinned to a contiguous slice of players and scores incoming draws, returning coincidence counts. `Pool.Count` fans the draw out to every worker's start channel, then fans the partial results back in.

Real-world example: `alpaca-take-home/internal/core/workers/pool.go` — a per-worker channel pool scoring draws against millions of players.

## Code

```go
type Draw struct{ Numbers [5]int }
type Player struct{ Numbers [5]int }
type job struct{ draw Draw }

// Worker scores its player slice against each draw, sends the counts
// (index = number of matches), and exits when starts is closed.
func Worker(players []Player, starts <-chan job, results chan<- [6]int) {
    for j := range starts {
        var counts [6]int
        for _, p := range players {
            if m := coincidences(p.Numbers, j.draw.Numbers); m >= 2 {
                counts[m]++
            }
        }
        results <- counts
    }
}

type Pool struct {
    workers int
    starts  []chan job
    results []chan [6]int
}

// Count fans the draw out to every worker, then fans the partial counts in.
func (p *Pool) Count(draw Draw) (c2, c3, c4, c5 int) {
    for w := range p.workers {
        p.starts[w] <- job{draw: draw} // fan-out
    }
    var total [6]int
    for w := range p.workers {
        partial := <-p.results[w]      // fan-in
        for k := 2; k <= 5; k++ {
            total[k] += partial[k]
        }
    }
    return total[2], total[3], total[4], total[5]
}

// Close stops every worker by closing its start channel.
func (p *Pool) Close() {
    for _, ch := range p.starts {
        close(ch)
    }
}
```

## Test

```go
type workerCheckFn func(*testing.T, [6]int)

var checkworker = func(fns ...workerCheckFn) []workerCheckFn { return fns }

func TestWorker(t *testing.T) {
    checkIndex := func(index int, want int) workerCheckFn {
        return func(t *testing.T, i [6]int) {
            t.Helper()
            assert.Equal(t, want, i[index])
        }
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
                {3, 2, 1, 10, 12},    // 3
                {5, 4, 2, 9, 6},      // 3
                {3, 2, 6, 7, 9},      // 2
                {1, 2, 3, 4, 5},      // 5
                {20, 21, 22, 23, 24}, // 0
            },
            draw: [5]int{1, 2, 3, 4, 5},
            checks: checkworker(
                checkIndex(2, 1),
                checkIndex(3, 2),
                checkIndex(4, 0),
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
            partial := <-results                        // collect it

            var total [6]int
            for k := 2; k <= 5; k++ {
                total[k] += partial[k]
            }
            for _, c := range tt.checks {
                c(t, total)
            }
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Worker
go-testgen gen . Pool.Count
```

## Testing Approach

Channel delivery tests:

1. **Per-worker channels, not a shared channel** — each worker owns its own `start` and `result` channel, so fan-out is `starts[w] <- job` and fan-in is `<-results[w]`. No lock, no contention on a single channel.
2. **Fan-out then fan-in in `Count`** — send the draw to every worker, then read exactly one result per worker. Collecting a different number of results means a worker lost or duplicated work.
3. **`close(starts)` as the shutdown signal** — a worker's `for j := range starts` exits when the channel is closed and drained. The test verifies a worker given a pre-closed channel returns cleanly, guarded by a `done` channel + timeout so a stuck worker fails the test instead of hanging it.
4. **Closure checks over the count array** — the worker emits a `[6]int` where the index is the match count, so `checkIndex(i, want)` is a tiny factory that asserts one bucket. Composing them with `checkworker(...)` keeps each test case's expectations declarative.
