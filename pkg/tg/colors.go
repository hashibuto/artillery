package tg

import "strings"

type Color string

const (
	Black       Color = "\033[30;m"
	DarkRed           = "\033[31;2m"
	DarkGreen         = "\033[32;2m"
	Brown             = "\033[33;2m"
	DarkBlue          = "\033[34;2m"
	DarkMagenta       = "\033[35;2m"
	DarkCyan          = "\033[36;2m"
	Grey              = "\033[37;2m"
	Red               = "\033[31m"
	Green             = "\033[32m"
	Yellow            = "\033[33m"
	Blue              = "\033[34m"
	Magenta           = "\033[35m"
	Cyan              = "\033[36m"
	White             = "\033[37m"
	Reset             = "\033[0m"
	Bold              = "\033[1m"
	Dim               = "\033[2m"
	Underline         = "\033[4m"
	Blink             = "\033[3m"
)

// BgColor returns the color as a background color codey
func BgColor(color Color) Color {
	return Color(strings.Replace(string(color), "[3", "[4", 1))
}
