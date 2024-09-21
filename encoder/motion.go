package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type Offset struct {
	X int
	Y int
}

var offsets = []Offset{
	{0, 0},
	{-1, -1},
	{0, -1},
	{1, -1},
	{1, 0},
	{1, 1},
	{0, 1},
	{-1, 1},
	{-1, 0},
	{-2, -2},
	{-1, -2},
	{0, -2},
	{1, -2},
	{2, -2},
	{2, -1},
	{2, 0},
	{2, 1},
	{2, 2},
	{1, 2},
	{0, 2},
	{-1, 2},
	{-2, 2},
	{-2, 1},
	{-2, 0},
	{-2, -1},
	{-3, -3},
	{-2, -3},
	{-1, -3},
	{0, -3},
	{1, -3},
	{2, -3},
	{3, -3},
	{3, -2},
	{3, -1},
	{3, 0},
	{3, 1},
	{3, 2},
	{3, 3},
	{2, 3},
	{1, 3},
	{0, 3},
	{-1, 3},
	{-2, 3},
	{-3, 3},
	{-3, 2},
	{-3, 1},
	{-3, 0},
	{-3, -1},
	{-3, -2},
	{-4, -4},
	{-3, -4},
	{-2, -4},
	{-1, -4},
	{0, -4},
	{1, -4},
	{2, -4},
	{3, -4},
	{4, -4},
	{4, -3},
	{4, -2},
	{4, -1},
	{4, 0},
	{4, 1},
	{4, 2},
	{4, 3},
	{4, 4},
	{3, 4},
	{2, 4},
	{1, 4},
	{0, 4},
	{-1, 4},
	{-2, 4},
	{-3, 4},
	{-4, 4},
	{-4, 3},
	{-4, 2},
	{-4, 1},
	{-4, 0},
	{-4, -1},
	{-4, -2},
	{-4, -3},
	{-5, -5},
	{-4, -5},
	{-3, -5},
	{-2, -5},
	{-1, -5},
	{0, -5},
	{1, -5},
	{2, -5},
	{3, -5},
	{4, -5},
	{5, -5},
	{5, -4},
	{5, -3},
	{5, -2},
	{5, -1},
	{5, 0},
	{5, 1},
	{5, 2},
	{5, 3},
	{5, 4},
	{5, 5},
	{4, 5},
	{3, 5},
	{2, 5},
	{1, 5},
	{0, 5},
	{-1, 5},
	{-2, 5},
	{-3, 5},
	{-4, 5},
	{-5, 5},
	{-5, 4},
	{-5, 3},
	{-5, 2},
	{-5, 1},
	{-5, 0},
	{-5, -1},
	{-5, -2},
	{-5, -3},
	{-5, -4},
	{-6, -6},
	{-5, -6},
	{-4, -6},
	{-3, -6},
	{-2, -6},
	{-1, -6},
	{0, -6},
	{1, -6},
	{2, -6},
	{3, -6},
	{4, -6},
	{5, -6},
	{6, -6},
	{6, -5},
	{6, -4},
	{6, -3},
	{6, -2},
	{6, -1},
	{6, 0},
	{6, 1},
	{6, 2},
	{6, 3},
	{6, 4},
	{6, 5},
	{6, 6},
	{5, 6},
	{4, 6},
	{3, 6},
	{2, 6},
	{1, 6},
	{0, 6},
	{-1, 6},
	{-2, 6},
	{-3, 6},
	{-4, 6},
	{-5, 6},
	{-6, 6},
	{-6, 5},
	{-6, 4},
	{-6, 3},
	{-6, 2},
	{-6, 1},
	{-6, 0},
	{-6, -1},
	{-6, -2},
	{-6, -3},
	{-6, -4},
	{-6, -5},
}

type Result struct {
	Cols    int         `json:"cols"`
	Rows    int         `json:"rows"`
	Vectors []BlockData `json:"vectors"`
}

type BlockData struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Error float64 `json:"err"`
}

func getXY(im []int, x int, y int, width int, height int) int {
	if x >= width {
		x = width - 1
	}
	if y >= height {
		y = height - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return im[x+y*width]
}

func compareRect(im1 []int, x1 int, y1 int, im2 []int, x2 int, y2 int, width int, height int, comp *PalComp) float64 {
	acc := float64(0)
	if x2 < 0 || x2+8 >= width || y2 < 0 || y2+8 >= height {
		return math.MaxFloat64
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			acc += comp.CompareColors(getXY(im1, x1+x, y1+y, width, height), getXY(im2, x2+x, y2+y, width, height))
		}
	}
	return acc / 4.0
}

func getMotion(i int) {
	pal := PaletteLoad(fmt.Sprintf("../data/motion/test%d.pal", i))
	imageColorData1, _, _, err := ImageLoad(fmt.Sprintf("../data/motion/test%d_1.png", i))
	if err != nil {
		panic(err)
	}
	imageColorData2, width, height, err := ImageLoad(fmt.Sprintf("../data/motion/test%d_2.png", i))
	if err != nil {
		panic(err)
	}
	palcomp := NewPalComp(pal)

	dithering := NewPatternDithering(bayerMatrixBig, -1, 0.5)
	dithering.Init(pal, palcomp, width, height)
	image1 := dithering.Process(imageColorData1, pal)
	image2 := dithering.Process(imageColorData2, pal)

	fmt.Printf("%d %d", len(image1), len(image2))

	rw := int(math.Ceil(float64(width) / 8))
	rh := int(math.Ceil(float64(height) / 8))

	result := Result{
		Rows:    rh,
		Cols:    rw,
		Vectors: make([]BlockData, 0),
	}

	for r := 0; r < rh; r++ {
		fmt.Printf("r:%d/%d\n", r, rh)
		for c := 0; c < rw; c++ {
			bestScore := math.MaxFloat64
			bestOffset := Offset{0, 0}
			for _, offset := range offsets {
				score := compareRect(image2, c*8, r*8, image1, c*8+offset.X, r*8+offset.Y, width, height, palcomp)
				if score < bestScore {
					bestScore = score
					bestOffset = offset
				}
			}
			result.Vectors = append(result.Vectors, BlockData{Error: bestScore, X: bestOffset.X, Y: bestOffset.Y})
		}
	}

	file, err := os.Create(fmt.Sprintf("../data/motion/test%d.json", i))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	if _, err = file.Write(data); err != nil {
		panic(err)
	}
}
