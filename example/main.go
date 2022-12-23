package main

import (
	"log"
	"sort"
	"strconv"

	"github.com/hashibuto/artillery"
	"github.com/hashibuto/artillery/pkg/tg"
)

type Animal struct {
	Type       string
	Age        int
	Attributes []string
}

type Zoo struct {
	Animals []*Animal
}

var TheZoo = &Zoo{
	Animals: []*Animal{},
}

var AnimalGroup = "animal commands"

var animalTypes = []string{"cat", "dog", "chicken", "horse"}

func makeAnimalCommand() *artillery.Command {
	return &artillery.Command{
		Name:        "animal",
		Group:       AnimalGroup,
		Description: "do an animal operation",
		SubCommands: []*artillery.Command{
			{
				Name:        "add",
				Description: "add an animal to the zoo",
				Arguments: []*artillery.Argument{
					{
						Name:        "animal",
						Description: "type of animal",
						MemberOf:    animalTypes,
					},
				},
				Options: []*artillery.Option{
					{
						Name:        "attribute",
						Description: "animal attribute",
						ShortName:   'a',
						IsArray:     true,
					},
					{
						Name:        "age",
						Description: "age of the animal",
						Type:        artillery.Int,
					},
				},
				OnExecute: func(ns artillery.Namespace, processor *artillery.Processor) error {
					var args struct {
						Animal    string
						Age       int
						Attribute []string
					}
					artillery.Reflect(ns, &args)

					animal := &Animal{
						Type:       args.Animal,
						Age:        args.Age,
						Attributes: args.Attribute,
					}
					TheZoo.Animals = append(TheZoo.Animals, animal)
					tg.Println(tg.Green, "Added a ", args.Animal, " to the zoo", tg.Reset)
					return nil
				},
			},
			{
				Name:        "rm",
				Description: "remove an animal from the zoo",
				Arguments: []*artillery.Argument{
					{
						Name:        "animal",
						Description: "type of animal",
						MemberOf:    animalTypes,
					},
				},
				OnExecute: func(ns artillery.Namespace, processor *artillery.Processor) error {
					var args struct {
						Animal string
					}
					artillery.Reflect(ns, &args)

					for idx, a := range TheZoo.Animals {
						if a.Type == args.Animal {
							tg.Println(tg.Green, "Removed one ", args.Animal, " from the zoo", tg.Reset)
							remaining := []*Animal{}
							remaining = append(remaining, TheZoo.Animals[:idx]...)
							remaining = append(remaining, TheZoo.Animals[idx+1:]...)
							break
						}
					}
					return nil
				},
			},
		},
	}
}

func makeListCommand() *artillery.Command {
	return &artillery.Command{
		Name:        "list",
		Group:       AnimalGroup,
		Description: "list all animals in the zoo",
		OnExecute: func(ns artillery.Namespace, processor *artillery.Processor) error {
			counts := map[string]int{}
			animalTypes := []string{}
			for _, animal := range TheZoo.Animals {
				curCount, ok := counts[animal.Type]
				if !ok {
					curCount = 0
				}
				curCount++
				counts[animal.Type] = curCount
			}
			for aType := range counts {
				animalTypes = append(animalTypes, aType)
			}
			sort.Slice(animalTypes, func(i, j int) bool {
				return animalTypes[i] < animalTypes[j]
			})

			table := tg.NewTable("type", "count")
			for _, aType := range animalTypes {
				table.Append(aType, strconv.Itoa(counts[aType]))
			}
			table.Render()

			return nil
		},
	}
}

func main() {
	processor := artillery.NewProcessor()
	processor.DefaultHeading = "uncategorized commands"

	cmds := []func() *artillery.Command{
		makeAnimalCommand,
		makeListCommand,
	}

	for _, c := range cmds {
		err := processor.AddCommand(c())
		if err != nil {
			log.Fatal(err)
		}
	}

	shell := processor.Shell()
	//shell.Prompt = "\033[33martillery \033[34m\033[1m$ \033[0m"
	shell.AutoCompleteSuggestStyle = "\033[32m"
	shell.ReadUntilTerm()
}
