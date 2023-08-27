package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
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
		files := listFiles(argInputString)
		fmt.Println(files)
		fmt.Println("command \"palette\" is not implemented")
	case "encode":
		fmt.Println("command \"encode\" is not implemented")
	case "compress":
		fmt.Println("command \"compress\" is not implemented")
	case "preview":
		fmt.Println("command \"preview\" is not implemented")
	default:
		fmt.Printf("Unknown command \"%s\"\n", command)
	}

	termClearLine()
	fmt.Printf("Attempt \033[33m%2d / %d\033[0m\t\tTotal distance \033[33m%-10.5g\033[0m\tElapsed \033[33m%s\033[0m\n", 1, 5, 6.44574, "1h 25m 34s")
	termClearLine()
	fmt.Printf("Step \033[33m%4d / %d\033[0m\tChanged points \033[33m%-10d\033[0m\tRemaining \033[33m%s\033[0m\n", 345, 1000, 10456, "2h 25m 32s")

	termJumpUp(2)
	termClearLine()
	fmt.Println("Test")
}
