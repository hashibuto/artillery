package artillery

import (
	"fmt"
	"regexp"
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

func parse(cmd string) ([]*OptionInput, []string, error) {
	options := []*OptionInput{}
	args := []string{}
	tokens, err := tokenize(cmd)
	if err != nil {
		return nil, nil, err
	}

	startedPosArgs := false
	for _, token := range tokens {
		matches := optionParser.FindAllStringSubmatch(token, -1)
		if matches == nil {
			startedPosArgs = true
		} else if startedPosArgs {
			return nil, nil, fmt.Errorf("Options cannot be preceded by positional arguments")
		}

		if matches != nil {
			inner := matches[0]
			// Single character option
			if inner[2] != "" {
				options = append(options, &OptionInput{
					Name: inner[2],
				})
			}

			// Single character option with equals
			if inner[4] != "" {
				options = append(options, &OptionInput{
					Name:  inner[4],
					Value: inner[5],
				})
			}

			// Grouped single options
			if inner[7] != "" {
				for _, c := range inner[7] {
					options = append(options, &OptionInput{
						Name: string(c),
					})
				}
			}

			// Long form option
			if inner[9] != "" {
				options = append(options, &OptionInput{
					Name: inner[9],
				})
			}

			// Long form option with equals
			if inner[11] != "" {
				options = append(options, &OptionInput{
					Name:  inner[11],
					Value: inner[12],
				})
			}
		} else {
			args = append(args, token)
		}
	}
	return options, args, nil
}

// tokenize breaks the command into individual tokens, preserving quoted areas
func tokenize(cmd string) ([]string, error) {
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

	if openQuote {
		return nil, fmt.Errorf("Command contains an unterminated quote")
	}

	return tokens, nil
}
