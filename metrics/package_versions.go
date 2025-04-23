package metrics

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var packageVersions = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "system_package_version",
		Help: "Version of installed packages",
	},
	[]string{"package", "version"},
)

func init() {
	prometheus.MustRegister(packageVersions)
}

func CollectPackageVersions() {
	switch runtime.GOOS {
	case "linux":
		collectLinuxPackageVersions()
	case "darwin":
		collectMacOSPackageVersions()
	default:
		log.Println("Unsupported operating system")
	}
}

func collectLinuxPackageVersions() {
	if _, err := os.Stat("/var/lib/dpkg/status"); err == nil {
		parseDpkgStatusFile("/var/lib/dpkg/status")
	} else if _, err := os.Stat("/var/lib/rpm"); err == nil {
		log.Println("Parsing RPM database is not implemented in pure Go")
	} else {
		log.Println("No supported package manager database found on this system")
	}
}

func parseDpkgStatusFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening dpkg status file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var packageName, version string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package: ") {
			packageName = strings.TrimPrefix(line, "Package: ")
		} else if strings.HasPrefix(line, "Version: ") {
			version = strings.TrimPrefix(line, "Version: ")
			if packageName != "" && version != "" {
				packageVersions.WithLabelValues(packageName, version).Set(1)
				packageName, version = "", ""
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading dpkg status file: %v", err)
	}
}

func collectMacOSPackageVersions() {
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
		return
	}

	entries, err := os.ReadDir(homebrewCellar)
	if err != nil {
		log.Printf("Error reading Homebrew cellar directory: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			packageName := entry.Name()
			versionEntries, err := os.ReadDir(homebrewCellar + "/" + packageName)
			if err != nil || len(versionEntries) == 0 {
				continue
			}
			version := versionEntries[0].Name()
			packageVersions.WithLabelValues(packageName, version).Set(1)
		}
	}
}
