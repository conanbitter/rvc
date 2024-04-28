package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func main() {
	//EncSaveRaw("12")
	//EncPreview("12")
	useHilbert := true
	debugOutput := true
	/*EncBlockTest2("01", useHilbert, debugOutput)
	EncBlockTest2("02", useHilbert, debugOutput)
	EncBlockTest2("03", useHilbert, debugOutput)
	EncBlockTest2("04", useHilbert, debugOutput)
	EncBlockTest2("05", useHilbert, debugOutput)
	EncBlockTest2("06", useHilbert, debugOutput)*/
	enc := EncBlockTest2("10", useHilbert, debugOutput, nil)
	EncBlockTest2("11", useHilbert, debugOutput, enc)
	EncBlockTest2("12", useHilbert, debugOutput, enc)
	//DebugDrawCurve(320, 240, "../data/enctest/hilbert.png")
	os.Exit(0)

	if len(os.Args) <= 1 {
		fmt.Println("Usage: rvc <command> <arguments> <input>")
		return
	}

	command := os.Args[1]

	flags := flag.NewFlagSet("", flag.ExitOnError)
	var (
		argOutput      string
		argPalFrom     string
		argPalSave     string
		argFrameRate   float64
		argDithering   string
		argCompression int
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
		RawEncode(argOutput, PaletteLoad(argPalFrom), listFiles(argInputString), float32(argFrameRate), FindDithering(argDithering))
	case "compress":
		fmt.Println("command \"compress\" is not implemented")
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
}

func RawEncode(filename string, palette Palette, files []string, frameRate float32, dithering ImageDithering) {
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

	rvf := NewRVFfile(filename, palette, width, height, len(files), frameRate)
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
		imageIndexData := dithering(imageColorData, palette, width, height)
		rvf.WriteRaw(imageIndexData)
		bar.Set(i + 1)
	}

	bar.Finish()
}
