package amdgpu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNumPlatforms(t *testing.T) {
	require := require.New(t)

	numPlatforms := getNumPlatforms()
	require.NotZero(numPlatforms)
}

func TestGetAMDDevices(t *testing.T) {
	require := require.New(t)

	numPlatforms := getNumPlatforms()
	//log.Infof("numPlatforms=%v", numPlatforms)

	fail := true
	for i := int(numPlatforms) - 1; i >= 0; i-- {
		devices := getAMDDevices(i)
		if devices != nil && len(devices) > 0 {
			fail = false
		}
	}
	require.False(fail, "Did not find any AMD GPUs")
}

func TestCLog(t *testing.T) {
	// This test should not crash. We don't really check if it logged, but
	// as long as it doesn't crash, everything is OK
	logFromC("Hello World")
}

func TestPrintPlatforms(t *testing.T) {
	t.Skip("Too primitive")
	printPlatforms()
}
