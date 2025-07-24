package domain

// ProcessFilter provides domain logic for determining which processes to include
type ProcessFilter struct{}

// NewProcessFilter creates a new ProcessFilter instance
func NewProcessFilter() *ProcessFilter {
	return &ProcessFilter{}
}

// ShouldIncludeProcess determines if a process should be included based on business rules
func (f *ProcessFilter) ShouldIncludeProcess(processName string) bool {
	// Skip kernel threads (names starting with '[' and ending with ']')
	if len(processName) > 0 && processName[0] == '[' {
		return false
	}
	
	// Skip empty process names
	if len(processName) == 0 {
		return false
	}
	
	return true
}

// FilterCPUProcesses filters a slice of CPU processes based on inclusion rules
func (f *ProcessFilter) FilterCPUProcesses(processes []CPUProcessInfo) []CPUProcessInfo {
	filtered := make([]CPUProcessInfo, 0, len(processes))
	for _, p := range processes {
		if f.ShouldIncludeProcess(p.Command) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// FilterGPUProcesses filters a slice of GPU processes based on inclusion rules
func (f *ProcessFilter) FilterGPUProcesses(processes []GPUProcessInfo) []GPUProcessInfo {
	filtered := make([]GPUProcessInfo, 0, len(processes))
	for _, p := range processes {
		// For GPU processes, we typically want to include all processes
		// since GPU usage is more rare and valuable to monitor
		// But we can still filter based on basic rules
		if len(processes) == 0 {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

// IsKernelThread checks if a process name indicates a kernel thread
func (f *ProcessFilter) IsKernelThread(processName string) bool {
	return len(processName) > 0 && processName[0] == '['
}

// IsValidProcessName checks if a process name is valid and non-empty
func (f *ProcessFilter) IsValidProcessName(processName string) bool {
	return len(processName) > 0
}
