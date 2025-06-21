// ui/stylefactory.go
package ui

import (
	"fmt"
	"math"
	"opforjellyfin/internal/logger"
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
func RenderRow(format string, isAlt bool, args ...interface{}) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	row := fmt.Sprintf(format, args...)

	// chatgpt hax solution
	bg := ""
	reset := "\x1b[0m"
	if isAlt {
		bg = "\x1b[48;5;234m"
	}

	var withBg strings.Builder
	inAnsi := false
	for i := 0; i < len(row); i++ {
		ch := row[i]
		withBg.WriteByte(ch)

		if ch == '\x1b' {
			inAnsi = true
		} else if inAnsi && ch == 'm' {
			inAnsi = false
			if bg != "" {
				withBg.WriteString(bg)
			}
		}
	}

	final := bg + withBg.String()

	visible := ansi.Strip(final)
	visibleWidth := runewidth.StringWidth(visible)

	if visibleWidth < width {
		final += strings.Repeat(" ", width-visibleWidth)
	}

	final += reset

	// debug
	logger.DebugLog(false, "[RenderRow] RAW: %q\n[RenderRow] VISIBLE: %q\n[RenderRow] RENDERED: %q\n", row, visible, final)

	return final
}
