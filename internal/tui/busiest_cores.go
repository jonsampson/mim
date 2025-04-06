package tui

import (
    "math"
    "sort"
    "strings"

    "github.com/jonsampson/mim/internal/domain"
)

type BusiestCores struct {
    coreUsages      []float64
    coreCharts      map[int]*CoreSparkline
    width           int
    height          int
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
        b.coreCharts[coreID].Resize(b.width-b.squareDimension-3, 1)
    }
}

// Update handles incoming messages and updates the internal state
func (b *BusiestCores) Update(msg interface{}) {
    switch msg := msg.(type) {
    case domain.CPUMemoryMetrics:
        b.updateMetrics(msg)
    }
}

// updateMetrics updates the core usage data and their sparkline charts.
func (b *BusiestCores) updateMetrics(metrics domain.CPUMemoryMetrics) {
    if b == nil {
        return
    }
    b.coreUsages = metrics.CPUUsagePerCore
    b.squareDimension = int(math.Ceil(math.Sqrt(float64(len(metrics.CPUUsagePerCore)))))
    for i, usage := range metrics.CPUUsagePerCore {
        b.initializeIfNeeded(i)
        b.coreCharts[i].Update(usage)
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
    views := []string{"Top 5 Busiest Cores:"}
    for _, coreID := range topCores {
        views = append(views, b.coreCharts[coreID].View())
    }
    return strings.Join(views, "\n")
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