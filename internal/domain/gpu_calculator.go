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

// CalculateProcessMemoryPercent calculates memory percentage for a specific process
// using the total GPU memory available
func (g *GPUCalculator) CalculateProcessMemoryPercent(processMemory, totalGPUMemory uint64) float64 {
	return g.CalculateMemoryPercent(processMemory, totalGPUMemory)
}