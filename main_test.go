package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Openscalers-dev/opencompute/internal/client"
	"github.com/Openscalers-dev/opencompute/internal/hardware"
	"github.com/Openscalers-dev/opencompute/internal/monitoring"
	"github.com/prometheus/client_golang/prometheus"
)

func TestReportRecoversAndKeepsSendingHeartbeats(t *testing.T) {
	node, err := hardware.Discover()
	if err != nil {
		t.Fatal(err)
	}
	registration := client.RegistrationRequest{
		NodeID:         "efdf75cd-4cef-418a-8274-ed3ef3bdc1e2",
		Hostname:       node.Hostname,
		OS:             node.OS,
		Architecture:   node.Architecture,
		CPUs:           node.CPUs,
		MemoryBytes:    node.MemoryBytes,
		DiskTotalBytes: node.DiskTotal,
	}
	const (
		registerPath   = "/v1/nodes/register"
		heartbeatPath  = "/v1/nodes/heartbeat"
		dropConnection = 0
	)
	responses := []struct {
		path   string
		status int
	}{
		{registerPath, dropConnection},
		{registerPath, http.StatusCreated},
		{heartbeatPath, dropConnection},
		{heartbeatPath, http.StatusOK},
		{heartbeatPath, http.StatusOK},
		{heartbeatPath, http.StatusNotFound},
		{registerPath, http.StatusCreated},
		{heartbeatPath, http.StatusOK},
		{heartbeatPath, http.StatusOK},
	}

	var received atomic.Int32
	lastRequest := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(received.Add(1)) - 1
		if i >= len(responses) {
			t.Errorf("unexpected extra request: %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		expected := responses[i]
		if r.Method != http.MethodPost || r.URL.Path != expected.path || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("request %d = %s %s (%s), want POST %s (application/json)", i+1, r.Method, r.URL.Path, r.Header.Get("Content-Type"), expected.path)
		}

		switch expected.path {
		case registerPath:
			var got client.RegistrationRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Errorf("decoding registration: %v", err)
			}
			if got != registration || got.OS != runtime.GOOS || got.Architecture != runtime.GOARCH {
				t.Errorf("registration = %+v, want %+v for %s/%s", got, registration, runtime.GOOS, runtime.GOARCH)
			}
		case heartbeatPath:
			var got client.HeartbeatRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Errorf("decoding heartbeat: %v", err)
			}
			if got.NodeID != registration.NodeID {
				t.Errorf("heartbeat node ID = %q, want %q", got.NodeID, registration.NodeID)
			}
			for _, usage := range []float64{got.CPUUsage, got.MemoryUsage, got.DiskUsage} {
				if !(usage >= 0 && usage <= 100) {
					t.Errorf("invalid heartbeat utilization percentage: %v", usage)
				}
			}
		}

		if expected.status == dropConnection {
			connection, _, err := http.NewResponseController(w).Hijack()
			if err != nil {
				t.Errorf("dropping connection: %v", err)
				return
			}
			connection.Close()
			return
		}
		w.WriteHeader(expected.status)
		if i == len(responses)-1 {
			close(lastRequest)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		report(ctx, client.NewClient(server.URL), registration, 10*time.Millisecond, 10*time.Millisecond)
	}()
	defer func() {
		cancel()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			t.Error("agent reporting did not stop promptly after cancellation")
		}
	}()

	select {
	case <-lastRequest:
	case <-ctx.Done():
		t.Fatalf("received %d of %d expected requests: %v", received.Load(), len(responses), ctx.Err())
	}
}

func TestReportCancelsInFlightRegistration(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var registration client.RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
			t.Errorf("decoding registration: %v", err)
		}
		close(requestStarted)
		select {
		case <-r.Context().Done():
			close(requestCanceled)
		case <-time.After(5 * time.Second):
			t.Error("registration request was not canceled")
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		report(ctx, client.NewClient(server.URL), client.RegistrationRequest{}, time.Minute, time.Minute)
	}()

	select {
	case <-requestStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("registration request did not start")
	}
	cancel()

	select {
	case <-requestCanceled:
	case <-time.After(3 * time.Second):
		t.Error("registration request continued after cancellation")
	}
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Error("agent reporting continued after cancellation")
	}
}

func TestReportKeepsSamplingDuringControlPlaneRequests(t *testing.T) {
	const registerPath = "/v1/nodes/register"
	const heartbeatPath = "/v1/nodes/heartbeat"

	started := make(chan string, 2)
	releaseRegistration := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			t.Errorf("reading request body: %v", err)
			return
		}
		select {
		case started <- r.URL.Path:
		case <-r.Context().Done():
			return
		}
		if r.URL.Path == registerPath {
			select {
			case <-releaseRegistration:
				w.WriteHeader(http.StatusCreated)
			case <-r.Context().Done():
			}
		} else {
			<-r.Context().Done()
		}
	}))
	defer server.Close()

	registry := prometheus.NewRegistry()
	monitoring.RegisterMetrics(registry)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		report(ctx, client.NewClient(server.URL), client.RegistrationRequest{}, 10*time.Millisecond, time.Minute)
	}()
	defer func() {
		cancel()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			t.Error("agent reporting did not stop after cancellation")
		}
	}()

	for _, path := range []string{registerPath, heartbeatPath} {
		select {
		case got := <-started:
			if got != path {
				t.Fatalf("started request %s, want %s", got, path)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("request %s did not start", path)
		}

		// A valid reading after this point must come from a new sample.
		monitoring.UpdateMetrics(-1, -1, -1)
		waitForMetricsRefresh(t, registry, path)
		if path == registerPath {
			close(releaseRegistration)
		}
	}
}

func waitForMetricsRefresh(t *testing.T, registry *prometheus.Registry, blockedPath string) {
	t.Helper()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case <-ticker.C:
			families, err := registry.Gather()
			if err != nil {
				t.Fatal(err)
			}
			valid := 0
			for _, family := range families {
				for _, measurement := range family.GetMetric() {
					if gauge := measurement.GetGauge(); gauge != nil {
						if value := gauge.GetValue(); value >= 0 && value <= 100 {
							valid++
						}
					}
				}
			}
			if valid == 3 {
				return
			}
		case <-deadline.C:
			t.Fatalf("local metrics did not refresh while %s was blocked", blockedPath)
		}
	}
}
