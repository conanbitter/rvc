#include <stdio.h>
#include <windows.h>
#define SDL_MAIN_HANDLED
#include <SDL.h>
#include "rvf_decode.h"

SDL_Window *window;
SDL_Renderer *renderer;
SDL_Texture *screen;
SDL_Rect screen_rect = {0, 0, 0, 0};
RVF_File *video;

int screen_width = 100, screen_height = 100, screen_scale = 1;

Uint32 *palette;

int debug = 0;
Uint32 debug_palette[16];
#define DEBUG_COLOR_SKIP 0, 0, 0
#define DEBUG_COLOR_REPEAT 96, 125, 139
#define DEBUG_COLOR_SOLID 46, 125, 50
// #define DEBUG_COLOR_PAL2 192, 202, 51
// #define DEBUG_COLOR_PAL4 205, 220, 57
// #define DEBUG_COLOR_PAL8 212, 225, 87
#define DEBUG_COLOR_PAL2 30, 136, 229
#define DEBUG_COLOR_PAL4 33, 150, 243
#define DEBUG_COLOR_PAL8 66, 165, 245
#define DEBUG_COLOR_RAW 255, 138, 101

void init_debug_palette(Uint32 format) {
    SDL_PixelFormat *pixel_format = SDL_AllocFormat(format);

    debug_palette[0x0] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_SKIP, 255);
    debug_palette[0x1] = debug_palette[0x0];
    debug_palette[0x2] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_SOLID, 255);
    debug_palette[0x3] = debug_palette[0x2];
    debug_palette[0x4] = debug_palette[0x2];
    debug_palette[0x5] = debug_palette[0x2];
    debug_palette[0x6] = debug_palette[0x2];
    debug_palette[0x7] = debug_palette[0x2];
    debug_palette[0x8] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_PAL2, 255);
    debug_palette[0x9] = debug_palette[0x8];
    debug_palette[0xA] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_PAL4, 255);
    debug_palette[0xB] = debug_palette[0xA];
    debug_palette[0xC] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_PAL8, 255);
    debug_palette[0xD] = debug_palette[0xC];
    debug_palette[0xE] = SDL_MapRGBA(pixel_format, DEBUG_COLOR_RAW, 255);
    debug_palette[0xF] = debug_palette[0xE];

    SDL_FreeFormat(pixel_format);
}

Uint32 *convert_palette(Uint32 format, RVF_Color *srcpalette, int colors) {
    Uint32 *result = calloc(colors, sizeof(Uint32));
    SDL_PixelFormat *pixel_format = SDL_AllocFormat(format);
    for (int i = 0; i < colors; i++) {
        result[i] = SDL_MapRGBA(pixel_format, srcpalette[i].r, srcpalette[i].g, srcpalette[i].b, 255);
    }
    SDL_FreeFormat(pixel_format);
    return result;
}

void convert_frame(uint8_t *data, int width, int height) {
    Uint32 *pixels;
    int pitch;
    SDL_LockTexture(screen, NULL, &pixels, &pitch);
    pitch /= sizeof(Uint32);

    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            pixels[y * pitch + x] = palette[data[y * width + x]];
        }
    }

    SDL_UnlockTexture(screen);
}

void convert_frame_debug(uint8_t *data, int width, int height) {
    Uint32 *pixels;
    int pitch;
    SDL_LockTexture(screen, NULL, &pixels, &pitch);
    pitch /= sizeof(Uint32);

    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            pixels[y * pitch + x] = debug_palette[data[y * width + x]];
        }
    }

    SDL_UnlockTexture(screen);
}

void adjust_frame(int width, int height) {
    float videoAR = (float)video->width / (float)video->height;
    float windowAR = (float)width / (float)height;
    if (videoAR < windowAR) {
        screen_rect.h = height;
        screen_rect.y = 0;
        screen_rect.w = height * videoAR;
        screen_rect.x = (width - screen_rect.w) / 2;
    } else {
        screen_rect.w = width;
        screen_rect.x = 0;
        screen_rect.h = width / videoAR;
        screen_rect.y = (height - screen_rect.h) / 2;
    }
}

void set_scale(int scale) {
    SDL_SetWindowSize(window, video->width * scale, video->height * scale);
    SDL_SetWindowPosition(window, SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED);
    screen_rect.x = 0;
    screen_rect.y = 0;
    screen_rect.w = video->width * scale;
    screen_rect.h = video->height * scale;
}

SDL_AudioDeviceID init_audio(RVF_Audio *audio) {
    int samples = audio->buffer_size / audio->channels;
    if (audio->bit_depth == 16) {
        samples /= 2;
    }
    SDL_AudioSpec spec = {
        .freq = audio->frequency,
        .format = audio->bit_depth == 16 ? AUDIO_S16 : AUDIO_U8,
        .channels = audio->channels,
        .samples = samples,
    };
    SDL_AudioDeviceID dev = SDL_OpenAudioDevice(NULL, 0, &spec, NULL, SDL_AUDIO_ALLOW_SAMPLES_CHANGE);
    if (dev != 0) {
        int success = SDL_QueueAudio(dev, audio->buffer, audio->buffer_size);
        if (success != 0) {
            return 0;
        }
    }
    return dev;
}

#ifdef NOCONSOLE
int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, PSTR lpCmdLine, int nCmdShow) {
    if (strlen(lpCmdLine) == 0) {
        SDL_ShowSimpleMessageBox(SDL_MESSAGEBOX_ERROR, "Error", "No file", NULL);
    }
    const char *filename = lpCmdLine;
#else
int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("Use player <filename>");
        return 0;
    }
    const char *filename = argv[1];
#endif

    video = rvf_open(filename);
    if (video == NULL) {
        return 0;
    }

    printf("colors: %d\nframe size: %dx%d\nframes total:%d\nfps: %f\n",
           video->colors,
           video->width,
           video->height,
           video->length,
           1.0f / video->frame_time);

    screen_width = video->width;
    screen_height = video->height;

    Uint32 init_flags = SDL_INIT_VIDEO;
    if (video->audio) {
        init_flags |= SDL_INIT_AUDIO;
    }

    SDL_Init(init_flags);
    window = SDL_CreateWindow("RVF", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, screen_width, screen_height, SDL_WINDOW_RESIZABLE);
    renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
    screen = SDL_CreateTexture(renderer, SDL_PIXELFORMAT_RGBA8888, SDL_TEXTUREACCESS_STREAMING, video->width, video->height);

    Uint32 format;
    SDL_QueryTexture(screen, &format, NULL, NULL, NULL);
    palette = convert_palette(format, video->palette, video->colors);
    init_debug_palette(format);
    screen_rect.w = screen_width;
    screen_rect.h = screen_height;

    SDL_AudioDeviceID audio_dev = 0;
    if (video->audio) {
        audio_dev = init_audio(video->audio);
        rvf_free_audio_buffer(video);
    }

    SDL_Event event;
    int working = 1;

    uint8_t *data = rvf_next_frame(video);
    if (data == NULL) {
        working = 0;
    }
    convert_frame(data, video->width, video->height);

    if (audio_dev > 0) {
        SDL_PauseAudioDevice(audio_dev, 0);
    }

    float elapsed = 0;
    float freq = SDL_GetPerformanceFrequency();
    Uint64 last_timer = SDL_GetPerformanceCounter();

    while (working) {
        while (SDL_PollEvent(&event)) {
            switch (event.type) {
                case SDL_QUIT:
                    working = 0;
                    break;
                case SDL_WINDOWEVENT:
                    if (event.window.event == SDL_WINDOWEVENT_RESIZED) {
                        int window_width = event.window.data1;
                        int window_height = event.window.data2;
                        adjust_frame(window_width, window_height);
                    }
                    break;
                case SDL_KEYDOWN:
                    switch (event.key.keysym.sym) {
                        case SDLK_ESCAPE:
                            working = 0;
                            break;
                        case SDLK_1:
                            set_scale(1);
                            break;
                        case SDLK_2:
                            set_scale(2);
                            break;
                        case SDLK_3:
                            set_scale(3);
                            break;
                        case SDLK_4:
                            set_scale(4);
                            break;
                        case SDLK_5:
                            set_scale(5);
                            break;
                        case SDLK_6:
                            set_scale(6);
                            break;
                        case SDLK_d:
                            debug = !debug;
                            rvf_debug(debug);
                            break;
                    }
                    break;
            }
        }
        SDL_RenderCopy(renderer, screen, NULL, &screen_rect);
        SDL_RenderPresent(renderer);

        elapsed += (float)(SDL_GetPerformanceCounter() - last_timer) / freq;
        last_timer = SDL_GetPerformanceCounter();

        if (elapsed > video->frame_time) {
            uint8_t *data = rvf_next_frame(video);
            if (data == NULL) {
                working = 0;
            }
            if (debug) {
                convert_frame_debug(data, video->width, video->height);
            } else {
                convert_frame(data, video->width, video->height);
            }
            elapsed -= video->frame_time;
        }

        SDL_Delay(5);
    }
    SDL_CloseAudioDevice(audio_dev);
    SDL_DestroyRenderer(renderer);
    SDL_DestroyWindow(window);
    SDL_Quit();
    rvf_close(&video);
    return 0;
}