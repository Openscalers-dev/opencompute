# OpenCompute
Open-source compute infrastructure for managing containerized workloads across distributed hardware

The agent runs on Linux, Windows, and macOS. It discovers host hardware, registers with the OpenScalers control plane, and reports CPU, memory, and disk utilization.

## Run the agent on macOS

The macOS builds require [macOS 12 Monterey or newer](https://go.dev/doc/go1.26#darwin). Build from source with the Go 1.26 toolchain specified in `go.mod`. GitHub Releases publishes Linux binaries only. The workflow builds macOS artifacts for validation while Developer ID signing and notarization are pending.

### Install and run

1. In the OpenCompute source checkout, build and install the agent.

   ```sh
   mkdir -p "$HOME/.local/bin"
   CGO_ENABLED=0 go build -o "$HOME/.local/bin/opencompute-agent" .
   ```

2. Run `"$HOME/.local/bin/opencompute-agent"` in a terminal. The agent needs write access to your user configuration directory, outbound HTTPS access to `api.openscalers.dev`, and an available loopback port 9101. A standard macOS account is sufficient.
3. Check the terminal for successful registration and heartbeat messages. To inspect local metrics, run `curl http://127.0.0.1:9101/metrics` from another terminal. Discovery and disk utilization use the writable startup Data volume at `/System/Volumes/Data`.

Press **Ctrl+C** to stop a foreground agent, or send **SIGTERM** to its process. The agent cancels in-flight control-plane requests and shuts down the metrics listener.

### Update or remove

To update, stop the agent, check out the desired source version, run the same build command, and start the agent again. The same node ID is loaded from `~/Library/Application Support/opencompute/identity.json` on startup.

To remove the agent, stop it and delete `~/.local/bin/opencompute-agent`. Delete the identity file as well if you want a future installation to register with a new node ID.

### Validation for macOS onboarding

Both architectures passed the Go test suite and executable smoke tests on macOS 26.5.2 with Go 1.26.8.

| Build | Runtime used |
| --- | --- |
| `darwin/arm64` | Native Apple Silicon |
| `darwin/amd64` | Rosetta on Apple Silicon |

The tests exercised registration, repeated heartbeats, recovery after dropped connections, and re-registration against a local test server. The executable smoke tests checked live Prometheus metrics, node ID reuse across four launches, and shutdown on SIGINT and SIGTERM. Linux and Windows agent and test binaries also cross-compiled for arm64 and amd64.

What's Next
- Tyler: add credentials for nodes authorization, onboard PiOS
- Ryan: Set up one click download and configuration
