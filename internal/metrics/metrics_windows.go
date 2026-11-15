package metrics

import (
	"fmt"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/hardware"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("getting CPU usage: %w", err)
	}

	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU usage data returned")
	}

	return percentages[0], nil
}

func GetMemoryUsage() (float64, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("getting memory usage: %w", err)
	}

	return memory.UsedPercent, nil
}

func GetDiskUsage() (float64, error) {
	path := hardware.SystemDiskPath()

	usage, err := disk.Usage(path)
	if err != nil {
		return 0, fmt.Errorf(
			"getting disk usage at %q: %w", path, err,
		)
	}

	return usage.UsedPercent, nil
}
