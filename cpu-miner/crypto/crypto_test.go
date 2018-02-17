package crypto

import (
	"encoding/binary"
	"testing"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSetupCryptonightContext(t *testing.T) {
	require := require.New(t)

	ctx, err := SetupCryptonightContext()
	require.Nil(err)
	require.NotNil(ctx)
}

func TestScanHashCryptonight(t *testing.T) {
	require := require.New(t)

	ctx, err := SetupCryptonightContext()
	require.Nil(err)
	require.NotNil(ctx)

	response, err := stratum.ParseResponse([]byte(stratum.TEST_JOB_STR))
	require.Nil(err)

	work, err := stratum.ParseWorkFromResponse(response)
	require.Nil(err)

	maxNonce := uint32(61)
	hashesDone := 0
	hashesDonePtr := unsafe.Pointer(&hashesDone)
	restart := 0
	restartPtr := unsafe.Pointer(&restart)
	foundNonce := false
	// FIXME: Fix this test. nonce should be changing?
	for foundNonce != true {
		foundNonce = ScanHashCryptonight(1, work, maxNonce, hashesDonePtr, ctx, restartPtr)
	}
}

func TestSolver(t *testing.T) {
	require := require.New(t)

	data := []byte("0505efcfdccb0506180897d587b02f9c97037e66ea638990b2b3a0efab7bab0bff4e3f3dfe1c7d00000000a6788e66eb9b82325f95fc7a2007d3fed7152a3590366cc2a9577dcadf3544a804")
	targetHex := []byte("e4a63d00")
	result := []byte("960A7A3A1826B0AA70E8043FFE7B9E23EE2E028BBA75F3D7557CCDFF9C7F1A00")
	_ = result

	b, err := stratum.HexToBin(string(targetHex), 4)
	target := binary.LittleEndian.Uint32(b)
	ctx, err := SetupCryptonightContext()
	require.Nil(err)
	require.NotNil(ctx)

	maxNonce := uint32(0xffffffff)/uint32(1*uint32(1+1)) - 0x20
	hashesDone := 0
	hashesDonePtr := unsafe.Pointer(&hashesDone)
	restart := 0
	restartPtr := unsafe.Pointer(&restart)
	foundNonce := false

	work := stratum.NewWork()
	copy(work.Data, data)
	work.Target[7] = target
	for foundNonce != true {
		foundNonce = ScanHashCryptonight(1, work, maxNonce, hashesDonePtr, ctx, restartPtr)
	}

	nonceStr, err := stratum.BinToHex(work.Data[39:43])
	require.Nil(err)
	hash := CryptonightHash(work.Data, len(work.Data))
	hashHex, err := stratum.BinToHex(hash)
	require.Nil(err)
	log.Infof("nonceStr: %v", nonceStr)
	log.Infof("hashHex: %v", hashHex)
}
