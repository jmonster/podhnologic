package main

import (
	"fmt"

	"github.com/muesli/termenv"
)

// ANSI color codes - will be initialized based on terminal background
var (
	colorReset  string
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorPurple string
	colorCyan   string
	colorWhite  string
	colorBold   string
	// Memphis design inspired colors
	colorHotPink   string
	colorTurquoise string
	colorOrange    string
)

func init() {
	initColors()
}

// initColors detects terminal background and sets appropriate colors
func initColors() {
	colorReset = "\033[0m"
	colorBold = "\033[1m"

	// Detect if terminal has light background
	output := termenv.DefaultOutput()
	isLight := !output.HasDarkBackground()

	if isLight {
		// Light mode: Memphis design with darker, saturated colors
		colorRed = "\033[38;5;160m"       // Dark red
		colorGreen = "\033[38;5;28m"      // Dark green
		colorYellow = "\033[38;5;178m"    // Memphis yellow
		colorBlue = "\033[38;5;26m"       // Dark blue
		colorPurple = "\033[38;5;90m"     // Dark purple
		colorCyan = "\033[38;5;30m"       // Dark cyan
		colorWhite = "\033[38;5;240m"     // Dark gray
		colorHotPink = "\033[38;5;198m"   // Hot pink
		colorTurquoise = "\033[38;5;44m"  // Turquoise
		colorOrange = "\033[38;5;208m"    // Orange
	} else {
		// Dark mode: Memphis design with bright, vibrant colors
		colorRed = "\033[38;5;196m"       // Bright red
		colorGreen = "\033[38;5;46m"      // Neon green
		colorYellow = "\033[38;5;226m"    // Bright yellow
		colorBlue = "\033[38;5;39m"       // Bright blue
		colorPurple = "\033[38;5;135m"    // Bright purple
		colorCyan = "\033[38;5;51m"       // Bright cyan
		colorWhite = "\033[37m"           // White
		colorHotPink = "\033[38;5;213m"   // Hot pink
		colorTurquoise = "\033[38;5;51m"  // Turquoise
		colorOrange = "\033[38;5;214m"    // Orange
	}
}

func printBanner() {
	fmt.Println()

	// Title with Memphis design colors - bold, contrasting, playful
	fmt.Printf("%s%s", colorBold, colorHotPink)
	fmt.Println("  ██████╗  ██████╗ ██████╗ ██╗  ██╗███╗   ██╗ ██████╗ ██╗      ██████╗  ██████╗ ██╗ ██████╗")
	fmt.Printf("%s", colorTurquoise)
	fmt.Println("  ██╔══██╗██╔═══██╗██╔══██╗██║  ██║████╗  ██║██╔═══██╗██║     ██╔═══██╗██╔════╝ ██║██╔════╝")
	fmt.Printf("%s", colorYellow)
	fmt.Println("  ██████╔╝██║   ██║██║  ██║███████║██╔██╗ ██║██║   ██║██║     ██║   ██║██║  ███╗██║██║     ")
	fmt.Printf("%s", colorOrange)
	fmt.Println("  ██╔═══╝ ██║   ██║██║  ██║██╔══██║██║╚██╗██║██║   ██║██║     ██║   ██║██║   ██║██║██║     ")
	fmt.Printf("%s", colorHotPink)
	fmt.Println("  ██║     ╚██████╔╝██████╔╝██║  ██║██║ ╚████║╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║╚██████╗")
	fmt.Printf("%s", colorTurquoise)
	fmt.Println("  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝")
	fmt.Printf("%s", colorReset)
	fmt.Println()

	// Music wave with Memphis colors
	fmt.Printf("%s          %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s\n",
		colorReset,
		colorHotPink, colorReset, colorTurquoise, colorReset, colorYellow, colorReset,
		colorOrange, colorReset, colorHotPink, colorReset, colorTurquoise, colorReset,
		colorYellow, colorReset, colorOrange, colorReset, colorHotPink, colorReset,
		colorTurquoise, colorReset, colorYellow, colorReset, colorOrange, colorReset,
		colorHotPink, colorReset, colorTurquoise, colorReset)

	fmt.Println()
}

func printSuccess(message string) {
	fmt.Printf("%s%s✓%s %s\n", colorBold, colorGreen, colorReset, message)
}

func printError(message string) {
	fmt.Printf("%s%s✗%s %s\n", colorBold, colorRed, colorReset, message)
}

func printInfo(message string) {
	fmt.Printf("%s%s→%s %s\n", colorBold, colorCyan, colorReset, message)
}

func printWarning(message string) {
	fmt.Printf("%s%s⚠%s  %s\n", colorBold, colorYellow, colorReset, message)
}
