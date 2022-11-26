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

		for idx, subCommand := range cmd.SubCommands {
			err := subCommand.Validate()
			if err != nil {
				return fmt.Errorf("Error in subcommand at position %d\n%w", idx, err)
			}
		}
	} else {
		if cmd.Options != nil {
			for idx, opt := range cmd.Options {
				err := opt.Validate()
				if err != nil {
					return fmt.Errorf("Error in command %s option %d\n%w", cmd.Name, idx, err)
				}
			}
		}

		if cmd.Arguments != nil {
			for idx, arg := range cmd.Arguments {
				err := arg.Validate(idx == len(cmd.Arguments)-1)
				if err != nil {
					return fmt.Errorf("Error in command %s argument %d\n%w", cmd.Name, idx, err)
				}
			}
		}
	}

	return nil
}

type Option struct {
	ShortName   byte
	Name        string
	Description string
	Type        ArgType
	Value       any // When value is specified, the option has an implicit value and cannot be provided with --opt=value
	Default     any
}

// Validate ensures the validity of the option
func (opt *Option) Validate() error {
	shortStr := string(opt.ShortName)
	if !validOptionName.MatchString(shortStr) {
		return fmt.Errorf("Option short names can only contain A-Z, a-z, 0-9 and _")
	}

	if len(opt.Name) < 2 {
		return fmt.Errorf("Option names must be at least 2 characters long")
	}

	if !validOptionName.MatchString(opt.Name) {
		return fmt.Errorf("Option names can only contain A-Z, a-z, 0-9 and _")
	}

	if opt.Description == "" {
		return fmt.Errorf("Option must have a description")
	}

	if opt.Value != nil {
		switch opt.Value.(type) {
		case int:
		case string:
		case bool:
		case float64:
		default:
			return fmt.Errorf("Default value must be one of int, string, bool, or float64 types")
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
