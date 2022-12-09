package artillery

func makeClearCommand() *Command {
	return &Command{
		Name:        "clear",
		Description: "clear the terminal",
		OnExecute: func(ns Namespace, processor *Processor) error {
			processor.nilShell.Clear()
			return nil
		},
	}
}
