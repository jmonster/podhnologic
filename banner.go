package main

import (
	"fmt"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

func printBanner() {
	fmt.Println()

	// Title with gradient effect
	fmt.Printf("%s%s", colorBold, colorCyan)
	fmt.Println("  ██████╗  ██████╗ ██████╗ ██╗  ██╗███╗   ██╗ ██████╗ ██╗      ██████╗  ██████╗ ██╗ ██████╗")
	fmt.Printf("%s", colorPurple)
	fmt.Println("  ██╔══██╗██╔═══██╗██╔══██╗██║  ██║████╗  ██║██╔═══██╗██║     ██╔═══██╗██╔════╝ ██║██╔════╝")
	fmt.Printf("%s", colorBlue)
	fmt.Println("  ██████╔╝██║   ██║██║  ██║███████║██╔██╗ ██║██║   ██║██║     ██║   ██║██║  ███╗██║██║     ")
	fmt.Printf("%s", colorPurple)
	fmt.Println("  ██╔═══╝ ██║   ██║██║  ██║██╔══██║██║╚██╗██║██║   ██║██║     ██║   ██║██║   ██║██║██║     ")
	fmt.Printf("%s", colorCyan)
	fmt.Println("  ██║     ╚██████╔╝██████╔╝██║  ██║██║ ╚████║╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║╚██████╗")
	fmt.Println("  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝")
	fmt.Printf("%s", colorReset)
	fmt.Println()

	// Music wave
	fmt.Printf("%s          %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s\n",
		colorReset,
		colorRed, colorReset, colorYellow, colorReset, colorGreen, colorReset,
		colorCyan, colorReset, colorBlue, colorReset, colorPurple, colorReset,
		colorRed, colorReset, colorYellow, colorReset, colorGreen, colorReset,
		colorCyan, colorReset, colorBlue, colorReset, colorPurple, colorReset,
		colorRed, colorReset, colorYellow, colorReset)

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
