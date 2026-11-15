package hardware

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func DiscoverMemory() (uint64, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("discovering memory: %w", err)
	}

	return memory.Total, nil
}

// DiscoverDisk Helper Function
func SystemDiskPath() string {
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}

	return drive + `\`
}

func DiscoverDisk() (uint64, uint64, error) {
	path := SystemDiskPath()

	usage, err := disk.Usage(path)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"discovering disk at %q: %w", path, err,
		)
	}

	return usage.Total, usage.Free, nil
}
