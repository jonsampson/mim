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