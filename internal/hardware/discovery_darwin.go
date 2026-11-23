package hardware

import (
	"fmt"

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

// SystemDiskPath returns the writable startup volume on macOS.
func SystemDiskPath() string {
	return "/System/Volumes/Data"
}

func DiscoverDisk() (uint64, uint64, error) {
	path := SystemDiskPath()
	usage, err := disk.Usage(path)
	if err != nil {
		return 0, 0, fmt.Errorf("discovering disk at %q: %w", path, err)
	}
	return usage.Total, usage.Free, nil
}
