package windows_utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseWindowsSystemInfo(output string) string {
	var osName, osVersion string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OS Name:") {
			osName = strings.TrimSpace(strings.TrimPrefix(line, "OS Name:"))
		}
		if strings.HasPrefix(line, "OS Version:") {
			osVersion = strings.TrimSpace(strings.TrimPrefix(line, "OS Version:"))
		}
	}
	if osName != "" && osVersion != "" {
		return fmt.Sprintf("%s (%s)", osName, osVersion)
	}
	return "Unknown OS Version"
}

func ParseWindowsMemory(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
			value := strings.TrimPrefix(line, "TotalVisibleMemorySize=")
			kb, err := strconv.Atoi(value)
			if err == nil {
				gb := float64(kb) / 1024 / 1024
				return fmt.Sprintf("%.1f", gb)
			}
		}
	}
	return "Unknown"
}
