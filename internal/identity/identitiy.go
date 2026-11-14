package identity

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	//go get github.com/google/uuid
	"github.com/google/uuid"
)

type NodeIdentity struct {
	NodeID string `json:"node_id"`
}

func LoadOrCreate() (NodeIdentity, error) {
	// Get the user's configuration directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return NodeIdentity{}, err
	}

	// Define where the identity will be stored
	identityDir := filepath.Join(configDir, "opencompute")
	identityPath := filepath.Join(identityDir, "identity.json")

	// Attempt to read an existing identity
	data, err := os.ReadFile(identityPath)

	if err == nil {
		var existing NodeIdentity

		// Convert JSON into our Go struct
		if err := json.Unmarshal(data, &existing); err != nil {
			return NodeIdentity{}, err
		}

		// Validate the existing UUID
		if _, err := uuid.Parse(existing.NodeID); err != nil {
			return NodeIdentity{}, err
		}

		return existing, nil
	}

	// Only generate a new identity if the file doesn't exist
	if !errors.Is(err, os.ErrNotExist) {
		return NodeIdentity{}, err
	}

	// Create the configuration directory if necessary
	if err := os.MkdirAll(identityDir, 0700); err != nil {
		return NodeIdentity{}, err
	}

	// Generate a new UUID
	newIdentity := NodeIdentity{
		NodeID: uuid.NewString(),
	}

	// Convert the identity into formatted JSON
	data, err = json.MarshalIndent(newIdentity, "", "  ")
	if err != nil {
		return NodeIdentity{}, err
	}

	// Save the identity to disk
	if err := os.WriteFile(identityPath, data, 0600); err != nil {
		return NodeIdentity{}, err
	}

	return newIdentity, nil
}
