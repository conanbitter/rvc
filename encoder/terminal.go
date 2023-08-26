package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

type TermColor int

const (
	TermBlack   TermColor = 30
	TermRed     TermColor = 31
	TermGreen   TermColor = 32
	TermYellow  TermColor = 33
	TermBlue    TermColor = 34
	TermMagenta TermColor = 35
	TermCyan    TermColor = 36
	TermWhite   TermColor = 37
	TermDefault TermColor = 39
	TermReset   TermColor = 0
)

func termInit() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	fmt.Print("\033[?25l")
}

func termReset() {
	fmt.Print("\033[?25h")
}

func termSetTitle(title string) {
	fmt.Printf("\033]0;%s\033\\", title)
}

func termSetColor(color TermColor) {
	fmt.Printf("\033[%dm", color)
}
