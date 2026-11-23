package identity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestLoadOrCreatePreservesIdentity(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("AppData", home)

	first, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("creating identity: %v", err)
	}
	if _, err := uuid.Parse(first.NodeID); err != nil {
		t.Fatalf("invalid persisted node ID: %v", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, "opencompute", "identity.json"))
	if err != nil {
		t.Fatalf("reading persisted identity: %v", err)
	}
	var persisted NodeIdentity
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("decoding persisted identity: %v", err)
	}

	second, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("loading existing identity: %v", err)
	}
	if first != persisted || first != second {
		t.Errorf("identity changed: created %+v, persisted %+v, loaded %+v", first, persisted, second)
	}
}
