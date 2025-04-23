package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"runtime"
	"time"

	"package_exporter/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Define command-line flags for address, port, interval, CPU (millicores), and memory limits
	address := flag.String("address", "0.0.0.0", "Address to bind the HTTP server")
	port := flag.String("port", "9101", "Port to bind the HTTP server")
	interval := flag.Int("interval", 30, "Interval (in minutes) to collect metrics")
	cpuMillicores := flag.Int("resource.cpu", 0, "Maximum CPU usage in millicores (0 for no limit)")
	memoryLimit := flag.Int64("resource.memory", 0, "Maximum memory usage in MB (0 for no limit)")
	flag.Parse()

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

	// Start a goroutine to periodically collect metrics based on the interval
	go func() {
		ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
		defer ticker.Stop()

		for {
			log.Println("Collecting metrics...")
			metrics.CollectPackageVersions()
			metrics.CollectOSInfo()
			metrics.CollectPackageUpdateAvailability()
			<-ticker.C
		}
	}()

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
}
