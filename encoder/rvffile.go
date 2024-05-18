package main

import (
	"encoding/binary"
	"io"
	"os"
)

type RVFfile struct {
	file *os.File
}

const (
	CompressionNone uint8 = 0b00000000
	CompressionFull uint8 = 0b00000001
	FrameRegular    uint8 = 0b00000000
	FrameIsKeyframe uint8 = 0b00000001
	FrameIsFirst    uint8 = 0b00000010
	FrameIsLast     uint8 = 0b00000100
)

var magic = [4]byte{'R', 'V', 'F', 2}

func write(file io.Writer, data interface{}) {
	binary.Write(file, binary.LittleEndian, data)
}

func NewRVFfile(filename string, palette Palette, width int, height int, frames int, frameRate float32, flags uint8) *RVFfile {
	result := &RVFfile{}
	var err error
	result.file, err = os.Create(filename)
	if err != nil {
		panic(err)
	}

	// Magic
	result.file.Write(magic[:])

	// Header
	write(result.file, uint32(width))
	write(result.file, uint32(height))
	write(result.file, uint32(frames))
	write(result.file, float32(1/frameRate))
	write(result.file, flags)

	//Palette
	write(result.file, uint8(palette.Len()-1))
	for _, color := range palette {
		write(result.file, uint8(color.R))
		write(result.file, uint8(color.G))
		write(result.file, uint8(color.B))
	}

	return result
}

func (rvf *RVFfile) WriteRaw(data []int) {
	for _, item := range data {
		write(rvf.file, byte(item))
	}
}

func (rvf *RVFfile) WriteCompressed(data []byte, flags uint8) {
	frameSize := len(data) + 1 + 4
	write(rvf.file, uint32(frameSize))
	write(rvf.file, flags)
	rvf.file.Write(data)
	write(rvf.file, uint32(frameSize))
}

func (rvf *RVFfile) Close() {
	rvf.file.Close()
}
