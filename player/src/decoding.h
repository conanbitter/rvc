#ifndef _DECODING_H
#define _DECODING_H

#include <stdint.h>

typedef uint8_t Block[16];

typedef struct Decoder {
    uint8_t* buffer;
    size_t buffer_size;
    size_t buffer_capacity;
    Block* blocks;
    Block* last_blocks;
    size_t block_data_size;
    int blocks_width;
    int blocks_height;
    int* curve;
} Decoder;

Decoder* dec_new(int frame_width, int frame_height);
void dec_free(Decoder** dec);

#endif