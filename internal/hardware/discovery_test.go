package hardware

import "testing"

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
}
