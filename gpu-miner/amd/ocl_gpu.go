package amdgpu

/*
#cgo CFLAGS: -Icl -I.
#cgo !darwin LDFLAGS: -lOpenCL
#cgo darwin LDFLAGS: -framework OpenCL

#include "ocl_gpu.h"
*/
import "C"

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
	"unsafe"

	amdgpu_cl "github.com/gurupras/go-cryptonight-miner/gpu-miner/amd/cl"
	"github.com/gurupras/go-cryptonight-miner/gpu-miner/gpucontext"
	cl "github.com/rainliu/gocl/cl"
	log "github.com/sirupsen/logrus"
)

var (
	UseC bool = false
)

const (
	setKernelArgError = "Error %s when calling clSetKernelArg for kernel %d, argument %d"
)

//export LOG
func LOG(logType int, msg string) {
	// FIXME: Currently broken on Linux
	if runtime.GOOS != "windows" {
		return
	}
	switch logType {
	case C.TYPE_DEBUG:
		log.Debugf(msg)
	case C.TYPE_INFO:
		log.Infof(msg)
	case C.TYPE_ERR:
		log.Errorf(msg)
	case C.TYPE_FATAL:
		log.Fatalf(msg)
	case C.TYPE_WARN:
		log.Warnf(msg)
	default:
		log.Warnf("Received message from C with unknown log level(%d): %v", logType, msg)
	}
}

func portSleep(sec int) {
	time.Sleep(time.Duration(sec) * time.Second)
}

func err_to_str(ret cl.CL_int) string {
	result := C.err_to_str(C.int(ret))
	return C.GoString(result)
}

func logFromC(msg string) {
	cStr := C.CString(msg)
	C.testCLog(cStr)
}

func getDeviceMaxComputeUnits(id cl.CL_device_id) cl.CL_uint {
	count := 0
	var info interface{}
	cl.CLGetDeviceInfo(id, cl.CL_DEVICE_MAX_COMPUTE_UNITS, cl.CL_size_t(unsafe.Sizeof(count)), &info, nil)
	return info.(cl.CL_uint)
}

func getNumPlatforms() cl.CL_uint {
	var (
		count cl.CL_uint = 0
		ret   cl.CL_int
	)

	if ret = cl.CLGetPlatformIDs(0, nil, &count); ret != cl.CL_SUCCESS {
		log.Errorf("Failed to call clGetPlatformIDs: %v", err_to_str(ret))
	}
	return count
}

func getDeviceInfoBytes(deviceId cl.CL_device_id, info cl.CL_device_info, size cl.CL_size_t) ([]byte, error) {
	var ret interface{}
	if err := cl.CLGetDeviceInfo(deviceId, info, size, &ret, nil); err != cl.CL_SUCCESS {
		return nil, fmt.Errorf("Failed to get device info: %v", err_to_str(err))
	}
	return []byte(ret.(string)), nil
}

func getAMDDevices(index int) (contexts []*gpucontext.GPUContext) {
	numPlatforms := getNumPlatforms()

	platforms := make([]cl.CL_platform_id, numPlatforms)
	cl.CLGetPlatformIDs(numPlatforms, platforms, nil)

	var numDevices cl.CL_uint
	cl.CLGetDeviceIDs(platforms[index], cl.CL_DEVICE_TYPE_GPU, 0, nil, &numDevices)

	deviceList := make([]cl.CL_device_id, numDevices)
	cl.CLGetDeviceIDs(platforms[index], cl.CL_DEVICE_TYPE_GPU, numDevices, deviceList, nil)

	contexts = make([]*gpucontext.GPUContext, 0)
	for i := cl.CL_uint(0); i < numDevices; i++ {
		data, err := getDeviceInfoBytes(deviceList[i], cl.CL_DEVICE_VENDOR, 256)
		if err != nil {
			log.Errorf("Failed to get CL_DEVICE_VENDOR: %v", err)
			continue
		}
		str := string(data)
		if !strings.Contains(str, "Advanced Micro Devices") {
			continue
		}

		ctx := gpucontext.New(int(i), 0, 0)
		ctx.DeviceID = deviceList[i]
		ctx.ComputeUnits = getDeviceMaxComputeUnits(ctx.DeviceID)

		var (
			maxMem  cl.CL_ulong
			freeMem cl.CL_ulong
			mmIface interface{}
			fmIface interface{}
		)
		cl.CLGetDeviceInfo(ctx.DeviceID, cl.CL_DEVICE_MAX_MEM_ALLOC_SIZE, cl.CL_size_t(4), &mmIface, nil)
		cl.CLGetDeviceInfo(ctx.DeviceID, cl.CL_DEVICE_GLOBAL_MEM_SIZE, cl.CL_size_t(4), &fmIface, nil)
		// log.Infof("Types: maxMem: %t  freeMem: %t", maxMem, freeMem)
		maxMem = mmIface.(cl.CL_ulong)
		freeMem = fmIface.(cl.CL_ulong)
		ctx.FreeMemory = cl.CL_ulong(math.Min(float64(maxMem), float64(freeMem)))

		friendlyNameBytes, err := getDeviceInfoBytes(deviceList[i], cl.CL_DEVICE_NAME, 256)
		if err != nil {
			log.Errorf("Failed to get device name: %v", err)
			continue
		}
		ctx.Name = string(friendlyNameBytes)
		log.Debugf("OpenCL GPU: %v, cpu: %d", ctx.Name, ctx.ComputeUnits)
		contexts = append(contexts, ctx)
	}
	return
}

func printPlatforms() {
	numPlatforms := getNumPlatforms()
	if numPlatforms == 0 {
		return
	}

	platforms := make([]cl.CL_platform_id, numPlatforms)
	cl.CLGetPlatformIDs(numPlatforms, platforms, nil)

	for i := 0; i < int(numPlatforms); i++ {
		var vendor interface{}
		if cl.CLGetPlatformInfo(platforms[i], cl.CL_PLATFORM_VENDOR, 256, &vendor, nil) != cl.CL_SUCCESS {
			continue
		}
		log.Infof("#%d: %v", i, vendor)
	}
}

func setKernelArgFromExtraBuffers(ctx *gpucontext.GPUContext, kernel int, argument cl.CL_uint, offset int) bool {
	buf := ctx.ExtraBuffers[offset]
	if ret := cl.CLSetKernelArg(ctx.Kernels[kernel], argument, clMemSize(), unsafe.Pointer(&buf)); ret != cl.CL_SUCCESS {
		return false
	}
	return true
}

func GoInitOpenCLGPU(index int, clCtx cl.CL_context, ctx *gpucontext.GPUContext, code [][]byte) error {

	var maxWorkSizeIntf interface{}

	if ret := cl.CLGetDeviceInfo(ctx.DeviceID, cl.CL_DEVICE_MAX_WORK_GROUP_SIZE, cl.CL_size_t(unsafe.Sizeof(index)), &maxWorkSizeIntf, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when querying device's max worksize: %v", err_to_str(ret))
	}

	deviceNameBytes, err := getDeviceInfoBytes(ctx.DeviceID, cl.CL_DEVICE_NAME, 256)
	if err != nil {
		log.Errorf("Failed to get device name: %v", err)
	}
	ctx.Name = string(deviceNameBytes)
	ctx.ComputeUnits = getDeviceMaxComputeUnits(ctx.DeviceID)

	log.Infof("#%d, GPU #%d %s, intensity: %d (%d/%v), cu: %d", index, ctx.DeviceIndex, ctx.Name, ctx.RawIntensity, ctx.WorkSize, maxWorkSizeIntf, ctx.ComputeUnits)

	var commandQueueProperties cl.CL_command_queue_properties
	var ret cl.CL_int
	// TODO: Add logic to do this differently for CL_VERSION_2_0
	// This is the non CL_VERSION_2_0 version
	ctx.CommandQueues = cl.CLCreateCommandQueue(clCtx, ctx.DeviceID, commandQueueProperties, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateCommandQueue: %v", err_to_str(ret))
	}

	ctx.InputBuffer = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_ONLY, 88, nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer: %v", err_to_str(ret))
	}

	//TODO: handle AEON?
	hashMemSize := MONERO_MEMORY
	threadMemMask := MONERO_MASK
	hasIterations := MONERO_ITER

	g_thd := ctx.RawIntensity
	ctx.ExtraBuffers[0] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(hashMemSize)*g_thd), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for scratchpads buffer: %v", err_to_str(ret))
	}

	ctx.ExtraBuffers[1] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(100*g_thd), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for hash states buffer: %v", err_to_str(ret))
	}

	// Blake-256 branches
	ctx.ExtraBuffers[2] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(unsafe.Sizeof(ret))*(g_thd+2)), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for branch-0 buffer: %v", err_to_str(ret))
	}

	// Groestl-256 branches
	ctx.ExtraBuffers[3] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(unsafe.Sizeof(ret))*(g_thd+2)), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for branch-1 buffer: %v", err_to_str(ret))
	}

	// JH-256 branches
	ctx.ExtraBuffers[4] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(unsafe.Sizeof(ret))*(g_thd+2)), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for branch-2 buffer: %v", err_to_str(ret))
	}

	// Skein-512 branches
	ctx.ExtraBuffers[5] = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(unsafe.Sizeof(ret))*(g_thd+2)), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for branch-3 buffer: %v", err_to_str(ret))
	}

	ctx.OutputBuffer = cl.CLCreateBuffer(clCtx, cl.CL_MEM_READ_WRITE, cl.CL_size_t(int(unsafe.Sizeof(ret))*0x100), nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateBuffer for output buffer: %v", err_to_str(ret))
	}

	ctx.Program = cl.CLCreateProgramWithSource(clCtx, 1, code, []cl.CL_size_t{cl.CL_size_t(len(code[0]))}, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateProgramWithSource: %v", err_to_str(ret))
	}

	options := fmt.Sprintf("-DITERATIONS=%d -DMASK=%d -DWORKSIZE=%d", hasIterations, threadMemMask, ctx.WorkSize)
	if ret = cl.CLBuildProgram(ctx.Program, 1, []cl.CL_device_id{ctx.DeviceID}, []byte(options), nil, nil); ret != cl.CL_SUCCESS {
		log.Errorf("Error when calling clBuildProgram: %v", err_to_str(ret))

		var len cl.CL_size_t
		if ret = cl.CLGetProgramBuildInfo(ctx.Program, ctx.DeviceID, cl.CL_PROGRAM_BUILD_LOG, 0, nil, &len); ret != cl.CL_SUCCESS {
			return fmt.Errorf("Error when calling clGetProgramBuildInfo for length of build log output: %v", err_to_str(ret))
		}

		var buildLog interface{}
		if ret = cl.CLGetProgramBuildInfo(ctx.Program, ctx.DeviceID, cl.CL_PROGRAM_BUILD_LOG, len, &buildLog, nil); ret != cl.CL_SUCCESS {
			return fmt.Errorf("Error when calling clGetProgramBuildInfo for build log: %v", err_to_str(ret))
		}
		log.Infof("Build log: \n")
		fmt.Printf("%v\n", buildLog)
		return fmt.Errorf("Failed to build program")
	}

	var statusIface interface{}
	var status cl.CL_build_status
	for {
		if ret = cl.CLGetProgramBuildInfo(ctx.Program, ctx.DeviceID, cl.CL_PROGRAM_BUILD_STATUS, cl.CL_size_t(unsafe.Sizeof(status)), &statusIface, nil); ret != cl.CL_SUCCESS {
			return fmt.Errorf("Error when calling clGetProgramBuildInfo for status of build: %v", err_to_str(ret))
		}
		status = statusIface.(cl.CL_build_status)
		if status != cl.CL_BUILD_IN_PROGRESS {
			break
		}
		portSleep(1)
	}

	kernelNames := []string{"cn0", "cn1", "cn2", "Blake", "Groestl", "JH", "Skein"}
	for i := 0; i < 7; i++ {
		ctx.Kernels[i] = cl.CLCreateKernel(ctx.Program, []byte(kernelNames[i]), &ret)
		if ret != cl.CL_SUCCESS {
			return fmt.Errorf("Error when calling clCreateKernel for kernel %s: %v", kernelNames[i], err_to_str(ret))
		}
	}
	ctx.Nonce = 0
	return nil
}

func getAMDPlatformIndex() int {
	numPlatforms := getNumPlatforms()
	if numPlatforms == 0 {
		return -1
	}

	platforms := make([]cl.CL_platform_id, numPlatforms)
	cl.CLGetPlatformIDs(numPlatforms, platforms, nil)

	for i := 0; i < int(numPlatforms); i++ {
		var vendor interface{}
		cl.CLGetPlatformInfo(platforms[i], cl.CL_PLATFORM_VENDOR, 256, &vendor, nil)
		vendorStr := vendor.(string)
		if strings.Contains(vendorStr, "Advanced Micro Devices") {
			log.Infof("Found AMD platform index: %d name: %v", i, vendor)
			return i
		}
	}
	return -1
}

func getCode() string {
	cryptonightCL := amdgpu_cl.Cryptonight_CL_STR
	blake256CL := amdgpu_cl.Blake256_CL_STR
	groestl256CL := amdgpu_cl.Groestl256_CL_STR
	jhCL := amdgpu_cl.JH_CL_STR
	wolfAesCL := amdgpu_cl.WolfAES_CL_STR
	skeinCL := amdgpu_cl.Skein_CL_STR

	code := cryptonightCL
	replacementMap := make(map[string]string)
	replacementMap["XMRIG_INCLUDE_WOLF_AES"] = wolfAesCL
	replacementMap["XMRIG_INCLUDE_WOLF_SKEIN"] = skeinCL
	replacementMap["XMRIG_INCLUDE_JH"] = jhCL
	replacementMap["XMRIG_INCLUDE_BLAKE256"] = blake256CL
	replacementMap["XMRIG_INCLUDE_GROESTL256"] = groestl256CL

	for k, v := range replacementMap {
		code = strings.Replace(code, k, v, -1)
		if strings.Contains(code, k) {
			log.Warnf("Failed to replace code: %v", k)
		}
	}
	return code
}
func InitOpenCL(gpuContexts []*gpucontext.GPUContext, numGPUs int, platformIndex int) error {
	if UseC {
		return CInitOpenCL(gpuContexts, numGPUs, platformIndex)
	} else {
		return GoInitOpenCL(gpuContexts, numGPUs, platformIndex)
	}
}

func CInitOpenCL(gpuContexts []*gpucontext.GPUContext, numGPUs int, platformIndex int) error {
	cContexts := make([]uint64, len(gpuContexts))

	code := getCode()
	cCode := C.CString(code)

	for i := 0; i < len(gpuContexts); i++ {
		cCtx := gpuContexts[i].AsCStruct()
		cContexts[i] = uint64(uintptr(unsafe.Pointer(cCtx)))
	}

	// contextPtrs := make([]uintptr, len(gpuContexts))
	// for i := 0; i < len(gpuContexts); i++ {
	// 	contextPtrs[i] = unsafe.Pointer(&cContexts[i])
	// }

	ctxPtr := unsafe.Pointer(&cContexts[0])
	if ret := C.InitOpenCL(ctxPtr, C.int(numGPUs), C.int(platformIndex), cCode); ret != 0 {
		return fmt.Errorf("Failed to initialize OpenCL: %v", ret)
	}
	return nil
}

func GoInitOpenCL(gpuContexts []*gpucontext.GPUContext, numGPUs int, platformIndex int) error {
	numPlatforms := getNumPlatforms()
	if numPlatforms == 0 {
		return fmt.Errorf("Did not find any OpenCL platforms")
	}

	if int(numPlatforms) <= platformIndex {
		return fmt.Errorf("Selected OpenCL platform index %d doesn't exist", platformIndex)
	}

	platforms := make([]cl.CL_platform_id, numPlatforms)
	cl.CLGetPlatformIDs(numPlatforms, platforms, nil)

	var vendorIface interface{}
	cl.CLGetPlatformInfo(platforms[platformIndex], cl.CL_PLATFORM_VENDOR, 256, &vendorIface, nil)
	vendorStr := vendorIface.(string)
	if !strings.Contains(vendorStr, "Advanced Micro Devices") {
		log.Warnf("Using non AMD devices: %s", vendorIface)
	}

	platformIdList := make([]cl.CL_platform_id, numPlatforms)
	cl.CLGetPlatformIDs(numPlatforms, platformIdList, nil)

	var numDevices cl.CL_uint
	if ret := cl.CLGetDeviceIDs(platformIdList[platformIndex], cl.CL_DEVICE_TYPE_GPU, 0, nil, &numDevices); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clGetDeviceIDs for number of devices: %v", err_to_str(ret))
	}

	for i := 0; i < numGPUs; i++ {
		if int(numDevices) <= gpuContexts[i].DeviceIndex {
			return fmt.Errorf("Selected OpenCL device index %d doesn't exist", gpuContexts[i].DeviceIndex)
		}
	}

	deviceIdList := make([]cl.CL_device_id, numDevices)

	if ret := cl.CLGetDeviceIDs(platformIdList[platformIndex], cl.CL_DEVICE_TYPE_GPU, numDevices, deviceIdList, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clGetDeviceIDs for device ID information: %v", err_to_str(ret))
	}

	tempDeviceList := make([]cl.CL_device_id, numGPUs)

	for i := 0; i < numGPUs; i++ {
		gpuContexts[i].DeviceID = deviceIdList[gpuContexts[i].DeviceIndex]
		tempDeviceList[i] = deviceIdList[gpuContexts[i].DeviceIndex]
	}

	var ret cl.CL_int
	clCtx := cl.CLCreateContext(nil, cl.CL_uint(numGPUs), tempDeviceList, nil, nil, &ret)
	if ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clCreateContext: %v", err_to_str(ret))
	}

	code := getCode()

	var codeBytes [1][]byte
	codeBytes[0] = []byte(code)
	//wg := sync.WaitGroup{}
	//failed := false
	for i := 0; i < numGPUs; i++ {
		/*
			wg.Add(1)
			go func(i int, clCtx cl.CL_context, ctx *gpucontext.GPUContext, code [][]byte) {
				defer wg.Done()
				if err := GoInitOpenCLGPU(i, clCtx, ctx, code); err != nil {
					failed = true
				}
			}(i, clCtx, gpuContexts[i], codeBytes[:])
		*/
		if err := GoInitOpenCLGPU(i, clCtx, gpuContexts[i], codeBytes[:]); err != nil {
			return err
		}
	}
	//wg.Wait()
	return nil
}

func SetWork(ctx *gpucontext.GPUContext, input []byte, workSize int, target uint64) error {
	if UseC {
		return CSetWork(ctx, input, workSize, target)
	}
	return GoSetWork(ctx, input, workSize, target)
}

func CSetWork(ctx *gpucontext.GPUContext, input []byte, workSize int, target uint64) error {
	cs := ctx.AsCStruct()
	ctxPtr := unsafe.Pointer(cs)

	inputPtr := unsafe.Pointer(&input[0])
	targetPtr := unsafe.Pointer(&target)

	if err := C.XMRSetWork(ctxPtr, inputPtr, C.int(workSize), targetPtr); err != 0 {
		return fmt.Errorf("Failed to run C XMRSetWork: %v", err)
	}
	return nil
}

func GoSetWork(ctx *gpucontext.GPUContext, input []byte, workSize int, target uint64) error {
	var ret cl.CL_int

	if workSize > 84 {
		return fmt.Errorf("Work size too long?")
	}

	log.Debugf("input length: %d", len(input))
	input[workSize] = 0x01
	for i := workSize + 1; i < (workSize+1)+(88-workSize-1); i++ {
		input[i] = 0
	}

	numThreads := ctx.RawIntensity
	ibuf := ctx.InputBuffer
	inputPtr := unsafe.Pointer(&input[0])
	if ret = cl.CLEnqueueWriteBuffer(ctx.CommandQueues, ctx.InputBuffer, cl.CL_TRUE, 0, 88, inputPtr, 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueWriteBuffer to fill input buffer: %v", err_to_str(ret))
	}

	if ret = cl.CLSetKernelArg(ctx.Kernels[0], 0, clMemSize(), unsafe.Pointer(&ibuf)); ret != cl.CL_SUCCESS {
		return fmt.Errorf(setKernelArgError, err_to_str(ret), 0, 0)
	}

	// Scratchpads, states
	if !setKernelArgFromExtraBuffers(ctx, 0, 1, 0) || !setKernelArgFromExtraBuffers(ctx, 0, 2, 1) {
		return fmt.Errorf("Failed to set kernel args from extra buffers")
	}

	// Threads
	if ret = cl.CLSetKernelArg(ctx.Kernels[0], 3, clLongSize(), unsafe.Pointer(&numThreads)); ret != cl.CL_SUCCESS {
		return fmt.Errorf(setKernelArgError, err_to_str(ret), 0, 3)
	}

	// CN2 Kernel
	// Scratchpads, states
	if !setKernelArgFromExtraBuffers(ctx, 1, 0, 0) || !setKernelArgFromExtraBuffers(ctx, 1, 1, 1) {
		return fmt.Errorf("Failed to set kernel args from extra buffers")
	}

	// Threads
	if ret = cl.CLSetKernelArg(ctx.Kernels[1], 2, clLongSize(), unsafe.Pointer(&numThreads)); ret != cl.CL_SUCCESS {
		return fmt.Errorf(setKernelArgError, err_to_str(ret), 1, 2)
	}

	// CN3 Kernel
	// Scratchpads, states
	if !setKernelArgFromExtraBuffers(ctx, 2, 0, 0) || !setKernelArgFromExtraBuffers(ctx, 2, 1, 1) {
		return fmt.Errorf("Failed to set kernel args from extra buffers")
	}

	// Branch 0-3
	for i := 0; i < 4; i++ {
		if !setKernelArgFromExtraBuffers(ctx, 2, cl.CL_uint(i+2), i+2) {
			return fmt.Errorf("Failed to set kernel args from extra buffers for branch: %v", i)
		}
	}

	// Threads
	if ret = cl.CLSetKernelArg(ctx.Kernels[2], 6, clLongSize(), unsafe.Pointer(&numThreads)); ret != cl.CL_SUCCESS {
		return fmt.Errorf(setKernelArgError, err_to_str(ret), 2, 6)
	}

	for i := 0; i < 4; i++ {
		// Nonce buffer, Output
		if !setKernelArgFromExtraBuffers(ctx, i+3, 0, 1) || !setKernelArgFromExtraBuffers(ctx, i+3, 1, i+2) {
			return fmt.Errorf("Failed while setting nonce buffer")
		}

		// Output
		obuf := ctx.OutputBuffer
		if ret = cl.CLSetKernelArg(ctx.Kernels[i+3], 2, clMemSize(), unsafe.Pointer(&obuf)); ret != cl.CL_SUCCESS {
			return fmt.Errorf(setKernelArgError, err_to_str(ret), i+3, 2)
		}

		// Target
		if ret = cl.CLSetKernelArg(ctx.Kernels[i+3], 3, clLongSize(), unsafe.Pointer(&target)); ret != cl.CL_SUCCESS {
			return fmt.Errorf(setKernelArgError, err_to_str(ret), i+3, 3)
		}
	}
	return nil
}

func clSizeWrap(v uintptr) cl.CL_size_t {
	return cl.CL_size_t(v)
}

func clIntSize() cl.CL_size_t {
	var v cl.CL_int
	return clSizeWrap(unsafe.Sizeof(v))
}
func clLongSize() cl.CL_size_t {
	var v cl.CL_long
	return clSizeWrap(unsafe.Sizeof(v))
}

func clMemSize() cl.CL_size_t {
	var v cl.CL_mem
	return clSizeWrap(unsafe.Sizeof(v))
}

func clGetSize(v interface{}) cl.CL_size_t {
	return clSizeWrap(unsafe.Sizeof(v))
}

func RunWork(ctx *gpucontext.GPUContext, hashResults []cl.CL_int) error {
	if UseC {
		return CRunWork(ctx, hashResults)
	}
	return GoRunWork(ctx, hashResults)
}

func CRunWork(ctx *gpucontext.GPUContext, hashResults []cl.CL_int) error {
	cs := ctx.AsCStruct()
	ctxPtr := unsafe.Pointer(cs)
	resultsPtr := unsafe.Pointer(&hashResults[0])
	if ret := C.XMRRunWork(ctxPtr, resultsPtr); ret != 0 {
		return fmt.Errorf("Failed to run work")
	}
	// We need to move Nonce since nonce is moved in the C-version
	ctx.Nonce = uint32(cs.Nonce)
	return nil
}

func GoRunWork(ctx *gpucontext.GPUContext, hashResults []cl.CL_int) error {
	var (
		ret  cl.CL_int
		zero cl.CL_uint = 0
	)
	branchNonces := make([]cl.CL_size_t, 4)

	gIntensity := ctx.RawIntensity
	workSize := ctx.WorkSize

	// Round up to next multiple of workSize
	g_thd := ((gIntensity + workSize - 1) / workSize) * workSize

	// number of global threads must be a multiple of the work group size (workSize)
	if g_thd%workSize != 0 {
		log.Fatalf("Number of global threads must be a multiple of the work group size %d %% %d != 0", g_thd, workSize)
	}

	for i := 2; i < 6; i++ {
		if ret = cl.CLEnqueueWriteBuffer(ctx.CommandQueues, ctx.ExtraBuffers[i], cl.CL_FALSE, clIntSize()*cl.CL_size_t(gIntensity), clIntSize(), unsafe.Pointer(&zero), 0, nil, nil); ret != cl.CL_SUCCESS {
			return fmt.Errorf("Error when calling clEnqueueWriteBuffer to zero branch buffer counter %d: %v", i-2, err_to_str(ret))
		}
	}

	if ret = cl.CLEnqueueWriteBuffer(ctx.CommandQueues, ctx.OutputBuffer, cl.CL_FALSE, clIntSize()*0xFF, clIntSize(), unsafe.Pointer(&zero), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueWriteBuffer to fetch results", err_to_str(ret))
	}

	cl.CLFinish(ctx.CommandQueues)

	nonce := []cl.CL_size_t{cl.CL_size_t(ctx.Nonce), 1}
	gThreads := []cl.CL_size_t{cl.CL_size_t(g_thd), 8}
	lThreads := []cl.CL_size_t{cl.CL_size_t(workSize), 8}

	if ret = cl.CLEnqueueNDRangeKernel(ctx.CommandQueues, ctx.Kernels[0], 2, nonce, gThreads, lThreads, 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueNDRangeKernel for kernel %d: %v", 0, err_to_str(ret))
	}

	tmpNonce := []cl.CL_size_t{cl.CL_size_t(ctx.Nonce)}
	wSizeSlice := []cl.CL_size_t{cl.CL_size_t(workSize)}
	g_thdSlice := []cl.CL_size_t{cl.CL_size_t(g_thd)}
	if ret = cl.CLEnqueueNDRangeKernel(ctx.CommandQueues, ctx.Kernels[1], 1, tmpNonce, g_thdSlice, wSizeSlice, 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueNDRangeKernel for kernel %d: %v", 1, err_to_str(ret))
	}

	if ret = cl.CLEnqueueNDRangeKernel(ctx.CommandQueues, ctx.Kernels[2], 2, nonce, gThreads, lThreads, 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueNDRangeKernel for kernel %d: %v", 2, err_to_str(ret))
	}

	if ret = cl.CLEnqueueReadBuffer(ctx.CommandQueues, ctx.ExtraBuffers[2], cl.CL_FALSE, clIntSize()*cl.CL_size_t(gIntensity), clIntSize(), unsafe.Pointer(&branchNonces[0]), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueReadBuffer to fetch results: %v", err_to_str(ret))
	}

	if ret = cl.CLEnqueueReadBuffer(ctx.CommandQueues, ctx.ExtraBuffers[3], cl.CL_FALSE, clIntSize()*cl.CL_size_t(gIntensity), clIntSize(), unsafe.Pointer(&branchNonces[1]), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueReadBuffer to fetch results: %v", err_to_str(ret))
	}

	if ret = cl.CLEnqueueReadBuffer(ctx.CommandQueues, ctx.ExtraBuffers[4], cl.CL_FALSE, clIntSize()*cl.CL_size_t(gIntensity), clIntSize(), unsafe.Pointer(&branchNonces[2]), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueReadBuffer to fetch results: %v", err_to_str(ret))
	}

	if ret = cl.CLEnqueueReadBuffer(ctx.CommandQueues, ctx.ExtraBuffers[5], cl.CL_FALSE, clIntSize()*cl.CL_size_t(gIntensity), clIntSize(), unsafe.Pointer(&branchNonces[3]), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueReadBuffer to fetch results: %v", err_to_str(ret))
	}

	cl.CLFinish(ctx.CommandQueues)

	for i := 0; i < 4; i++ {
		if branchNonces[0] != 0 {
			// Threads
			if ret = cl.CLSetKernelArg(ctx.Kernels[i+3], 4, clLongSize(), unsafe.Pointer(&branchNonces[i])); ret != cl.CL_SUCCESS {
				return fmt.Errorf(setKernelArgError, err_to_str(ret), i+3, 4)
			}

			// Round up to next multiple of workSize
			branchNonces[i] = ((branchNonces[i] + cl.CL_size_t(workSize) - 1) / cl.CL_size_t(workSize)) * cl.CL_size_t(workSize)
			// number of global threads must be a multiple of the work group size (workSize)
			if int(branchNonces[i])%workSize != 0 {
				log.Fatalf("Number of global threads must be a multiple of the work group size (%d)", workSize)
			}

			tmpNonceSlice := []cl.CL_size_t{cl.CL_size_t(ctx.Nonce)}
			if ret = cl.CLEnqueueNDRangeKernel(ctx.CommandQueues, ctx.Kernels[i+3], 1, tmpNonceSlice, branchNonces[i:], wSizeSlice, 0, nil, nil); ret != cl.CL_SUCCESS {
				return fmt.Errorf("Error when calling clEnqueueNDRangeKernel for kernel %d: %v", i+3, err_to_str(ret))
			}
		}
	}

	if ret = cl.CLEnqueueReadBuffer(ctx.CommandQueues, ctx.OutputBuffer, cl.CL_TRUE, 0, clIntSize()*0x100, unsafe.Pointer(&hashResults[0]), 0, nil, nil); ret != cl.CL_SUCCESS {
		return fmt.Errorf("Error when calling clEnqueueReadBuffer to fetch results: %v", err_to_str(ret))
	}

	cl.CLFinish(ctx.CommandQueues)
	numHashValues := hashResults[0xFF]
	// Avoid out of memory read, we only have storage for 0xFF reads
	if numHashValues > 0xFF {
		numHashValues = 0xFF
	}

	ctx.Nonce += uint32(gIntensity)
	return nil
}

func testCContext(ctx *gpucontext.GPUContext) error {
	cs := ctx.AsCStruct()
	ptr := unsafe.Pointer(cs)
	cu := 0
	if ret := C.testCContext(ptr, unsafe.Pointer(&cu)); ret != C.CL_SUCCESS {
		return fmt.Errorf("Failed to properly issue command on C context: %v", ret)
	} else {
		if cu != int(ctx.ComputeUnits) {
			return fmt.Errorf("Compute units did not match: go=%d c=%d", ctx.ComputeUnits, cu)
		}
	}
	return nil
}
