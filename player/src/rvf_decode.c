#include "rvf_decode.h"
#include <stdlib.h>

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

    return result;
}

void rvf_close(RVF_File** file) {
    fclose((*file)->file);
    free(*file);
    *file = NULL;
}