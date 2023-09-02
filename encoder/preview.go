package main

import (
	"fmt"
	"path/filepath"

	"github.com/veandco/go-sdl2/sdl"
)

type Viewer struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	files     []string
	pal       Palette
	dithering ImageDithering
	current   int
	texture   *sdl.Texture
	rect      *sdl.Rect
	scale     int
	w         int
	h         int
}

func ViewerNew(files []string, palette Palette, dithering ImageDithering) (*Viewer, error) {
	result := &Viewer{files: files, pal: palette, dithering: dithering, current: 0, texture: nil, scale: 1}
	var err error
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}
	result.window, err = sdl.CreateWindow("Image Viewer", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 100, 100, sdl.WINDOW_SHOWN)
	if err != nil {
		sdl.Quit()
		return nil, err
	}
	result.renderer, err = sdl.CreateRenderer(result.window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		result.window.Destroy()
		sdl.Quit()
		return nil, err
	}
	return result, nil
}

func (v *Viewer) Free() {
	if v.texture != nil {
		v.texture.Destroy()
	}
	v.renderer.Destroy()
	v.window.Destroy()
	sdl.Quit()
}

func (v *Viewer) SetTitle(title string) {
	v.window.SetTitle(title)
}

func (v *Viewer) LoadImage(index int) {
	v.SetTitle(fmt.Sprintf("Loading \"%s\"...", filepath.Base(v.files[index])))
	var score1 float64 = 0
	var score2 float64 = 0
	if v.texture != nil {
		v.texture.Destroy()
	}
	imageColorData, width, height, err := ImageLoad(v.files[index])
	if err != nil {
		panic(err)
	}

	imageIndexData := v.dithering(imageColorData, v.pal, width, height)

	v.texture, err = v.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	if err != nil {
		panic(err)
	}
	outputData, pitch, err := v.texture.Lock(nil)
	if err != nil {
		panic(err)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := v.pal[imageIndexData[y*width+x]]
			score1 += color.ToFloatColor().Distance(imageColorData[y*width+x].ToFloatColor())
			score2 += color.ToFloatColor().Difference(imageColorData[y*width+x].ToFloatColor())
			outputData[y*pitch+x*4] = byte(color.R)
			outputData[y*pitch+x*4+1] = byte(color.G)
			outputData[y*pitch+x*4+2] = byte(color.B)
			outputData[y*pitch+x*4+3] = 255
		}
	}
	v.texture.Unlock()

	v.w = width
	v.h = height
	v.SetScale(v.scale)
	v.SetTitle(fmt.Sprintf("[%d/%d] %s (score: %f / %f )", index+1, len(v.files), filepath.Base(v.files[index]), score1, score2))
}

func (v *Viewer) SetScale(newScale int) {
	v.scale = newScale
	v.rect = &sdl.Rect{X: 0, Y: 0, W: int32(v.w * v.scale), H: int32(v.h * v.scale)}
	v.window.SetSize(int32(v.w*v.scale), int32(v.h*v.scale))
	v.window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
}

func (v *Viewer) Run() error {
	v.LoadImage(0)
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.State == sdl.PRESSED {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					case sdl.K_1:
						v.SetScale(1)
					case sdl.K_2:
						v.SetScale(2)
					case sdl.K_3:
						v.SetScale(3)
					case sdl.K_4:
						v.SetScale(4)
					case sdl.K_5:
						v.SetScale(5)
					case sdl.K_RIGHT:
						if (e.Keysym.Mod&sdl.KMOD_SHIFT > 0) && len(v.files) > 10 {
							v.current = (v.current + 10) % len(v.files)
						} else {
							v.current++
							if v.current >= len(v.files) {
								v.current = 0
							}
						}
						v.LoadImage(v.current)
					case sdl.K_LEFT:
						if (e.Keysym.Mod&sdl.KMOD_SHIFT > 0) && len(v.files) > 10 {
							v.current -= 10
							if v.current < 0 {
								v.current += len(v.files)
							}
						} else {
							v.current--
							if v.current < 0 {
								v.current = len(v.files) - 1
							}
						}
						v.LoadImage(v.current)
					}
				}
			}
		}
		v.renderer.Copy(v.texture, nil, v.rect)
		v.renderer.Present()
		sdl.Delay(5)
	}
	return nil
}

func Preview(files []string, palette Palette, dithering ImageDithering) {
	palette.Sort()
	viewer, err := ViewerNew(files, palette, dithering)
	if err != nil {
		panic(err)
	}
	defer viewer.Free()
	viewer.Run()
}
