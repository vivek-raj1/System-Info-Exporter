package metrics

import (
	"bufio"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	systemUserMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_user_info",
			Help: "Information about system users, including username, home directory, UID, and active status",
		},
		[]string{"username", "home_directory", "uid", "active"},
	)
)

func init() {
	prometheus.MustRegister(systemUserMetrics)
}

func CollectSystemUserMetrics() {
	users, err := fetchAllUsers()
	if err != nil {
		log.Printf("Error fetching user information: %v", err)
		return
	}

	for _, user := range users {
		uid, err := strconv.Atoi(user.Uid)
		if err != nil {
			log.Printf("Error parsing UID for user %s: %v", user.Username, err)
			continue
		}

		active := "0"
		if isUserActive(uid) {
			active = "1"
		}

		systemUserMetrics.WithLabelValues(user.Username, user.HomeDir, user.Uid, active).Set(1)
	}
}

func fetchAllUsers() ([]*user.User, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []*user.User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 7 {
			users = append(users, &user.User{
				Username: fields[0],
				Uid:      fields[2],
				Gid:      fields[3],
				HomeDir:  fields[5],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func isUserActive(uid int) bool {
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		log.Printf("Error reading %s: %v", procDir, err)
		return false
	}

	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err == nil {
			stat := &syscall.Stat_t{}
			if err := syscall.Stat(procDir+"/"+entry.Name(), stat); err == nil && int(stat.Uid) == uid {
				return true
			}
		}
	}
	return false
}
