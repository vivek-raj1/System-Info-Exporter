package metrics

import (
	"log"
	"os/user"
)

func CollectUserMetrics() {
	users, err := user.Current()
	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		return
	}
	// Use the metric defined in user_info.go
	systemUserMetrics.WithLabelValues(users.Username, users.Uid, users.Gid, users.HomeDir).Set(1)
}
