package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client communicates with the OpenScalers Control Plane.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a client for the control plane.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegistrationRequest contains the information sent to the control plane.
type RegistrationRequest struct {
	NodeID         string `json:"node_id"`
	Hostname       string `json:"hostname"`
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	CPUs           int    `json:"cpus"`
	MemoryBytes    uint64 `json:"memory_bytes"`
	DiskTotalBytes uint64 `json:"disk_total_bytes"`
}

// Register sends the node's information to the control plane.
func (c *Client) Register(node RegistrationRequest) error {
	body, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("encoding registration request: %w", err)
	}

	url := c.baseURL + "/v1/nodes/register"

	resp, err := c.httpClient.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("sending registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed: HTTP %d", resp.StatusCode)
	}

	return nil
}


// HeartbeatRequest contains the node's latest utilization metrics.
type HeartbeatRequest struct {
	NodeID      string  `json:"node_id"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

// Heartbeat sends the node's latest metrics to the control plane.
func (c *Client) Heartbeat(heartbeat HeartbeatRequest) error {
	body, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("encoding heartbeat: %w", err)
	}

	url := c.baseURL + "/v1/nodes/heartbeat"

	resp, err := c.httpClient.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("sending heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: HTTP %d", resp.StatusCode)
	}

	return nil
}