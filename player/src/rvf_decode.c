#include "rvf_decode.h"
#include <stdlib.h>

#define FRAME_TYPE_RAW 0
#define FRAME_TYPE_EOF 1

RVF_File* rvf_open(const char* filename) {
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
    if (version != 1) {
        printf("Wrong file format version.");
        free(result);
        return NULL;
    }
    uint32_t width, height, length;
    float frame_time;
    fread(&width, 4, 1, result->file);
    fread(&height, 4, 1, result->file);
    fread(&length, 4, 1, result->file);
    fread(&frame_time, 4, 1, result->file);
    result->width = width;
    result->height = height;
    result->length = length;
    result->frame_time = frame_time;

    uint8_t color_count;
    fread(&color_count, 1, 1, result->file);
    result->colors = (int)color_count + 1;

    result->palette = calloc(result->colors, sizeof(RVF_Color));
    fread(result->palette, sizeof(RVF_Color), result->colors, result->file);

    result->frames_offset = ftell(result->file);

    result->data = malloc(result->width * result->height);
    return result;
}

void rvf_close(RVF_File** file) {
    fclose((*file)->file);
    free((*file)->data);
    free((*file)->palette);
    free(*file);
    *file = NULL;
}

uint8_t* rvf_next_frame(RVF_File* file) {
    file->current_frame++;
    uint32_t block_length;
    fread(&block_length, 4, 1, file->file);
    uint8_t block_type;
    fread(&block_type, 1, 1, file->file);
    if (block_type == FRAME_TYPE_EOF) {
        fseek(file->file, file->frames_offset, SEEK_SET);
        file->current_frame = 0;
        fread(&block_length, 4, 1, file->file);
        fread(&block_type, 1, 1, file->file);
    }
    if (block_type != FRAME_TYPE_RAW || block_length > (file->width * file->height + 1 + 4)) {
        printf("Wrong frame data");
        return NULL;
    }
    fread(file->data, block_length - 4 - 1, 1, file->file);
    uint32_t block_length_tail;
    fread(&block_length_tail, 4, 1, file->file);
    if (block_length != block_length_tail) {
        printf("Head and tail lengths mismatch");
        return NULL;
    }
    return file->data;
}