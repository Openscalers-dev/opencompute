package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/client"
	"github.com/Openscalers-dev/opencompute/internal/hardware"
	"github.com/Openscalers-dev/opencompute/internal/identity"
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

	nodeIdentity, err := identity.LoadOrCreate()
	if err != nil {
		fmt.Println("Error loading node identity:", err)
		return
	}
	fmt.Println("Node ID:", nodeIdentity.NodeID)

	node, err := hardware.Discover()
	if err != nil {
		fmt.Println("Error discovering node:", err)
		return
	}

	// Connect to the OpenScalers Control Plane
	apiClient := client.NewClient("https://api.openscalers.dev")

	// Prepare node registration
	registration := client.RegistrationRequest{
		NodeID:         nodeIdentity.NodeID,
		Hostname:       node.Hostname,
		OS:             node.OS,
		Architecture:   node.Architecture,
		CPUs:           node.CPUs,
		MemoryBytes:    node.MemoryBytes,
		DiskTotalBytes: node.DiskTotal,
	}

	// Register the node
	if err := apiClient.Register(registration); err != nil {
		log.Printf("Failed to register node: %v", err)
	} else {
		log.Println("Successfully registered node with OpenScalers!")
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

	// Start HTTP server for Prometheus metrics
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
	lastHeartbeat := time.Time{}

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
			// Send heartbeat every 15 seconds.
			if time.Since(lastHeartbeat) >= 15*time.Second {
				heartbeat := client.HeartbeatRequest{
					NodeID:      nodeIdentity.NodeID,
					CPUUsage:    systemMetrics.CPUUsagePercent,
					MemoryUsage: systemMetrics.MemoryUsagePercent,
					DiskUsage:   systemMetrics.DiskUsagePercent,
				}

				if err := apiClient.Heartbeat(heartbeat); err != nil {
					log.Printf("Heartbeat failed: %v", err)
				} else {
					log.Println("Heartbeat sent successfully!")
					lastHeartbeat = time.Now()
				}
			}

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
