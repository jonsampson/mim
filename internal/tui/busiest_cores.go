package tui

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
)

type BusiestCores struct {
	coreUsages []float64
	coreCharts map[int]*sparkline.Model
	width      int
	height     int
}

// NewBusiestCores initializes a BusiestCores instance.
func NewBusiestCores() *BusiestCores {
	return &BusiestCores{
		coreUsages: make([]float64, 0),
		coreCharts: make(map[int]*sparkline.Model),
	}
}

func (b *BusiestCores) initializeIfNeeded(coreID int) {
	if b.coreCharts[coreID] == nil {
		chart := sparkline.New(10, 1,
			sparkline.WithMaxValue(100),
		)
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
	if b == nil {
		return
	}
	b.coreUsages = metrics.CPUUsagePerCore
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
	coresPerRow := 6
	rows := (len(b.coreUsages) + coresPerRow - 1) / coresPerRow

	for row := 0; row < rows; row++ {
		var rowViews []string
		for col := 0; col < coresPerRow; col++ {
			coreID := row*coresPerRow + col
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

	color := lipgloss.Color("6") // Cyan
	if usage > 50 {
		color = lipgloss.Color("3") // Yellow
	}
	if usage > 80 {
		color = lipgloss.Color("1") // Red
	}

	style := lipgloss.NewStyle().Foreground(color)
	coreLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(fmt.Sprintf("@%2d", coreID))

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
	b.width = width
	b.height = height
	graphWidth := (width / 6) - 7 // Assuming 6 cores per row, with some padding
	for _, chart := range b.coreCharts {
		chart.Resize(graphWidth, 1)
	}
}
