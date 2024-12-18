This repository has been archived, please consider the project's successor: https://github.com/hashibuto/commander

# artillery
An interactive shell/REPL and CLI parser for golang

Artillery can function both as an interactive shell or REPL, as well as a single command CLI parser.  In interactive mode, it uses [NilShell](https://github.com/hashibuto/nilshell) for the command line editor, completion, and history.

## A basic interactive shell/REPL

```
import (
    "github.com/hashibuto/artillery"
)

	processor := artillery.NewProcessor()

	cmds := []*artillery.Command{
		&artillery.Command{
            Name:        "hello",
            Description: "prints hello to the console",
            Arguments: []*artillery.Argument{
                {
                    Name:        "name",
                    Description: "name of the person to which to say hello",
                },
            },
            Options: []*artillery.Option{
                {
                    Name:        "shout",
                    Description: "shout mode",
                    ShortName:   's',
                    Type:        artillary.Bool
                    Value:       true,
                },
            },
            OnExecute: func(ns artillery.Namespace, processor *artillery.Processor) error {
                var args struct {
                    Name  string
                    Shout bool
                }
                err := artillery.Reflect(ns, &args)
                if err != nil {
                    return err
                }

                message := fmt.Sprintf("hello %s!", args.Name)
                if args.Shout {
                    message = strings.ToUpper(message)
                }
                return nil
            },
        },
	}

	for _, c := range cmds {
		err := processor.AddCommand(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	processor.Shell().ReadUntilTerm()
```

## Parse a single CLI command (non-interactive)

```
import (
    "github.com/hashibuto/artillery"
    "os"
)

helloCmd := &artillery.Command{
    Name:        "hello",
    Description: "prints hello to the console",
    Arguments: []*artillery.Argument{
        {
            Name:        "name",
            Description: "name of the person to which to say hello",
        },
    },
    Options: []*artillery.Option{
        {
            Name:        "shout",
            Description: "shout mode",
            ShortName:   's',
            Type:        artillary.Bool
            Value:       true,
        },
    },
    OnExecute: func(ns artillery.Namespace, processor *artillery.Processor) error {
        var args struct {
            Name  string
            Shout bool
        }
        err := artillery.Reflect(ns, &args)
        if err != nil {
            return err
        }

        message := fmt.Sprintf("hello %s!", args.Name)
        if args.Shout {
            message = strings.ToUpper(message)
        }
        return nil
    },
}

processor := artillery.NewProcessor()
processor.RemoveBuiltins(false)
err := processor.AddCommands(helloCmd)
if err != nil {
    panic(err)
}

err = processor.Process(os.Args[1:])
```

## Special commands / keystrokes
- `clear` clears the terminal
- `!<command>` execs the command ie `!cat /home/user/something` for bash do `!bash -c "cat /home/user/something | grep whatever"`
- `exit` exits
- `<ctrl+r>` reverse search
- `<up>` move up backwards through the command history
- `<down>` move forwards through the command history
- `<ctrl+d>` exit
- `<tab>` autocomplete

## Example
[Example showcasing most features](https://github.com/hashibuto/artillery/blob/master/example/main.go)
