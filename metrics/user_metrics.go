package metrics

import (
	"log"
	"os/user"
)

func CollectUserMetrics() {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error fetching current user info: %v", err)
		return
	}

	// Add GID to match the updated metric definition
	systemUserMetrics.WithLabelValues(
		currentUser.Username,
		currentUser.HomeDir,
		currentUser.Uid,
		currentUser.Gid,
		"1", // Assume the current user is always active
	).Set(1)
}
