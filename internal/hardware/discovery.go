package hardware

import (
	"os"
	"runtime"
)

type Node struct {
	Hostname	 string
	OS           string
	Architecture string
	CPUs         int
}

func Discover() (Node, error){
	hostname, err := os.Hostname()
	if err != nil {
		return Node{}, err
	}



	return Node{
		Hostname: hostname,
		OS: runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUs: runtime.NumCPU(),
	}, nil

}