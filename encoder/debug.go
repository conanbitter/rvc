package main

import (
	"encoding/binary"
	"fmt"
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

func EncBlockTest2(filename string) {
	fmt.Println("Loading palette")
	pal := PaletteLoad("../data/enctest/common.pal")
	fmt.Println("Loading image")
	img, imwidth, imheight := EncLoadRaw(filename)
	fmt.Println("Unwrapping")
	blocks, bw, bh := ImageToBlocks(img, imwidth, imheight)

	fmt.Println("Encoding")
	encoder := NewEncoder(pal)
	encoder.Encode(blocks)
	blocksRes := encoder.Decode()

	fmt.Println("Wrapping")
	outimg, outw, outh := BlocksToImage(blocksRes, bw, bh)
	fmt.Println("Saving image")
	ImageSave("../data/enctest/"+filename+"_enc.png", outimg, outw, outh, pal)
}
