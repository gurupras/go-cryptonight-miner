#ifndef __GPU_CONTEXT_H_
#define __GPU_CONTEXT_H_

#ifdef __APPLE__
#include "OpenCL/opencl.h"
#else
#include "CL/opencl.h"
#endif

struct gpu_context {
  int DeviceIndex;
  int RawIntensity;
  int WorkSize;
  cl_device_id DeviceID;
  cl_command_queue CommandQueues;
  cl_mem InputBuffer;
  cl_mem OutputBuffer;
  cl_mem ExtraBuffers[6];
  cl_program Program;
  cl_kernel Kernels[7];
  cl_ulong FreeMemory;
  cl_uint ComputeUnits;
  char *Name;
  unsigned int Nonce;
};

char* err_to_str(int ret);
int XMRRunWork(void *ctx_ptr, void *results_ptr);
int testCContext(void *ctx_ptr, void *result);
#endif
