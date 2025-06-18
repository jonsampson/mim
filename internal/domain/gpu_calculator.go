package domain

// GPUCalculator provides domain logic for GPU-related calculations
type GPUCalculator struct{}

// NewGPUCalculator creates a new GPUCalculator instance
func NewGPUCalculator() *GPUCalculator {
	return &GPUCalculator{}
}

// CalculateMemoryPercent calculates GPU memory usage percentage
// Returns 0 if total is 0 to avoid division by zero
func (g *GPUCalculator) CalculateMemoryPercent(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return (float64(used) / float64(total)) * 100.0
}

// CalculateUtilizationPercent converts GPU utilization from uint32 to float64
// and validates the range
func (g *GPUCalculator) CalculateUtilizationPercent(utilization uint32) float64 {
	percent := float64(utilization)
	// Clamp to valid range (0-100)
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// ValidateMemoryPercent validates that a memory percentage is within expected bounds
func (g *GPUCalculator) ValidateMemoryPercent(percent float64) bool {
	return percent >= 0 && percent <= 100
}

// ValidateUtilizationPercent validates that a utilization percentage is within expected bounds
func (g *GPUCalculator) ValidateUtilizationPercent(percent float64) bool {
	return percent >= 0 && percent <= 100
}

// CalculateProcessMemoryPercent calculates memory percentage for a specific process
// using the total GPU memory available
func (g *GPUCalculator) CalculateProcessMemoryPercent(processMemory, totalGPUMemory uint64) float64 {
	return g.CalculateMemoryPercent(processMemory, totalGPUMemory)
}