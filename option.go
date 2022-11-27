package artillery

import "fmt"

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
func (opt *Option) ApplyDefault(target map[string]any) {
	if opt.IsArray {
		target[opt.Name] = []string{}
	} else {
		target[opt.Name] = opt.Default
	}
}

// Apply will apply the input to the target.  If input is nil then the default will be applied
func (opt *Option) Apply(inp *OptionInput, target map[string]any) error {
	if opt.IsArray {
		if inp.Value == "" {
			return fmt.Errorf("Value must be specified for option --%s", opt.Name)
		}

		targ := target[opt.Name].([]string)
		target[opt.Name] = append(targ, inp.Value)
		return nil
	}

	if opt.Value != nil {
		if inp.Value != "" {
			return fmt.Errorf("Option -%s/--%s does not accept an \"=\" assigment operator", string(opt.ShortName), opt.Name)
		}
		target[opt.Name] = opt.Value
	} else {
		if inp.Value == "" {
			return fmt.Errorf("Option -%s/--%s must specify a value by use of an \"=\" assigment operator", string(opt.ShortName), opt.Name)
		}

		if opt.Type == "" || opt.Type == String {
			target[opt.Name] = inp.Value
		} else {
			switch opt.Type {
			case Int:
				// Do atoi
			}
		}

	}

	return nil
}
