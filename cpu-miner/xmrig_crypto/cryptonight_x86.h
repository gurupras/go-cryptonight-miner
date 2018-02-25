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

#ifndef __CRYPTONIGHT_X86_H__
#define __CRYPTONIGHT_X86_H__


#ifdef __GNUC__
#   include <x86intrin.h>
#else
#   include <intrin.h>
#   define __restrict__ __restrict
#endif


#include "cryptonight.h"

#include "c_keccak.h"
#include "c_groestl.h"
#include "c_blake256.h"
#include "c_jh.h"
#include "c_skein.h"

extern size_t ITERATIONS;
extern size_t MEM;
extern size_t MASK;
extern bool SOFT_AES;

void arch_cryptonight_hash(const void *__restrict__ input, size_t size, const void *__restrict__ output, cryptonight_ctx *__restrict__ ctx);
void arch_cryptonight_double_hash(const void *__restrict__ input, size_t size, const void *__restrict__ output, cryptonight_ctx *__restrict__ ctx);
#endif /* __CRYPTONIGHT_X86_H__ */
