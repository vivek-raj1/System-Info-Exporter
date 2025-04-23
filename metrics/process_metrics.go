package metrics

import (
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/process"
)

// Ensure the gopsutil library is added to your project dependencies:
// Run the following command in your terminal:
// go get github.com/shirou/gopsutil/process

var (
	processMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_process_info",
			Help: "Information about running processes",
		},
		[]string{"pid", "name", "user"},
	)
)

func RegisterProcessMetrics() {
	prometheus.MustRegister(processMetrics)
}

func CollectProcessMetrics() {
	if runtime.GOOS == "linux" {
		collectLinuxProcessMetrics()
	} else if runtime.GOOS == "darwin" {
		collectMacOSProcessMetrics()
	} else {
		log.Printf("Process metrics collection is not supported on %s", runtime.GOOS)
	}
}

func collectLinuxProcessMetrics() {
	procDir := "/proc"
	files, err := ioutil.ReadDir(procDir)
	if err != nil {
		log.Printf("Error reading /proc directory: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		pid := file.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue // Skip non-numeric directories
		}

		statusPath := procDir + "/" + pid + "/status"
		data, err := ioutil.ReadFile(statusPath)
		if err != nil {
			log.Printf("Error reading process status for PID %s: %v", pid, err)
			continue
		}

		var name, user string
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Name:") {
				name = strings.TrimSpace(strings.Split(line, ":")[1])
			} else if strings.HasPrefix(line, "Uid:") {
				uid := strings.Fields(strings.Split(line, ":")[1])[0]
				user = uid // You can resolve UID to username if needed
			}
		}

		processMetrics.WithLabelValues(pid, name, user).Set(1)
	}
}

func collectMacOSProcessMetrics() {
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error fetching process list: %v", err)
		return
	}

	for _, proc := range processes {
		pid := strconv.Itoa(int(proc.Pid))
		name, err := proc.Name()
		if err != nil {
			log.Printf("Error fetching process name for PID %s: %v", pid, err)
			continue
		}
		username, err := proc.Username()
		if err != nil {
			log.Printf("Error fetching username for PID %s: %v", pid, err)
			continue
		}

		processMetrics.WithLabelValues(pid, name, username).Set(1)
	}
}
