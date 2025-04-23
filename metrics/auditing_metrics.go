package metrics

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	auditingMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_auditing_info",
			Help: "Information about auditing files",
		},
		[]string{"file_path", "last_modified", "size"},
	)
)

func RegisterAuditingMetrics() {
	prometheus.MustRegister(auditingMetrics)
}

func CollectAuditingMetrics() {
	// Define directories to audit
	directoriesToAudit := []string{"/var/log", "/etc"}

	for _, dir := range directoriesToAudit {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v", path, err)
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Collect file details
			lastModified := info.ModTime().Format(time.RFC3339)
			size := info.Size()

			// Add file details to Prometheus metric
			auditingMetrics.WithLabelValues(path, lastModified, formatBytes(size)).Set(1)
			return nil
		})

		if err != nil {
			log.Printf("Error walking directory %s: %v", dir, err)
		}
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
