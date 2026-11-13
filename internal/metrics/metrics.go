package metrics

import "os"

func ReadCPUStats() ([]byte, error) {
    return os.ReadFile("/proc/stat")
}