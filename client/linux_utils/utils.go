package linux_utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseLinuxMemory(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.Atoi(fields[1])
				if err == nil {
					gb := float64(kb) / 1024 / 1024
					return fmt.Sprintf("%.1f", gb)
				}
			}
		}
	}
	return "Unknown"
}
