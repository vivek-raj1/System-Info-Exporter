package metrics

import "time"

// CollectSystemMetrics periodically collects all system metrics
func CollectSystemMetrics() {
	go func() {
		for {
			CollectSystemUserMetrics() // Ensure user metrics are collected
			CollectUserMetrics()
			CollectProcessMetrics()
			CollectNetworkMetrics()
			CollectFilesystemMetrics()
			time.Sleep(5 * time.Minute) // Adjust the interval as needed
		}
	}()
}
