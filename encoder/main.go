package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/schollz/progressbar/v3"
)

var compressionLevels []float64 = []float64{
	0.0001,
	0.001,
	0.005,
	0.01,
	0.02,
	0.05,
	0.1,
	0.2,
	0.5,
	1.0,
}

func main() {
	/*
		//EncSaveRaw("12")
		//EncPreview("12")
		useHilbert := true
		debugOutput := true
		EncBlockTest2("01", useHilbert, debugOutput, nil)
		/*EncBlockTest2("02", useHilbert, debugOutput, nil)
		EncBlockTest2("03", useHilbert, debugOutput, nil)
		EncBlockTest2("04", useHilbert, debugOutput, nil)
		EncBlockTest2("05", useHilbert, debugOutput, nil)
		EncBlockTest2("06", useHilbert, debugOutput, nil)
		enc := EncBlockTest2("10", useHilbert, debugOutput, nil)
		EncBlockTest2("11", useHilbert, debugOutput, enc)
		EncBlockTest2("12", useHilbert, debugOutput, enc)
		//DebugDrawCurve(320, 240, "../data/enctest/hilbert.png")
		EncBlockTest3("01", useHilbert, nil)
		EncBlockTest3("02", useHilbert, nil)
		EncBlockTest3("03", useHilbert, nil)
		EncBlockTest3("04", useHilbert, nil)
		EncBlockTest3("05", useHilbert, nil)
		EncBlockTest3("06", useHilbert, nil)
		enc := EncBlockTest3("10", useHilbert, nil)
		EncBlockTest3("11", useHilbert, enc)
		EncBlockTest3("12", useHilbert, enc)
		os.Exit(0)*/

	if len(os.Args) <= 1 {
		fmt.Println("Usage: rvc <command> <arguments> <input>")
		return
	}

	fc, err := os.Create("profiling/cpuprof.profile")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer fc.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(fc); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	command := os.Args[1]

	flags := flag.NewFlagSet("", flag.ExitOnError)
	var (
		argOutput      string
		argPalFrom     string
		argPalSave     string
		argFrameRate   float64
		argDithering   string
		argCompression int
		argAudio       string
	)

	flags.StringVar(&argOutput, "o", "", "output file")
	flags.StringVar(&argOutput, "output", "", "output file")
	flags.StringVar(&argPalFrom, "pf", "", "loading palette file")
	flags.StringVar(&argPalFrom, "pal-from", "", "loading palette file")
	flags.StringVar(&argPalSave, "ps", "", "saving palette file")
	flags.StringVar(&argPalSave, "pal-save", "", "saving palette file")
	flags.Float64Var(&argFrameRate, "fr", 30.0, "video frame rate")
	flags.Float64Var(&argFrameRate, "frame-rate", 30.0, "video frame rate")
	flags.StringVar(&argDithering, "d", "none", "dithering method")
	flags.StringVar(&argDithering, "dithering", "none", "dithering method")
	flags.IntVar(&argCompression, "c", 0, "compression level")
	flags.IntVar(&argCompression, "compression", 0, "compression level")
	flags.StringVar(&argAudio, "audio", "", "compression level")
	flags.StringVar(&argAudio, "a", "", "compression level")

	flags.Parse(os.Args[2:])
	argInput := flags.Args()
	if len(argInput) < 1 {
		fmt.Println("Usage: rvc <command> <arguments> <input>")
		return
	}
	argInputString := strings.Join(argInput, " ")

	termInit()
	defer termReset()
	termSetTitle("Retro Video Codec")

	termSetColor(TermBlue)

	fmt.Println(`
    8888888b.   888     888   .d8888b.
    888   Y88b  888     888  d88P  Y88b
    888    888  888     888  888    888
    888   d88P  Y88b   d88P  888
    8888888P"    Y88b d88P   888
    888 T88b      Y88o88P    888    888
    888  T88b      Y888P     Y88b  d88P
    888   T88b      Y8P       "Y8888P"

    R E T R O    V I D E O    C O D E C
 `)
	termSetColor(TermReset)

	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Input: \"%s\"\n", argInputString)
	fmt.Printf("Output: %s\n", argOutput)
	fmt.Printf("Input palette: %s\n", argPalFrom)
	fmt.Printf("Output palette: %s\n", argPalSave)
	fmt.Printf("Frame rate: %f\n", argFrameRate)
	fmt.Printf("Ditherig: %s\n", argDithering)
	fmt.Printf("Compression level: %d\n", argCompression)
	fmt.Printf("Audio: %s\n", argAudio)

	switch command {
	case "palette":
		if argOutput == "" {
			fmt.Println("Must specify output filename (-o, --output)")
		} else {
			files := listFiles(argInputString)
			if len(files) == 0 {
				fmt.Println("Can't find any files")
			} else {
				pal := CalcPalette(files)
				pal.Save(argOutput)
			}
		}
	case "encode":
		var audioFile *WAVfile = nil
		if argAudio != "" {
			audioFile = OpenWAV(argAudio)
		}
		if argCompression == 0 {
			RawEncode(argOutput,
				PaletteLoad(argPalFrom),
				listFiles(argInputString),
				float32(argFrameRate),
				FindDithering(argDithering),
				audioFile)
		} else {
			comp := argCompression
			if comp < 0 {
				comp = 0
			}
			if comp >= len(compressionLevels) {
				comp = len(compressionLevels) - 1
			}

			Encode(argOutput,
				PaletteLoad(argPalFrom),
				listFiles(argInputString),
				float32(argFrameRate),
				FindDithering(argDithering),
				compressionLevels[comp], //0.02
				audioFile)
		}
	case "preview":
		if argPalFrom == "" {
			fmt.Println("Must specify palette filename (-pf, --pal-from)")
		} else {
			files := listFiles(argInputString)
			if len(files) == 0 {
				fmt.Println("Can't find any files")
			} else {
				dithering := FindDithering(argDithering)
				if dithering == nil {
					fmt.Println("Wrong dithering method")
				} else {
					pal := PaletteLoad(argPalFrom)
					Preview(files, pal, dithering)
				}
			}
		}
	default:
		fmt.Printf("Unknown command \"%s\"\n", command)
	}

	fm, err := os.Create("profiling/memprof.profile")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer fm.Close() // error handling omitted for example
	runtime.GC()     // get up-to-date statistics
	if err := pprof.WriteHeapProfile(fm); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

}

func RawEncode(filename string, palette Palette, files []string, frameRate float32, dithering DitheringMethod, audio *WAVfile) {
	if len(files) == 0 {
		return
	}

	bar := progressbar.NewOptions(len(files),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowCount(),
		progressbar.OptionUseANSICodes(true))

	_, width, height, err := ImageLoad(files[0])
	if err != nil {
		panic(err)
	}

	palComp := NewPalComp(palette)

	dithering.Init(palette, palComp, width, height)

	rvf := NewRVFfile(filename, palette, width, height, len(files), frameRate, CompressionNone, audio)
	defer rvf.Close()

	bar.Set(0)

	for i, file := range files {
		imageColorData, fwidth, fheight, err := ImageLoad(file)
		if err != nil {
			panic(err)
		}
		if fwidth != width || fheight != height {
			panic(fmt.Errorf("frame size incorrect (%dx%d  must be %dx%d)", fwidth, fheight, width, height))
		}
		imageIndexData := dithering.Process(imageColorData, palette)
		rvf.WriteRaw(imageIndexData)
		bar.Set(i + 1)
	}

	bar.Finish()
}

func Encode(filename string, palette Palette, files []string, frameRate float32, dithering DitheringMethod, treshold float64, audio *WAVfile) {
	if len(files) == 0 {
		return
	}

	bar := progressbar.NewOptions(len(files),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowCount(),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("frames"),
	)

	bar.Set(0)

	_, width, height, err := ImageLoad(files[0])
	if err != nil {
		panic(err)
	}

	palComp := NewPalComp(palette)

	dithering.Init(palette, palComp, width, height)

	rvf := NewRVFfile(filename, palette, width, height, len(files), frameRate, CompressionFull, audio)
	defer rvf.Close()

	bw := int(math.Ceil(float64(width) / 4))
	bh := int(math.Ceil(float64(height) / 4))

	curve := GetHilbertCurve(bw, bh)
	encoder := NewEncoder(palette, palComp, treshold)

	totalSize := uint64(0)

	for i, file := range files {
		imageColorData, fwidth, fheight, err := ImageLoad(file)
		if err != nil {
			panic(err)
		}
		if fwidth != width || fheight != height {
			panic(fmt.Errorf("frame size incorrect (%dx%d  must be %dx%d)", fwidth, fheight, width, height))
		}
		imageIndexData := dithering.Process(imageColorData, palette)
		blocks, _, _ := ImageToBlocks(imageIndexData, fwidth, fheight)
		hblocks := ApplyCurve(blocks, curve)
		encoder.Encode(hblocks)
		packdata := encoder.Pack()
		flags := FrameRegular
		if i == 0 {
			flags |= FrameIsFirst
		}
		if i == len(files)-1 {
			flags |= FrameIsLast
		}
		if encoder.IsClean() {
			flags |= FrameIsKeyframe
		}
		rvf.WriteCompressed(packdata, flags)
		totalSize += uint64(len(packdata))

		bar.Set(i + 1)
	}

	compression := float64(totalSize) / float64(width*height*len(files)) * 100

	bar.Finish()
	fmt.Printf("\nCompression: %.f %%\nEncoding statistics:\n", compression)
	encoder.PrintStats()
}
