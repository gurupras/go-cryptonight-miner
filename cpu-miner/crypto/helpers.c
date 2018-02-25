#include "helpers.h"
#include "miner.h"

#include <stdlib.h>
#include <stdio.h>

#if defined __unix__ && (!defined __APPLE__) && (!defined DISABLE_LINUX_HUGEPAGES)
#include <sys/mman.h>
#include <unistd.h>
#include <sys/types.h>
#elif defined _WIN32
#include <windows.h>
#endif

void *setup_persistent_ctx() {
  struct cryptonight_ctx *persistentctx = NULL;
	#if defined __unix__ && (!defined __APPLE__) && (!defined DISABLE_LINUX_HUGEPAGES)
	persistentctx = (struct cryptonight_ctx *)mmap(0, sizeof(struct cryptonight_ctx), PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB | MAP_POPULATE, 0, 0);
	if(persistentctx == MAP_FAILED) persistentctx = (struct cryptonight_ctx *)malloc(sizeof(struct cryptonight_ctx));
	madvise(persistentctx, sizeof(struct cryptonight_ctx), MADV_RANDOM | MADV_WILLNEED | MADV_HUGEPAGE);
	if(!geteuid()) mlock(persistentctx, sizeof(struct cryptonight_ctx));
	#elif defined _WIN32
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

int scanhash_cryptonight_wrapper(int thr_id, void *pdata, const void *ptarget,
		unsigned long max_nonce, void *hashes_done, void *persistentctx, void *restart)
{
  return scanhash_cryptonight(thr_id, (uint32_t *) pdata, (uint32_t *) ptarget, (uint32_t) max_nonce, (unsigned long *) hashes_done, (struct cryptonight_ctx *) persistentctx, (int *) restart);
}

void cryptonight_hash_wrapper(void *output, void *input, int length)
{
  cryptonight_hash(output, input, (size_t) length);
}
