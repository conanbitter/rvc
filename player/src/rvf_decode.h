#ifndef _RVFDECODE_H
#define _RVFDECODE_H

#include <stdint.h>
#include <stdio.h>
#include "decoding.h"

#pragma pack(push, 1)
typedef struct RVF_Color {
    uint8_t r;
    uint8_t g;
    uint8_t b;
} RVF_Color;
#pragma pack(pop)

typedef struct RVF_Audio {
    int channels;
    unsigned int frequency;
    int bit_depth;
    char* buffer;
    size_t buffer_size;
} RVF_Audio;

typedef struct RVF_File {
    // Header
    int format_version;
    int width;
    int height;
    int colors;
    int length;
    // Other data
    int is_compressed;
    FILE* file;
    float frame_time;
    RVF_Color* palette;
    uint8_t* data;
    int current_frame;
    long frames_offset;
    int frame_size;
    Decoder* decoder;
    RVF_Audio* audio;
} RVF_File;

RVF_File* rvf_open(const char* filename);
void rvf_close(RVF_File** file);
uint8_t* rvf_next_frame(RVF_File* file);
void rvf_debug(int enabled);
// char* rvf_prev_frame(RVF_File* file);
//  char* rvf_seek(RVF_File* file, float seconds, int relative, int precise);
//  char* rvf_seek(RVF_File* file, int frames, int relative, int precise);
void rvf_free_audio_buffer(RVF_File* file);

#endif