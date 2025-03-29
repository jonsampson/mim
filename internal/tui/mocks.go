package tui

import (
    "github.com/jonsampson/mim/internal/infra"
    "github.com/stretchr/testify/mock"
)

// MockMetricsCollector is a mock implementation of the metricsCollector interface
type MockMetricsCollector struct {
    mock.Mock
    metricsChan chan infra.SystemMetrics
}

// Ensure MockMetricsCollector implements the metricsCollector interface
var _ metricsCollector = (*MockMetricsCollector)(nil)
func (m *MockMetricsCollector) Start() {
    m.Called()
}

func (m *MockMetricsCollector) Stop() {
    m.Called()
}

func (m *MockMetricsCollector) Metrics() <-chan infra.SystemMetrics {
    return m.metricsChan
}
