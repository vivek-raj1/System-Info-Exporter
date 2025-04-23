package metrics

import (
	"fmt"
	"log"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	EXT4_SUPER_MAGIC = 0xEF53     // Magic number for ext4 filesystem
	TMPFS_MAGIC      = 0x01021994 // Magic number for tmpfs filesystem
)

var (
	filesystemMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_filesystem_info",
			Help: "Information about mounted filesystems",
		},
		[]string{"mount_point", "filesystem_type", "total_space", "used_space"},
	)
)

func RegisterFilesystemMetrics() {
	prometheus.MustRegister(filesystemMetrics)
}

func CollectFilesystemMetrics() {
	mountPoints := []string{"/", "/home", "/var"} // Add more mount points as needed
	for _, mountPoint := range mountPoints {
		var stat syscall.Statfs_t
		err := syscall.Statfs(mountPoint, &stat)
		if err != nil {
			log.Printf("Error collecting filesystem metrics for %s: %v", mountPoint, err)
			continue
		}

		totalSpace := int64(stat.Blocks * uint64(stat.Bsize))
		freeSpace := int64(stat.Bfree * uint64(stat.Bsize))
		usedSpace := totalSpace - freeSpace
		filesystemType := getFilesystemType(stat.Type)

		filesystemMetrics.WithLabelValues(
			mountPoint,
			filesystemType,
			formatBytesFilesystem(totalSpace),
			formatBytesFilesystem(usedSpace),
		).Set(1)
	}
}

func getFilesystemType(fsType uint32) string {
	switch fsType {
	case EXT4_SUPER_MAGIC:
		return "ext4"
	case TMPFS_MAGIC:
		return "tmpfs"
	// Add more filesystem types as needed
	default:
		return "unknown"
	}
}

func formatBytesFilesystem(bytes int64) string {
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
