package tui

import (
    "github.com/stretchr/testify/mock"
)

// MockMetricsCollector is a mock implementation of the metricsCollector interface
type MockMetricsCollector[T any] struct {
    mock.Mock
    metricsChan chan T
}

// Ensure MockMetricsCollector implements the metricsCollector interface
var _ metricsCollector[any] = (*MockMetricsCollector[any])(nil)

func (m *MockMetricsCollector[T]) Start() {
    m.Called()
}

func (m *MockMetricsCollector[T]) Stop() {
    m.Called()
}

func (m *MockMetricsCollector[T]) Metrics() <-chan T {
    args := m.Called()
    return args.Get(0).(chan T)
}
