package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/hardware"
	"github.com/Openscalers-dev/opencompute/internal/metrics"
	"github.com/Openscalers-dev/opencompute/internal/monitoring"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create a context that will be canceled on SIGINT or SIGTERM with graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	fmt.Println("Opencompute Agent Starting... \n Version 0.1.0 \n Status: Initializing")

	node, err := hardware.Discover()

	if err != nil {
		fmt.Println("Error discovering node:", err)
		return
	}

	//Print node information
	fmt.Println("Node:", node.Hostname)
	fmt.Println("Operating System:", node.OS)
	fmt.Println("Architecture:", node.Architecture)
	fmt.Println("CPU Count:", node.CPUs)
	fmt.Printf(
		"Total Memory: %.2f GiB\n",
		float64(node.MemoryBytes)/(1024*1024*1024),
	)
	fmt.Printf(
		"Total Disk: %.2f GiB\n",
		float64(node.DiskTotal)/(1024*1024*1024),
	)
	fmt.Printf(
		"Available Disk: %.2f GiB\n",
		float64(node.DiskAvailable)/(1024*1024*1024),
	)

	monitoring.RegisterMetrics()

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		fmt.Println("Prometheus metrics available at http://127.0.0.1:9101/metrics")

		err := http.ListenAndServe("127.0.0.1:9101", mux)
		if err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	//Heartbeat ticket
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			systemMetrics, err := metrics.GetSystemMetrics()
			if err != nil {
				fmt.Println("Error getting system metrics:", err)
				continue
			}

			monitoring.UpdateMetrics(
				systemMetrics.CPUUsagePercent,
				systemMetrics.MemoryUsagePercent,
				systemMetrics.DiskUsagePercent,
			)

			fmt.Printf(
				"CPU: %.2f%% | RAM: %.2f%% | Disk: %.2f%%\n",
				systemMetrics.CPUUsagePercent,
				systemMetrics.MemoryUsagePercent,
				systemMetrics.DiskUsagePercent,
			)
		case <-ctx.Done():
			fmt.Println("OpenCompute Agent Shutting Down...")
			return
		}
	}
}
