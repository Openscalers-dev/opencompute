package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/hardware"
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

	//Heartbeat ticket
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("OpenCompute heartbeat:", time.Now().Format("15:04:05"))
		case <-ctx.Done():
			fmt.Println("OpenCompute Agent Shutting Down...")
			return
		}
	}
}
