package artillery

import (
	"fmt"
	"sort"

	"github.com/hashibuto/artillery/pkg/tg"
)

type helpCommandArgs struct {
	Verbose bool
	Command []any
}

func makeHelpCommand() *Command {
	return &Command{
		Name:        "help",
		Description: "display the command set, and contextual help",
		Arguments: []*Argument{
			{
				Name:        "command",
				Description: "command and subcommand if available",
				IsArray:     true,
			},
		},
		OnExecute: func(ns Namespace, processor *Processor) error {
			helpArgs := &helpCommandArgs{}
			err := Reflect(ns, helpArgs)
			if err != nil {
				return err
			}

			if len(helpArgs.Command) == 0 {
				fmt.Println()
				groups := []string{}
				byGroup := map[string][]*Command{}
				for _, cmd := range processor.commandLookup {
					_, ok := byGroup[cmd.Group]
					if !ok {
						groups = append(groups, cmd.Group)
						byGroup[cmd.Group] = []*Command{}
					}
					byGroup[cmd.Group] = append(byGroup[cmd.Group], cmd)
				}

				// Alphabetize the groups
				sort.Slice(groups, func(i, j int) bool {
					return i < j
				})

				// Alphabetize within the groups
				for _, group := range byGroup {
					sort.Slice(group, func(i, j int) bool {
						return group[i].Name < group[j].Name
					})
				}

				for _, groupName := range groups {
					group := byGroup[groupName]
					if groupName == "" {
						if processor.DefaultHeading == "" {
							groupName = "commands"
						} else {
							groupName = processor.DefaultHeading
						}
					}
					tg.Print(tg.Bold, tg.Blue, groupName, "\n\n", tg.Reset)
					table := tg.NewTable("command", "description")
					table.HideHeading = true
					for _, cmd := range group {
						table.Append(cmd.Name, cmd.Description)
					}
					table.Render()
					fmt.Println()
				}
			} else {
				var curCommand *Command
				var ok bool
				curLookup := processor.commandLookup
				for _, cmdName := range helpArgs.Command {
					cmdNameStr := cmdName.(string)
					curCommand, ok = curLookup[cmdNameStr]
					if !ok {
						return fmt.Errorf("Unknown command or subcommand \"%s\"", cmdNameStr)
					}
					curLookup = curCommand.subCommandLookup
				}
				curCommand.DisplayHelp()
			}

			return nil
		},
	}
}
