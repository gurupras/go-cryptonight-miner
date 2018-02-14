package stratum

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHex2Bin(t *testing.T) {
	require := require.New(t)

	blob := "01019ce68cd4053fbee525ef7689f4a76a7d6c3257b2f59ec578bfe3ef61d14db996adcd78d76f000000004a12bb9b9980cd600e6b531396bc2270f9bbe6d1f295a7188b8c744395cd10b511"
	blobLen := len(blob)
	b, err := HexToBin(blob, blobLen/2)
	// fmt.Println(b)
	require.Nil(err)
	hexStr, err := BinToHex(b)
	// fmt.Println(hexStr)
	require.Nil(err)
	require.Equal(blob, hexStr)
}
