package artillery

import (
	"fmt"
)

type Option struct {
	ShortName   byte
	Name        string
	Description string
	Type        ArgType
	Value       any // When value is specified, the option has an implicit value and cannot be provided with --opt=value
	Default     any
	IsArray     bool // When true, argument can be reused multiple times
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

// ApplyDefault applies the default value to the target
func (opt *Option) ApplyDefault(namespace Namespace) {
	if opt.IsArray {
		namespace[opt.Name] = []any{}
	} else {
		namespace[opt.Name] = opt.Default
	}
}

// Apply will apply the input to the namespace.  If input is nil then the default will be applied
func (opt *Option) Apply(inp *OptionInput, namespace Namespace) error {
	if opt.IsArray {
		if inp.Value == "" {
			return fmt.Errorf("Value must be specified for option --%s", opt.Name)
		}

		arr := namespace[opt.Name].([]any)
		val, err := convert(inp.Value, opt.Type)
		if err != nil {
			return fmt.Errorf("Option -%s/--%s - %s", string(opt.ShortName), opt.Name, err)
		}
		namespace[opt.Name] = append(arr, val)
		return nil
	}

	if opt.Value != nil {
		if inp.Value != "" {
			return fmt.Errorf("Option -%s/--%s does not accept an \"=\" assigment operator", string(opt.ShortName), opt.Name)
		}
		namespace[opt.Name] = opt.Value
	} else {
		if inp.Value == "" && opt.Default == nil {
			return fmt.Errorf("Option -%s/--%s must specify a value by use of an \"=\" assigment operator", string(opt.ShortName), opt.Name)
		}

		val, err := convert(inp.Value, opt.Type)
		if err != nil {
			return fmt.Errorf("Option -%s/--%s - %s", string(opt.ShortName), opt.Name, err)
		}

		namespace[opt.Name] = val
	}

	return nil
}
