package mineros

import (
	"fmt"
	"runtime"

	"github.com/gurupras/minerconfig/pcie"
)

func GetPCITopology(deviceInstanceID string) (*pcie.Topology, error) {
	if runtime.GOOS == "windows" {
		return winGetPCITopology(deviceInstanceID)
	} else {
		return nil, fmt.Errorf("Unimplemented for OS '%v'", runtime.GOOS)
	}
}
