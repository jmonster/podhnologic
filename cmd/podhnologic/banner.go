package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/muesli/termenv"
)

const (
	appleBlack = "#000000"
	appleWhite = "#f4ffe8"

	applePhosphorDark   = "#0f7a00"
	applePhosphorMid    = "#22cc00"
	applePhosphorBright = "#7cff00"
	applePhosphorGlow   = "#b6ff00"
	applePhosphorDim    = "#4d8f00"

	appleRainbowGreen  = "#61bb46"
	appleRainbowYellow = "#fdb827"
	appleRainbowOrange = "#f5821f"
	appleRainbowRed    = "#e03a3e"
	appleRainbowBlue   = "#009ddc"
	appleGrayDim       = "#6f7f6c"
	appleGrayDark      = "#4f5d4c"

	appleLightPhosphorDark = "#006d1f"
	appleLightPhosphorMid  = "#178a00"
)

// ANSI color codes - will be initialized based on terminal background
var (
	colorReset   string
	colorRed     string
	colorGreen   string
	colorWarning string
	colorInfo    string
	colorBold    string
	// Banner accent colors
	colorPhosphorGlow   string
	colorPhosphorBright string
	colorPhosphorMid    string
	colorPhosphorDim    string
)

func ansiHex(color string) string {
	hex := strings.TrimPrefix(color, "#")
	if len(hex) != 6 {
		return ""
	}

	r, errR := strconv.ParseInt(hex[0:2], 16, 0)
	g, errG := strconv.ParseInt(hex[2:4], 16, 0)
	b, errB := strconv.ParseInt(hex[4:6], 16, 0)
	if errR != nil || errG != nil || errB != nil {
		return ""
	}

	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

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
		// Light mode: darker Apple green terminal colors
		colorRed = ansiHex(appleRainbowRed)
		colorGreen = ansiHex(appleLightPhosphorMid)
		colorWarning = ansiHex(appleRainbowOrange)
		colorInfo = ansiHex(appleRainbowBlue)
		colorPhosphorGlow = ansiHex(appleLightPhosphorMid)
		colorPhosphorBright = ansiHex(appleRainbowGreen)
		colorPhosphorMid = ansiHex(appleLightPhosphorDark)
		colorPhosphorDim = ansiHex(appleGrayDark)
	} else {
		// Dark mode: Apple II phosphor greens
		colorRed = ansiHex(appleRainbowRed)
		colorGreen = ansiHex(applePhosphorBright)
		colorWarning = ansiHex(appleRainbowOrange)
		colorInfo = ansiHex(appleRainbowBlue)
		colorPhosphorGlow = ansiHex(applePhosphorGlow)
		colorPhosphorBright = ansiHex(applePhosphorBright)
		colorPhosphorMid = ansiHex(applePhosphorMid)
		colorPhosphorDim = ansiHex(applePhosphorDim)
	}
}

func printBanner() {
	fmt.Println()

	// Title with Apple II phosphor green tones
	fmt.Printf("%s%s", colorBold, colorPhosphorGlow)
	fmt.Println("  ██████╗  ██████╗ ██████╗ ██╗  ██╗███╗   ██╗ ██████╗ ██╗      ██████╗  ██████╗ ██╗ ██████╗")
	fmt.Printf("%s", colorPhosphorMid)
	fmt.Println("  ██╔══██╗██╔═══██╗██╔══██╗██║  ██║████╗  ██║██╔═══██╗██║     ██╔═══██╗██╔════╝ ██║██╔════╝")
	fmt.Printf("%s", colorPhosphorBright)
	fmt.Println("  ██████╔╝██║   ██║██║  ██║███████║██╔██╗ ██║██║   ██║██║     ██║   ██║██║  ███╗██║██║     ")
	fmt.Printf("%s", colorPhosphorDim)
	fmt.Println("  ██╔═══╝ ██║   ██║██║  ██║██╔══██║██║╚██╗██║██║   ██║██║     ██║   ██║██║   ██║██║██║     ")
	fmt.Printf("%s", colorPhosphorBright)
	fmt.Println("  ██║     ╚██████╔╝██████╔╝██║  ██║██║ ╚████║╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║╚██████╗")
	fmt.Printf("%s", colorPhosphorMid)
	fmt.Println("  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝")
	fmt.Printf("%s", colorReset)
	fmt.Println()

	// Music wave with phosphor greens
	fmt.Printf("%s          %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s\n",
		colorReset,
		colorPhosphorDim, colorReset, colorPhosphorMid, colorReset, colorPhosphorBright, colorReset,
		colorPhosphorGlow, colorReset, colorPhosphorDim, colorReset, colorPhosphorMid, colorReset,
		colorPhosphorBright, colorReset, colorPhosphorGlow, colorReset, colorPhosphorDim, colorReset,
		colorPhosphorMid, colorReset, colorPhosphorBright, colorReset, colorPhosphorGlow, colorReset,
		colorPhosphorDim, colorReset, colorPhosphorMid, colorReset)

	fmt.Println()
}

func printSuccess(message string) {
	fmt.Printf("%s%s✓%s %s\n", colorBold, colorGreen, colorReset, message)
}

func printError(message string) {
	fmt.Printf("%s%s✗%s %s\n", colorBold, colorRed, colorReset, message)
}

func printInfo(message string) {
	fmt.Printf("%s%s→%s %s\n", colorBold, colorInfo, colorReset, message)
}

func printWarning(message string) {
	fmt.Printf("%s%s⚠%s  %s\n", colorBold, colorWarning, colorReset, message)
}
