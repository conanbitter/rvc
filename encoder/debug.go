package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
)

func EncSaveRaw(filename string) {
	pal := PaletteLoad("../data/enctest/common.pal")
	img, imwidth, imheight, _ := ImageLoad("../data/enctest/" + filename + ".png")
	dith := DitheringPattern4(img, pal, imwidth, imheight)
	file, _ := os.Create("../data/enctest/" + filename + ".raw")
	defer file.Close()
	binary.Write(file, binary.LittleEndian, uint16(imwidth))
	binary.Write(file, binary.LittleEndian, uint16(imheight))
	for _, item := range dith {
		write(file, byte(item))
	}
}

func EncLoadRaw(filename string) ([]int, int, int) {
	file, _ := os.Open("../data/enctest/" + filename + ".raw")
	defer file.Close()
	var width uint16
	var height uint16
	binary.Read(file, binary.LittleEndian, &width)
	binary.Read(file, binary.LittleEndian, &height)

	data := make([]int, int(width)*int(height))
	pixel := make([]byte, 1)
	for i := 0; i < int(width)*int(height); i++ {
		file.Read(pixel)
		data[i] = int(pixel[0])
	}
	return data, int(width), int(height)
}

func EncPreview(filename string) {
	pal := PaletteLoad("../data/enctest/common.pal")
	img, imwidth, imheight := EncLoadRaw(filename)
	ImageSave("../data/enctest/"+filename+"_prev.png", img, imwidth, imheight, pal)
}

func EncBlockTest(filename string) {
	pal := PaletteLoad("../data/enctest/common.pal")
	img, imwidth, imheight := EncLoadRaw(filename)
	blocks, bw, bh := ImageToBlocks(img, imwidth, imheight)
	outimg, outw, outh := BlocksToImage(blocks, bw, bh)
	ImageSave("../data/enctest/"+filename+"_blocks.png", outimg, outw, outh, pal)
}

func EncBlockTest2(filename string, useHilbert bool, debugOutput bool, useEncoder *FrameEncoder) *FrameEncoder {
	fmt.Println("Loading palette")
	pal := PaletteLoad("../data/enctest/common.pal")
	fmt.Println("Loading image")
	img, imwidth, imheight := EncLoadRaw(filename)
	fmt.Println("Unwrapping")
	blocks, bw, bh := ImageToBlocks(img, imwidth, imheight)
	var curve []int
	if useHilbert {
		fmt.Println("Applying Hilbert curve")
		curve = GetHilbertCurve(bw, bh)
		hblocks := make([]ImageBlock, len(blocks))
		for i, n := range curve {
			hblocks[i] = blocks[n]
		}
		blocks = hblocks
	}

	fmt.Println("Encoding")
	var encoder *FrameEncoder
	if useEncoder != nil {
		encoder = useEncoder
	} else {
		encoder = NewEncoder(pal)
	}
	lastFrame := encoder.lastFrame
	encoder.Encode(blocks)
	blocksRes := encoder.Decode(lastFrame)

	if useHilbert {
		fmt.Println("Applying inverse Hilbert curve")
		hblocksRes := make([]ImageBlock, len(blocksRes))
		for i, n := range curve {
			hblocksRes[n] = blocksRes[i]
		}
		blocksRes = hblocksRes
	}

	if debugOutput {
		debugImage := encoder.DebugDecode()
		if useHilbert {
			hdeb := make([]int, len(debugImage))
			for i, n := range curve {
				hdeb[n] = debugImage[i]
			}
			debugImage = hdeb
		}
		DebugDecodeSave(debugImage, bw, bh, "../data/enctest/"+filename+"_deb.png")

	}

	fmt.Println("Wrapping")
	fmt.Println(len(blocksRes), len(blocks))
	outimg, outw, outh := BlocksToImage(blocksRes, bw, bh)
	fmt.Println("Saving image")
	ImageSave("../data/enctest/"+filename+"_enc.png", outimg, outw, outh, pal)
	return encoder
}

var BlockColors = [][]byte{
	{128, 128, 128}, // SKIP       CONT.
	{128, 0, 0},     // REPEAT     CONT.
	{0, 128, 0},     // SOLID      CONT.
	{50, 128, 50},   // SOLID SEP  CONT.
	{128, 128, 0},   // PAL2       CONT.
	{64, 128, 0},    // PAL2 CACHE CONT.
	{128, 0, 128},   // PAL4       CONT.
	{64, 0, 128},    // PAL4 CACHE CONT.
	{0, 128, 128},   // PAL8       CONT.
	{0, 64, 128},    // PAL8 CACHE CONT.
	{0, 0, 128},     // RAW        CONT.
	{0, 0, 0},       // SKIP       FIRST
	{255, 0, 0},     // REPEAT     FIRST
	{0, 255, 0},     // SOLID      FIRST
	{100, 255, 100}, // SOLID SEP  FIRST
	{255, 255, 0},   // PAL2       FIRST
	{128, 255, 0},   // PAL2 CACHE FIRST
	{255, 0, 255},   // PAL4       FIRST
	{128, 0, 255},   // PAL4 CACHE FIRST
	{0, 255, 255},   // PAL8       FIRST
	{0, 128, 255},   // PAL8 CACHE FIRST
	{0, 0, 255},     // RAW        FIRST
}

func DebugDecodeSave(data []int, width int, height int, filename string) {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	ind := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := BlockColors[data[ind]]
			ind++
			offset := img.PixOffset(x, y)
			img.Pix[offset] = color[0]
			img.Pix[offset+1] = color[1]
			img.Pix[offset+2] = color[2]
			img.Pix[offset+3] = 255
		}
	}
	imgFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()
	if err = png.Encode(imgFile, img); err != nil {
		panic(err)
	}
}
