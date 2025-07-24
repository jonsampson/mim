package infra

import (
	"log"
	"sync"
	"time"
)

type MetricsCollector[T any] interface {
	Start()
	Stop()
	Metrics() <-chan T
}

type BaseCollector[T any] struct {
	metrics        chan T
	stop           chan struct{}
	getMetricsFunc func() (T, error)
	stopped        bool
	mu             sync.Mutex
}

func NewBaseCollector[T any](getMetricsFunc func() (T, error)) *BaseCollector[T] {
	return &BaseCollector[T]{
		metrics:        make(chan T),
		stop:           make(chan struct{}),
		getMetricsFunc: getMetricsFunc,
		stopped:        false,
	}
}

func (bc *BaseCollector[T]) Start() {
	go bc.collectMetrics()
}

func (bc *BaseCollector[T]) Stop() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if !bc.stopped {
		close(bc.stop)
		close(bc.metrics)
		bc.stopped = true
	}
}

func (bc *BaseCollector[T]) Metrics() <-chan T {
	return bc.metrics
}

func (bc *BaseCollector[T]) collectMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics, err := bc.getMetricsFunc()
			if err == nil {
				if bc.stopped {
					return
				}
				select {
				case bc.metrics <- metrics:
				case <-bc.stop:
					return
				}
			}
			if err != nil {
				log.Printf("Error collecting metrics, skipping cycle: %v", err)
				// Continue to next collection cycle instead of crashing
				continue
			}
		case <-bc.stop:
			return
		}
	}
}
