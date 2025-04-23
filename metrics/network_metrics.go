package metrics

import (
	"log"
	"net"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	networkMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_network_info",
			Help: "Information about network interfaces",
		},
		[]string{"interface", "ip_address", "mac_address"},
	)
)

func init() {
	prometheus.MustRegister(networkMetrics)
}

func CollectNetworkMetrics() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error fetching network interfaces: %v", err)
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			// Skip interfaces that are down
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Error fetching addresses for interface %s: %v", iface.Name, err)
			continue
		}

		ipAddress := "unknown"
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err == nil && ip.To4() != nil {
				ipAddress = ip.String()
				break
			}
		}

		macAddress := iface.HardwareAddr.String()
		if macAddress == "" {
			macAddress = "unknown"
		}

		networkMetrics.WithLabelValues(iface.Name, ipAddress, macAddress).Set(1)
	}
}
