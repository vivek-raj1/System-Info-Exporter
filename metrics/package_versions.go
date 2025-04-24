package metrics

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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

func CollectPackageVersions(debug bool) {
	if debug {
		log.Println("Debug: Starting package version collection")
	}

	switch runtime.GOOS {
	case "linux":
		if debug {
			log.Println("Debug: Collecting package versions for Linux")
		}
		collectLinuxPackageVersions(debug)
	case "darwin":
		if debug {
			log.Println("Debug: Collecting package versions for macOS")
		}
		collectMacOSPackageVersions(debug)
	default:
		log.Println("Unsupported operating system")
	}

	if debug {
		log.Println("Debug: Finished package version collection")
	}
}

func collectLinuxPackageVersions(debug bool) {
	if _, err := os.Stat("/var/lib/dpkg/status"); err == nil {
		// Use dpkg for Debian/Ubuntu-based systems
		parseDpkgStatusFile("/var/lib/dpkg/status")
		return
	}

	if _, err := exec.LookPath("yum"); err == nil {
		// Use yum for Red Hat-based systems
		output, err := exec.Command("yum", "list", "installed").Output()
		if err != nil {
			log.Printf("Error querying installed packages using yum: %v", err)
			return
		}
		parseYumOutput(output)
		return
	}

	if _, err := exec.LookPath("dnf"); err == nil {
		// Use dnf for newer Red Hat-based systems
		output, err := exec.Command("dnf", "list", "installed").Output()
		if err != nil {
			log.Printf("Error querying installed packages using dnf: %v", err)
			return
		}
		parseYumOutput(output)
		return
	}

	log.Println("No supported package manager found on this system")
}

func parseYumOutput(output []byte) {
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip headers and irrelevant lines
		if strings.HasPrefix(line, "Installed Packages") || strings.HasPrefix(line, "Loaded plugins:") || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse package name and version
		fields := strings.Fields(line)
		if len(fields) >= 3 { // Ensure there are enough fields for package and version
			packageName := fields[0]
			version := fields[1]
			packageVersions.WithLabelValues(packageName, version).Set(1) // Provide both labels
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading yum/dnf list output: %v", err)
	}
}

func parseVersionToFloat(version string) float64 {
	// Replace non-numeric characters with dots and parse as float
	version = strings.ReplaceAll(version, "-", ".")
	version = strings.ReplaceAll(version, "_", ".")
	parts := strings.Split(version, ".")
	var numericVersion string
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			numericVersion += part
		} else {
			break
		}
	}
	floatVersion, err := strconv.ParseFloat(numericVersion, 64)
	if err != nil {
		log.Printf("Error parsing version %s: %v", version, err)
		return 0
	}
	return floatVersion
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

func collectMacOSPackageVersions(debug bool) {
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
