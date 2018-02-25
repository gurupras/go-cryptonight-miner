#include "helpers.h"

#include <stdlib.h>
#include <stdio.h>

#if defined __unix__ && (!defined __APPLE__) && (!defined DISABLE_LINUX_HUGEPAGES)
#include <sys/mman.h>
#include <unistd.h>
#include <sys/types.h>
#elif defined _WIN32
#include <windows.h>
#endif

void *xmrig_setup_persistent_ctx() {
  struct cryptonight_ctx *persistentctx = NULL;
	#if defined __unix__ && (!defined __APPLE__) && (!defined DISABLE_LINUX_HUGEPAGES)
	persistentctx = (struct cryptonight_ctx *)mmap(0, sizeof(struct cryptonight_ctx), PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB | MAP_POPULATE, 0, 0);
	if(persistentctx == MAP_FAILED) persistentctx = (struct cryptonight_ctx *)malloc(sizeof(struct cryptonight_ctx));
	madvise(persistentctx, sizeof(struct cryptonight_ctx), MADV_RANDOM | MADV_WILLNEED | MADV_HUGEPAGE);
	if(!geteuid()) mlock(persistentctx, sizeof(struct cryptonight_ctx));
	#elif defined _WIN32
  printf("%s: _WIN32", __func__);
	persistentctx = VirtualAlloc(NULL, sizeof(struct cryptonight_ctx), MEM_COMMIT | MEM_RESERVE | MEM_LARGE_PAGES, PAGE_READWRITE);
	if(!persistentctx) {
    printf("Failed to call VirtualAlloc()");
    persistentctx = (struct cryptonight_ctx *)malloc(sizeof(struct cryptonight_ctx));
  }
	#else
	persistentctx = (struct cryptonight_ctx *)malloc(sizeof(struct cryptonight_ctx));
	#endif
  return persistentctx;
}

int xmrig_cryptonight_hash_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx)
{
  return xmrig_cryptonight_hash(input, size, output, target, (struct cryptonight_ctx *) ctx);
}
void xmrig_cryptonight_hash_void_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx)
{
  xmrig_cryptonight_hash_void(input, size, output, target, (struct cryptonight_ctx *) ctx);
}
