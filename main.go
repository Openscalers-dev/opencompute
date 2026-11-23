package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/client"
	"github.com/Openscalers-dev/opencompute/internal/hardware"
	"github.com/Openscalers-dev/opencompute/internal/identity"
	"github.com/Openscalers-dev/opencompute/internal/metrics"
	"github.com/Openscalers-dev/opencompute/internal/monitoring"

	"github.com/prometheus/client_golang/prometheus"
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

	monitoring.RegisterMetrics(prometheus.DefaultRegisterer)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    "127.0.0.1:9101",
		Handler: mux,
	}
	go func() {
		fmt.Println("Prometheus metrics available at http://127.0.0.1:9101/metrics")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server error:", err)
		}
	}()

	report(ctx, apiClient, registration, 5*time.Second, 15*time.Second)

	fmt.Println("OpenCompute Agent Shutting Down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutting down metrics server: %v", err)
	}
}

func report(ctx context.Context, apiClient *client.Client, registration client.RegistrationRequest, sampleInterval, heartbeatInterval time.Duration) {
	samples := make(chan *metrics.SystemMetrics, 1)
	var sender sync.WaitGroup
	sender.Go(func() {
		reportToControlPlane(ctx, apiClient, registration, samples, sampleInterval, heartbeatInterval)
	})
	defer sender.Wait()

	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
		}

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

		// Keep only the latest pending sample while a request is in flight.
		select {
		case <-samples:
		default:
		}
		samples <- systemMetrics

		fmt.Printf(
			"CPU: %.2f%% | RAM: %.2f%% | Disk: %.2f%%\n",
			systemMetrics.CPUUsagePercent,
			systemMetrics.MemoryUsagePercent,
			systemMetrics.DiskUsagePercent,
		)
	}
}

func reportToControlPlane(ctx context.Context, apiClient *client.Client, registration client.RegistrationRequest, samples <-chan *metrics.SystemMetrics, retryInterval, heartbeatInterval time.Duration) {
	registered := registerNode(ctx, apiClient, registration)
	retry := time.NewTicker(retryInterval)
	defer retry.Stop()

	var lastHeartbeat time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-retry.C:
			if !registered {
				registered = registerNode(ctx, apiClient, registration)
			}
		case systemMetrics := <-samples:
			if !registered || time.Since(lastHeartbeat) < heartbeatInterval {
				continue
			}
			heartbeat := client.HeartbeatRequest{
				NodeID:      registration.NodeID,
				CPUUsage:    systemMetrics.CPUUsagePercent,
				MemoryUsage: systemMetrics.MemoryUsagePercent,
				DiskUsage:   systemMetrics.DiskUsagePercent,
			}
			if err := apiClient.Heartbeat(ctx, heartbeat); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Heartbeat failed: %v", err)
				if errors.Is(err, client.ErrNodeNotRegistered) {
					registered = false
				}
				continue
			}
			log.Println("Heartbeat sent successfully!")
			lastHeartbeat = time.Now()
		}
	}
}

func registerNode(ctx context.Context, apiClient *client.Client, registration client.RegistrationRequest) bool {
	if err := apiClient.Register(ctx, registration); err != nil {
		if ctx.Err() == nil {
			log.Printf("Failed to register node: %v", err)
		}
		return false
	}
	log.Println("Successfully registered node with OpenScalers!")
	return true
}
