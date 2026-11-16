# OpenCompute
Open-source compute infrastructure for managing containerized workloads across distributed hardware

Currently, the agent runs on Linux and automatically discovers system hardware, including CPU count, memory, disk capacity, operating system, and architecture. It also continuously monitors CPU, memory, and disk utilization using gopsutil, collecting updated metrics every five seconds with support for graceful shutdown.

What's Next
- Tyler: add credentials for nodes authorization, onboard PiOS
- Ryan: Set up one click download and configuration
