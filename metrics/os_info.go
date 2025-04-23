package metrics

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var osInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "system_os_info",
		Help: "Operating system name, version, architecture, platform, and kernel version",
	},
	[]string{"os_name", "os_version", "architecture", "platform", "kernel_version"},
)

func init() {
	prometheus.MustRegister(osInfo)
}

func CollectOSInfo() {
	architecture := runtime.GOARCH
	platform := runtime.GOOS

	// Fetch kernel version
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error fetching kernel version: %v", err)
		return
	}
	kernelVersion := strings.TrimSpace(string(output))

	switch platform {
	case "linux":
		collectLinuxOSInfo(architecture, platform, kernelVersion)
	case "darwin":
		collectMacOSInfo(architecture, platform, kernelVersion)
	default:
		log.Println("OS information collection is not supported on this platform")
	}
}

func collectLinuxOSInfo(architecture, platform, kernelVersion string) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		log.Printf("Error opening /etc/os-release: %v", err)
		return
	}
	defer file.Close()

	var osName, osVersion string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "NAME=") && osName == "" {
			osName = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") && osVersion == "" {
			osVersion = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading /etc/os-release: %v", err)
		return
	}

	if osName == "" || osVersion == "" {
		log.Println("Failed to extract OS name or version from /etc/os-release")
		return
	}

	osInfo.WithLabelValues(osName, osVersion, architecture, platform, kernelVersion).Set(1)
}

func collectMacOSInfo(architecture, platform, kernelVersion string) {
	cmd := exec.Command("sw_vers")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error fetching macOS version: %v", err)
		return
	}

	var osName, osVersion string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ProductName:") {
			osName = strings.TrimSpace(strings.TrimPrefix(line, "ProductName:"))
		} else if strings.HasPrefix(line, "ProductVersion:") {
			osVersion = strings.TrimSpace(strings.TrimPrefix(line, "ProductVersion:"))
		}
	}

	if osName == "" || osVersion == "" {
		log.Println("Failed to extract macOS name or version")
		return
	}

	osInfo.WithLabelValues(osName, osVersion, architecture, platform, kernelVersion).Set(1)
}
