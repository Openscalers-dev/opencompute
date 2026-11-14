package hardware

import (
	"os"
	"runtime"
)

type Node struct {
	Hostname      string
	OS            string
	Architecture  string
	CPUs          int
	MemoryBytes   uint64
	DiskTotal     uint64
	DiskAvailable uint64
}

func Discover() (Node, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return Node{}, err
	}

	//Discover Memory
	memoryBytes, err := DiscoverMemory()
	if err != nil {
		return Node{}, err
	}

	//Discover Disk
	diskTotal, diskAvailable, err := DiscoverDisk()
	if err != nil {
		return Node{}, err
	}

	return Node{
		Hostname:      hostname,
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
		CPUs:          runtime.NumCPU(),
		MemoryBytes:   memoryBytes,
		DiskTotal:     diskTotal,
		DiskAvailable: diskAvailable,
	}, nil

}
