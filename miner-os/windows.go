package mineros

import (
	"fmt"
	"regexp"
	"strconv"

	ps "github.com/gorillalabs/go-powershell"
	"github.com/gorillalabs/go-powershell/backend"
	gocm "github.com/gurupras/go-cryptonight-miner"
)

func winGetPCITopology(deviceInstanceID string) (*gocm.Topology, error) {
	cmdStr := fmt.Sprintf(`gwmi Win32_PnPEntity | where {$_.DeviceID -eq '%v'} | foreach { $_.GetDeviceProperties('DEVPKEY_Device_LocationInfo').deviceProperties.Data }`, deviceInstanceID)

	back := &backend.Local{}
	shell, err := ps.New(back)
	if err != nil {
		return nil, err
	}
	defer shell.Exit()

	stdout, stderr, err := shell.Execute(cmdStr)
	if err != nil {
		fmt.Printf("Err: %v\n", stderr)
		return nil, err
	}
	// fmt.Printf("pci topology=%v\n", stdout)
	regex := regexp.MustCompile(`PCI bus (?P<bus>\d+), device (?P<device>\d+), function (?P<function>\d+)`)
	match := regex.FindStringSubmatch(stdout)
	var (
		bus      int
		device   int
		function int
	)
	if bus, err = strconv.Atoi(match[1]); err != nil {
		return nil, err
	}
	if device, err = strconv.Atoi(match[2]); err != nil {
		return nil, err
	}
	if function, err = strconv.Atoi(match[3]); err != nil {
		return nil, err
	}
	topology := &gocm.Topology{
		bus,
		device,
		function,
	}
	return topology, nil
}
