package nodeinfo

import (
	"os"
	"runtime"
)

type NodeInfo struct {
	Hostname      string
	Distro        string
	DistroVersion string
}

func Read() *NodeInfo {
	hostname, _ := os.Hostname()

	distro := runtime.GOOS // fallback useful for development
	var distroVersion string
	// TODO Detect those values. https://jira.percona.com/browse/PMM-3453

	return &NodeInfo{
		Hostname:      hostname,
		Distro:        distro,
		DistroVersion: distroVersion,
	}
}
