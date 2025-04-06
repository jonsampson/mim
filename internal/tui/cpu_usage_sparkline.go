package tui

import (
    "fmt"

    "github.com/NimbleMarkets/ntcharts/sparkline"
    "github.com/charmbracelet/lipgloss"
    "github.com/jonsampson/mim/internal/domain"
)

type CPUUsageSparkline struct {
    sl        *sparkline.Model
    width     int
    height    int
    lastValue float64
}

func NewCPUUsageSparkline() *CPUUsageSparkline {
    sparkline := sparkline.New(
        20,
        1,
        sparkline.WithMaxValue(100),
    )
    return &CPUUsageSparkline{
        width:     50, // Set a default width
        height:    3,
        lastValue: 0,
        sl:        &sparkline,
    }
}

func (c *CPUUsageSparkline) Update(metrics domain.CPUMemoryMetrics) {
    c.lastValue = metrics.CPUUsageTotal
    if c.sl != nil {
        c.sl.Push(c.lastValue)
        c.sl.Draw()
    }
}

func (c *CPUUsageSparkline) View() string {
    sparklineView := c.sl.View()
    currentUsage := fmt.Sprintf("Current CPU Usage: %.1f%%", c.lastValue)
    return lipgloss.JoinVertical(lipgloss.Left, currentUsage, sparklineView)
}

func (c *CPUUsageSparkline) Resize(width, height int) {
    c.width = width
    c.height = height
    if c.sl != nil {
        c.sl.Resize(width, height)
    }
}
