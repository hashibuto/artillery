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
	OnExecute func(any, Namespace)

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

	return nil
}

// Execute attempts to execute the supplied argument tokens after evaluating the input against the
// specified rules.
func (cmd *Command) Execute(tokens []any) error {
	namespace := Namespace{}

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

		default:
			return fmt.Errorf("Option --%s is not recognized", optName)
		}
	}

	return nil
}

type Argument struct {
	Name           string
	Description    string
	Type           ArgType                      // String is the default argument type
	Default        any                          // Default value (only valid in the final argument position)
	MemberOf       []string                     // When value must be a member of a limited collection (strings only)
	CompletionFunc func(prefix string) []string // Used to autocomplete if set (cannot be used together with MemberOf)
	NArgs          bool                         // When true, argument becomes an array (must be in the final argument position)
}

// Validate ensures the validity of the argument
func (arg *Argument) Validate(isLast bool) error {
	if len(arg.Name) < 2 {
		return fmt.Errorf("Option names must be at least 2 characters long")
	}

	if arg.MemberOf != nil && len(arg.MemberOf) > 0 && arg.CompletionFunc != nil {
		return fmt.Errorf("MemberOf and CompletionFunc cannot be used together")
	}

	if !isLast {
		if arg.Default != nil {
			return fmt.Errorf("Default may only be specified when it's the final argument")
		}

		if arg.NArgs != false {
			return fmt.Errorf("NArgs can only be true when in the final argument position")
		}
	}

	if arg.Description == "" {
		return fmt.Errorf("Argument must have a description")
	}

	return nil
}

// Suggest returns options which match the filter string
func (cmd *Command) Suggest(filter string) []prompt.Suggest {
	return nil
}
