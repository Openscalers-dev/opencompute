package metrics

import "testing"

func TestGetSystemMetrics(t *testing.T) {
	metrics, err := GetSystemMetrics()
	if err != nil {
		t.Fatalf("GetSystemMetrics() failed: %v", err)
	}
	if metrics == nil {
		t.Fatal("GetSystemMetrics() returned nil metrics")
	}

	for _, measurement := range []struct {
		name  string
		value float64
	}{
		{"CPU", metrics.CPUUsagePercent},
		{"memory", metrics.MemoryUsagePercent},
		{"disk", metrics.DiskUsagePercent},
	} {
		if !(measurement.value >= 0 && measurement.value <= 100) {
			t.Errorf("%s usage = %v, want a percentage in [0, 100]", measurement.name, measurement.value)
		}
	}
}
