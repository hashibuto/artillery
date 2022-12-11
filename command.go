package artillery

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashibuto/artillery/pkg/tg"
	ns "github.com/hashibuto/nilshell"
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
	Options            []*Option
	Arguments          []*Argument
	OnExecute          func(Namespace, *Processor) error
	OnCompleteOverride func(cmd *Command, tokens []any, processor *Processor) []*ns.AutoComplete

	// These are computed when they are added to the shell
	subCommandLookup  map[string]*Command
	shortNameToName   map[string]string
	nameToArgOrOption map[string]any
	isInitialized     bool
}

// Prepare establishes the validity of the command as well as prepares various optimizations, and returns an
// error on the first validation violation
func (cmd *Command) Prepare() error {
	// Make the data safer, and sort everything so that we only need to do it once
	if cmd.SubCommands == nil {
		cmd.SubCommands = []*Command{}
	}
	sort.Slice(cmd.SubCommands, func(i, j int) bool {
		return cmd.SubCommands[i].Name < cmd.SubCommands[j].Name
	})
	if cmd.Options == nil {
		cmd.Options = []*Option{}
	}
	sort.Slice(cmd.Options, func(i, j int) bool {
		return cmd.Options[i].Name < cmd.Options[j].Name
	})
	if cmd.Arguments == nil {
		cmd.Arguments = []*Argument{}
	}
	sort.Slice(cmd.Arguments, func(i, j int) bool {
		return cmd.Arguments[i].Name < cmd.Arguments[j].Name
	})

	if cmd.Name == "" {
		return fmt.Errorf("Commmand requires a name")
	}
	if cmd.Description == "" {
		return fmt.Errorf("Command requires a description")
	}

	if len(cmd.SubCommands) > 0 {
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
			err := subCommand.Prepare()
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

		if len(cmd.Options) > 0 {
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

		if len(cmd.Arguments) > 0 {
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
	tg.Print(tg.Blue, cmd.Description, tg.Reset, "\n\n")
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
		table := tg.NewTable("subcommand", "description")
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
			table := tg.NewTable("", "name", "description")
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
			table := tg.NewTable("", "name", "description")
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

// Process processes the supplied cliArgs as though this were a standalone commmand.  This is useful for processing arguments directly from
// the cli
func (cmd *Command) Process(cliArgs []string) error {
	catTokens := categorizeTokens(cliArgs)
	return cmd.Execute(catTokens, nil)
}

// Execute attempts to execute the supplied argument tokens after evaluating the input against the
// specified rules.
func (cmd *Command) Execute(tokens []any, processor *Processor) error {
	namespace := Namespace{}
	for _, arg := range cmd.Arguments {
		arg.ApplyDefault(namespace)
	}
	for _, opt := range cmd.Options {
		opt.ApplyDefault(namespace)
	}

	if len(cmd.SubCommands) > 0 {
		// Attempt to look up a subcommand
		subCmdStr, tokens, err := extractCommand(tokens)
		if err != nil {
			return err
		}
		subCmd, ok := cmd.subCommandLookup[subCmdStr]
		if !ok {
			return fmt.Errorf("%s is not a valid subcommand of %s", subCmdStr, cmd.Name)
		}
		return subCmd.Execute(tokens, processor)
	}

	var err error
	tokens, err = cmd.CompressTokens(tokens)
	if err != nil {
		return err
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

	return cmd.OnExecute(namespace, processor)
}

func (cmd *Command) OnComplete(tokens []any, processor *Processor) []*ns.AutoComplete {
	if cmd.OnCompleteOverride != nil {
		return cmd.OnCompleteOverride(cmd, tokens, processor)
	}

	return cmd.onComplete(tokens, processor)
}

func (cmd *Command) onComplete(tokens []any, processor *Processor) []*ns.AutoComplete {
	sug := []*ns.AutoComplete{}

	// We only operate on arguments
	if len(cmd.Arguments) == 0 {
		return sug
	}

	// Is the current input token an arg?
	isArg := false

	finalToken := tokens[len(tokens)-1]
	switch finalToken.(type) {
	case string:
		isArg = true
	}

	if !isArg {
		return sug
	}

	// if it's an arg, which arg is it
	count := 0
	for _, token := range tokens {
		switch token.(type) {
		case string:
			count++
		}
	}
	if count > len(cmd.Arguments) {
		if !cmd.Arguments[len(cmd.Arguments)-1].IsArray {
			// Empty
			return sug
		}
		count = len(cmd.Arguments)
	}

	cmdArg := cmd.Arguments[count-1]
	if cmdArg.CompletionFunc != nil {
		results := cmdArg.CompletionFunc(finalToken.(string), processor)
		for _, result := range results {
			sug = append(sug, &ns.AutoComplete{
				Name: result,
			})
		}
	} else if cmdArg.MemberOf != nil {
		for _, result := range cmdArg.MemberOf {
			if strings.HasPrefix(result, finalToken.(string)) {
				sug = append(sug, &ns.AutoComplete{
					Name: result,
				})
			}
		}
	}

	return sug
}

// CompressTokens compresses any token/value pairs where required into a single *Option.
func (cmd *Command) CompressTokens(tokens []any) ([]any, error) {
	shortNameToName := cmd.shortNameToName
	if shortNameToName == nil {
		shortNameToName = map[string]string{}
	}

	compressed := []any{}
	idx := 0
	for idx < len(tokens) {
		token := tokens[idx]
		switch t := token.(type) {
		case *OptionInput:
			name := t.Name
			var optAny any
			var ok bool
			if len(t.Name) == 1 {
				name, ok = shortNameToName[t.Name]
				if !ok {
					return nil, fmt.Errorf("Unknown option %s", t.Name)
				}
			}

			optAny, ok = cmd.nameToArgOrOption[name]
			if !ok {
				return nil, fmt.Errorf("Unknown option %s", t.Name)
			}

			switch o := optAny.(type) {
			case *Option:
				if o.Value == nil && t.Value == "" {
					if idx < len(tokens)-1 {
						switch oo := tokens[idx+1].(type) {
						case string:
							t.Value = oo
							compressed = append(compressed, t)
							idx += 2
							continue
						default:
							return nil, fmt.Errorf("Option %s requires a companion argument", o.InvocationDisplay())
						}
					} else {
						return nil, fmt.Errorf("Option %s requires a companion argument", o.InvocationDisplay())
					}
				}
			}
		}
		compressed = append(compressed, token)
		idx++
	}

	return compressed, nil
}
