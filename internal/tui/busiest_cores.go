package tui

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
)

type BusiestCores struct {
	coreUsages      []float64
	coreCharts      map[int]*sparkline.Model
	width           int
	height          int
	squareDimension int
	// Cached styles
	labelStyle  lipgloss.Style
	lowStyle    lipgloss.Style    // usage <= 50%
	mediumStyle lipgloss.Style    // 50% < usage <= 80%
	highStyle   lipgloss.Style    // usage > 80%
}

// NewBusiestCores initializes a BusiestCores instance.
func NewBusiestCores() *BusiestCores {
	return &BusiestCores{
		coreUsages:  make([]float64, 0),
		coreCharts:  make(map[int]*sparkline.Model),
		// Initialize cached styles once
		labelStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		lowStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("6")),    // Cyan
		mediumStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),    // Yellow
		highStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),    // Red
	}
}

func (b *BusiestCores) initializeIfNeeded(coreID int) {
	if b.coreCharts[coreID] == nil {
		graphWidth := max((b.width / (b.squareDimension * 2)), 1)
		chart := sparkline.New(graphWidth, 1,
			sparkline.WithMaxValue(100),
		)
		chart.PushAll(make([]float64, graphWidth))
		b.coreCharts[coreID] = &chart
	}
}

// Update handles incoming messages and updates the internal state
func (b *BusiestCores) Update(msg interface{}) {
	switch msg := msg.(type) {
	case domain.CPUMemoryMetrics:
		b.updateMetrics(msg)
	}
}

// updateMetrics updates the core usage data and their braille graphs.
func (b *BusiestCores) updateMetrics(metrics domain.CPUMemoryMetrics) {
	if b == nil || b.width == 0 {
		return
	}
	b.coreUsages = metrics.CPUUsagePerCore
	b.squareDimension = int(math.Ceil(math.Sqrt(float64(len(b.coreUsages)))))
	for i, usage := range metrics.CPUUsagePerCore {
		b.initializeIfNeeded(i)
		b.coreCharts[i].Push(usage)
	}
}

// View renders the braille graphs for all cores.
func (b *BusiestCores) View() string {
	if b == nil || len(b.coreUsages) == 0 {
		return ""
	}

	var views []string
	for row := 0; row < b.squareDimension; row++ {
		var rowViews []string
		for col := 0; col < b.squareDimension; col++ {
			coreID := row*b.squareDimension + col
			if coreID >= len(b.coreUsages) {
				break
			}
			rowViews = append(rowViews, b.renderCore(coreID))
		}
		views = append(views, strings.Join(rowViews, " "))
	}

	return strings.Join(views, "\n")
}

func (b *BusiestCores) renderCore(coreID int) string {
	chart := b.coreCharts[coreID]
	chart.DrawBraille()
	usage := b.coreUsages[coreID]

	// Select appropriate cached style based on usage
	var style lipgloss.Style
	if usage > 80 {
		style = b.highStyle
	} else if usage > 50 {
		style = b.mediumStyle
	} else {
		style = b.lowStyle
	}

	coreLabel := b.labelStyle.Render(fmt.Sprintf("@%2d", coreID))

	return fmt.Sprintf("%s %s %3.0f%%",
		coreLabel,
		style.Render(chart.View()),
		usage,
	)
}

// Resize adjusts the size of the braille graphs based on available space.
func (b *BusiestCores) Resize(width, height int) {
	if b == nil {
		return
	}
	log.Printf("Resizing BusiestCores to %d x %d", width, height)
	b.width = width
	b.height = height
	if b.squareDimension == 0 {
		return
	}
	graphWidth := max((b.width / (b.squareDimension * 2)), 1) // Adjust based on square dimension, minimum 1
	for _, chart := range b.coreCharts {
		chart.Resize(graphWidth, 1)
	}
}

// max returns the larger of x or y
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
