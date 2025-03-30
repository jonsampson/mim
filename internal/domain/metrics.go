package domain

type CPUMemoryMetrics struct {
    CPUUsagePerCore []float64
    CPUUsageTotal   float64
    MemoryUsage     float64
}

type GPUMetrics struct {
	GPUUsage       float64
	GPUMemoryUsage float64
}
