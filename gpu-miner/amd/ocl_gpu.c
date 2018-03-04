#include "ocl_gpu.h"
#include <stdio.h>
#include <assert.h>

#define OCL_ERR_SUCCESS (0)
#define OCL_ERR_API (2)
#define OCL_ERR_BAD_PARAMS (1)

typedef unsigned int uint;

//#define assert(cond) if(!(cond)) {printf("Assertion " ##cond "failed"); exit(-1);}

const char *kSetKernelArgErr = "Error %s when calling clSetKernelArg for kernel %d, argument %d.";

char* err_to_str(int ret)
{
        switch(ret)
        {
        case CL_SUCCESS:
                return "CL_SUCCESS";
        case CL_DEVICE_NOT_FOUND:
                return "CL_DEVICE_NOT_FOUND";
        case CL_DEVICE_NOT_AVAILABLE:
                return "CL_DEVICE_NOT_AVAILABLE";
        case CL_COMPILER_NOT_AVAILABLE:
                return "CL_COMPILER_NOT_AVAILABLE";
        case CL_MEM_OBJECT_ALLOCATION_FAILURE:
                return "CL_MEM_OBJECT_ALLOCATION_FAILURE";
        case CL_OUT_OF_RESOURCES:
                return "CL_OUT_OF_RESOURCES";
        case CL_OUT_OF_HOST_MEMORY:
                return "CL_OUT_OF_HOST_MEMORY";
        case CL_PROFILING_INFO_NOT_AVAILABLE:
                return "CL_PROFILING_INFO_NOT_AVAILABLE";
        case CL_MEM_COPY_OVERLAP:
                return "CL_MEM_COPY_OVERLAP";
        case CL_IMAGE_FORMAT_MISMATCH:
                return "CL_IMAGE_FORMAT_MISMATCH";
        case CL_IMAGE_FORMAT_NOT_SUPPORTED:
                return "CL_IMAGE_FORMAT_NOT_SUPPORTED";
        case CL_BUILD_PROGRAM_FAILURE:
                return "CL_BUILD_PROGRAM_FAILURE";
        case CL_MAP_FAILURE:
                return "CL_MAP_FAILURE";
        case CL_MISALIGNED_SUB_BUFFER_OFFSET:
                return "CL_MISALIGNED_SUB_BUFFER_OFFSET";
        case CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST:
                return "CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST";
        case CL_COMPILE_PROGRAM_FAILURE:
                return "CL_COMPILE_PROGRAM_FAILURE";
        case CL_LINKER_NOT_AVAILABLE:
                return "CL_LINKER_NOT_AVAILABLE";
        case CL_LINK_PROGRAM_FAILURE:
                return "CL_LINK_PROGRAM_FAILURE";
        case CL_DEVICE_PARTITION_FAILED:
                return "CL_DEVICE_PARTITION_FAILED";
        case CL_KERNEL_ARG_INFO_NOT_AVAILABLE:
                return "CL_KERNEL_ARG_INFO_NOT_AVAILABLE";
        case CL_INVALID_VALUE:
                return "CL_INVALID_VALUE";
        case CL_INVALID_DEVICE_TYPE:
                return "CL_INVALID_DEVICE_TYPE";
        case CL_INVALID_PLATFORM:
                return "CL_INVALID_PLATFORM";
        case CL_INVALID_DEVICE:
                return "CL_INVALID_DEVICE";
        case CL_INVALID_CONTEXT:
                return "CL_INVALID_CONTEXT";
        case CL_INVALID_QUEUE_PROPERTIES:
                return "CL_INVALID_QUEUE_PROPERTIES";
        case CL_INVALID_COMMAND_QUEUE:
                return "CL_INVALID_COMMAND_QUEUE";
        case CL_INVALID_HOST_PTR:
                return "CL_INVALID_HOST_PTR";
        case CL_INVALID_MEM_OBJECT:
                return "CL_INVALID_MEM_OBJECT";
        case CL_INVALID_IMAGE_FORMAT_DESCRIPTOR:
                return "CL_INVALID_IMAGE_FORMAT_DESCRIPTOR";
        case CL_INVALID_IMAGE_SIZE:
                return "CL_INVALID_IMAGE_SIZE";
        case CL_INVALID_SAMPLER:
                return "CL_INVALID_SAMPLER";
        case CL_INVALID_BINARY:
                return "CL_INVALID_BINARY";
        case CL_INVALID_BUILD_OPTIONS:
                return "CL_INVALID_BUILD_OPTIONS";
        case CL_INVALID_PROGRAM:
                return "CL_INVALID_PROGRAM";
        case CL_INVALID_PROGRAM_EXECUTABLE:
                return "CL_INVALID_PROGRAM_EXECUTABLE";
        case CL_INVALID_KERNEL_NAME:
                return "CL_INVALID_KERNEL_NAME";
        case CL_INVALID_KERNEL_DEFINITION:
                return "CL_INVALID_KERNEL_DEFINITION";
        case CL_INVALID_KERNEL:
                return "CL_INVALID_KERNEL";
        case CL_INVALID_ARG_INDEX:
                return "CL_INVALID_ARG_INDEX";
        case CL_INVALID_ARG_VALUE:
                return "CL_INVALID_ARG_VALUE";
        case CL_INVALID_ARG_SIZE:
                return "CL_INVALID_ARG_SIZE";
        case CL_INVALID_KERNEL_ARGS:
                return "CL_INVALID_KERNEL_ARGS";
        case CL_INVALID_WORK_DIMENSION:
                return "CL_INVALID_WORK_DIMENSION";
        case CL_INVALID_WORK_GROUP_SIZE:
                return "CL_INVALID_WORK_GROUP_SIZE";
        case CL_INVALID_WORK_ITEM_SIZE:
                return "CL_INVALID_WORK_ITEM_SIZE";
        case CL_INVALID_GLOBAL_OFFSET:
                return "CL_INVALID_GLOBAL_OFFSET";
        case CL_INVALID_EVENT_WAIT_LIST:
                return "CL_INVALID_EVENT_WAIT_LIST";
        case CL_INVALID_EVENT:
                return "CL_INVALID_EVENT";
        case CL_INVALID_OPERATION:
                return "CL_INVALID_OPERATION";
        case CL_INVALID_GL_OBJECT:
                return "CL_INVALID_GL_OBJECT";
        case CL_INVALID_BUFFER_SIZE:
                return "CL_INVALID_BUFFER_SIZE";
        case CL_INVALID_MIP_LEVEL:
                return "CL_INVALID_MIP_LEVEL";
        case CL_INVALID_GLOBAL_WORK_SIZE:
                return "CL_INVALID_GLOBAL_WORK_SIZE";
        case CL_INVALID_PROPERTY:
                return "CL_INVALID_PROPERTY";
        case CL_INVALID_IMAGE_DESCRIPTOR:
                return "CL_INVALID_IMAGE_DESCRIPTOR";
        case CL_INVALID_COMPILER_OPTIONS:
                return "CL_INVALID_COMPILER_OPTIONS";
        case CL_INVALID_LINKER_OPTIONS:
                return "CL_INVALID_LINKER_OPTIONS";
        case CL_INVALID_DEVICE_PARTITION_COUNT:
                return "CL_INVALID_DEVICE_PARTITION_COUNT";
#ifdef CL_VERSION_2_0
        case CL_INVALID_PIPE_SIZE:
                return "CL_INVALID_PIPE_SIZE";
        case CL_INVALID_DEVICE_QUEUE:
                return "CL_INVALID_DEVICE_QUEUE";
#endif
        default:
                return "UNKNOWN_ERROR";
        }
}


int XMRRunWork(void *ctx_ptr, void *results_ptr) {
        struct gpu_context *ctx = (struct gpu_context *)ctx_ptr;
        int *hashResults = (int *)results_ptr;

        cl_int ret;
        cl_uint zero = 0;
        size_t BranchNonces[4];
        memset(BranchNonces, 0, sizeof(size_t) * 4);

        size_t g_intensity = ctx->RawIntensity;
        size_t w_size = ctx->WorkSize;
        // round up to next multiple of w_size
        size_t g_thd = ((g_intensity + w_size - 1u) / w_size) * w_size;
        // number of global threads must be a multiple of the work group size (w_size)
        assert(g_thd % w_size == 0);

        for (int i = 2; i < 6; ++i) {
                if ((ret = clEnqueueWriteBuffer(ctx->CommandQueues, ctx->ExtraBuffers[i],
                                                CL_FALSE, sizeof(cl_uint) * g_intensity,
                                                sizeof(cl_uint), &zero, 0, NULL, NULL)) !=
                    CL_SUCCESS) {
                        printf("Error %s when calling clEnqueueWriteBuffer to zero branch "
                               "buffer counter %d.",
                               err_to_str(ret), i - 2);
                        return OCL_ERR_API;
                }
        }

        if ((ret = clEnqueueWriteBuffer(ctx->CommandQueues, ctx->OutputBuffer,
                                        CL_FALSE, sizeof(cl_uint) * 0xFF,
                                        sizeof(cl_uint), &zero, 0, NULL, NULL)) !=
            CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        clFinish(ctx->CommandQueues);

        size_t Nonce[2] = {ctx->Nonce, 1}, gthreads[2] = {g_thd, 8},
               lthreads[2] = {w_size, 8};
        if ((ret = clEnqueueNDRangeKernel(ctx->CommandQueues, ctx->Kernels[0], 2,
                                          Nonce, gthreads, lthreads, 0, NULL,
                                          NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueNDRangeKernel for kernel %d.",
                       err_to_str(ret), 0);
                return OCL_ERR_API;
        }

        /*for(int i = 1; i < 3; ++i)
           {
            if((ret = clEnqueueNDRangeKernel(*ctx->CommandQueues, ctx->Kernels[i], 1,
           &ctx->Nonce, &g_thd, &w_size, 0, NULL, NULL)) != CL_SUCCESS)
            {
                Log(LOG_CRITICAL, "Error %s when calling clEnqueueNDRangeKernel for
           kernel %d.", err_to_str(ret), i);
                return(ERR_OCL_API);
            }
           }*/

        size_t tmpNonce = ctx->Nonce;
        if ((ret = clEnqueueNDRangeKernel(ctx->CommandQueues, ctx->Kernels[1], 1,
                                          &tmpNonce, &g_thd, &w_size, 0, NULL,
                                          NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueNDRangeKernel for kernel %d.",
                       err_to_str(ret), 1);
                return OCL_ERR_API;
        }

        if ((ret = clEnqueueNDRangeKernel(ctx->CommandQueues, ctx->Kernels[2], 2,
                                          Nonce, gthreads, lthreads, 0, NULL,
                                          NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueNDRangeKernel for kernel %d.",
                       err_to_str(ret), 2);
                return OCL_ERR_API;
        }

        if ((ret = clEnqueueReadBuffer(ctx->CommandQueues, ctx->ExtraBuffers[2],
                                       CL_FALSE, sizeof(cl_uint) * g_intensity,
                                       sizeof(cl_uint), BranchNonces, 0, NULL,
                                       NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        if ((ret = clEnqueueReadBuffer(ctx->CommandQueues, ctx->ExtraBuffers[3],
                                       CL_FALSE, sizeof(cl_uint) * g_intensity,
                                       sizeof(cl_uint), BranchNonces + 1, 0, NULL,
                                       NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        if ((ret = clEnqueueReadBuffer(ctx->CommandQueues, ctx->ExtraBuffers[4],
                                       CL_FALSE, sizeof(cl_uint) * g_intensity,
                                       sizeof(cl_uint), BranchNonces + 2, 0, NULL,
                                       NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        if ((ret = clEnqueueReadBuffer(ctx->CommandQueues, ctx->ExtraBuffers[5],
                                       CL_FALSE, sizeof(cl_uint) * g_intensity,
                                       sizeof(cl_uint), BranchNonces + 3, 0, NULL,
                                       NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        clFinish(ctx->CommandQueues);

        for (int i = 0; i < 4; ++i) {
                if (BranchNonces[i]) {
                        // Threads
                        if ((clSetKernelArg(ctx->Kernels[i + 3], 4, sizeof(cl_ulong),
                                            BranchNonces + i)) != CL_SUCCESS) {
                                printf(kSetKernelArgErr, err_to_str(ret), i + 3, 4);
                                return OCL_ERR_API;
                        }

                        // round up to next multiple of w_size
                        BranchNonces[i] = ((BranchNonces[i] + w_size - 1u) / w_size) * w_size;
                        // number of global threads must be a multiple of the work group size
                        // (w_size)
                        assert(BranchNonces[i] % w_size == 0);
                        size_t tmpNonce = ctx->Nonce;
                        if ((ret = clEnqueueNDRangeKernel(ctx->CommandQueues, ctx->Kernels[i + 3],
                                                          1, &tmpNonce, BranchNonces + i, &w_size,
                                                          0, NULL, NULL)) != CL_SUCCESS) {
                                printf("Error %s when calling clEnqueueNDRangeKernel for kernel %d.",
                                       err_to_str(ret), i + 3);
                                return OCL_ERR_API;
                        }
                }
        }

        if ((ret = clEnqueueReadBuffer(ctx->CommandQueues, ctx->OutputBuffer, CL_TRUE,
                                       0, sizeof(cl_uint) * 0x100, hashResults, 0,
                                       NULL, NULL)) != CL_SUCCESS) {
                printf("Error %s when calling clEnqueueReadBuffer to fetch results.",
                       err_to_str(ret));
                return OCL_ERR_API;
        }

        clFinish(ctx->CommandQueues);
        uint numHashValues = hashResults[0xFF];
        // avoid out of memory read, we have only storage for 0xFF results
        if (numHashValues > 0xFF) {
                numHashValues = 0xFF;
        }

        ctx->Nonce += (uint32_t)g_intensity;

        return OCL_ERR_SUCCESS;
}

int testCContext(void *ctx_ptr, void *result) {
        struct gpu_context *ctx = (struct gpu_context *)ctx_ptr;
        int ret;
        size_t ret_size;
        ret = clGetDeviceInfo(ctx->DeviceID, CL_DEVICE_MAX_COMPUTE_UNITS, 4, result,
                           &ret_size);
        return ret;
}
