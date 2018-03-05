#include "helpers.h"

#include <stdio.h>
#include <stdlib.h>

#if defined __unix__ && (!defined __APPLE__) &&                                \
        (!defined DISABLE_LINUX_HUGEPAGES)
#include <sys/mman.h>
#include <sys/types.h>
#include <unistd.h>
#include "aligned_malloc.h"
#elif defined _WIN32
#include "mem_win.h"
#include <windows.h>
#endif

#include <stdio.h>

static int USING_HUGEPAGES = 1;
static int size = 0;
#ifndef _WIN32
static int LOCKED = 0;
#endif
void *xmrig_setup_hugepages(int nthreads) {
	void *ret;
	size = MEMORY * (nthreads * 1 + 1);
#if defined _WIN32
	TrySetLockPagesPrivilege();
	ret = VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE | MEM_LARGE_PAGES,
	                   PAGE_READWRITE);
	if (!ret) {
		printf("Failed to call VirtualAlloc()\n");
		USING_HUGEPAGES = 0;
		ret = _mm_malloc(size, 16);
	}
#else
// POSIX-like
#if defined(__APPLE__)
	ret = mmap(0, size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANON,
	           VM_FLAGS_SUPERPAGE_SIZE_2MB, 0);
#elif defined(__FreeBSD__)
	ret =
		mmap(0, size, PROT_READ | PROT_WRITE,
		     MAP_PRIVATE | MAP_ANONYMOUS | MAP_ALIGNED_SUPER | MAP_PREFAULT_READ,
		     -1, 0);
#else
	ret = mmap(0, size, PROT_READ | PROT_WRITE,
	           MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB | MAP_POPULATE, 0, 0);
#endif
	if (ret == MAP_FAILED) {
		USING_HUGEPAGES = 0;
		ret = _mm_malloc(size, 16);
		goto out;
	}

	if (madvise(ret, size, MADV_RANDOM | MADV_WILLNEED) != 0) {
		printf("madvise failed");
	}

	if (mlock(ret, size) == 0) {
		LOCKED = 1;
	}
out:
#endif
	return ret;
}

void release_hugepages(void *ptr) {
	if (USING_HUGEPAGES) {
#ifdef _WIN32
		VirtualFree(ptr, 0, MEM_RELEASE);
#else
		if (LOCKED) {
			munlock(ptr, size);
		}
		munmap(ptr, size);
#endif
	} else {
		_mm_free(ptr);
	}
}

void *xmrig_thread_persistent_ctx(void *memptr, int thread_id) {
	uint8_t *mem = (uint8_t *)memptr;
	struct cryptonight_ctx *persistent_ctx;
	persistent_ctx =
		(void *)&mem[MEMORY - sizeof(struct cryptonight_ctx) * (thread_id + 1)];
	persistent_ctx->memory = (void *)&mem[MEMORY * (thread_id * 1 + 1)];
	return persistent_ctx;
}

void *xmrig_simple_cryptonight_context() {
	return _mm_malloc(sizeof(struct cryptonight_ctx), 16);
}

int xmrig_cryptonight_hash_wrapper(const void *input, int size,
                                   const void *output, const void *target,
                                   void *ctx) {
	return xmrig_cryptonight_hash(input, size, output, target,
	                              (struct cryptonight_ctx *)ctx);
}
void xmrig_cryptonight_hash_void_wrapper(const void *input, int size,
                                         const void *output, const void *target,
                                         void *ctx) {
	xmrig_cryptonight_hash_void(input, size, output, target,
	                            (struct cryptonight_ctx *)ctx);
}
