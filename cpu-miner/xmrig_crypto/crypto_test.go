package xmrig_crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelf(t *testing.T) {
	require := require.New(t)

	err := SelfTest()
	require.Nil(err)
}
