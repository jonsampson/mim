package domain

// CPUTimes represents CPU time measurements for a process
type CPUTimes struct {
	User   float64
	System float64
}

// Total returns the total CPU time (user + system)
func (c CPUTimes) Total() float64 {
	return c.User + c.System
}

// CPUCalculator provides domain logic for CPU-related calculations
type CPUCalculator struct{}

// NewCPUCalculator creates a new CPUCalculator instance
func NewCPUCalculator() *CPUCalculator {
	return &CPUCalculator{}
}

// CalculateCPUPercent calculates CPU percentage using delta time calculations
// Returns 0 if deltaTimeSeconds is <= 0 or if there's no meaningful delta
func (c *CPUCalculator) CalculateCPUPercent(
	currentTimes CPUTimes,
	lastTimes CPUTimes,
	deltaTimeSeconds float64,
) float64 {
	if deltaTimeSeconds <= 0 {
		return 0
	}

	currentTotal := currentTimes.Total()
	lastTotal := lastTimes.Total()
	cpuDelta := currentTotal - lastTotal

	// Avoid negative deltas or very small deltas that could be noise
	if cpuDelta <= 0 {
		return 0
	}

	return (cpuDelta / deltaTimeSeconds) * 100.0
}

// CalculateOverallCPUPercent calculates total CPU usage across all cores
// Takes per-core usage percentages and returns weighted average
func (c *CPUCalculator) CalculateOverallCPUPercent(coreUsages []float64) float64 {
	if len(coreUsages) == 0 {
		return 0
	}

	var total float64
	for _, usage := range coreUsages {
		total += usage
	}

	return total / float64(len(coreUsages))
}

// ValidateCPUPercent validates that a CPU percentage is within expected bounds
func (c *CPUCalculator) ValidateCPUPercent(percent float64) bool {
	return percent >= 0 && percent <= 100
}