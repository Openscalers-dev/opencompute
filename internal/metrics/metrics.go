package metrics

type SystemMetrics struct {
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	DiskUsagePercent   float64
}

func GetSystemMetrics() (*SystemMetrics, error) {
	cpuUsage, err := GetCPUUsage()
	if err != nil {
		return nil, err
	}

	memoryUsage, err := GetMemoryUsage()
	if err != nil {
		return nil, err
	}

	diskUsage, err := GetDiskUsage()
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		CPUUsagePercent:    cpuUsage,
		MemoryUsagePercent: memoryUsage,
		DiskUsagePercent:   diskUsage,
	}, nil
}
