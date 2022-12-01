package artillery

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/hashibuto/artillery/pkg/term"
	"github.com/hashibuto/go-prompt"
)

type ArgType string

var validOptionName = regexp.MustCompile("^[A-Za-z0-9_]+")

const (
	String ArgType = "string"
	Int    ArgType = "int"
	Bool   ArgType = "bool"
	Float  ArgType = "float"
)

type Namespace map[string]any

type Command struct {
	Name        string
	Group       string // If specified, group will be presented in the help and similar items will be displayed together
	Description string
	SubCommands []*Command

	// Commands which have subcommands cannot have any of the following
	Options   []*Option
	Arguments []*Argument
	OnExecute func(Namespace) error

	// These are computed when they are added to the shell
	subCommandLookup  map[string]*Command
	shortNameToName   map[string]string
	nameToArgOrOption map[string]any
}

// Validate establishes the validity of the command and returns an error on the first violation
func (cmd *Command) Validate() error {
	if cmd.Name == "" {
		return fmt.Errorf("Commmand requires a name")
	}
	if cmd.Description == "" {
		return fmt.Errorf("Command requires a description")
	}

	if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
		if cmd.OnExecute != nil {
			return fmt.Errorf("Commands with subcommands cannot declare an OnExecute function")
		}
		if cmd.Options != nil && len(cmd.Options) > 0 {
			return fmt.Errorf("Commands with subcommands cannot have their own options")
		}
		if cmd.Arguments != nil && len(cmd.Arguments) > 0 {
			return fmt.Errorf("Commands with subcommands cannot declare their own arguments")
		}

		cmd.subCommandLookup = map[string]*Command{}
		for idx, subCommand := range cmd.SubCommands {
			err := subCommand.Validate()
			if err != nil {
				return fmt.Errorf("Error in subcommand at position %d\n%w", idx, err)
			}

			if _, exists := cmd.subCommandLookup[subCommand.Name]; exists {
				return fmt.Errorf("Subcommand with name \"%s\" already present on command \"%s\"", subCommand.Name, cmd.Name)
			}
			cmd.subCommandLookup[subCommand.Name] = subCommand
		}
	} else {
		nameToArgOrOption := map[string]any{}
		shortNameToName := map[string]string{}

		if cmd.OnExecute == nil {
			return fmt.Errorf("OnExecute method is required")
		}

		if cmd.Options != nil && len(cmd.Options) > 0 {
			cmd.nameToArgOrOption = nameToArgOrOption
			cmd.shortNameToName = shortNameToName
			for idx, opt := range cmd.Options {
				err := opt.Validate()
				if err != nil {
					return fmt.Errorf("Error in command %s option %d\n%w", cmd.Name, idx, err)
				}

				if _, exists := nameToArgOrOption[opt.Name]; exists {
					return fmt.Errorf("Argument name already exists for option \"%s\"", opt.Name)
				}
				nameToArgOrOption[opt.Name] = opt
				if _, exists := shortNameToName[string(opt.ShortName)]; exists {
					return fmt.Errorf("Short name already exists for option \"%s\"", opt.Name)
				}
				shortNameToName[string(opt.ShortName)] = opt.Name
			}
		}

		if cmd.Arguments != nil && len(cmd.Arguments) > 0 {
			cmd.nameToArgOrOption = nameToArgOrOption
			cmd.shortNameToName = shortNameToName
			for idx, arg := range cmd.Arguments {
				err := arg.Validate(idx == len(cmd.Arguments)-1)
				if err != nil {
					return fmt.Errorf("Error in command %s argument %d\n%w", cmd.Name, idx, err)
				}

				if _, exists := nameToArgOrOption[arg.Name]; exists {
					return fmt.Errorf("Argument name already exists for argument \"%s\"", arg.Name)
				}
				nameToArgOrOption[arg.Name] = arg
			}
		}
	}

	return nil
}

func (cmd *Command) DisplayHelp() {
	term.Print(term.Blue, cmd.Description, term.Reset, "\n\n")
	fmt.Println("usage:")
	fmt.Print(cmd.Name)
	if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
		fmt.Printf(" <subcommand>\n\n")
		subCommands := make([]*Command, len(cmd.SubCommands))
		for idx, sub := range cmd.SubCommands {
			subCommands[idx] = sub
		}
		sort.Slice(subCommands, func(i, j int) bool {
			return subCommands[i].Name < subCommands[j].Name
		})
		fmt.Println("subcommands:")
		table := term.NewTable("subcommand", "description")
		table.HideHeading = true
		for _, subCommand := range subCommands {
			table.Append(subCommand.Name, subCommand.Description)
		}
		table.Render()
	} else {
		options := []*Option{}
		args := []*Argument{}

		if cmd.Options != nil && len(cmd.Options) > 0 {
			fmt.Print(" [<options...>]")
			for _, opt := range cmd.Options {
				options = append(options, opt)
			}
		}
		if cmd.Arguments != nil && len(cmd.Arguments) > 0 {
			for _, arg := range cmd.Arguments {
				fmt.Printf(" %s", arg.Usage())
				args = append(args, arg)
			}
		}
		fmt.Printf("\n\n")

		if len(args) > 0 {
			fmt.Println("arguments:")
			sort.Slice(args, func(i, j int) bool {
				return args[i].Name < args[j].Name
			})
			table := term.NewTable("", "name", "description")
			table.HideHeading = true
			for _, arg := range args {
				table.Append("", arg.Name, arg.Description)
			}
			table.Render()
			fmt.Println()
		}

		if len(options) > 0 {
			fmt.Println("options:")
			sort.Slice(options, func(i, j int) bool {
				return options[i].Name < options[j].Name
			})
			table := term.NewTable("", "name", "description")
			table.HideHeading = true
			for _, opt := range cmd.Options {
				table.Append("", opt.InvocationDisplay(), opt.Description)
			}
			table.Render()
			fmt.Println()
		}
	}
	fmt.Println()
}

// Execute attempts to execute the supplied argument tokens after evaluating the input against the
// specified rules.
func (cmd *Command) Execute(tokens []any) error {
	namespace := Namespace{}
	if cmd.Arguments != nil {
		for _, arg := range cmd.Arguments {
			arg.ApplyDefault(namespace)
		}
	}
	if cmd.Options != nil {
		for _, opt := range cmd.Options {
			opt.ApplyDefault(namespace)
		}
	}

	if cmd.SubCommands != nil && len(cmd.SubCommands) > 0 {
		// Attempt to look up a subcommand
		subCmdStr, tokens, err := extractCommand(tokens)
		if err != nil {
			return err
		}
		subCmd, ok := cmd.subCommandLookup[subCmdStr]
		if !ok {
			return fmt.Errorf("%s is not a valid subcommand of %s", subCmdStr, cmd.Name)
		}
		return subCmd.Execute(tokens)
	}

	// This branch of code is on a terminal command (ie. no further subcommands), so evaluate args
	opts, args, err := group(tokens)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		var optName string
		var ok bool
		if len(opt.Name) == 1 {
			optName, ok = cmd.shortNameToName[opt.Name]
			if !ok {
				return fmt.Errorf("Option -%s is not recognized", opt.Name)
			}
		} else {
			optName = opt.Name
		}

		optDef, ok := cmd.nameToArgOrOption[optName]
		if !ok {
			return fmt.Errorf("Option --%s is not recognized", optName)
		}

		switch t := optDef.(type) {
		case *Option:
			err = t.Apply(opt, namespace)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Option --%s is not recognized", optName)
		}
	}

	for idx, arg := range args {
		if cmd.Arguments != nil {
			ix := idx
			if ix >= len(cmd.Arguments) {
				ix = len(cmd.Arguments) - 1
				if !cmd.Arguments[ix].IsArray {
					return fmt.Errorf("Unexpected argument \"%s\"", arg)
				}
			}

			argDef := cmd.Arguments[ix]
			argDef.Apply(arg, namespace)
		} else {
			return fmt.Errorf("Unexpected argument \"%s\"", arg)
		}
	}

	if cmd.Arguments != nil {
		for _, arg := range cmd.Arguments {
			v := namespace[arg.Name]
			if v == nil {
				return fmt.Errorf("Expected argument \"%s\"", arg.Name)
			}
		}
	}

	return cmd.OnExecute(namespace)
}

// Suggest returns options which match the filter string
func (cmd *Command) Suggest(filter string) []prompt.Suggest {
	return nil
}
