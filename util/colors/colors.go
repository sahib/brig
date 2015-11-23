// Package colors implement easy printing of terminal colors.
package colors

import "fmt"

const (
	Cyan          = 36
	Green         = 32
	Magenta       = 35
	Red           = 31
	Yellow        = 33
	BackgroundRed = 41
)

// Reset sequence
var ColorResetEscape = "\033[0m"

// ColorResetEscape translates a ANSI color number to a color escape.
func ColorEscape(color int) string {
	return fmt.Sprintf("\033[0;%dm", color)
}

// Colorize the msg using ANSI color escapes
func Colorize(msg string, color int) string {
	return ColorEscape(color) + msg + ColorResetEscape
}
