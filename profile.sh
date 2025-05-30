#!/bin/bash

# Simple profiling helper script for Mim

echo "Mim Profiling Helper"
echo "==================="
echo ""
echo "1. Run with CPU profiling to file"
echo "2. Run with web-based profiling"
echo "3. Analyze existing cpu.prof"
echo ""
read -p "Select option (1-3): " option

case $option in
    1)
        echo "Starting Mim with CPU profiling..."
        echo "Run for at least 30 seconds, then press 'q' to quit"
        echo ""
        ./mim -cpuprofile=cpu.prof
        echo ""
        echo "Profile saved to cpu.prof"
        echo "To analyze: go tool pprof -http=:8080 cpu.prof"
        ;;
    2)
        echo "Starting Mim with web profiling on http://localhost:6060/debug/pprof"
        echo "Press Ctrl+C to stop"
        echo ""
        echo "While running, you can:"
        echo "  - Get CPU profile: curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof"
        echo "  - View in browser: go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30"
        echo ""
        ./mim -webpprof
        ;;
    3)
        if [ -f "cpu.prof" ]; then
            echo "Opening cpu.prof in web browser..."
            go tool pprof -http=:8080 cpu.prof
        else
            echo "Error: cpu.prof not found. Run option 1 first."
        fi
        ;;
    *)
        echo "Invalid option"
        ;;
esac