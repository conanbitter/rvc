package main

import (
	"fmt"
	"math"
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
	{-7, -7},
	{-6, -7},
	{-5, -7},
	{-4, -7},
	{-3, -7},
	{-2, -7},
	{-1, -7},
	{0, -7},
	{1, -7},
	{2, -7},
	{3, -7},
	{4, -7},
	{5, -7},
	{6, -7},
	{7, -7},
	{7, -6},
	{7, -5},
	{7, -4},
	{7, -3},
	{7, -2},
	{7, -1},
	{7, 0},
	{7, 1},
	{7, 2},
	{7, 3},
	{7, 4},
	{7, 5},
	{7, 6},
	{7, 7},
	{6, 7},
	{5, 7},
	{4, 7},
	{3, 7},
	{2, 7},
	{1, 7},
	{0, 7},
	{-1, 7},
	{-2, 7},
	{-3, 7},
	{-4, 7},
	{-5, 7},
	{-6, 7},
	{-7, 7},
	{-7, 6},
	{-7, 5},
	{-7, 4},
	{-7, 3},
	{-7, 2},
	{-7, 1},
	{-7, 0},
	{-7, -1},
	{-7, -2},
	{-7, -3},
	{-7, -4},
	{-7, -5},
	{-7, -6},
}

type Result struct {
	Cols    int            `json:"cols"`
	Rows    int            `json:"rows"`
	Vectors []MotionVector `json:"vectors"`
}

type MotionVector struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Error  float64 `json:"err"`
	Result *ImageBlock
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

func copyRect(src []int, x1 int, y1 int, dst []int, x2 int, y2 int, width int, height int) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x2+x < 0 || x2+x >= width || y2+y < 0 || y2+y >= height {
				continue
			}
			dst[x2+x+width*(y2+y)] = src[x1+x+width*(y1+y)]
		}
	}
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
		Vectors: make([]MotionVector, 0),
	}

	for r := 0; r < rh; r++ {
		fmt.Printf("\r calc r:%d/%d", r, rh)
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
			result.Vectors = append(result.Vectors, MotionVector{Error: bestScore, X: bestOffset.X, Y: bestOffset.Y})
		}
	}

	/*file, err := os.Create(fmt.Sprintf("../data/motion/test%d.json", i))
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
	}*/
	resImage := make([]int, len(image1))
	moved := 0
	for r := 0; r < rh; r++ {
		fmt.Printf("\r draw r:%d/%d", r, rh)
		for c := 0; c < rw; c++ {
			offset := result.Vectors[c+r*rw]
			cx := c * 8
			cy := r * 8
			//fmt.Printf("= cx=%d cy=%d dx=%d dy=%d id=%d is=%d\n", cx, cy, offset.X, offset.Y, cx+7+(cy+7)*width, cx+7+offset.X+(cy+7+offset.Y)*width)
			if offset.Error < 0.1 {
				copyRect(image1, cx+offset.X, cy+offset.Y, resImage, cx, cy, width, height)
				moved++
			} else {
				copyRect(image2, cx, cy, resImage, cx, cy, width, height)
			}
		}
	}
	fmt.Printf("  Moved: %d%%\n", int(math.Round(float64(moved)/float64(rw*rh)*100.0)))
	ImageSave(fmt.Sprintf("../data/motion/test%d_res1.png", i), image2, width, height, pal)
	ImageSave(fmt.Sprintf("../data/motion/test%d_res2.png", i), resImage, width, height, pal)
}

func compareBlock(block ImageBlock, image []int, x int, y int, width int, height int, comp *PalComp) float64 {
	if x < 0 || x+4 >= width || y < 0 || y+4 >= height {
		return math.MaxFloat64
	}
	acc := float64(0)
	for iy := 0; iy < 4; iy++ {
		for ix := 0; ix < 4; ix++ {
			blockIndex := ix + 4*iy
			mx := ix + x
			my := iy + y
			if mx < 0 {
				mx = 0
			}
			if mx >= width {
				mx = width - 1
			}
			if my < 0 {
				my = 0
			}
			if my >= height {
				my = height - 1
			}
			imageIndex := mx + width*(my)
			acc += comp.CompareColors(block[blockIndex], image[imageIndex])
		}
	}
	return acc
}

func copyBlock(block *ImageBlock, image []int, x int, y int, width int, height int) {
	if x < 0 || x+4 >= width || y < 0 || y+4 >= height {
		return
	}
	for iy := 0; iy < 4; iy++ {
		for ix := 0; ix < 4; ix++ {
			blockIndex := ix + 4*iy
			mx := ix + x
			my := iy + y
			if mx < 0 {
				mx = 0
			}
			if mx >= width {
				mx = width - 1
			}
			if my < 0 {
				my = 0
			}
			if my >= height {
				my = height - 1
			}
			imageIndex := mx + width*(my)
			block[blockIndex] = image[imageIndex]
		}
	}
}

func findBestOffset(block ImageBlock, bx int, by int, image []int, width int, height int, comp *PalComp) MotionVector {
	bestScore := math.MaxFloat64
	bestOffset := Offset{0, 0}
	bestBlock := ImageBlock{}
	for _, offset := range offsets {
		score := compareBlock(block, image, bx*4+offset.X, by*4+offset.Y, width, height, comp)
		if score < bestScore {
			bestScore = score
			bestOffset = offset
		}
	}
	copyBlock(&bestBlock, image, bx*4+bestOffset.X, by*4+bestOffset.Y, width, height)
	return MotionVector{
		X:      bestOffset.X,
		Y:      bestOffset.Y,
		Error:  bestScore,
		Result: &bestBlock,
	}
}

func CalculateMotionVectors(imageBlocks []ImageBlock, bw int, bh int, image []int, width int, height int, comp *PalComp) []MotionVector {
	result := make([]MotionVector, len(imageBlocks))

	if image == nil {
		for i := range result {
			result[i].Error = math.MaxFloat64
		}
		return result
	}

	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			index := bx + by*bw
			result[index] = findBestOffset(imageBlocks[index], bx, by, image, width, height, comp)
		}
	}
	return result
}

func ApplyCurveMotion(vectors []MotionVector, curve []int) []MotionVector {
	result := make([]MotionVector, len(vectors))
	for i, n := range curve {
		result[i] = vectors[n]
	}
	return result
}
