package goroutine_run_loops

import (
	"context"
	"fmt"
	"sync"
)

type Event string

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (b *EventBus) Subscribe(topic string, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], ch)
}

func (b *EventBus) Publish(topic string, event Event) {
	b.mu.RLock()
	chs := b.subscribers[topic]
	b.mu.RUnlock()
	for _, ch := range chs {
		ch <- event
	}
}

type Processor struct {
	input  <-chan int
	output chan<- int
	done   chan struct{}
}

func NewProcessor(input <-chan int, output chan<- int) *Processor {
	return &Processor{
		input:  input,
		output: output,
		done:   make(chan struct{}),
	}
}

func (p *Processor) Run(ctx context.Context) {
	defer close(p.done)
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-p.input:
			if !ok {
				return
			}
			p.output <- v * 2
		}
	}
}

func (p *Processor) Wait() {
	<-p.done
}

type Merger struct {
	started chan struct{}
	output  chan<- int
	wg      sync.WaitGroup
	inputs  []<-chan int
}

func NewMerger(output chan<- int, inputs ...<-chan int) *Merger {
	return &Merger{
		inputs:  inputs,
		output:  output,
		started: make(chan struct{}),
	}
}

func (m *Merger) Run() {
	for _, ch := range m.inputs {
		m.wg.Add(1)
		go func(c <-chan int) {
			defer m.wg.Done()
			for v := range c {
				m.output <- v
			}
		}(ch)
	}
	close(m.started)
}

func (m *Merger) Wait() {
	m.wg.Wait()
	close(m.output)
}

type Generator struct {
	output chan<- int
	stop   chan struct{}
}

func NewGenerator(output chan<- int) *Generator {
	return &Generator{
		output: output,
		stop:   make(chan struct{}),
	}
}

func (g *Generator) Run(ctx context.Context, count int) {
	defer close(g.output)
	for i := range count {
		select {
		case <-ctx.Done():
			return
		case <-g.stop:
			return
		case g.output <- i:
		}
	}
}

func (g *Generator) Stop() {
	close(g.stop)
}

func FanOut(input <-chan int, workers int) []<-chan int {
	outs := make([]chan int, workers)
	for i := range workers {
		outs[i] = make(chan int)
	}

	for i := range workers {
		i := i
		go func() {
			for v := range input {
				outs[i] <- v * 10
			}
			close(outs[i])
		}()
	}

	result := make([]<-chan int, workers)
	for i, ch := range outs {
		result[i] = ch
	}
	return result
}

type RunLoop struct {
	ch      chan string
	done    chan struct{}
	mu      sync.Mutex
	handled []string
}

func NewRunLoop() *RunLoop {
	return &RunLoop{
		ch:   make(chan string, 10),
		done: make(chan struct{}),
	}
}

func (r *RunLoop) Start() {
	go func() {
		for msg := range r.ch {
			r.mu.Lock()
			r.handled = append(r.handled, msg)
			r.mu.Unlock()
		}
		close(r.done)
	}()
}

func (r *RunLoop) Submit(msg string) {
	r.ch <- msg
}

func (r *RunLoop) Stop() {
	close(r.ch)
	<-r.done
}

func (r *RunLoop) Handled() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.handled))
	copy(out, r.handled)
	return out
}

func (r *RunLoop) ProcessMessages(ctx context.Context, msgs []string) []string {
	if ctx.Err() != nil {
		return nil
	}

	results := make([]string, 0, len(msgs))
	resultsCh := make(chan string, len(msgs))

	go func() {
		defer close(resultsCh)
		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return
			case resultsCh <- fmt.Sprintf("processed: %s", msg):
			}
		}
	}()

	for res := range resultsCh {
		results = append(results, res)
	}
	return results
}
