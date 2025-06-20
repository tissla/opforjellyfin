// internal/stylefactory.go
package internal

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

type ColorStyle = lipgloss.Style

// mapping colors to structs so i dont have to remember what the numbers mean
var Style = struct {
	Pink  ColorStyle
	Red   ColorStyle
	Green ColorStyle
	LBlue ColorStyle
}{
	Pink:  ColorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("201"))),
	Red:   ColorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("9"))),
	Green: ColorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("10"))),
	LBlue: ColorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("12"))),
}

// simple style wrapper
func StyleFactory(text string, style lipgloss.Style) string {
	return style.Render(text)
}

// tone a number between green and red depending on its value in relation to min, max (i didnt find this in lipgloss)
func StyleByRange(value, min, max int) string {
	text := fmt.Sprintf("%d", value)

	if value <= min {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render(text)
	}
	if value >= max {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Render(text)
	}

	ratio := float64(value-min) / float64(max-min)

	interp := func(a, b int, t float64) int {
		return int(math.Round(float64(a)*(1-t) + float64(b)*t))
	}

	r := interp(255, 0, ratio)
	g := interp(0, 255, ratio)
	b := interp(0, 0, ratio)

	color := fmt.Sprintf("#%02x%02x%02x", r, g, b)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(text)
}

// zebra lines
func RenderRow(line string, isAlt bool) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	visible := ansi.Strip(line)
	visibleWidth := runewidth.StringWidth(visible)

	padding := width - visibleWidth
	if padding < 0 {
		padding = 0
	}
	padded := line + strings.Repeat(" ", padding)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(width)

	if isAlt {
		style = style.Background(lipgloss.Color("236"))
	}

	return style.Render(padded)
}
