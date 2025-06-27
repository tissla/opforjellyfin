// ui/stylefactory.go
package ui

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
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
func StyleByRange(input interface{}, min, max int) string {

	var (
		text  string
		value int
	)

	switch v := input.(type) {
	case int:
		value = v
		text = strconv.Itoa(v)
	case string:
		text = v
		var err error
		value, err = findIntValueInString(v)
		if err != nil {
			return text
		}
	default:
		return fmt.Sprintf("%v", input) // fallback
	}

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
	g := 255 - r //rest of color
	b := 0       //always zero

	color := fmt.Sprintf("#%02x%02x%02x", r, g, b)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(text)
}

// get width func
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	return width
}

// zebra lines
func RenderRow(format string, isAlt bool, args ...interface{}) string {

	width := GetTerminalWidth()
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

	// debug, this spits out ALOT of lines when used with 'list'.
	//logger.DebugLog(false, "[RenderRow] RAW: %q\n[RenderRow] VISIBLE: %q\n[RenderRow] RENDERED: %q\n", row, visible, final)

	return final
}

// AnsiPadLeft pads text with chosen filler on the left
func AnsiPadLeft(text string, width int, taila ...string) string {
	tail := ""
	if len(taila) > 0 {
		tail = taila[0]
	}
	trunc := ansi.Truncate(text, width, tail)
	visible := runewidth.StringWidth(ansi.Strip(trunc))
	if visible < width {
		padding := strings.Repeat(" ", width-visible)
		return padding + trunc
	}
	return trunc
}

// AnsiPadsRight pads text with chosen filler on the right
func AnsiPadRight(text string, width int, taila ...string) string {
	tail := ""
	if len(taila) > 0 {
		tail = taila[0]
	}
	trunc := ansi.Truncate(text, width, tail)
	visible := runewidth.StringWidth(trunc)
	if visible < width {
		trunc += strings.Repeat(" ", width-visible)
	}
	return trunc
}

// helper for StyleByRange
func findIntValueInString(text string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(text)
	if match == "" {
		return 0, errors.New("no number found in string")
	}
	value, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}
	return value, nil
}
