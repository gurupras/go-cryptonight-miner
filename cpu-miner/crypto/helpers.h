#ifndef __HELPERS_H_
#define __HELPERS_H_
#include "cryptonight.h"

void *setup_persistent_ctx();
int scanhash_cryptonight_wrapper(int thr_id, void *pdata, const void *ptarget,
		unsigned long max_nonce, void *hashes_done, void *persistentctx, void *restart);
void cryptonight_hash_wrapper(void *output, void *input, int length);
void simple_fn(const char *data, int len);
#endif
