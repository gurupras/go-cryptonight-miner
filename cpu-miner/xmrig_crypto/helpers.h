#ifndef __HELPERS_H_
#define __HELPERS_H_
#include "cryptonight.h"

void *xmrig_setup_hugepages(int nthreads);
void *xmrig_thread_persistent_ctx(void *mem, int thread_id);
int xmrig_cryptonight_hash_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx);
void xmrig_cryptonight_hash_void_wrapper(const void *input, int size, const void *output, const  void *target, void *ctx);

#endif
