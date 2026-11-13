package main

import (
	"fmt"
	"github.com/Openscalers-dev/opencompute/internal/hardware"
)


func main() {
	fmt.Println("Opencompute Agent Starting... \n Version 0.1.0 \n Status: Initializing \n")

	node, err := hardware.Discover()

	if err != nil {
		fmt.Println("Error discovering node:", err)
		return
	}

	fmt.Println("Node:", node.Hostname)
	fmt.Println("Operating System:", node.OS)
	fmt.Println("Architecture:", node.Architecture)
	fmt.Println("CPU Count:", node.CPUs)
}