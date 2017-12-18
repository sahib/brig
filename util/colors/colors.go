// Package colors implement easy printing of terminal colors.
package colors

import "fmt"

const (
	// Cyan should be used for debug messages.
	Cyan = 36
	// Green should be used for informational/success messages.
	Green = 32
	// Magenta should be used for critical errors.
	Magenta = 35
	// Red should be used for normal errors.
	Red = 31
	// White can be used for detailed differences
	White = 37
	// Yellow should be used for warnings.
	Yellow = 33
	// BackgroundRed should be used for panic.
	BackgroundRed = 41
)

// ColorResetEscape terminates all previous colors.
var ColorResetEscape = "\033[0m"

// ColorEscape translates a ANSI color number to a color escape.
func ColorEscape(color int) string {
	return fmt.Sprintf("\033[0;%dm", color)
}

// Colorize the msg using ANSI color escapes
func Colorize(msg string, color int) string {
	return ColorEscape(color) + msg + ColorResetEscape
}
