package metrics

import "time"

// CollectSystemMetrics periodically collects all system metrics
func CollectSystemMetrics() {
	go func() {
		for {
			CollectUserMetrics()
			CollectProcessMetrics()
			CollectNetworkMetrics()
			CollectFilesystemMetrics()
			time.Sleep(5 * time.Minute) // Adjust the interval as needed
		}
	}()
}
