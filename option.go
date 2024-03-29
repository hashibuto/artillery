package artillery

import (
	"fmt"
)

func CreateEmptyArrayOfType(arrType ArgType) any {
	switch arrType {
	case String:
		return []string{}
	case Int:
		return []int{}
	case Float:
		return []float64{}
	case Bool:
		return []bool{}
	default:
		return []string{}
	}
}

type Option struct {
	ShortName   byte
	Name        string
	Description string
	Type        ArgType
	Value       any // When value is specified, the option has an implicit value and cannot be provided with --opt=value
	Default     any
	IsArray     bool // When true, argument can be reused multiple times
	IsRequired  bool // When true a value is required to be set
}

// Validate ensures the validity of the option
func (opt *Option) Validate() error {
	shortStr := string(opt.ShortName)
	if opt.ShortName != 0 {
		if !validOptionName.MatchString(shortStr) {
			return fmt.Errorf("Option short names can only contain A-Z, a-z, 0-9 and _")
		}
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
		namespace[opt.Name] = CreateEmptyArrayOfType(opt.Type)
	} else {
		namespace[opt.Name] = opt.Default
	}
}

// ApplyArrayDefaults applies array defaults to the target if empty after processing
func (opt *Option) ApplyArrayDefaults(namespace Namespace) {
	if opt.IsArray {
		val := namespace[opt.Name]
		var length int
		switch t := val.(type) {
		case []string:
			length = len(t)
		case []int:
			length = len(t)
		case []float64:
			length = len(t)
		case []bool:
			length = len(t)
		}

		if length == 0 && opt.Default != nil {
			namespace[opt.Name] = opt.Default
		}
	}
}

// Apply will apply the input to the namespace.  If input is nil then the default will be applied
func (opt *Option) Apply(inp *OptionInput, namespace Namespace) error {
	if opt.IsArray {
		if inp.Value == "" {
			return fmt.Errorf("Value must be specified for option %s", opt.InvocationDisplay())
		}

		val, err := convert(inp.Value, opt.Type)
		if err != nil {
			return fmt.Errorf("Option %s - %s", opt.InvocationDisplay(), err)
		}

		switch t := namespace[opt.Name].(type) {
		case []string:
			namespace[opt.Name] = append(t, val.(string))
		case []int:
			namespace[opt.Name] = append(t, val.(int))
		case []float64:
			namespace[opt.Name] = append(t, val.(float64))
		case []bool:
			namespace[opt.Name] = append(t, val.(bool))
		}
		return nil
	}

	if inp.Value == "" && opt.Value == nil && opt.Default == nil && opt.IsRequired {
		return fmt.Errorf("Option %s is required", opt.InvocationDisplay())
	}

	if opt.Value != nil {
		if inp.Value != "" {
			return fmt.Errorf("Option %s does not accept an \"=\" assigment operator", opt.InvocationDisplay())
		}
		namespace[opt.Name] = opt.Value
	} else {
		if inp.Value == "" && opt.Default == nil {
			return fmt.Errorf("Option %s must specify a value by use of an \"=\" assigment operator", opt.InvocationDisplay())
		}

		val, err := convert(inp.Value, opt.Type)
		if err != nil {
			return fmt.Errorf("Option %s - %s", opt.InvocationDisplay(), err)
		}

		namespace[opt.Name] = val
	}

	return nil
}

// InvocationDisplay returns the help name for the option
func (opt *Option) InvocationDisplay() string {
	extra := ""
	if opt.Default != nil {
		extra = fmt.Sprintf("=%s", opt.DefaultValueDisplay())
	} else {
		extra = fmt.Sprintf("=<%s>", opt.ArgTypeDisplay())
	}

	if opt.ShortName != 0 {
		return fmt.Sprintf("-%s, --%s%s", string(opt.ShortName), opt.Name, extra)
	}
	return fmt.Sprintf("--%s%s", opt.Name, extra)
}

// ArgTypeDisplay returns the argument data type for display
func (opt *Option) ArgTypeDisplay() string {
	switch opt.Type {
	case "":
		return string(String)
	default:
		return string(opt.Type)
	}
}

// DefaultValueDisplay returns the default value for display purposes
func (opt *Option) DefaultValueDisplay() string {
	switch t := opt.Default.(type) {
	case string:
		return fmt.Sprintf("'%s'", t)
	case int:
		return fmt.Sprintf("%d", t)
	case float64:
		return fmt.Sprintf("%0.3f", t)
	case bool:
		if opt.Default == true {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", t)
	}
}
