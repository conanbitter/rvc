package main

import (
	"encoding/binary"
	"io"
	"os"
)

type RVFfile struct {
	file *os.File
}

type FrameType uint8

const (
	FrameTypeRaw FrameType = iota
	FrameTypeEof
)

var magic = [4]byte{'R', 'V', 'F', 1}

func write(file io.Writer, data interface{}) {
	binary.Write(file, binary.LittleEndian, data)
}

func NewRVFfile(filename string, palette Palette, width int, height int, frames int, frameRate float32) *RVFfile {
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
	var datasize uint32 = uint32(len(data)) + 1 + 4
	write(rvf.file, datasize)
	write(rvf.file, FrameTypeRaw)
	for _, item := range data {
		write(rvf.file, byte(item))
	}
	write(rvf.file, datasize)
}

func (rvf *RVFfile) Close() {
	var datasize uint32 = 1 + 4
	write(rvf.file, datasize)
	write(rvf.file, FrameTypeEof)
	write(rvf.file, datasize)
	rvf.file.Close()
}
