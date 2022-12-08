package tg

import (
	"fmt"
	"strings"
)

// Print displays output to the stdout
func Print(args ...any) {
	fmt.Print(Sprint(args...))
}

// Println displays output to the stdout followed by a newline
func Println(args ...any) {
	newArgs := append(args, "\n")
	fmt.Print(Sprint(newArgs...))
}

// Sprint prints output to a string
func Sprint(args ...any) string {
	formatters := []string{}
	for _, arg := range args {
		var fmtSpec string
		switch arg.(type) {
		case int:
			fmtSpec = "%i"
		case string:
			fmtSpec = "%s"
		default:
			fmtSpec = "%v"
		}

		formatters = append(formatters, fmtSpec)
	}

	formatter := strings.Join(formatters, "")
	return fmt.Sprintf(formatter, args...)
}
