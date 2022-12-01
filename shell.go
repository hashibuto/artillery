package artillery

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/hashibuto/artillery/pkg/term"
	"github.com/hashibuto/go-prompt"
)

var (
	inst *Shell      = nil
	lock *sync.Mutex = &sync.Mutex{}
)

type Shell struct {
	defaultHeading string
	promptText     string
	commandLookup  map[string]*Command
}

// SetPrompt sets the prompt text, which will be displayed the next time the prompt is rendered
func (s *Shell) SetPrompt(promptText string) {
	s.promptText = promptText
}

// SetDefaultHeading sets the default heading for ungrouped commands
func (s *Shell) SetDefaultHeading(defaultHeading string) {
	s.defaultHeading = defaultHeading
}

// GetInstance returns a singleton shell instance
func GetInstance() *Shell {
	lock.Lock()
	defer lock.Unlock()

	if inst == nil {
		inst = &Shell{
			promptText:    term.Sprint(term.Red, "🩥  ", term.Reset),
			commandLookup: map[string]*Command{},
		}

		err := inst.AddCommand(makeHelpCommand())
		if err != nil {
			log.Fatal(err)
		}

		err = inst.AddCommand(makeExitCommand())
		if err != nil {
			log.Fatal(err)
		}
	}

	return inst
}

// AddCommand adds a command to the shell.  If the command is in some way invalid, an error will be returne here.
func (s *Shell) AddCommand(cmd *Command) error {
	err := cmd.Validate()
	if err != nil {
		return err
	}

	if _, exists := s.commandLookup[cmd.Name]; exists {
		return fmt.Errorf("Command named %s already declared", cmd.Name)
	}

	s.commandLookup[cmd.Name] = cmd

	return nil
}

// Repl performs the Read/Eval/Print/Loop and blocks until exited
func (s *Shell) Repl() {
	p := prompt.New(func(string) {}, s.completer, prompt.OptionPrefix(""))
	for {
		fmt.Print(s.promptText)
		input, shouldExit := p.InputWithExit()
		if shouldExit {
			break
		}

		// Parse input
		tokens, err := parse(input)
		if err != nil {
			term.Println(term.Red, err, term.Reset)
			continue
		}

		if len(tokens) == 0 {
			continue
		}

		cmdStr, tokens, err := extractCommand(tokens)
		if err != nil {
			term.Println(term.Red, err, term.Reset)
			continue
		}

		cmd, ok := s.commandLookup[cmdStr]
		if !ok {
			term.Println(term.Red, "Command \"", cmdStr, "\" not found", term.Reset)
			continue
		}
		err = cmd.Execute(tokens)
		if err != nil {
			term.Println(term.Red, err, term.Reset)
			continue
		}
	}
	s.Exit(0)
}

// Exit exits the shell, cleaning up the terminal in the process
func (s *Shell) Exit(statusCode int) {
	cmd := exec.Command("reset")
	cmd.Run()
	os.Exit(statusCode)
}

func (s *Shell) executor(cmd string) {
}

func (s *Shell) completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix([]prompt.Suggest{}, d.GetWordBeforeCursor(), true)
}
