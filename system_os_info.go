package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"runtime"
	"time"

	"system_os_info/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Define command-line flags for address, port, interval, CPU (millicores), and memory limits
	address := flag.String("address", "0.0.0.0", "Address to bind the HTTP server")
	port := flag.String("port", "9101", "Port to bind the HTTP server")
	interval := flag.Int("interval", 30, "Interval (in minutes) to collect metrics")
	cpuMillicores := flag.Int("resource.cpu", 0, "Maximum CPU usage in millicores (0 for no limit)")
	memoryLimit := flag.Int64("resource.memory", 0, "Maximum memory usage in MB (0 for no limit)")

	// Add flags for enabling filesystem and process metrics
	enableFilesystem := flag.Bool("filesystem", false, "Enable collection of filesystem metrics")
	enableProcess := flag.Bool("process", false, "Enable collection of process metrics")

	// Add a flag for enabling debug mode
	debugMode := flag.Bool("debug", false, "Enable debug mode with detailed logs")

	// Add flags for enabling auditing files and scheduled jobs metrics
	enableAuditing := flag.Bool("auditing", false, "Enable collection of auditing files metrics")
	enableScheduledJobs := flag.Bool("scheduled-jobs", false, "Enable collection of scheduled jobs metrics")
	flag.Parse()

	// Set log level based on debug mode
	if *debugMode {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Debug mode enabled")
	}

	// Set CPU limit based on millicores
	if *cpuMillicores > 0 {
		numCores := int(math.Ceil(float64(*cpuMillicores) / 1000.0))
		runtime.GOMAXPROCS(numCores)
		log.Printf("CPU usage limited to %d millicores (~%d cores)", *cpuMillicores, numCores)
	}

	// Start a goroutine to monitor memory usage if a limit is set
	if *memoryLimit > 0 {
		go func() {
			memoryLimitBytes := *memoryLimit * 1024 * 1024
			for {
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				if memStats.Alloc > uint64(memoryLimitBytes) {
					log.Fatalf("Memory usage exceeded the limit of %d MB", *memoryLimit)
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	// Example debug log
	if *debugMode {
		log.Println("Debug: Starting metric collection loop")
	}

	// Conditionally register metrics based on flags
	if *enableFilesystem {
		log.Println("Registering filesystem metrics...")
		metrics.RegisterFilesystemMetrics()
	}

	if *enableProcess {
		log.Println("Registering process metrics...")
		metrics.RegisterProcessMetrics()
	}

	if *enableAuditing {
		log.Println("Registering auditing files metrics...")
		metrics.RegisterAuditingMetrics()
	}

	if *enableScheduledJobs {
		log.Println("Registering scheduled jobs metrics...")
		metrics.RegisterScheduledJobsMetrics()
	}

	// Start a goroutine to periodically collect metrics based on the interval
	go func() {
		ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
		defer ticker.Stop()

		for {
			if *debugMode {
				log.Println("Debug: Collecting metrics...")
			}
			metrics.CollectSystemUserMetrics(*debugMode) // Pass debug flag
			metrics.CollectPackageVersions(*debugMode)   // Pass debug flag
			metrics.CollectOSInfo()
			metrics.CollectPackageUpdateAvailability()

			// Collect filesystem metrics if enabled
			if *enableFilesystem {
				if *debugMode {
					log.Println("Debug: Collecting filesystem metrics...")
				}
				metrics.CollectFilesystemMetrics()
			}

			// Collect process metrics if enabled
			if *enableProcess {
				if *debugMode {
					log.Println("Debug: Collecting process metrics...")
				}
				metrics.CollectProcessMetrics()
			}

			if *enableAuditing {
				if *debugMode {
					log.Println("Debug: Collecting auditing files metrics...")
				}
				metrics.CollectAuditingMetrics()
			}

			if *enableScheduledJobs {
				if *debugMode {
					log.Println("Debug: Collecting scheduled jobs metrics...")
				}
				metrics.CollectScheduledJobsMetrics()
			}

			<-ticker.C
		}
	}()

	// Start a goroutine to collect system metrics
	go metrics.CollectSystemMetrics(*debugMode) // Pass debug flag

	// Call any initialization logic from other files
	// metrics.InitializeSystemMetrics()

	// Expose metrics
	http.Handle("/metrics", promhttp.Handler())

	// Add a handler for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>System OS info Exporter</title>
			</head>
			<body>
				<h1>System OS info Exporter</h1>
				<p>This exporter collects and exposes system information and package details as Prometheus metrics.</p>
				<p>Visit the <a href="/metrics">/metrics</a> page to view the metrics.</p>
			</body>
			</html>
		`))
	})

	serverAddress := *address + ":" + *port
	log.Printf("Starting Prometheus exporter on %s", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}

	log.Println("Application started")
}
