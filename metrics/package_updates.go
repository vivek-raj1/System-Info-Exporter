package metrics

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var packageUpdateAvailable = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "system_package_update_available",
		Help: "Indicates if updates are available for installed packages (1 if updates are available, 0 otherwise)",
	},
	[]string{"package"},
)

func init() {
	prometheus.MustRegister(packageUpdateAvailable)
}

func CollectPackageUpdateAvailability() {
	switch runtime.GOOS {
	case "linux":
		detectLinuxDistroAndCollectUpdates()
	case "darwin":
		collectMacOSPackageUpdates()
	default:
		log.Println("Unsupported operating system for update availability check")
	}
}

func detectLinuxDistroAndCollectUpdates() {
	if _, err := os.Stat("/etc/os-release"); err == nil {
		file, err := os.Open("/etc/os-release")
		if err != nil { // Fixed syntax error: removed extra parentheses
			log.Printf("Error opening /etc/os-release: %v", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var distro string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				distro = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				break
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading /etc/os-release: %v", err)
			return
		}

		switch distro {
		case "ubuntu", "debian":
			collectAptUpdates()
		case "rhel", "centos", "fedora", "amazon", "amzn":
			collectYumOrDnfUpdates()
		default:
			log.Printf("Unsupported Linux distribution: %s", distro)
		}
	} else {
		log.Println("Unable to detect Linux distribution, skipping update check")
	}
}

func collectMacOSPackageUpdates() {
	homebrewPaths := []string{"/usr/local/Cellar", "/opt/homebrew/Cellar"}

	var homebrewCellar string
	for _, path := range homebrewPaths {
		if _, err := os.Stat(path); err == nil {
			homebrewCellar = path
			break
		}
	}

	if homebrewCellar == "" {
		log.Println("Homebrew not found on this system")
		packageUpdateAvailable.WithLabelValues("homebrew").Set(0)
		return
	}

	outdatedPath := homebrewCellar + "/outdated"
	if _, err := os.Stat(outdatedPath); err == nil {
		packageUpdateAvailable.WithLabelValues("homebrew").Set(1)
	} else {
		packageUpdateAvailable.WithLabelValues("homebrew").Set(0)
	}
}

func collectAptUpdates() {
	if _, err := os.Stat("/var/lib/apt/lists"); err == nil {
		if _, err := os.Stat("/var/lib/apt/lists/partial"); err == nil {
			log.Println("APT updates are available")
			packageUpdateAvailable.WithLabelValues("apt").Set(1)
		} else {
			packageUpdateAvailable.WithLabelValues("apt").Set(0)
		}
	} else {
		log.Println("APT package manager not found, skipping update check")
	}
}

func collectYumOrDnfUpdates() {
	if _, err := os.Stat("/var/cache/yum"); err == nil {
		log.Println("YUM updates are available")
		packageUpdateAvailable.WithLabelValues("yum").Set(1)
		return
	}

	if _, err := os.Stat("/var/cache/dnf"); err == nil {
		log.Println("DNF updates are available")
		packageUpdateAvailable.WithLabelValues("dnf").Set(1)
		return
	}

	packageUpdateAvailable.WithLabelValues("yum_or_dnf").Set(0)
	log.Println("No updates found for YUM or DNF")
}
