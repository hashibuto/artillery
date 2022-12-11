package artillery

import (
	"fmt"
	"strings"

	"github.com/hashibuto/artillery/pkg/tg"
)

func makeSetCommand() *Command {
	return &Command{
		Name:        "set",
		Description: "modify a setting",
		Arguments: []*Argument{
			{
				Name:        "setting",
				Description: "setting to change",
				MemberOf:    []string{"debug"},
			},
			{
				Name:        "value",
				Description: "new value of setting",
			},
		},
		OnExecute: func(ns Namespace, processor *Processor) error {
			var args struct {
				Setting string
				Value   string
			}
			err := Reflect(ns, &args)
			if err != nil {
				return err
			}

			switch args.Setting {
			case "debug":
				lower := strings.ToLower(args.Value)
				if lower == "true" {
					processor.nilShell.Debug = true
					fmt.Println("debug mode on")
				} else if lower == "false" {
					processor.nilShell.Debug = false
					fmt.Println("debug mode off")
				} else {
					tg.Println(tg.Red, "Debug setting must be true/false", tg.Reset)
				}
			}

			return nil
		},
	}
}
