package monitoring

//go get github.com/prometheus/client_golang/prometheus
//go get github.com/prometheus/client_golang/prometheus/promhttp
import "github.com/prometheus/client_golang/prometheus"

var (
	cpuusage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "opencompute_cpu_usage_percent",
			Help: "Current CPU utilization percentage.",
		},
	)

	memoryusage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "opencompute_memory_usage_percent",
			Help: "Current memory utilization percentage.",
		},
	)

	diskusage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "opencompute_disk_usage_percent",
			Help: "Current disk utilization percentage.",
		},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(cpuusage, memoryusage, diskusage)
}

func UpdateMetrics(cpu, memory, disk float64) {
	cpuusage.Set(cpu)
	memoryusage.Set(memory)
	diskusage.Set(disk)
}
