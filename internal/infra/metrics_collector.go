package infra

import "time"

type MetricsCollector[T any] interface {
    Start()
    Stop()
    Metrics() <-chan T
}

type BaseCollector[T any] struct {
    metrics chan T
    stop    chan struct{}
    getMetricsFunc func() (T, error)
}

func NewBaseCollector[T any](getMetricsFunc func() (T, error)) *BaseCollector[T] {
    return &BaseCollector[T]{
        metrics:       make(chan T),
        stop:          make(chan struct{}),
        getMetricsFunc: getMetricsFunc,
    }
}

func (bc *BaseCollector[T]) Start() {
    go bc.collectMetrics()
}

func (bc *BaseCollector[T]) Stop() {
    close(bc.stop)
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
                bc.metrics <- metrics
            }
        case <-bc.stop:
            return
        }
    }
}
