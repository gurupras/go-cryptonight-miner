package mineros

import (
	"fmt"
	"runtime"

	gocm "github.com/gurupras/go-cryptonight-miner"
)

func GetPCITopology(deviceInstanceID string) (*gocm.Topology, error) {
	if runtime.GOOS == "windows" {
		return winGetPCITopology(deviceInstanceID)
	} else {
		return nil, fmt.Errorf("Unimplemented for OS '%v'", runtime.GOOS)
	}
}
