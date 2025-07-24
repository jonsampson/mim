package infra

import (
	"time"

	"github.com/jonsampson/mim/internal/domain"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type CPUMemoryCollector struct {
	*BaseCollector[domain.CPUMemoryMetrics]
	lastProcessTimes map[int32]*cpu.TimesStat
	lastCollectTime  time.Time
	cpuCalculator    *domain.CPUCalculator
	processFilter    *domain.ProcessFilter
	usernameCache    *UsernameCache
	// Pre-allocated buffers to reduce GC pressure
	processInfoBuffer []domain.CPUProcessInfo
}

func NewCPUMemoryCollector() *CPUMemoryCollector {
	collector := &CPUMemoryCollector{
		lastProcessTimes:  make(map[int32]*cpu.TimesStat),
		lastCollectTime:   time.Now(),
		cpuCalculator:     domain.NewCPUCalculator(),
		processFilter:     domain.NewProcessFilter(),
		usernameCache:     NewUsernameCache(),
		processInfoBuffer: make([]domain.CPUProcessInfo, 0, 1000), // Pre-allocate for ~1000 processes
	}
	collector.BaseCollector = NewBaseCollector(collector.getMetrics)
	return collector
}

func (c *CPUMemoryCollector) getMetrics() (domain.CPUMemoryMetrics, error) {
	type result struct {
		value any
		err   error
	}

	perCoreChan := make(chan result)
	totalChan := make(chan result)
	memChan := make(chan result)
	processesChan := make(chan result)

	// Collect per-core CPU usage
	go func() {
		cpuUsagePerCore, err := cpu.Percent(0, true) // Non-blocking using cached measurements
		perCoreChan <- result{cpuUsagePerCore, err}
	}()

	// Collect total CPU usage
	go func() {
		cpuUsageTotal, err := cpu.Percent(0, false) // Non-blocking using cached measurements
		if err != nil {
			totalChan <- result{nil, err}
		} else {
			totalChan <- result{cpuUsageTotal[0], nil}
		}
	}()

	// Get memory stats
	go func() {
		memStat, err := mem.VirtualMemory()
		if err != nil {
			memChan <- result{nil, err}
		} else {
			memChan <- result{memStat.UsedPercent, nil}
		}
	}()

	// Get process information
	go func() {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(c.lastCollectTime).Seconds()
		
		// Get all processes (more efficient than Pids() + individual NewProcess calls)
		processes, err := process.Processes()
		if err != nil {
			processesChan <- result{nil, err}
			return
		}
		
		// Reuse pre-allocated buffer to reduce GC pressure
		c.processInfoBuffer = c.processInfoBuffer[:0] // Reset length but keep capacity
		newProcessTimes := make(map[int32]*cpu.TimesStat)
		
		for _, proc := range processes {
			pid := proc.Pid
			
			// Get current CPU times (single /proc read per process)
			currentTimes, err := proc.Times()
			if err != nil {
				continue
			}
			
			// Store for next iteration
			newProcessTimes[pid] = currentTimes
			
			// Get process name for filtering
			name, _ := proc.Name()
			// Apply domain filtering rules
			if !c.processFilter.ShouldIncludeProcess(name) {
				continue
			}
			
			// Calculate CPU percentage using domain service
			var cpuPercent float64
			if lastTimes, exists := c.lastProcessTimes[pid]; exists {
				currentCPUTimes := domain.CPUTimes{
					User:   currentTimes.User,
					System: currentTimes.System,
				}
				lastCPUTimes := domain.CPUTimes{
					User:   lastTimes.User,
					System: lastTimes.System,
				}
				cpuPercent = c.cpuCalculator.CalculateCPUPercent(currentCPUTimes, lastCPUTimes, deltaTime)
			}
			
			// Get memory percentage for all processes (needed for proper sorting)
			memPercent, err := proc.MemoryPercent()
			if err != nil {
				continue
			}
			
			// Get username using cache (fast after first few lookups due to UID deduplication)
			username := c.usernameCache.GetUsername(uint32(pid))
			
			c.processInfoBuffer = append(c.processInfoBuffer, domain.CPUProcessInfo{
				Pid:           uint32(pid),
				CPUPercent:    cpuPercent,
				MemoryPercent: float64(memPercent),
				Command:       name,
				User:          username,
			})
		}
		
		// Update stored times and timestamp
		c.lastProcessTimes = newProcessTimes
		c.lastCollectTime = currentTime
		
		processesChan <- result{c.processInfoBuffer, nil}
	}()

	// Collect results and handle potential errors
	metrics := domain.CPUMemoryMetrics{}
	var err error

	for range 4 {
		select {
		case r := <-perCoreChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.CPUUsagePerCore = r.value.([]float64)
		case r := <-totalChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.CPUUsageTotal = r.value.(float64)
		case r := <-memChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.MemoryUsage = r.value.(float64)
		case r := <-processesChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.Processes = r.value.([]domain.CPUProcessInfo)
		}
	}

	// Return the collected metrics
	return metrics, err
}
