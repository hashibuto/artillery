package artillery

import (
	"fmt"
	"regexp"

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
	OnExecute func(Namespace)

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

	if cmd.OnExecute == nil {
		return fmt.Errorf("OnExecute method is required")
	}

	return nil
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

	cmd.OnExecute(namespace)

	return nil
}

// Suggest returns options which match the filter string
func (cmd *Command) Suggest(filter string) []prompt.Suggest {
	return nil
}
