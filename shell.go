package artillery

import (
	"fmt"
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
	promptText string
}

// SetPrompt sets the prompt text, which will be displayed the next time the prompt is rendered
func (s *Shell) SetPrompt(promptText string) {
	s.promptText = promptText
}

// GetInstance returns a singleton shell instance
func GetInstance() *Shell {
	lock.Lock()
	defer lock.Unlock()

	if inst == nil {
		inst = &Shell{
			promptText: term.Sprint(term.Red, "🩥  ", term.Reset),
		}
	}

	return inst
}

// AddCommand adds a command to the shell.  If the command is in some way invalid, an error will be returne here.
func (s *Shell) AddCommand(cmd *Command) error {
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
		if input == "exit" {
			break
		}
	}

	cmd := exec.Command("reset")
	cmd.Run()
}

func (s *Shell) executor(cmd string) {
}

func (s *Shell) completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix([]prompt.Suggest{}, d.GetWordBeforeCursor(), true)
}
