package tui

import (
	"math"
	"sort"

	"github.com/jonsampson/mim/internal/domain"
)

type BusiestCores struct {
	coreUsages []float64
	coreCharts map[int]*CoreSparkline
	width      int
	height     int
	squareDimension int
}

// NewBusiestCores initializes a BusiestCores instance.
func NewBusiestCores() *BusiestCores {
	return &BusiestCores{
		coreUsages: make([]float64, 0),
		coreCharts: make(map[int]*CoreSparkline),
	}
}

func (b *BusiestCores) initializeIfNeeded(coreID int) {
	if b.coreCharts[coreID] == nil {
		b.coreCharts[coreID] = NewCoreSparkline(coreID)
		b.coreCharts[coreID].Resize(b.width-b.squareDimension-3, b.height)
	}
}

// Update updates the core usage data and their sparkline charts.
func (b *BusiestCores) Update(metrics domain.CPUMemoryMetrics) {
	if b == nil {
		return
	}
	b.coreUsages = metrics.CPUUsagePerCore
	b.squareDimension = int(math.Ceil(math.Sqrt(float64(len(metrics.CPUUsagePerCore)))))
	for i, usage := range metrics.CPUUsagePerCore {
		b.initializeIfNeeded(i)
		b.coreCharts[i].Update(usage)
		b.coreCharts[i].sparkline.Draw()
	}
}

// GetTopBusiestCores returns the top 5 busiest cores and their usage percentages.
func (b *BusiestCores) GetTopBusiestCores() []int {
	if b == nil || len(b.coreUsages) == 0 {
		return []int{}
	}
	type coreUsage struct {
		Core  int
		Usage float64
	}

	var coreUsages []coreUsage
	for i, usage := range b.coreUsages {
		coreUsages = append(coreUsages, coreUsage{Core: i, Usage: usage})
	}

	sort.Slice(coreUsages, func(i, j int) bool {
		return coreUsages[i].Usage > coreUsages[j].Usage
	})

	topCores := coreUsages
	if len(coreUsages) > 5 {
		topCores = coreUsages[:5]
	}

	result := make([]int, len(topCores))
	for i, core := range topCores {
		result[i] = core.Core
	}

	return result
}

// View renders the sparkline charts for the top 5 busiest cores.
func (b *BusiestCores) View() string {
	topCores := b.GetTopBusiestCores()
	view := "Top 5 Busiest Cores:\n"
	for _, coreID := range topCores {
		view += b.coreCharts[coreID].View() + "\n"
	}
	return view
}

// Resize adjusts the size of the sparkline charts based on available space.
func (b *BusiestCores) Resize(width, height int) {
	if b == nil {
		return
	}
	b.width = width
	b.height = height
	for _, sparkline := range b.coreCharts {
		sparkline.Resize(width-b.squareDimension-3, 1)
	}
}
