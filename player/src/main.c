#include <stdio.h>
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

Uint32 *convert_palette(Uint32 format, RVF_Color *srcpalette, int colors) {
    Uint32 *result = calloc(colors, sizeof(Uint32));
    SDL_PixelFormat *pixel_format = SDL_AllocFormat(format);
    for (int i = 0; i < colors; i++) {
        result[i] = SDL_MapRGBA(pixel_format, srcpalette[i].r, srcpalette[i].g, srcpalette[i].b, 255);
    }
    return result;
}

void main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("Use player <filename>");
        return;
    }

    video = rvf_open(argv[1]);
    if (video == NULL) {
        return;
    }

    printf("colors: %d\nframe size: %dx%d\nframes total:%d\nfps: %f\n",
           video->colors,
           video->width,
           video->height,
           video->length,
           1.0f / video->frame_time);

    SDL_Init(SDL_INIT_VIDEO);
    window = SDL_CreateWindow("RVF", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, screen_width, screen_height, SDL_WINDOW_RESIZABLE);
    renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
    screen = SDL_CreateTexture(renderer, SDL_PIXELFORMAT_RGBA8888, SDL_TEXTUREACCESS_STREAMING, 100, 100);

    Uint32 format;
    SDL_QueryTexture(screen, &format, NULL, NULL, NULL);
    palette = convert_palette(format, video->palette, video->colors);
    screen_rect.w = 100;
    screen_rect.h = 100;

    SDL_Event event;
    int working = 1;
    while (working) {
        while (SDL_PollEvent(&event)) {
            switch (event.type) {
                case SDL_QUIT:
                    working = 0;
                    break;
            }
        }
        SDL_RenderCopy(renderer, screen, NULL, &screen_rect);
        SDL_RenderPresent(renderer);
        SDL_Delay(5);
    }

    SDL_DestroyRenderer(renderer);
    SDL_DestroyWindow(window);
    SDL_Quit();
    rvf_close(&video);
}