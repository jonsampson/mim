# Profiling Guide for Mim

This guide explains how to profile the Mim application to identify performance bottlenecks.

## CPU Profiling

### Method 1: File-based CPU Profile

1. Run the application with CPU profiling enabled:
```bash
./mim -cpuprofile=cpu.prof
```

2. Let the application run for at least 30 seconds to collect meaningful data, then quit (press 'q')

3. Analyze the profile:
```bash
# Text-based analysis
go tool pprof cpu.prof

# Web-based visualization (opens in browser)
go tool pprof -http=:8080 cpu.prof
```

### Method 2: Web-based Live Profiling

1. Run the application with web pprof enabled:
```bash
./mim -webpprof
```

2. While the application is running, access profiles at:
- CPU Profile: http://localhost:6060/debug/pprof/profile?seconds=30
- Heap Profile: http://localhost:6060/debug/pprof/heap
- Goroutine Profile: http://localhost:6060/debug/pprof/goroutine
- All profiles: http://localhost:6060/debug/pprof/

3. Download and analyze a 30-second CPU profile:
```bash
# Download profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Or analyze directly
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

## Common pprof Commands

Once in the pprof interactive mode:
- `top` - Show top functions by CPU time
- `top -cum` - Show top functions by cumulative CPU time
- `list <function>` - Show source code for a function
- `web` - Generate a graph (requires graphviz)
- `pdf` - Generate a PDF report

## What to Look For

1. **High Self Time**: Functions that consume CPU directly
2. **High Cumulative Time**: Functions whose callees consume CPU
3. **Frequent Allocations**: Check the heap profile for memory pressure
4. **Lock Contention**: Check the mutex profile if enabled

## Example Analysis Session

```bash
# Generate a CPU profile
./mim -cpuprofile=cpu.prof

# In another terminal, after 30+ seconds
go tool pprof cpu.prof
(pprof) top 10
(pprof) list <suspicious_function>
(pprof) web
```

## Tips for Effective Profiling

1. Profile under realistic load - ensure all collectors are running
2. Profile for at least 30 seconds to get statistically significant data
3. Compare profiles before and after optimizations
4. Focus on the hottest paths first (highest CPU consumers)
5. Don't forget to check memory allocations - they can cause GC pressure

## Known Performance Areas

Based on the architecture, pay attention to:
- Metric collection loops (1-second intervals)
- View rendering functions (called every update)
- String formatting and concatenation
- Sorting operations
- Terminal rendering overhead