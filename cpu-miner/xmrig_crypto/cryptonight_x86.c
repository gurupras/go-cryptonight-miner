#include "cryptonight_x86.h"
#include "soft_aes.h"

size_t ITERATIONS = 0x80000;
size_t MEM = MEMORY;
size_t MASK = 0x1FFFF0;
bool SOFT_AES = false;

static inline void do_blake_hash(const void* input, size_t len, char* output) {
    blake256_hash((uint8_t *) output, (uint8_t *) input, len);
}


static inline void do_groestl_hash(const void* input, size_t len, char* output) {
    groestl((uint8_t *) input, len * 8, (uint8_t *) output);
}


static inline void do_jh_hash(const void* input, size_t len, char* output) {
    jh_hash(32 * 8, (uint8_t *) input, 8 * len, (uint8_t *) output);
}


static inline void do_skein_hash(const void* input, size_t len, char* output) {
    xmr_skein((uint8_t *) input, (uint8_t *) output);
}


static void (* const extra_hashes[4])(const void *, size_t, char *) = {do_blake_hash, do_groestl_hash, do_jh_hash, do_skein_hash};



#if defined(__x86_64__) || defined(_M_AMD64)
#   define EXTRACT64(X) _mm_cvtsi128_si64(X)

#   ifdef __GNUC__
static inline uint64_t __umul128(uint64_t a, uint64_t b, uint64_t* hi)
{
    unsigned __int128 r = (unsigned __int128) a * (unsigned __int128) b;
    *hi = r >> 64;
    return (uint64_t) r;
}
#   else
    #define __umul128 _umul128
#   endif
#elif defined(__i386__) || defined(_M_IX86)
#   define HI32(X) \
    _mm_srli_si128((X), 4)


#   define EXTRACT64(X) \
    ((uint64_t)(uint32_t)_mm_cvtsi128_si32(X) | \
    ((uint64_t)(uint32_t)_mm_cvtsi128_si32(HI32(X)) << 32))

static inline uint64_t __umul128(uint64_t multiplier, uint64_t multiplicand, uint64_t *product_hi) {
    // multiplier   = ab = a * 2^32 + b
    // multiplicand = cd = c * 2^32 + d
    // ab * cd = a * c * 2^64 + (a * d + b * c) * 2^32 + b * d
    uint64_t a = multiplier >> 32;
    uint64_t b = multiplier & 0xFFFFFFFF;
    uint64_t c = multiplicand >> 32;
    uint64_t d = multiplicand & 0xFFFFFFFF;

    //uint64_t ac = a * c;
    uint64_t ad = a * d;
    //uint64_t bc = b * c;
    uint64_t bd = b * d;

    uint64_t adbc = ad + (b * c);
    uint64_t adbc_carry = adbc < ad ? 1 : 0;

    // multiplier * multiplicand = product_hi * 2^64 + product_lo
    uint64_t product_lo = bd + (adbc << 32);
    uint64_t product_lo_carry = product_lo < bd ? 1 : 0;
    *product_hi = (a * c) + (adbc >> 32) + (adbc_carry << 32) + product_lo_carry;

    return product_lo;
}
#endif


// This will shift and xor tmp1 into itself as 4 32-bit vals such as
// sl_xor(a1 a2 a3 a4) = a1 (a2^a1) (a3^a2^a1) (a4^a3^a2^a1)
static inline __m128i sl_xor(__m128i tmp1)
{
    __m128i tmp4;
    tmp4 = _mm_slli_si128(tmp1, 0x04);
    tmp1 = _mm_xor_si128(tmp1, tmp4);
    tmp4 = _mm_slli_si128(tmp4, 0x04);
    tmp1 = _mm_xor_si128(tmp1, tmp4);
    tmp4 = _mm_slli_si128(tmp4, 0x04);
    tmp1 = _mm_xor_si128(tmp1, tmp4);
    return tmp1;
}


static inline void aes_genkey_sub(__m128i* xout0, __m128i* xout2, uint8_t rcon)
{
    __m128i xout1 = _mm_aeskeygenassist_si128(*xout2, rcon);
    xout1  = _mm_shuffle_epi32(xout1, 0xFF); // see PSHUFD, set all elems to 4th elem
    *xout0 = sl_xor(*xout0);
    *xout0 = _mm_xor_si128(*xout0, xout1);
    xout1  = _mm_aeskeygenassist_si128(*xout0, 0x00);
    xout1  = _mm_shuffle_epi32(xout1, 0xAA); // see PSHUFD, set all elems to 3rd elem
    *xout2 = sl_xor(*xout2);
    *xout2 = _mm_xor_si128(*xout2, xout1);
}


static inline void soft_aes_genkey_sub(__m128i* xout0, __m128i* xout2, uint8_t rcon)
{
    __m128i xout1 = soft_aeskeygenassist(*xout2, rcon);
    xout1  = _mm_shuffle_epi32(xout1, 0xFF); // see PSHUFD, set all elems to 4th elem
    *xout0 = sl_xor(*xout0);
    *xout0 = _mm_xor_si128(*xout0, xout1);
    xout1  = soft_aeskeygenassist(*xout0, 0x00);
    xout1  = _mm_shuffle_epi32(xout1, 0xAA); // see PSHUFD, set all elems to 3rd elem
    *xout2 = sl_xor(*xout2);
    *xout2 = _mm_xor_si128(*xout2, xout1);
}


static inline void aes_genkey(const __m128i* memory, __m128i* k0, __m128i* k1, __m128i* k2, __m128i* k3, __m128i* k4, __m128i* k5, __m128i* k6, __m128i* k7, __m128i* k8, __m128i* k9, bool SOFT_AES)
{
    __m128i xout0 = _mm_load_si128(memory);
    __m128i xout2 = _mm_load_si128(memory + 1);
    *k0 = xout0;
    *k1 = xout2;

    SOFT_AES ? soft_aes_genkey_sub(&xout0, &xout2, 0x01) : aes_genkey_sub(&xout0, &xout2, 0x01);
    *k2 = xout0;
    *k3 = xout2;

    SOFT_AES ? soft_aes_genkey_sub(&xout0, &xout2, 0x02) : aes_genkey_sub(&xout0, &xout2, 0x02);
    *k4 = xout0;
    *k5 = xout2;

    SOFT_AES ? soft_aes_genkey_sub(&xout0, &xout2, 0x04) : aes_genkey_sub(&xout0, &xout2, 0x04);
    *k6 = xout0;
    *k7 = xout2;

    SOFT_AES ? soft_aes_genkey_sub(&xout0, &xout2, 0x08) : aes_genkey_sub(&xout0, &xout2, 0x08);
    *k8 = xout0;
    *k9 = xout2;
}


static inline void aes_round(__m128i key, __m128i* x0, __m128i* x1, __m128i* x2, __m128i* x3, __m128i* x4, __m128i* x5, __m128i* x6, __m128i* x7, bool SOFT_AES)
{
    if (SOFT_AES) {
        *x0 = soft_aesenc(*x0, key);
        *x1 = soft_aesenc(*x1, key);
        *x2 = soft_aesenc(*x2, key);
        *x3 = soft_aesenc(*x3, key);
        *x4 = soft_aesenc(*x4, key);
        *x5 = soft_aesenc(*x5, key);
        *x6 = soft_aesenc(*x6, key);
        *x7 = soft_aesenc(*x7, key);
    }
    else {
        *x0 = _mm_aesenc_si128(*x0, key);
        *x1 = _mm_aesenc_si128(*x1, key);
        *x2 = _mm_aesenc_si128(*x2, key);
        *x3 = _mm_aesenc_si128(*x3, key);
        *x4 = _mm_aesenc_si128(*x4, key);
        *x5 = _mm_aesenc_si128(*x5, key);
        *x6 = _mm_aesenc_si128(*x6, key);
        *x7 = _mm_aesenc_si128(*x7, key);
    }
}

static inline void cn_explode_scratchpad(const __m128i *input, __m128i *output)
{
    __m128i xin0, xin1, xin2, xin3, xin4, xin5, xin6, xin7;
    __m128i k0, k1, k2, k3, k4, k5, k6, k7, k8, k9;

    aes_genkey(input, &k0, &k1, &k2, &k3, &k4, &k5, &k6, &k7, &k8, &k9, SOFT_AES);

    xin0 = _mm_load_si128(input + 4);
    xin1 = _mm_load_si128(input + 5);
    xin2 = _mm_load_si128(input + 6);
    xin3 = _mm_load_si128(input + 7);
    xin4 = _mm_load_si128(input + 8);
    xin5 = _mm_load_si128(input + 9);
    xin6 = _mm_load_si128(input + 10);
    xin7 = _mm_load_si128(input + 11);

    for (size_t i = 0; i < MEM / sizeof(__m128i); i += 8) {
        aes_round(k0, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k1, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k2, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k3, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k4, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k5, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k6, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k7, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k8, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);
        aes_round(k9, &xin0, &xin1, &xin2, &xin3, &xin4, &xin5, &xin6, &xin7, SOFT_AES);

        _mm_store_si128(output + i + 0, xin0);
        _mm_store_si128(output + i + 1, xin1);
        _mm_store_si128(output + i + 2, xin2);
        _mm_store_si128(output + i + 3, xin3);
        _mm_store_si128(output + i + 4, xin4);
        _mm_store_si128(output + i + 5, xin5);
        _mm_store_si128(output + i + 6, xin6);
        _mm_store_si128(output + i + 7, xin7);
    }
}

static inline void cn_implode_scratchpad(const __m128i *input, __m128i *output)
{
    __m128i xout0, xout1, xout2, xout3, xout4, xout5, xout6, xout7;
    __m128i k0, k1, k2, k3, k4, k5, k6, k7, k8, k9;

    aes_genkey(output + 2, &k0, &k1, &k2, &k3, &k4, &k5, &k6, &k7, &k8, &k9, SOFT_AES);

    xout0 = _mm_load_si128(output + 4);
    xout1 = _mm_load_si128(output + 5);
    xout2 = _mm_load_si128(output + 6);
    xout3 = _mm_load_si128(output + 7);
    xout4 = _mm_load_si128(output + 8);
    xout5 = _mm_load_si128(output + 9);
    xout6 = _mm_load_si128(output + 10);
    xout7 = _mm_load_si128(output + 11);

    for (size_t i = 0; i < MEM / sizeof(__m128i); i += 8)
    {
        xout0 = _mm_xor_si128(_mm_load_si128(input + i + 0), xout0);
        xout1 = _mm_xor_si128(_mm_load_si128(input + i + 1), xout1);
        xout2 = _mm_xor_si128(_mm_load_si128(input + i + 2), xout2);
        xout3 = _mm_xor_si128(_mm_load_si128(input + i + 3), xout3);
        xout4 = _mm_xor_si128(_mm_load_si128(input + i + 4), xout4);
        xout5 = _mm_xor_si128(_mm_load_si128(input + i + 5), xout5);
        xout6 = _mm_xor_si128(_mm_load_si128(input + i + 6), xout6);
        xout7 = _mm_xor_si128(_mm_load_si128(input + i + 7), xout7);

        aes_round(k0, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k1, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k2, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k3, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k4, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k5, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k6, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k7, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k8, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
        aes_round(k9, &xout0, &xout1, &xout2, &xout3, &xout4, &xout5, &xout6, &xout7, SOFT_AES);
    }

    _mm_store_si128(output + 4, xout0);
    _mm_store_si128(output + 5, xout1);
    _mm_store_si128(output + 6, xout2);
    _mm_store_si128(output + 7, xout3);
    _mm_store_si128(output + 8, xout4);
    _mm_store_si128(output + 9, xout5);
    _mm_store_si128(output + 10, xout6);
    _mm_store_si128(output + 11, xout7);
}

inline void arch_cryptonight_hash(const void *__restrict__ input, size_t size, const void *__restrict__ output, cryptonight_ctx *__restrict__ ctx)
{
    keccak((uint8_t *) input, (int) size, ctx->state0, 200);

    cn_explode_scratchpad((__m128i*) ctx->state0, (__m128i*) ctx->memory);

    const uint8_t* l0 = ctx->memory;
    uint64_t* h0 = (uint64_t*) ctx->state0;

    uint64_t al0 = h0[0] ^ h0[4];
    uint64_t ah0 = h0[1] ^ h0[5];
    __m128i bx0 = _mm_set_epi64x(h0[3] ^ h0[7], h0[2] ^ h0[6]);

    uint64_t idx0 = h0[0] ^ h0[4];

    for (size_t i = 0; i < ITERATIONS; i++) {
        __m128i cx;
        cx = _mm_load_si128((__m128i *) &l0[idx0 & MASK]);

        if (SOFT_AES) {
            cx = soft_aesenc(cx, _mm_set_epi64x(ah0, al0));
        }
        else {
            cx = _mm_aesenc_si128(cx, _mm_set_epi64x(ah0, al0));
        }

        _mm_store_si128((__m128i *) &l0[idx0 & MASK], _mm_xor_si128(bx0, cx));
        idx0 = EXTRACT64(cx);
        bx0 = cx;

        uint64_t hi, lo, cl, ch;
        cl = ((uint64_t*) &l0[idx0 & MASK])[0];
        ch = ((uint64_t*) &l0[idx0 & MASK])[1];
        lo = __umul128(idx0, cl, &hi);

        al0 += hi;
        ah0 += lo;

        ((uint64_t*)&l0[idx0 & MASK])[0] = al0;
        ((uint64_t*)&l0[idx0 & MASK])[1] = ah0;

        ah0 ^= ch;
        al0 ^= cl;
        idx0 = al0;
    }

    cn_implode_scratchpad((__m128i*) ctx->memory, (__m128i*) ctx->state0);

    keccakf(h0, 24);
    extra_hashes[ctx->state0[0] & 3](ctx->state0, 200, (char*) output);
}


inline void arch_cryptonight_double_hash(const void *__restrict__ input, size_t size, const void *__restrict__ output, struct cryptonight_ctx *__restrict__ ctx)
{
    keccak((const uint8_t *) input,        (int) size, ctx->state0, 200);
    keccak((const uint8_t *) input + size, (int) size, ctx->state1, 200);

    const uint8_t* l0 = ctx->memory;
    const uint8_t* l1 = ctx->memory + MEM;
    uint64_t* h0 = (uint64_t*) ctx->state0;
    uint64_t* h1 = (uint64_t*) ctx->state1;

    cn_explode_scratchpad((__m128i*) h0, (__m128i*) l0);
    cn_explode_scratchpad((__m128i*) h1, (__m128i*) l1);

    uint64_t al0 = h0[0] ^ h0[4];
    uint64_t al1 = h1[0] ^ h1[4];
    uint64_t ah0 = h0[1] ^ h0[5];
    uint64_t ah1 = h1[1] ^ h1[5];

    __m128i bx0 = _mm_set_epi64x(h0[3] ^ h0[7], h0[2] ^ h0[6]);
    __m128i bx1 = _mm_set_epi64x(h1[3] ^ h1[7], h1[2] ^ h1[6]);

    uint64_t idx0 = h0[0] ^ h0[4];
    uint64_t idx1 = h1[0] ^ h1[4];

    for (size_t i = 0; i < ITERATIONS; i++) {
        __m128i cx0 = _mm_load_si128((__m128i *) &l0[idx0 & MASK]);
        __m128i cx1 = _mm_load_si128((__m128i *) &l1[idx1 & MASK]);

        if (SOFT_AES) {
            cx0 = soft_aesenc(cx0, _mm_set_epi64x(ah0, al0));
            cx1 = soft_aesenc(cx1, _mm_set_epi64x(ah1, al1));
        }
        else {
            cx0 = _mm_aesenc_si128(cx0, _mm_set_epi64x(ah0, al0));
            cx1 = _mm_aesenc_si128(cx1, _mm_set_epi64x(ah1, al1));
        }

        _mm_store_si128((__m128i *) &l0[idx0 & MASK], _mm_xor_si128(bx0, cx0));
        _mm_store_si128((__m128i *) &l1[idx1 & MASK], _mm_xor_si128(bx1, cx1));

        idx0 = EXTRACT64(cx0);
        idx1 = EXTRACT64(cx1);

        bx0 = cx0;
        bx1 = cx1;

        uint64_t hi, lo, cl, ch;
        cl = ((uint64_t*) &l0[idx0 & MASK])[0];
        ch = ((uint64_t*) &l0[idx0 & MASK])[1];
        lo = __umul128(idx0, cl, &hi);

        al0 += hi;
        ah0 += lo;

        ((uint64_t*) &l0[idx0 & MASK])[0] = al0;
        ((uint64_t*) &l0[idx0 & MASK])[1] = ah0;

        ah0 ^= ch;
        al0 ^= cl;
        idx0 = al0;

        cl = ((uint64_t*) &l1[idx1 & MASK])[0];
        ch = ((uint64_t*) &l1[idx1 & MASK])[1];
        lo = __umul128(idx1, cl, &hi);

        al1 += hi;
        ah1 += lo;

        ((uint64_t*) &l1[idx1 & MASK])[0] = al1;
        ((uint64_t*) &l1[idx1 & MASK])[1] = ah1;

        ah1 ^= ch;
        al1 ^= cl;
        idx1 = al1;
    }

    cn_implode_scratchpad((__m128i*) l0, (__m128i*) h0);
    cn_implode_scratchpad((__m128i*) l1, (__m128i*) h1);

    keccakf(h0, 24);
    keccakf(h1, 24);

    extra_hashes[ctx->state0[0] & 3](ctx->state0, 200, (char*) output);
    extra_hashes[ctx->state1[0] & 3](ctx->state1, 200, (char*) output + 32);
}
