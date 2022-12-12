package artillery

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/hashibuto/artillery/pkg/tg"
	ns "github.com/hashibuto/nilshell"
)

type Processor struct {
	DefaultHeading string
	nilShell       *ns.NilShell
	commandLookup  map[string]*Command

	beforeAndCursor string
	afterCursor     string
	full            string
}

func NewProcessor() *Processor {
	proc := &Processor{
		DefaultHeading: "commands",
		commandLookup:  map[string]*Command{},
	}
	proc.nilShell = ns.NewShell("» ", proc.OnComplete, proc.OnExecute)
	err := proc.AddCommand(makeHelpCommand())
	if err != nil {
		panic(fmt.Sprintf("Problem with the help command\n%v", err))
	}
	err = proc.AddCommand(makeClearCommand())
	if err != nil {
		panic(fmt.Sprintf("Problem with the clear command\n%v", err))
	}
	err = proc.AddCommand(makeSetCommand())
	if err != nil {
		panic(fmt.Sprintf("Problem with the set command\n%v", err))
	}
	err = proc.AddCommand(makeExitCommand())
	if err != nil {
		panic(fmt.Sprintf("Problem with the exit command\n%v", err))
	}
	return proc
}

// Shell returns the underlying NilShell instance
func (p *Processor) Shell() *ns.NilShell {
	return p.nilShell
}

// AddCommand adds a command to the processor.  If the command is in some way invalid, an error will be returned here.
func (p *Processor) AddCommand(cmd *Command) error {
	err := cmd.Prepare()
	if err != nil {
		return err
	}

	if _, exists := p.commandLookup[cmd.Name]; exists {
		return fmt.Errorf("Command named %s already declared", cmd.Name)
	}

	p.commandLookup[cmd.Name] = cmd

	return nil
}

func (p *Processor) OnExecute(nilShell *ns.NilShell, input string) {
	if len(input) == 0 {
		return
	}

	if input[0] == '!' {
		input = input[1:]
		tokens, openQuote := tokenize(input)
		if openQuote {
			tg.Println(tg.Red, "Unterminated quotation", tg.Reset)
			return
		}
		if len(tokens) == 0 {
			return
		}
		args := []string{}
		if len(tokens) > 1 {
			args = append(args, tokens[1:]...)
		}
		cmd := exec.Command(tokens[0], args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// Parse input
	tokens, err := parse(input)
	if err != nil {
		tg.Println(tg.Red, err, tg.Reset)
		return
	}

	if len(tokens) == 0 {
		return
	}

	cmdStr, tokens, err := extractCommand(tokens)
	if err != nil {
		tg.Println(tg.Red, err, tg.Reset)
		return
	}

	cmd, ok := p.commandLookup[cmdStr]
	if !ok {
		tg.Println(tg.Red, "Command \"", cmdStr, "\" not found", tg.Reset)
		return
	}
	err = cmd.Execute(tokens, p)
	if err != nil {
		tg.Println(tg.Red, err, tg.Reset)
		return
	}
}

func (p *Processor) OnComplete(beforeAndCursor string, afterCursor string, full string) []*ns.AutoComplete {
	p.beforeAndCursor = beforeAndCursor
	p.afterCursor = afterCursor
	p.full = full

	sug := []*ns.AutoComplete{}
	tokens, openQuote := tokenize(beforeAndCursor)
	if openQuote {
		return []*ns.AutoComplete{}
	}

	var finalChar byte
	if len(beforeAndCursor) == 0 {
		finalChar = ' '
	} else {
		finalChar = beforeAndCursor[len(beforeAndCursor)-1]
	}
	if finalChar == ' ' {
		tokens = append(tokens, "")
	}

	curLookup := p.commandLookup
	for idx, arg := range tokens {
		prefix := idx == (len(tokens) - 1)
		if !prefix {
			cmd, ok := curLookup[arg]
			if !ok {
				// Will be empty
				return sug
			}

			if len(cmd.SubCommands) == 0 {
				// No more subcommands, let's try suggestions within the command now
				remainingTokens := tokens[idx+1:]
				categorizedTokens := categorizeTokens(remainingTokens)
				compressedTokens, err := cmd.CompressTokens(categorizedTokens)
				if err != nil {
					return sug
				}

				return cmd.OnComplete(compressedTokens, p)
			}

			curLookup = cmd.subCommandLookup
		} else {
			for name := range curLookup {
				if strings.HasPrefix(name, arg) {
					sug = append(sug, &ns.AutoComplete{
						Name: name,
					})
				}
			}
		}
	}

	if len(sug) == 1 && len(tokens) > 0 && tokens[len(tokens)-1] == sug[0].Name {
		return []*ns.AutoComplete{}
	}

	sort.Slice(sug, func(i, j int) bool {
		return sug[i].Name < sug[j].Name
	})
	return sug
}
