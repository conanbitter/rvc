package main

import (
	"encoding/gob"
	"os"
)

type Pattern struct {
	Width  int
	Height int
	Order  int
	Data   []int
}

var bayerMatrixBig = &Pattern{
	Width:  8,
	Height: 8,
	Order:  64,
	Data: []int{
		0, 32, 8, 40, 2, 34, 10, 42,
		48, 16, 56, 24, 50, 18, 58, 26,
		12, 44, 4, 36, 14, 46, 6, 38,
		60, 28, 52, 20, 62, 30, 54, 22,
		3, 35, 11, 43, 1, 33, 9, 41,
		51, 19, 59, 27, 49, 17, 57, 25,
		15, 47, 7, 39, 13, 45, 5, 37,
		63, 31, 55, 23, 61, 29, 53, 21},
}

var bayerMatrixSmall = &Pattern{
	Width:  4,
	Height: 4,
	Order:  16,
	Data: []int{
		0, 8, 2, 10,
		12, 4, 14, 6,
		3, 11, 1, 9,
		15, 7, 13, 5},
}

func NewPattern(width int, height int, order int) *Pattern {
	if order < 0 {
		order = width * height
	}
	return &Pattern{
		Width:  width,
		Height: height,
		Order:  order,
		Data:   make([]int, width*height)}
}

func LoadPattern(filename string) *Pattern {
	var result *Pattern
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	err = dec.Decode(result)
	if err != nil {
		panic(err)
	}
	return result
}

func (pattern *Pattern) Save(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(pattern)
	if err != nil {
		panic(err)
	}
}

func (pattern *Pattern) Reshape(width int, height int) *Pattern {
	result := NewPattern(width, height, pattern.Order)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idst := x + y*width
			isrc := (x % pattern.Width) + (y%pattern.Height)*pattern.Width
			result.Data[idst] = pattern.Data[isrc]
		}
	}
	return result
}
