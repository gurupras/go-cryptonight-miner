package stratum

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestParseWork(t *testing.T) {
	// t.Skip("Usually skip since this is an infinite loop test")
	require := require.New(t)

	response, err := ParseResponse([]byte(TEST_JOB_STR))
	require.Nil(err)

	work, err := ParseWorkFromResponse(response)
	require.Nil(err)

	_ = work
}

func TestNoncePtr(t *testing.T) {
	require := require.New(t)

	response, err := ParseResponse([]byte(TEST_JOB_STR))
	require.Nil(err)

	work, err := ParseWorkFromResponse(response)
	require.Nil(err)

	noncePtr := (*uint32)(unsafe.Pointer(&work.Data[39]))
	require.Zero(int(*noncePtr))

	*noncePtr += 10
	require.Equal(10, int(*noncePtr))
}
