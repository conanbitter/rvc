package main

import (
	"image"
	"math"
	"os"
	"sort"

	_ "image/jpeg"
	"image/png"

	_ "golang.org/x/image/tiff"
)

type FloatColor struct {
	R, G, B float64
}

type IntColor struct {
	R, G, B int
}

type Palette []IntColor

func clipFloat(val float64) float64 {
	if val > 1.0 {
		return 1.0
	}
	if val < 0.0 {
		return 0.0
	}
	return val
}

func clipInt(val int) int {
	if val > 255 {
		return 255
	}
	if val < 0 {
		return 0
	}
	return val
}

//===== FLOAT COLOR =======

func (color FloatColor) Normalized() FloatColor {
	return FloatColor{
		clipFloat(color.R),
		clipFloat(color.G),
		clipFloat(color.B)}
}

func (color FloatColor) Luma() float64 {
	return 0.299*color.R + 0.587*color.G + 0.114*color.B
}

func (color FloatColor) Distance(other FloatColor) float64 {
	return math.Sqrt((color.R-other.R)*(color.R-other.R) +
		(color.G-other.G)*(color.G-other.G) +
		(color.B-other.B)*(color.B-other.B))
}

func (color FloatColor) Difference(other FloatColor) float64 {
	diffR := color.R - other.R
	diffG := color.G - other.G
	diffB := color.B - other.B
	diffcolor := FloatColor{R: diffR * diffR, G: diffG * diffG, B: diffB * diffB}
	diffluma := color.Luma() - other.Luma()
	return diffcolor.Luma()*0.75 + diffluma*diffluma
}

func (color FloatColor) DistanceSquared(other FloatColor) float64 {
	return (color.R-other.R)*(color.R-other.R) +
		(color.G-other.G)*(color.G-other.G) +
		(color.B-other.B)*(color.B-other.B)
}

func (color FloatColor) ToIntColor() IntColor {
	norm := color.Normalized()
	return IntColor{
		clipInt(int(math.Round(norm.R * 255))),
		clipInt(int(math.Round(norm.G * 255))),
		clipInt(int(math.Round(norm.B * 255)))}
}

//===== INT COLOR =======

func (color IntColor) Normalized() IntColor {
	return IntColor{
		clipInt(color.R),
		clipInt(color.G),
		clipInt(color.B)}
}

func (color IntColor) Luma() float64 {
	return 0.299*float64(color.R) + 0.587*float64(color.G) + 0.114*float64(color.B)
}

func (color IntColor) Distance(other IntColor) uint64 {
	r := uint64(color.R) - uint64(other.R)
	g := uint64(color.G) - uint64(other.G)
	b := uint64(color.B) - uint64(other.B)
	return r*r + g*g + b*b
}

func (color IntColor) ToFloatColor() FloatColor {
	return FloatColor{float64(color.R) / 255.0, float64(color.G) / 255.0, float64(color.B) / 255.0}
}

//===== PALETTE =======

func (pal Palette) Len() int {
	return len(pal)
}

func (pal Palette) Sort() {
	sort.Slice(pal, func(i, j int) bool { return pal[i].Luma() < pal[j].Luma() })
}

func (pal Palette) Save(filename string) {
	pal.Sort()
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	data := make([]byte, 0, pal.Len()*3)
	for _, color := range pal {
		data = append(data, byte(color.R), byte(color.G), byte(color.B))
	}
	_, err = fo.Write(data)
	if err != nil {
		panic(err)
	}

}

func (pal Palette) GetIntColorIndex(color IntColor) int {
	var minDist uint64 = math.MaxUint64
	minIndex := 0
	for i, other := range pal {
		dist := color.Distance(other)
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}
	return minIndex
}

func (pal Palette) GetFloatColorIndex(color FloatColor) int {
	var minDist float64 = math.MaxFloat64
	minIndex := 0
	for i, other := range pal {
		dist := color.Difference(other.ToFloatColor())
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}
	return minIndex
}

func PaletteLoad(filename string) Palette {
	fi, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	size := len(fi) / 3
	result := make(Palette, size)
	for i := range result {
		result[i].R = int(fi[i*3])
		result[i].G = int(fi[i*3+1])
		result[i].B = int(fi[i*3+2])
	}

	return result
}

func ImageLoad(filename string) ([]IntColor, int, int, error) {
	imgFile, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, 0, 0, err
	}
	bounds := img.Bounds()
	width := bounds.Size().X
	height := bounds.Size().Y
	result := make([]IntColor, 0, width*height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			result = append(result, IntColor{int(r / 257), int(g / 257), int(b / 257)}.Normalized())
		}
	}
	return result, width, height, nil
}

func ImageSave(filename string, data []int, width int, height int, pal Palette) error {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	ind := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			colorind := data[ind]
			color := pal[colorind]
			ind++
			offset := img.PixOffset(x, y)
			img.Pix[offset] = uint8(color.R)
			img.Pix[offset+1] = uint8(color.G)
			img.Pix[offset+2] = uint8(color.B)
			img.Pix[offset+3] = 255
		}
	}
	imgFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer imgFile.Close()
	if err = png.Encode(imgFile, img); err != nil {
		return err
	}
	return nil
}

func CalcPalette(input []string) Palette {
	calc := NewColorCalc(256, 1000, 5)
	calc.Input(input)
	calc.Run()
	return calc.GetPalette()
}
