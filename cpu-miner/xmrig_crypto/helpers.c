#include "helpers.h"

#include <stdlib.h>
#include <stdio.h>

#if defined __unix__ && (!defined __APPLE__) && (!defined DISABLE_LINUX_HUGEPAGES)
#include <sys/mman.h>
#include <unistd.h>
#include <sys/types.h>
#elif defined _WIN32
#include <windows.h>
#include "mem_win.h"
#endif

#include <stdio.h>

void *xmrig_setup_hugepages(int nthreads)
{
  void *ret;
  int size = MEMORY * (nthreads * 1 + 1);
#if defined _WIN32
  TrySetLockPagesPrivilege();
  ret = VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE | MEM_LARGE_PAGES, PAGE_READWRITE);
  if(!ret) {
    printf("Failed to call VirtualAlloc()\n");
    ret = _mm_malloc(size, 16);
  }
#endif
  return ret;
}

void *xmrig_thread_persistent_ctx(void *memptr, int thread_id) {
  uint8_t *mem = (uint8_t *)memptr;
  struct cryptonight_ctx *persistent_ctx;
#if defined _WIN32
  persistent_ctx = (void *) &mem[MEMORY - sizeof(struct cryptonight_ctx) * (thread_id + 1)];
  persistent_ctx->memory = (void *) &mem[MEMORY * (thread_id * 1 + 1)];
#endif
  return persistent_ctx;
}

int xmrig_cryptonight_hash_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx)
{
  return xmrig_cryptonight_hash(input, size, output, target, (struct cryptonight_ctx *) ctx);
}
void xmrig_cryptonight_hash_void_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx)
{
  xmrig_cryptonight_hash_void(input, size, output, target, (struct cryptonight_ctx *) ctx);
}
