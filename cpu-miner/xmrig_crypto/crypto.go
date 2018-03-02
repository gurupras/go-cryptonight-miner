package xmrig_crypto

/*
#cgo CFLAGS: -I. -Ofast -fuse-linker-plugin -funroll-loops -fvariable-expansion-in-unroller -ftree-loop-if-convert-stores -fmerge-all-constants -fbranch-target-load-optimize2 -fsched2-use-superblocks -falign-loops=16 -falign-functions=16 -falign-jumps=16 -falign-labels=16 -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native -Wl,--stack,10485760
#cgo LDFLAGS:
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

type XMRigWork struct {
	*stratum.Work
	Cdata *XMRigCData
}

func NewXMRigWork() *XMRigWork {
	return &XMRigWork{
		stratum.NewWork(),
		nil,
	}
}

type XMRigCData struct {
	Input        unsafe.Pointer
	Target       unsafe.Pointer
	Size         C.int
	HashBytes    []byte
	HashBytesPtr unsafe.Pointer
}

func (work *XMRigWork) UpdateCData() {
	if work.Cdata == nil {
		work.Cdata = &XMRigCData{
			unsafe.Pointer(&work.Data[0]),
			nil,
			C.int(work.Size),
			make([]byte, 32),
			nil,
		}
		work.Cdata.HashBytesPtr = unsafe.Pointer(&work.Cdata.HashBytes[0])
	} else {
		work.Cdata.Size = C.int(work.Size)
	}
}

func SetupHugePages(totalMiners uint32) (unsafe.Pointer, error) {
	totalMinersCint := C.int(int(totalMiners))
	ptr := C.xmrig_setup_hugepages(totalMinersCint)
	if ptr != nil {
		return ptr, nil
	} else {
		return nil, fmt.Errorf("Failed to set up hugepages")
	}
}

func SetupCryptonightContext(memPtr unsafe.Pointer, threadId uint32) (unsafe.Pointer, error) {
	threadIdCint := C.int(int(threadId))
	ptr := C.xmrig_thread_persistent_ctx(memPtr, threadIdCint)
	if ptr != nil {
		return ptr, nil
	} else {
		return nil, fmt.Errorf("Failed to get cryptonight context for thread: %d", threadId)
	}
}

func SetupSimpleCryptonightContext() (unsafe.Pointer, error) {
	ptr := C.xmrig_simple_cryptonight_context()
	if ptr != nil {
		return ptr, nil
	} else {
		return nil, fmt.Errorf("malloc cryptonight_ctx failed")
	}
}

func CryptonightHash(work *XMRigWork, ctx unsafe.Pointer) ([]byte, bool) {
	target := work.Work.Target
	targetPtr := unsafe.Pointer(&target)
	// blobBytes := stratum.BinToStr(work.Data)
	// log.Infof("blob: %v\n", blobBytes)
	// log.Infof("size: %v", work.Size)

	found := C.xmrig_cryptonight_hash_wrapper(work.Cdata.Input, work.Cdata.Size, work.Cdata.HashBytesPtr, targetPtr, ctx)
	return work.Cdata.HashBytes, found == 1
}

func SelfTest() error {
	ret := C.xmrig_self_test()
	if ret != 0 {
		return fmt.Errorf("Failed self test")
	} else {
		return nil
	}
}
