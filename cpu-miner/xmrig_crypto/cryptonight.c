/* XMRig
 * Copyright 2010      Jeff Garzik <jgarzik@pobox.com>
 * Copyright 2012-2014 pooler      <pooler@litecoinpool.org>
 * Copyright 2014      Lucas Jones <https://github.com/lucasjones>
 * Copyright 2014-2016 Wolf9466    <https://github.com/OhGodAPet>
 * Copyright 2016      Jay D Dee   <jayddee246@gmail.com>
 * Copyright 2016-2017 XMRig       <support@xmrig.com>
 *
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program. If not, see <http://www.gnu.org/licenses/>.
 */


#include "cryptonight.h"

#if defined(XMRIG_ARM)
#   include "cryptonight_arm.h"
#else
#   include "cryptonight_x86.h"
#endif

#include "cryptonight_test.h"

#include <stdio.h>

static void cryptonight_av1_aesni(const void *input, size_t size, const void *output, struct cryptonight_ctx *ctx) {
#   if !defined(XMRIG_ARMv7)
    arch_cryptonight_hash(input, size, output, ctx);
#   endif
}


static void cryptonight_av2_aesni_double(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
#   if !defined(XMRIG_ARMv7)
    arch_cryptonight_double_hash(input, size, output, ctx);
#   endif
}


static void cryptonight_av3_softaes(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
    arch_cryptonight_hash(input, size, output, ctx);
}


static void cryptonight_av4_softaes_double(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
    arch_cryptonight_double_hash(input, size, output, ctx);
}


#ifndef XMRIG_NO_AEON
static void cryptonight_lite_av1_aesni(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
    #   if !defined(XMRIG_ARMv7)
    arch_cryptonight_hash(input, size, output, ctx);
#endif
}


static void cryptonight_lite_av2_aesni_double(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
#   if !defined(XMRIG_ARMv7)
    arch_cryptonight_double_hash(input, size, output, ctx);
#   endif
}


static void cryptonight_lite_av3_softaes(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
    arch_cryptonight_hash(input, size, output, ctx);
}


static void cryptonight_lite_av4_softaes_double(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) {
    arch_cryptonight_double_hash(input, size, output, ctx);
}

void (*cryptonight_variations[8])(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) = {
            cryptonight_av1_aesni,
            cryptonight_av2_aesni_double,
            cryptonight_av3_softaes,
            cryptonight_av4_softaes_double,
            cryptonight_lite_av1_aesni,
            cryptonight_lite_av2_aesni_double,
            cryptonight_lite_av3_softaes,
            cryptonight_lite_av4_softaes_double
        };
#else
void (*cryptonight_variations[4])(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) = {
            cryptonight_av1_aesni,
            cryptonight_av2_aesni_double,
            cryptonight_av3_softaes,
            cryptonight_av4_softaes_double
        };
#endif

void (*cryptonight_hash_ctx)(const void *input, size_t size, const void *output, cryptonight_ctx *ctx) = cryptonight_av1_aesni;

bool xmrig_cryptonight_hash(const void *input, int size, const void *output, const void *target, cryptonight_ctx *ctx)
{
    cryptonight_hash_ctx(input, size, output, ctx);

    // TODO: Check this
    return (*(uint64_t *)(output + 24)) < (*(uint64_t *) target);
}


void xmrig_cryptonight_hash_void(const void *input, size_t size, const void *output, const void *target, cryptonight_ctx *ctx)
{
    cryptonight_hash_ctx(input, size, output, ctx);
}


static char *print_bin(char *dest, const void *buf, size_t len) {
    int offset = 0;
    for(int i = 0; i < len; i++) {
        offset += sprintf(dest+offset, "%d", (long) (*(unsigned char *) (buf+i)));
        if(i+1 != len) {
            offset += sprintf(dest+offset, " ");
        }
    }
    return dest;
}

int xmrig_self_test() {
    char output[64];

    struct cryptonight_ctx *ctx = (struct cryptonight_ctx*) _mm_malloc(sizeof(struct cryptonight_ctx), 16);
    ctx->memory = (uint8_t *) _mm_malloc(MEMORY * 2, 16);

    cryptonight_hash_ctx(test_input, 76, output, ctx);

    _mm_free(ctx->memory);
    _mm_free(ctx);

    int ret = memcmp(output, test_output0, 32) == 0;

    if(!ret) {
      char buf[256];
      memset(buf, 0, sizeof buf);
      print_bin(buf, test_output0, 32);
      printf("Failed self test.. Expected:\n%s\n", buf);

      memset(buf, 0, sizeof buf);
      print_bin(buf, output, 32);
      printf("Got:\n%s\n", buf);
    }
    return !ret;
}
