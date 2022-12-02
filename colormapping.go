package artillery

import (
	"github.com/hashibuto/artillery/pkg/term"
	"github.com/hashibuto/go-prompt"
)

var ColorMapping = map[term.Color]prompt.Color{
	term.Black:       prompt.Black,
	term.DarkRed:     prompt.DarkRed,
	term.DarkGreen:   prompt.DarkGreen,
	term.Brown:       prompt.Brown,
	term.DarkBlue:    prompt.DarkBlue,
	term.DarkMagenta: prompt.Purple,
	term.DarkCyan:    prompt.Cyan,
	term.Grey:        prompt.LightGray,
	term.Red:         prompt.Red,
	term.Green:       prompt.Green,
	term.Yellow:      prompt.Yellow,
	term.Blue:        prompt.Blue,
	term.Magenta:     prompt.Fuchsia,
	term.Cyan:        prompt.Turquoise,
	term.White:       prompt.LightGray,
}
