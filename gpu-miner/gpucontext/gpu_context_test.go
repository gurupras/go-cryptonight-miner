package gpucontext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsCStruct(t *testing.T) {
	require := require.New(t)

	ctxs := getAMDDevices(0)
	require.NotZero(len(ctxs))

	ctx := ctxs[0]
	err := testCContext(ctx)
	require.Nil(err)
}
