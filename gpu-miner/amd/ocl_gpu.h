#ifndef __OCL_GPU_H_
#define __OCL_GPU_H_

#ifdef __APPLE__
#include "OpenCL/opencl.h"
#else
#include "CL/opencl.h"
#include "CL/cl_ext.h"
#endif

#define MONERO_MEMORY 2097152
#define MONERO_MASK   0x1FFFF0
#define MONERO_ITER   0x80000

enum LOG_TYPE {
  TYPE_DEBUG = 0,
  TYPE_INFO = 1,
  TYPE_WARN = 2,
  TYPE_ERR = 3,
  TYPE_FATAL = 4,
};

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

struct topology {
  int bus;
  int device;
  int function;
};

char* err_to_str(int ret);
void testCLog(char *msg);
void GetTopology(void *, void *);
int InitOpenCL(void *ctx_ptr, int num_gpus, int platform_idx, const char *code);
int XMRSetWork(void *ctx_ptr, void *input_vptr, int input_len, void *target_ptr);
int XMRRunWork(void *ctx_ptr, void *results_ptr);
int testCContext(void *ctx_ptr, void *result);
#endif
