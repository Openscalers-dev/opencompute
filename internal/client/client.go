package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
func (c *Client) Register(ctx context.Context, node RegistrationRequest) error {
	body, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("encoding registration request: %w", err)
	}

	status, err := c.post(ctx, "/v1/nodes/register", body)
	if err != nil {
		return fmt.Errorf("sending registration request: %w", err)
	}
	if status != http.StatusCreated {
		return fmt.Errorf("registration failed: HTTP %d", status)
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

// ErrNodeNotRegistered signals that the control plane no longer knows this node.
var ErrNodeNotRegistered = errors.New("node is not registered")

// Heartbeat sends the node's latest metrics to the control plane.
func (c *Client) Heartbeat(ctx context.Context, heartbeat HeartbeatRequest) error {
	body, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("encoding heartbeat: %w", err)
	}

	status, err := c.post(ctx, "/v1/nodes/heartbeat", body)
	if err != nil {
		return fmt.Errorf("sending heartbeat: %w", err)
	}
	if status == http.StatusNotFound {
		return fmt.Errorf("heartbeat failed: HTTP %d: %w", status, ErrNodeNotRegistered)
	}
	if status != http.StatusOK {
		return fmt.Errorf("heartbeat failed: HTTP %d", status)
	}

	return nil
}

func (c *Client) post(ctx context.Context, path string, body []byte) (int, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	return response.StatusCode, nil
}
