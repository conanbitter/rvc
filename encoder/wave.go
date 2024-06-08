package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

type WAVfile struct {
	Cannels     int
	SampleRate  uint
	IsHiQuality bool
	Data        []byte
}

func readChunkName(file *os.File) string {
	chunkName := make([]byte, 4)
	file.Read(chunkName)
	return string(chunkName)
}

func assertChunkName(file *os.File, name string) {
	readedName := readChunkName(file)
	if readedName != name {
		panic(fmt.Errorf("wrong chunk name: \"%s\" (must be \"%s\")", readedName, name))
	}
}

func OpenWAV(filename string) *WAVfile {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	assertChunkName(file, "RIFF")
	var fileSize, fmtSize uint32
	binary.Read(file, binary.LittleEndian, &fileSize)
	fileSize += 8
	assertChunkName(file, "WAVE")
	assertChunkName(file, "fmt ")
	binary.Read(file, binary.LittleEndian, &fmtSize)
	var formatTag, numChannels, numBits uint16
	binary.Read(file, binary.LittleEndian, &formatTag)
	if formatTag != 1 {
		panic("File format must be PCM Integer")
	}
	binary.Read(file, binary.LittleEndian, &numChannels)
	if numChannels > 2 {
		panic("Number of channels must be 1 or 2")
	}
	var sampleRate uint32
	binary.Read(file, binary.LittleEndian, &sampleRate)
	if sampleRate < 8000 {
		panic("Sample rate must be at least 8 kHz")
	}
	file.Seek(6, 1)
	binary.Read(file, binary.LittleEndian, &numBits)
	if numBits != 8 && numBits != 16 {
		panic("Wrong bit depth (must be 8 or 16)")
	}
	if fmtSize > 16 {
		file.Seek(int64(fmtSize)-16, 1)
	}
	assertChunkName(file, "data")
	var dataSize uint32
	binary.Read(file, binary.LittleEndian, &dataSize)
	data := make([]byte, dataSize)
	file.Read(data)

	return &WAVfile{
		Cannels:     int(numChannels),
		SampleRate:  uint(sampleRate),
		IsHiQuality: numBits == 16,
		Data:        data,
	}
}
