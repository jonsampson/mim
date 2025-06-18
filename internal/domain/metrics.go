package domain

type CPUMemoryMetrics struct {
	CPUUsagePerCore []float64
	CPUUsageTotal   float64
	MemoryUsage     float64
	Processes       []CPUProcessInfo
}

type CPUProcessInfo struct {
	Pid           uint32
	CPUPercent    float64
	MemoryPercent float64
	Command       string
	User          string
}

type GPUMetrics struct {
	GPUUsage       float64
	GPUMemoryUsage float64
	Processes      []GPUProcessInfo
}

type GPUProcessInfo struct {
	Pid           uint32
	SmUtil        uint32
	UsedGpuMemory float64
	User          string
}
