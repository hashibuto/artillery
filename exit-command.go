package artillery

func makeExitCommand() *Command {
	return &Command{
		Name:        "exit",
		Description: "exit the shell",
		OnExecute: func(ns Namespace) error {
			shell := GetInstance()
			shell.Exit(0)
			return nil
		},
	}
}
