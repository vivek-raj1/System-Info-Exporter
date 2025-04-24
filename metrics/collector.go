package metrics

import "time"

// CollectSystemMetrics periodically collects all system metrics
func CollectSystemMetrics(debug bool) {
	go func() {
		for {
			CollectSystemUserMetrics(debug) // Pass debug flag
			CollectUserMetrics()
			CollectProcessMetrics()
			CollectNetworkMetrics()
			CollectFilesystemMetrics()
			time.Sleep(5 * time.Minute) // Adjust the interval as needed
		}
	}()
}
