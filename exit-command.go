package artillery

func makeExitCommand() *Command {
	return &Command{
		Name:        "exit",
		Description: "exit the shell",
		OnExecute: func(ns Namespace, processor *Processor) error {
			processor.Shell().Shutdown()
			return nil
		},
	}
}
