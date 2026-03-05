package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Verbose controls whether debug/trace messages are printed.
var Verbose bool

var (
	// Colors
	colorCyan    = lipgloss.Color("#00D7FF")
	colorGreen   = lipgloss.Color("#00FF87")
	colorYellow  = lipgloss.Color("#FFD700")
	colorRed     = lipgloss.Color("#FF5F5F")
	colorDim     = lipgloss.Color("#6C6C6C")
	colorMagenta = lipgloss.Color("#FF87FF")
	colorWhite   = lipgloss.Color("#FFFFFF")

	// Styles
	styleBanner = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorMagenta).
			PaddingLeft(1).
			PaddingRight(1)

	styleSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	styleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorRed)

	styleWarn = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)

	styleInfo = lipgloss.NewStyle().
			Foreground(colorWhite)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleKey = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	styleVal = lipgloss.NewStyle().
			Foreground(colorWhite)

	styleStructLine = lipgloss.NewStyle().
			Foreground(colorDim)
)

// Icons for status messages.
const (
	iconSuccess = "✔"
	iconError   = "✘"
	iconWarn    = "⚠"
	iconInfo    = "●"
	iconArrow   = "→"
	iconDot     = "·"
)

// PrintBanner prints the styled application banner.
func PrintBanner() {
	banner := `
   ╔═══════════════════════════════════════╗
   ║           C L A Y   O V E N           ║
   ╚═══════════════════════════════════════╝`
	fmt.Println(styleBanner.Render(banner))
	fmt.Println()
}

// PrintHeader prints a styled section header.
func PrintHeader(msg string) {
	bar := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	fmt.Println()
	fmt.Println(styleDim.Render(bar))
	fmt.Println(styleHeader.Render("▸ " + msg))
	fmt.Println(styleDim.Render(bar))
}

// PrintSuccess prints a green success message.
func PrintSuccess(msg string) {
	fmt.Println(styleSuccess.Render("  " + iconSuccess + " " + msg))
}

// PrintError prints a red error message.
func PrintError(msg string) {
	fmt.Fprintln(os.Stderr, styleError.Render("  "+iconError+" "+msg))
}

// PrintWarn prints a yellow warning message.
func PrintWarn(msg string) {
	fmt.Println(styleWarn.Render("  " + iconWarn + " " + msg))
}

// PrintInfo prints a neutral info message.
func PrintInfo(msg string) {
	fmt.Println(styleInfo.Render("  " + iconInfo + " " + msg))
}

// PrintKeyVal prints a key-value pair with aligned formatting.
func PrintKeyVal(key, val string) {
	formatted := fmt.Sprintf("  %-20s %s %s", styleKey.Render(key), styleDim.Render(iconArrow), styleVal.Render(val))
	fmt.Println(formatted)
}

// PrintVerbose prints a dim debug message only when Verbose is true.
func PrintVerbose(msg string) {
	if Verbose {
		fmt.Println(styleDim.Render("  " + iconDot + " " + msg))
	}
}

// PrintStructLine prints a single line of the generated structure file.
func PrintStructLine(line string) {
	line = strings.TrimRight(line, "\n")
	fmt.Println(styleStructLine.Render(line))
}
