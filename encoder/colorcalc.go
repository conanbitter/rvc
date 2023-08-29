package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type ColorPoint struct {
	color    FloatColor
	segment  int
	count    uint64
	distance float64
}

type ColorCalc struct {
	points    []ColorPoint
	centroids []FloatColor

	colors    int
	poinCount uint64

	totalDistance float64
	pointsChanged uint64

	workers     int
	pointRanges [][]ColorPoint

	bestError   float64
	bestPalette Palette
	bestAtt     int
	errors      []float64

	maxSteps   int
	maxAttempt int
}

func swapPoints(left, right *ColorPoint) {
	*left, *right = *right, *left
}

func NewColorCalc(colors int, steps int, attempts int) *ColorCalc {
	if colors > 256 {
		colors = 256
	}
	if colors < 1 {
		colors = 1
	}
	return &ColorCalc{colors: colors, maxSteps: steps, maxAttempt: attempts}
}

func (cc *ColorCalc) Input(images []string) {
	var cube [256][256][256]uint64
	fmt.Println("Loading images...")

	bar := progressbar.NewOptions(len(images),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowCount(),
		progressbar.OptionUseANSICodes(true))

	bar.Set(0)

	for i, imgName := range images {
		img, _, _, err := ImageLoad(imgName)
		if err != nil {
			panic(err)
		}
		for _, data := range img {
			cube[data.R][data.G][data.B]++
		}
		bar.Set(i + 1)
	}

	bar.Finish()

	colors_total := uint64(0)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if cube[r][g][b] > 0 {
					colors_total++
				}
			}
		}
	}

	fmt.Printf("\n\nTotal number of colors: %d\n", colors_total)
	if uint64(cc.colors) > colors_total {
		cc.colors = int(colors_total)
	}
	cc.poinCount = colors_total
	if colors_total == 0 {
		panic(errors.New("wrong input"))
	}

	cc.points = make([]ColorPoint, 0, colors_total)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if cube[r][g][b] > 0 {
					cc.points = append(cc.points, ColorPoint{
						color:    FloatColor{float64(r) / 255, float64(g) / 255, float64(b) / 255},
						segment:  0,
						count:    cube[r][g][b],
						distance: math.MaxFloat64})
				}
			}
		}
	}

	cc.workers = runtime.NumCPU()
	if cc.workers > 1 {
		cc.pointRanges = make([][]ColorPoint, cc.workers)
		rangeSize := len(cc.points) / cc.workers
		for i := 0; i < cc.workers-1; i++ {
			cc.pointRanges[i] = cc.points[i*rangeSize : (i+1)*rangeSize]
		}
		cc.pointRanges[cc.workers-1] = cc.points[(cc.workers-1)*rangeSize:]
	}
}

func (point *ColorPoint) pointDistance(center *ColorPoint) float64 {
	dist := point.color.Distance(center.color)
	if dist < point.distance {
		point.distance = dist
		return dist
	}
	return point.distance
}

func (cc *ColorCalc) initCentroids() {
	centInd := 0
	swapPoints(&cc.points[0], &cc.points[rand.Uint64()%cc.poinCount])
	for centInd < cc.colors-1 {
		var sum float64 = 0
		for i := uint64(centInd + 1); i < cc.poinCount; i++ {
			sum += cc.points[i].pointDistance(&cc.points[centInd])
		}
		rnd := rand.Float64() * sum
		centInd++
		sum = 0
		next := cc.poinCount - 1
		for i := uint64(centInd + 1); i < cc.poinCount; i++ {
			sum += cc.points[i].distance
			if sum > rnd {
				next = i
				break
			}
		}
		swapPoints(&cc.points[centInd], &cc.points[next])
	}

	cc.centroids = make([]FloatColor, cc.colors)
	for i := 0; i < cc.colors; i++ {
		cc.centroids[i] = cc.points[i].color
	}
}

func (cc *ColorCalc) calcCentroids() {
	//start := time.Now()
	newCentroids := make([]FloatColor, cc.colors)
	sizes := make([]uint64, cc.colors)
	for _, point := range cc.points {
		sizes[point.segment] += point.count
		c := &newCentroids[point.segment]
		c.R += point.color.R * float64(point.count)
		c.G += point.color.G * float64(point.count)
		c.B += point.color.B * float64(point.count)
	}
	cc.totalDistance = 0
	for i := range cc.centroids {
		if sizes[i] == 0 {
			continue
		}
		size := float64(sizes[i])
		newCentroids[i].R /= size
		newCentroids[i].G /= size
		newCentroids[i].B /= size
		cc.totalDistance += math.Sqrt(newCentroids[i].Distance(cc.centroids[i]))
		cc.centroids[i] = newCentroids[i]
	}
	//fmt.Printf("Centroids: %s   ", time.Since(start))
}

func (cc *ColorCalc) calcSegments() {
	var (
		mt sync.Mutex
		wg sync.WaitGroup
	)

	//start := time.Now()
	cc.pointsChanged = 0
	for _, task := range cc.pointRanges {
		wg.Add(1)
		go func(chunk []ColorPoint) {
			for i := range chunk {
				oldSeg := chunk[i].segment
				newSeg := oldSeg
				minDist := chunk[i].color.Distance(cc.centroids[oldSeg])
				for c := range cc.centroids {
					dist := chunk[i].color.Distance(cc.centroids[c])
					if dist < minDist {
						minDist = dist
						newSeg = c
					}
				}
				if oldSeg != newSeg {
					chunk[i].segment = newSeg
					mt.Lock()
					cc.pointsChanged++
					mt.Unlock()
				}
			}
			wg.Done()
		}(task)
	}
	wg.Wait()

	//fmt.Printf("SegmentsMt: %s\n", time.Since(start))
}

func formatTime(dur time.Duration) string {
	var result strings.Builder
	d := dur.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		fmt.Fprintf(&result, "%2d h ", int(h))
	} else {
		fmt.Fprint(&result, "     ")
	}
	if m > 0 {
		fmt.Fprintf(&result, "%2d m ", int(m))
	} else {
		fmt.Fprint(&result, "     ")
	}
	fmt.Fprintf(&result, "%2d s", int(s))
	return result.String()
}

func (cc *ColorCalc) printState(attempt int, step int, start time.Time) {
	elapsed := time.Since(start)
	remainingSteps := cc.maxSteps*cc.maxAttempt - step - (attempt-1)*cc.maxSteps
	remaining := elapsed * time.Duration(remainingSteps) / time.Duration(step+(attempt-1)*cc.maxSteps)

	termJumpUp(2)
	termClearLine()
	fmt.Printf("Attempt \033[33m%2d / %d\033[0m\t\tTotal distance \033[33m%-10.5g\033[0m\tElapsed   \033[33m%s\033[0m\n",
		attempt,
		cc.maxAttempt,
		cc.totalDistance,
		formatTime(elapsed))
	termClearLine()
	fmt.Printf("Step \033[33m%4d / %d\033[0m\tChanged points \033[33m%-10d\033[0m\tRemaining \033[33m%s\033[0m\n",
		step,
		cc.maxSteps,
		cc.pointsChanged,
		formatTime(remaining))
}

func (cc *ColorCalc) CalcError() float64 {
	score := float64(0)
	for _, point := range cc.points {
		score += math.Sqrt(point.color.Distance(cc.centroids[point.segment])) * float64(point.count)
	}
	return score
}

func (cc *ColorCalc) Run() {
	cc.errors = make([]float64, 0, cc.maxAttempt)
	fmt.Print("Calculating...\n\n\n")
	startTime := time.Now()
	for a := 1; a < cc.maxAttempt+1; a++ {
		cc.initCentroids()
		for i := 1; i < cc.maxSteps+1; i++ {
			cc.calcSegments()
			if cc.pointsChanged == 0 {
				cc.printState(a, i, startTime)
				break
			}
			cc.calcCentroids()
			cc.printState(a, i, startTime)
		}
		cc.calcSegments()
		colorErr := cc.CalcError()
		if a == 1 || colorErr < cc.bestError {
			cc.bestAtt = a
			cc.bestError = colorErr
			cc.bestPalette = cc.calcPalette()
		}
		cc.errors = append(cc.errors, colorErr)
	}
	fmt.Printf("\nMost successful attempt is %d\n", cc.bestAtt)
	fmt.Print(cc.errors)
}

func (km *ColorCalc) calcPalette() Palette {
	result := make(Palette, km.colors)
	for i, c := range km.centroids {
		result[i] = c.ToIntColor()
	}
	result.Sort()
	return result
}

func (km *ColorCalc) GetPalette() Palette {
	return km.bestPalette
}
