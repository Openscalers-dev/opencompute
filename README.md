# OpenCompute
Open-source compute infrastructure for managing containerized workloads across distributed hardware

Currently, the agent runs on Linux and automatically discovers system hardware, including CPU count, memory, disk capacity, operating system, and architecture. It also continuously monitors CPU, memory, and disk utilization using gopsutil, collecting updated metrics every five seconds with support for graceful shutdown.

What's Next
- The next step is integrating Prometheus to expose system metrics through an HTTP endpoint. Eventually, OpenCompute will connect to the OpenScalers control plane, allowing users to register their machines, monitor resources remotely, and manage distributed compute workloads.
- Implement this same creation, but for PiOS.
