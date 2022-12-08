package artillery

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var validNameChars = "[a-zA-Z0-9_]"
var optionParser = regexp.MustCompile(fmt.Sprintf(
	"(^-(%s)$)|(^-(%s)=(.*)$)|(^-(%s+)$)|(^--(%s+)$)|(^--(%s+)=(.*)$)",
	validNameChars,
	validNameChars,
	validNameChars,
	validNameChars,
	validNameChars,
))

type OptionInput struct {
	Name  string
	Value string
}

// group attempts to group tokens into either options or positional arguments.  group will error if
// positional arguments precede options.
func group(tokens []any) ([]*OptionInput, []string, error) {
	options := []*OptionInput{}
	args := []string{}

	argsStarted := false
	for _, token := range tokens {
		switch t := token.(type) {
		case string:
			argsStarted = true
			args = append(args, t)
		case *OptionInput:
			if argsStarted {
				return nil, nil, fmt.Errorf("Options must precede positional arguments")
			}
			options = append(options, t)
		default:
			return nil, nil, fmt.Errorf("Unexpected token type %T", t)
		}
	}

	return options, args, nil
}

// extractCommand attempts to extract a command from the token collection and returns the command
// along with the remaining tokens
func extractCommand(tokens []any) (string, []any, error) {
	if len(tokens) == 0 {
		return "", nil, fmt.Errorf("No command available")
	}

	cmdStr := tokens[0]
	switch t := cmdStr.(type) {
	case string:
		return t, tokens[1:], nil
	default:
		return "", nil, fmt.Errorf("Argument was not a command")
	}
}

// parse linearly extracts options and arguments from the provided command string
func parse(cmd string) ([]any, error) {
	tokens, openQuote := tokenize(cmd)
	if openQuote {
		return nil, fmt.Errorf("Unterminated quotation mark")
	}

	return categorizeTokens(tokens), nil
}

// categorizeTokens categorizes parsed tokens into options or arguments
func categorizeTokens(tokens []string) []any {
	output := []any{}
	for _, token := range tokens {
		matches := optionParser.FindAllStringSubmatch(token, -1)
		if matches != nil {
			inner := matches[0]
			// Single character option
			if inner[2] != "" {
				output = append(output, &OptionInput{
					Name: inner[2],
				})
			}

			// Single character option with equals
			if inner[4] != "" {
				output = append(output, &OptionInput{
					Name:  inner[4],
					Value: inner[5],
				})
			}

			// Grouped single options
			if inner[7] != "" {
				for _, c := range inner[7] {
					output = append(output, &OptionInput{
						Name: string(c),
					})
				}
			}

			// Long form option
			if inner[9] != "" {
				output = append(output, &OptionInput{
					Name: inner[9],
				})
			}

			// Long form option with equals
			if inner[11] != "" {
				output = append(output, &OptionInput{
					Name:  inner[11],
					Value: inner[12],
				})
			}
		} else {
			output = append(output, token)
		}
	}
	return output
}

// tokenize breaks the command into individual tokens, preserving quoted areas
func tokenize(cmd string) ([]string, bool) {
	tokens := []string{}
	openQuote := false
	var quoteChar byte
	tokenStart := 0
	for i := 0; i < len(cmd); i++ {
		if cmd[i] == ' ' && !openQuote {
			token := strings.Trim(cmd[tokenStart:i], " ")
			if (strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"")) || (strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'")) {
				token = token[1 : len(token)-1]
			}
			if len(token) > 0 {
				tokens = append(tokens, token)
			}
			tokenStart = i
		}

		if cmd[i] == '"' || cmd[i] == '\'' {
			if openQuote {
				if quoteChar == cmd[i] {
					openQuote = false
				}
			} else {
				quoteChar = cmd[i]
				openQuote = true
			}
		}
	}
	token := strings.Trim(cmd[tokenStart:], " ")
	if (strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"")) || (strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'")) {
		token = token[1 : len(token)-1]
	}
	if len(token) > 0 {
		tokens = append(tokens, token)
	}

	return tokens, openQuote
}

// convert converts the provided input value to the specified argument type
func convert(value string, argType ArgType) (any, error) {
	if argType == "" || argType == String {
		return value, nil
	}
	switch argType {
	case Int:
		val, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("expected an integer value")
		}
		return val, nil
	case Float:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("expected a floating point value")
		}
		return val, nil
	case Bool:
		val := strings.ToLower(value)
		if val == "true" {
			return true, nil
		}
		if val == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("expected a boolean value")
	default:
		return nil, fmt.Errorf("unexpected data type")
	}
}
