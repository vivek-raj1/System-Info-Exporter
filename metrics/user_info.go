package metrics

import (
	"bufio"
	"log"
	"os"
	"os/exec"
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
			Help: "Information about system users, including username, home directory, UID, GID, and active status",
		},
		[]string{"username", "home_directory", "uid", "gid", "active"},
	)
)

func init() {
	prometheus.MustRegister(systemUserMetrics)
}

func CollectSystemUserMetrics(debug bool) {
	if debug {
		log.Println("Debug: Starting system user metrics collection")
	}

	users, err := fetchAllUsers(debug) // Pass debug flag
	if err != nil {
		log.Printf("Error fetching user information: %v", err)
		return
	}

	if debug {
		log.Printf("Debug: Fetched %d users from /etc/passwd", len(users))
	}

	for _, user := range users {
		if debug {
			log.Printf("Debug: Processing user: %s (UID: %s, GID: %s, HomeDir: %s)", user.Username, user.Uid, user.Gid, user.HomeDir)
		}

		uid, err := strconv.Atoi(user.Uid)
		if err != nil {
			log.Printf("Error parsing UID for user %s: %v", user.Username, err)
			continue
		}

		active := "0"
		if isUserActive(uid) {
			active = "1"
		}

		systemUserMetrics.WithLabelValues(user.Username, user.HomeDir, user.Uid, user.Gid, active).Set(1)

		if debug {
			log.Printf("Debug: Updated metric for user: %s (Active: %s)", user.Username, active)
		}
	}

	if debug {
		log.Println("Debug: All users have been processed and metrics updated.")
	}
}

func fetchAllUsers(debug bool) ([]*user.User, error) {
	users := []*user.User{}

	// Open /etc/passwd to read all users
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 7 {
			// Filter out system/application users by UID
			uid, err := strconv.Atoi(fields[2])
			if err != nil || uid < 1000 {
				continue
			}

			// Filter users by shell
			shell := fields[6]
			if shell != "/bin/bash" && shell != "/bin/sh" {
				continue
			}

			// Use os/user package to fetch user details
			usr, err := user.Lookup(fields[0])
			if err != nil {
				log.Printf("Error looking up user %s: %v", fields[0], err)
				continue
			}

			users = append(users, &user.User{
				Username: usr.Username,
				Uid:      usr.Uid,
				Gid:      fields[3], // Extract GID directly from /etc/passwd
				HomeDir:  usr.HomeDir,
			})

			if debug {
				log.Printf("Debug: Fetched user: %s (UID: %s, GID: %s, HomeDir: %s, Shell: %s)", usr.Username, usr.Uid, fields[3], usr.HomeDir, shell)
			}
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

	// Check active processes in /proc
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err == nil {
			stat := &syscall.Stat_t{}
			if err := syscall.Stat(procDir+"/"+entry.Name(), stat); err == nil && int(stat.Uid) == uid {
				return true
			}
		}
	}

	// Fallback: Check active users using the `w` command
	output, err := exec.Command("w", "-h").Output()
	if err != nil {
		log.Printf("Error executing 'w' command: %v", err)
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] != "" {
			activeUser := fields[0]
			userInfo, err := user.Lookup(activeUser)
			if err == nil && strconv.Itoa(uid) == userInfo.Uid {
				return true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading 'w' command output: %v", err)
	}
	return false
}
