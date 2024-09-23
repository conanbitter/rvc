package main

import (
	"image"
	"image/png"
	"os"
)

type Point struct {
	X int
	Y int
}

var InitPoints = [4]Point{
	{X: 0, Y: 0},
	{X: 0, Y: 1},
	{X: 1, Y: 1},
	{X: 1, Y: 0},
}

func hindex2xy(hindex int, n int) Point {
	p := InitPoints[hindex&0b11]
	hindex >>= 2

	for i := 4; i <= n; i *= 2 {
		i2 := i / 2
		switch hindex & 0b11 {
		case 0:
			p = Point{X: p.Y, Y: p.X}
		case 1:
			p.Y += i2
		case 2:
			p.X += i2
			p.Y += i2
		case 3:
			p = Point{
				X: i2 - 1 - p.Y + i2,
				Y: i2 - 1 - p.X,
			}
		}
		hindex >>= 2
	}
	return p
}

func GetHilbertCurve(width int, height int) []int {
	var size int
	if width > height {
		size = width
	} else {
		size = height
	}
	n := 1
	for n < size {
		n *= 2
	}
	size = n
	offsetx := (size - width) / 2
	offsety := (size - height) / 2
	//fmt.Println(size, offsetx, offsety)

	curveInd := 0

	result := make([]int, width*height)
	for i := range result {
		var p Point
		for {
			p = hindex2xy(curveInd, size)
			curveInd++
			if (p.X >= offsetx &&
				p.X < offsetx+width &&
				p.Y >= offsety &&
				p.Y < offsety+height) ||
				curveInd >= size*size {
				break
			}
		}
		result[i] = p.X - offsetx + (p.Y-offsety)*width
		//result[i] = p.X + p.Y*width
	}
	return result
}

func ApplyCurve(blocks []ImageBlock, curve []int) []ImageBlock {
	result := make([]ImageBlock, len(blocks))
	for i, n := range curve {
		result[i] = blocks[n]
	}
	return result
}

func UnwrapCurve(blocks []ImageBlock, curve []int) []ImageBlock {
	result := make([]ImageBlock, len(blocks))
	for i, n := range curve {
		result[n] = blocks[i]
	}
	return result
}

func drawLine(x1 int, y1 int, x2 int, y2 int, image *image.Gray) {
	var (
		p1 int
		p2 int
	)
	if x1 == x2 {
		if y1 < y2 {
			p1 = y1
			p2 = y2
		} else {
			p1 = y2
			p2 = y1
		}
		for i := p1; i <= p2; i++ {
			image.Pix[image.PixOffset(x1, i)] = 0
		}
	} else if y1 == y2 {
		if x1 < x2 {
			p1 = x1
			p2 = x2
		} else {
			p1 = x2
			p2 = x1
		}
		for i := p1; i <= p2; i++ {
			image.Pix[image.PixOffset(i, y1)] = 0
		}
	}
}

func DebugDrawCurve(width int, height int, filename string) {
	imwidth := width * 3   //+ 10
	imheight := height * 3 //+ 10
	img := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{imwidth, imheight}})
	for i := 0; i < imwidth*imheight; i++ {
		img.Pix[i] = 255
	}

	curve := GetHilbertCurve(width, height)

	x0 := (curve[0]%width)*3 + 1
	y0 := (curve[0]/width)*3 + 1

	for _, p := range curve[1:] {
		x1 := (p%width)*3 + 1
		y1 := (p/width)*3 + 1
		drawLine(x0, y0, x1, y1, img)
		x0 = x1
		y0 = y1
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
