package gpucontext

/*
#include "gpu_context.h"
*/
import "C"

import (
	"unsafe"

	"github.com/rainliu/gocl/cl"
)

type GPUContext struct {
	DeviceIndex   int
	RawIntensity  int
	WorkSize      int
	DeviceID      cl.CL_device_id     `cl_device_id`
	CommandQueues cl.CL_command_queue `cl_command_queue`
	InputBuffer   cl.CL_mem           `cl_mem`
	OutputBuffer  cl.CL_mem           `cl_mem`
	ExtraBuffers  [6]cl.CL_mem        `cl_mem`
	Program       cl.CL_program       `cl_program`
	Kernels       [7]cl.CL_kernel     `cl_kernel`
	FreeMemory    cl.CL_ulong
	ComputeUnits  cl.CL_uint
	Name          string
	Nonce         uint32
	cStruct       *C.struct_gpu_context
}

func (ctx *GPUContext) AsCStruct() *C.struct_gpu_context {
	if ctx.cStruct == nil {
		ret := &C.struct_gpu_context{}
		ret.DeviceIndex = C.int(ctx.DeviceIndex)
		ret.RawIntensity = C.int(ctx.RawIntensity)
		ret.WorkSize = C.int(ctx.WorkSize)
		ret.DeviceID = *(*C.cl_device_id)(unsafe.Pointer(&ctx.DeviceID))
		ret.CommandQueues = *(*C.cl_command_queue)(unsafe.Pointer(&ctx.CommandQueues))
		ret.InputBuffer = *(*C.cl_mem)(unsafe.Pointer(&ctx.InputBuffer))
		ret.OutputBuffer = *(*C.cl_mem)(unsafe.Pointer(&ctx.OutputBuffer))
		for i := 0; i < len(ctx.ExtraBuffers); i++ {
			ret.ExtraBuffers[i] = *(*C.cl_mem)(unsafe.Pointer(&ctx.ExtraBuffers[i]))
		}
		ret.Program = *(*C.cl_program)(unsafe.Pointer(&ctx.Program))
		for i := 0; i < len(ctx.Kernels); i++ {
			ret.Kernels[i] = *(*C.cl_kernel)(unsafe.Pointer(&ctx.Kernels[i]))
		}
		ret.FreeMemory = *(*C.cl_ulong)(unsafe.Pointer(&ctx.FreeMemory))
		ret.ComputeUnits = *(*C.cl_uint)(unsafe.Pointer(&ctx.ComputeUnits))
		ret.Name = C.CString(ctx.Name)
		ctx.cStruct = ret
	}
	// Nonce may change
	ctx.cStruct.Nonce = C.uint(ctx.Nonce)
	return ctx.cStruct
}

func New(index, intensity, worksize int) *GPUContext {
	gc := &GPUContext{}
	gc.DeviceIndex = index
	gc.RawIntensity = intensity
	gc.WorkSize = worksize
	return gc
}
