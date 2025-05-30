package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsampson/mim/internal/infra"
	"github.com/jonsampson/mim/internal/tui"
)

func main() {
	// Parse command line flags
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var webpprof = flag.Bool("webpprof", false, "enable web-based pprof on :6060")
	flag.Parse()

	// Start web-based pprof if requested
	if *webpprof {
		go func() {
			log.Println("Starting pprof server on http://localhost:6060/debug/pprof")
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Start CPU profiling if requested
	var cpuProfileFile *os.File
	if *cpuprofile != "" {
		var err error
		cpuProfileFile, err = os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
			cpuProfileFile.Close()
			log.Fatal("could not start CPU profile: ", err)
		}
		log.Printf("CPU profiling started, writing to %s", *cpuprofile)
		
		// Ensure profiling is stopped on exit
		defer func() {
			pprof.StopCPUProfile()
			cpuProfileFile.Close()
			log.Printf("CPU profile written to %s", *cpuprofile)
		}()
	}

	// Setup signal handler for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if cpuProfileFile != nil {
			pprof.StopCPUProfile()
			cpuProfileFile.Close()
			log.Printf("CPU profile written to %s (interrupted)", *cpuprofile)
		}
		os.Exit(0)
	}()

	// Open a log file for debugging
	logFile, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set up the log package to write to the log file
	log.SetOutput(logFile)

	factory := infra.CollectorFactory{}
	collectors := factory.CreateCollectors()

	// Initialize the model without specifying the initial size
	model, err := tui.InitialModel(collectors...)
	if err != nil {
		log.Printf("Error initializing model: %v\n", err)
		fmt.Printf("Error initializing model: %v\n", err)
		os.Exit(1)
	}

	// Initialize the Bubble Tea program
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Printf("Alas, there's been an error: %v", err)
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
