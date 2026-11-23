package hardware

import (
	"runtime"
	"testing"
)

func TestDiscover(t *testing.T) {
	node, err := Discover()

	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if node.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	if node.CPUs <= 0 {
		t.Error("CPU count should be greater than zero")
	}

	if node.MemoryBytes == 0 {
		t.Error("Expected total memory to be greater than zero")
	}
	if node.OS != runtime.GOOS || node.Architecture != runtime.GOARCH {
		t.Errorf("discovered platform = %s/%s, want %s/%s", node.OS, node.Architecture, runtime.GOOS, runtime.GOARCH)
	}
	if node.DiskTotal == 0 || node.DiskAvailable > node.DiskTotal {
		t.Errorf("invalid disk capacity: %d available of %d total bytes", node.DiskAvailable, node.DiskTotal)
	}
}
