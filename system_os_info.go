package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"package_exporter/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Define command-line flags for address, port, and interval
	address := flag.String("address", "0.0.0.0", "Address to bind the HTTP server")
	port := flag.String("port", "9101", "Port to bind the HTTP server")
	interval := flag.Int("interval", 30, "Interval (in minutes) to collect metrics")
	flag.Parse()

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
