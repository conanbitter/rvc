#ifndef _DECODING_H
#define _DECODING_H

#include <stdint.h>
#include <stdio.h>

typedef uint8_t Block[16];

typedef struct PaletteCache {
    uint8_t* pals;
    int count;
    int head;
    int colors;
} PaletteCache;

typedef struct Decoder {
    int width;
    int height;
    uint8_t* buffer;
    size_t buffer_size;
    size_t buffer_capacity;
    Block* blocks;
    Block* last_blocks;
    size_t block_data_size;
    int blocks_width;
    int blocks_height;
    int* curve;
    PaletteCache cache[3];
} Decoder;

Decoder* dec_new(int frame_width, int frame_height);
void dec_free(Decoder** dec);
void dec_decode(Decoder* dec, FILE* file, uint32_t length, uint8_t* dest, int debug);

#endif