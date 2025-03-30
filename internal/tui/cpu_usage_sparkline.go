package tui

import (
    "fmt"

    "github.com/NimbleMarkets/ntcharts/sparkline"
    "github.com/charmbracelet/lipgloss"
    "github.com/jonsampson/mim/internal/domain"
)

type CPUUsageSparkline struct {
    sl         *sparkline.Model
    width      int
    height     int
    lastValue  float64
}

func NewCPUUsageSparkline(width, height int) *CPUUsageSparkline {
    return &CPUUsageSparkline{
        width:     width,
        height:    height,
        lastValue: 0,
    }
}

func (c *CPUUsageSparkline) initializeIfNeeded() {
    if c.sl == nil {
        sparkline := sparkline.New(
			c.width, 
			c.height,
			sparkline.WithMaxValue(100),
		)
        c.sl = &sparkline
    }
}

func (c *CPUUsageSparkline) Update(metrics domain.CPUMemoryMetrics) {
    c.initializeIfNeeded()
    c.lastValue = metrics.CPUUsageTotal
    c.sl.Push(c.lastValue)
    c.sl.Draw()
}

func (c *CPUUsageSparkline) View() string {
    c.initializeIfNeeded()
    sparklineView := c.sl.View()
    currentUsage := fmt.Sprintf("CPU Usage: %.1f%%", c.lastValue)
    return lipgloss.JoinVertical(lipgloss.Left, sparklineView, currentUsage)
}
