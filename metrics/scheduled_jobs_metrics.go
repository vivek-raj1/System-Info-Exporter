package metrics

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	scheduledJobsMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_scheduled_jobs_info",
			Help: "Information about scheduled jobs",
		},
		[]string{"job_name", "schedule", "last_run_status"},
	)
)

func RegisterScheduledJobsMetrics() {
	prometheus.MustRegister(scheduledJobsMetrics)
}

func CollectScheduledJobsMetrics() {
	cronFile := "/etc/crontab" // Path to the system crontab file
	file, err := os.Open(cronFile)
	if err != nil {
		log.Printf("Error opening crontab file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			log.Printf("Invalid crontab entry: %s", line)
			continue
		}

		// Extract schedule and command
		schedule := strings.Join(fields[:5], " ")
		jobName := fields[5]
		lastRunStatus := "unknown" // Placeholder for last run status

		// Add the job details to the Prometheus metric
		scheduledJobsMetrics.WithLabelValues(jobName, schedule, lastRunStatus).Set(1)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading crontab file: %v", err)
	}
}
