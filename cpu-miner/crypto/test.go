package crypto

/*
#cgo CFLAGS: -I /c:\msys64\boost -I /c/msys64/boost -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native
#include "helpers.h"
*/
import "C"

func test() {
	data := "hello"
	ptr := C.CString(data)
	C.simple_fn(ptr, C.int(len(data)))
}
