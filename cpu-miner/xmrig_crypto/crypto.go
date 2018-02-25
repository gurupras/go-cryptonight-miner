package xmrig_crypto

/*
#cgo CFLAGS: -I. -I../cryptolib -I${SRCDIR}/cpu-miner/cryptolib -I${SRCDIR}/cryptolib -L${SRCDIR}/cpu-miner/cryptolib -L${SRCDIR}/cryptolib -lcryptonight -Ofast -fuse-linker-plugin -funroll-loops -fvariable-expansion-in-unroller -ftree-loop-if-convert-stores -fmerge-all-constants -fbranch-target-load-optimize2 -fsched2-use-superblocks -falign-loops=16 -falign-functions=16 -falign-jumps=16 -falign-labels=16 -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native -Wl,--stack,10485760
#cgo LDFLAGS: -L. -L../cryptolib -L${SRCDIR}/cpu-miner/cryptolib -L${SRCDIR}/cryptolib -lcryptonight
#include "helpers.h"
#include "cryptonight.h"
#include "hash.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
)

func SetupCryptonightContext() (unsafe.Pointer, error) {
	ctxPtr := C.xmrig_setup_persistent_ctx()
	if ctxPtr == nil {
		return nil, fmt.Errorf("Failed to setup cryptonight context.")
	} else {
		return ctxPtr, nil
	}
}

func CryptonightHash(work *stratum.Work, ctx unsafe.Pointer) ([]byte, bool) {
	inputPtr := unsafe.Pointer(&work.Data[0])
	workLength := C.int(len(work.Data))
	hashBytes := make([]byte, 32)
	hashPtr := unsafe.Pointer(&hashBytes[0])
	targetPtr := unsafe.Pointer(&work.Target[0])

	found := C.xmrig_cryptonight_hash_wrapper(inputPtr, workLength, hashPtr, targetPtr, ctx)
	return hashBytes, found == 1
}

func SelfTest() error {
	ret := C.xmrig_self_test()
	if ret != 0 {
		return fmt.Errorf("Failed self test")
	} else {
		return nil
	}
}
