package hardware

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func DiscoverMemory() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)

			if len(fields) < 3 || fields[2] != "kB" {
				return 0, fmt.Errorf("invalid MemTotal format")
			}

			memoryKB, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, err
			}

			return memoryKB * 1024, nil
		}
	}

	return 0, fmt.Errorf("MemTotal not found")
}

// SystemDiskPath returns the mount point used for disk discovery and monitoring.
func SystemDiskPath() string {
	return "/"
}

func DiscoverDisk() (uint64, uint64, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(SystemDiskPath(), &stat)
	if err != nil {
		return 0, 0, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)

	return total, available, nil
}
