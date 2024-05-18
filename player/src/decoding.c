#include <stdlib.h>
#include "decoding.h"

typedef struct Point {
    int x;
    int y;
} Point;

const Point INIT_POINTS[4] = {
    {0, 0},
    {0, 1},
    {1, 1},
    {1, 0},
};

#define ENC_SKIP 0x00
#define ENC_SKIP_LONG 0x10
#define ENC_REPEAT 0x20
#define ENC_REPEAT_LONG 0x30
#define ENC_SOLID 0x40
#define ENC_SOLID_LONG 0x50
#define ENC_SOLID_SEP 0x60
#define ENC_SOLID_SEP_LONG 0x70
#define ENC_PAL2 0x80
#define ENC_PAL2_CACHE 0x90
#define ENC_PAL4 0xA0
#define ENC_PAL4_CACHE 0xB0
#define ENC_PAL8 0xC0
#define ENC_PAL8_CACHE 0xD0
#define ENC_RAW 0xE0
#define ENC_RAW_LONG 0xF0

static Point hindex2xy(int hindex, int n) {
    Point p = INIT_POINTS[hindex & 0b11];
    hindex >>= 2;
    for (int i = 4; i <= n; i *= 2) {
        int i2 = i / 2;
        switch (hindex & 0b11) {
            case 0: {
                int temp = p.x;
                p.x = p.y;
                p.y = temp;
            } break;
            case 1:
                p.y += i2;
                break;
            case 2:
                p.x += i2;
                p.y += i2;
                break;
            case 3: {
                int temp = p.x;
                p.x = i2 - 1 - p.y + i2;
                p.y = i2 - 1 - temp;
            } break;
        }
        hindex >>= 2;
    }
    return p;
}

static int* get_hilbert_curve(int width, int height) {
    int* curve = calloc(width * height, sizeof(int));

    int size;
    if (width > height) {
        size = width;
    } else {
        size = height;
    }
    int n = 1;
    while (n < size) {
        n *= 2;
    }
    size = n;
    int offsetx = (size - width) / 2;
    int offsety = (size - height) / 2;

    int curveInd = 0;

    for (int i = 0; i < width * height; i++) {
        Point p;
        while (1) {
            p = hindex2xy(curveInd, size);
            curveInd++;
            if ((p.x >= offsetx &&
                 p.x < offsetx + width &&
                 p.y >= offsety &&
                 p.y < offsety + height) ||
                curveInd >= size * size) {
                break;
            }
        }
        curve[p.x - offsetx + (p.y - offsety) * width] = i;
    }
    return curve;
}

static void unpack_bits2(uint8_t* src, uint8_t* dst) {
    dst[0] = (src[0] >> 7) & 0b1;
    dst[1] = (src[0] >> 6) & 0b1;
    dst[2] = (src[0] >> 5) & 0b1;
    dst[3] = (src[0] >> 4) & 0b1;
    dst[4] = (src[0] >> 3) & 0b1;
    dst[5] = (src[0] >> 2) & 0b1;
    dst[6] = (src[0] >> 1) & 0b1;
    dst[7] = src[0] & 0b1;
    dst[8] = (src[1] >> 7) & 0b1;
    dst[9] = (src[1] >> 6) & 0b1;
    dst[10] = (src[1] >> 5) & 0b1;
    dst[11] = (src[1] >> 4) & 0b1;
    dst[12] = (src[1] >> 3) & 0b1;
    dst[13] = (src[1] >> 2) & 0b1;
    dst[14] = (src[1] >> 1) & 0b1;
    dst[15] = src[1] & 0b1;
}

static void unpack_bits4(uint8_t* src, uint8_t* dst) {
    dst[0] = (src[0] >> 6) & 0b11;
    dst[1] = (src[0] >> 4) & 0b11;
    dst[2] = (src[0] >> 2) & 0b11;
    dst[3] = src[0] & 0b11;
    dst[4] = (src[1] >> 6) & 0b11;
    dst[5] = (src[1] >> 4) & 0b11;
    dst[6] = (src[1] >> 2) & 0b11;
    dst[7] = src[1] & 0b11;
    dst[8] = (src[2] >> 6) & 0b11;
    dst[9] = (src[2] >> 4) & 0b11;
    dst[10] = (src[2] >> 2) & 0b11;
    dst[11] = src[2] & 0b11;
    dst[12] = (src[3] >> 6) & 0b11;
    dst[13] = (src[4] >> 4) & 0b11;
    dst[14] = (src[4] >> 2) & 0b11;
    dst[15] = src[4] & 0b11;
}

static void unpack_bits8(uint8_t* src, uint8_t* dst) {
    dst[0] = (src[0] >> 5) & 0b111;
    dst[1] = (src[0] >> 2) & 0b111;
    dst[2] = (src[0] & 0b11) << 1 + (src[1] >> 7) & 0b1;
    dst[3] = (src[1] >> 4) & 0b111;
    dst[4] = (src[1] >> 1) & 0b111;
    dst[5] = (src[1] & 0b1) << 2 + (src[2] >> 6) & 0b11;
    dst[6] = (src[2] >> 3) & 0b111;
    dst[7] = src[2] & 0b111;
    dst[8] = (src[3] >> 5) & 0b111;
    dst[9] = (src[3] >> 2) & 0b111;
    dst[10] = (src[3] & 0b11) << 1 + (src[4] >> 7) & 0b1;
    dst[11] = (src[4] >> 4) & 0b111;
    dst[12] = (src[4] >> 1) & 0b111;
    dst[13] = (src[4] & 0b1) << 2 + (src[5] >> 6) & 0b11;
    dst[14] = (src[5] >> 3) & 0b111;
    dst[15] = src[5] & 0b111;
}

Decoder* dec_new(int frame_width, int frame_height) {
    Decoder* dec = malloc(sizeof(Decoder));
    dec->buffer = NULL;
    dec->buffer_size = 0;
    dec->buffer_capacity = 0;

    dec->blocks_width = ceil((float)frame_width / 4.0);
    dec->blocks_height = ceil((float)frame_height / 4.0);
    dec->blocks = calloc(dec->blocks_width * dec->blocks_height, sizeof(Block));
    // result->last_blocks = malloc(result->block_data_size);
    dec->curve = get_hilbert_curve(dec->blocks_width, dec->blocks_height);
    return dec;
}

void dec_free(Decoder** dec) {
    free((*dec)->buffer);
    free((*dec)->curve);
    free((*dec)->blocks);
    free(*dec);
    *dec = NULL;
}