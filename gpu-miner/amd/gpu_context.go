package amdgpu

import "github.com/rainliu/gocl/cl"

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
}

func New(index, intensity, worksize int) *GPUContext {
	gc := &GPUContext{}
	gc.DeviceIndex = index
	gc.RawIntensity = intensity
	gc.WorkSize = worksize
	return gc
}
