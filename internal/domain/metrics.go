package domain

type CPUMemoryMetrics struct {
	CPUUsage    []float64
	MemoryUsage float64
}

type GPUMetrics struct {
	GPUUsage       float64
	GPUMemoryUsage float64
}
