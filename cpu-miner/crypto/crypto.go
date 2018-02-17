package crypto

/*
#cgo CFLAGS: -I. -Ofast -fuse-linker-plugin -funroll-loops -fvariable-expansion-in-unroller -ftree-loop-if-convert-stores -fmerge-all-constants -fbranch-target-load-optimize2 -fsched2-use-superblocks -falign-loops=16 -falign-functions=16 -falign-jumps=16 -falign-labels=16 -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native -Wl,--stack,10485760
#include "helpers.h"
#include "cryptonight.h"
#include "miner.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
)

const ()

func SetupCryptonightContext() (unsafe.Pointer, error) {
	ctxPtr := C.setup_persistent_ctx()
	if ctxPtr == nil {
		return nil, fmt.Errorf("Failed to setup cryptonight context.")
	} else {
		return ctxPtr, nil
	}
}

func ScanHashCryptonight(id uint32, work *stratum.Work, maxNonce uint32, hashesDonePtr unsafe.Pointer, ctx unsafe.Pointer, restart unsafe.Pointer) bool {
	workPtr := unsafe.Pointer(&work.Data[0])
	targetPtr := unsafe.Pointer(&work.Target[0])
	cId := C.int(int(id))
	cMaxNonce := C.ulong(maxNonce)
	// log.Debugf("Calling C.scanhash_cryptonight_wrapper()")
	return C.scanhash_cryptonight_wrapper(cId, workPtr, targetPtr, cMaxNonce, hashesDonePtr, ctx, restart) != 0
}

func CryptonightHash(b []byte, length int) []byte {
	workPtr := unsafe.Pointer(&b[0])
	hash := make([]byte, 32)
	hashPtr := unsafe.Pointer(&hash[0])
	C.cryptonight_hash_wrapper(hashPtr, workPtr, C.int(length))
	return hash
}
