package metrics

import (
	"fmt"
	"time"

	//go get github.com/shirou/gopsutil/v4/cpu
	"github.com/shirou/gopsutil/v4/cpu"

	//go get github.com/shirou/gopsutil/v4/mem
	"github.com/shirou/gopsutil/v4/mem"

	//go get github.com/shirou/gopsutil/v4/disk
	"github.com/shirou/gopsutil/v4/disk"
)

func GetCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU usage data returned")
	}

	return percentages[0], nil
}

func GetMemoryUsage() (float64, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory usage: %w", err)
	}

	return memory.UsedPercent, nil
}

func GetDiskUsage() (float64, error) {
	usage, err := disk.Usage("/")
	if err != nil {
		return 0, fmt.Errorf("failed to get disk usage: %w", err)
	}
	return usage.UsedPercent, nil
}
