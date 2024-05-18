#include "rvf_decode.h"
#include <stdlib.h>

#define COMPRESSION_NONE 0b00000000
#define COMPRESSION_FULL 0b00000001
#define FRAME_REGULAR 0b00000000
#define FRAME_IS_KEYFRAME 0b00000001
#define FRAME_IS_FIRST 0b00000010
#define FRAME_IS_LAS 0b00000100

RVF_File*
rvf_open(const char* filename) {
    RVF_File* result = malloc(sizeof(RVF_File));
    result->file = fopen(filename, "rb");

    uint32_t magic = 0;
    fread(&magic, 3, 1, result->file);
    if (magic != 0x465652) {
        printf("Wrong file format.");
        free(result);
        return NULL;
    }
    uint8_t version = 0;
    fread(&version, 1, 1, result->file);
    if (version != 2) {
        printf("Wrong file format version.");
        free(result);
        return NULL;
    }
    uint32_t width, height, length;
    float frame_time;
    uint8_t flags;
    fread(&width, 4, 1, result->file);
    fread(&height, 4, 1, result->file);
    fread(&length, 4, 1, result->file);
    fread(&frame_time, 4, 1, result->file);
    fread(&flags, 1, 1, result->file);
    result->width = width;
    result->height = height;
    result->length = length;
    result->frame_time = frame_time;
    result->is_compressed = (flags & COMPRESSION_FULL) > 0;

    uint8_t color_count;
    fread(&color_count, 1, 1, result->file);
    result->colors = (int)color_count + 1;

    result->palette = calloc(result->colors, sizeof(RVF_Color));
    fread(result->palette, sizeof(RVF_Color), result->colors, result->file);

    result->frames_offset = ftell(result->file);
    result->current_frame = -1;

    result->frame_size = result->width * result->height;
    result->data = malloc(result->frame_size);

    result->decoder = dec_new(result->width, result->height);
    return result;
}

void rvf_close(RVF_File** file) {
    fclose((*file)->file);
    dec_free(&((*file)->decoder));
    free((*file)->data);
    free((*file)->palette);
    free(*file);
    *file = NULL;
}

uint8_t* rvf_next_frame(RVF_File* file) {
    file->current_frame++;
    if (file->current_frame >= file->length) {
        file->current_frame = 0;
        fseek(file->file, file->frames_offset, SEEK_SET);
    }

    if (file->is_compressed) {
        uint32_t data_length;
        int flags;
        fread(&data_length, 4, 1, file->file);
        fread(&flags, 1, 1, file->file);
        dec_decode(file->decoder, file->file, data_length, file->data);
    } else {
        fread(file->data, file->frame_size, 1, file->file);
    }
    return file->data;
}