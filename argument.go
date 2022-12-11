package artillery

import (
	"fmt"
)

type CompletionFunc func(prefix string, processor *Processor) []string

type Argument struct {
	Name           string
	Description    string
	Type           ArgType        // String is the default argument type
	Default        any            // Default value (only valid in the final argument position)
	MemberOf       []string       // When value must be a member of a limited collection (strings only)
	CompletionFunc CompletionFunc // Used to dynamically list member values, with a prefix for optimization
	IsArray        bool           // When true, argument becomes an array (must be in the final argument position)
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

		if arg.IsArray != false {
			return fmt.Errorf("IsArray can only be true when in the final argument position")
		}
	}

	if arg.Description == "" {
		return fmt.Errorf("Argument must have a description")
	}

	return nil
}

// ApplyDefault applies the default value to the target
func (arg *Argument) ApplyDefault(namespace Namespace) {
	if arg.IsArray {
		namespace[arg.Name] = []any{}
	} else {
		namespace[arg.Name] = arg.Default
	}
}

// Usage displays the usage pattern string
func (arg *Argument) Usage() string {
	if arg.Default != nil {
		return fmt.Sprintf("[%s]", arg.Name)
	}

	if arg.IsArray == true {
		return fmt.Sprintf("<%s...>", arg.Name)
	}

	return fmt.Sprintf("<%s>", arg.Name)
}

// Apply will apply the input to the target.  If input is nil then the default will be applied
func (arg *Argument) Apply(inp string, namespace Namespace) error {
	val, err := convert(inp, arg.Type)
	if err != nil {
		return fmt.Errorf("Argument %s - %w", arg.Name, err)
	}

	if arg.IsArray {
		targ := namespace[arg.Name].([]any)
		namespace[arg.Name] = append(targ, val)
		return nil
	}

	namespace[arg.Name] = val

	return nil
}
