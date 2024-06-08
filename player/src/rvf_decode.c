#include "rvf_decode.h"
#include <stdlib.h>

#define COMPRESSION_NONE 0b00000000
#define COMPRESSION_FULL 0b00000001
#define AUDIO_BLOCK 0b00000010
#define AUDIO_STREAM 0b00000100
#define FRAME_REGULAR 0b00000000
#define FRAME_IS_KEYFRAME 0b00000001
#define FRAME_IS_FIRST 0b00000010
#define FRAME_IS_LAST 0b00000100

static int debug = 0;

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
    if (version != 3) {
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

    if (flags & AUDIO_BLOCK || flags & AUDIO_STREAM) {
        uint32_t frequency, buffer_size;
        uint8_t channels, quality;
        fread(&channels, 1, 1, result->file);
        fread(&frequency, 4, 1, result->file);
        fread(&quality, 1, 1, result->file);

        result->audio = malloc(sizeof(RVF_Audio));
        result->audio->channels = channels;
        result->audio->frequency = frequency;
        result->audio->bit_depth = quality ? 16 : 8;
    }

    uint8_t color_count;
    fread(&color_count, 1, 1, result->file);
    result->colors = (int)color_count + 1;

    result->palette = calloc(result->colors, sizeof(RVF_Color));
    fread(result->palette, sizeof(RVF_Color), result->colors, result->file);

    if (flags & AUDIO_BLOCK) {
        uint32_t buffer_size;
        fread(&buffer_size, 4, 1, result->file);
        result->audio->buffer_size = buffer_size;
        result->audio->buffer = malloc(buffer_size);
        fread(result->audio->buffer, buffer_size, 1, result->file);
    }

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
        uint8_t flags;
        fread(&data_length, 4, 1, file->file);
        fread(&flags, 1, 1, file->file);
        dec_decode(file->decoder, file->file, data_length - 4 - 1, file->data, debug);
        fseek(file->file, 4, SEEK_CUR);
    } else {
        fread(file->data, file->frame_size, 1, file->file);
    }
    return file->data;
}

void rvf_debug(int enabled) {
    debug = enabled;
}

void rvf_free_audio_buffer(RVF_File* file) {
    if (file->audio) {
        free(file->audio->buffer);
    }
}