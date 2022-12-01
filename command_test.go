package artillery

import (
	"testing"
)

func TestCommandMissingArg(t *testing.T) {
	input := "tiger"
	cmd := Command{
		Name:        "add",
		Description: "add an animal to the zoo",
		Arguments: []*Argument{
			{
				Name:        "animal",
				Description: "type of animal",
			},
			{
				Name:        "subtype",
				Description: "subtype of animal",
			},
		},
		OnExecute: func(ns Namespace) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens)
	if err == nil {
		t.Errorf("Should have thrown an error due to insufficient arguments (missing subtype)")
		return
	}

}

func TestCommandExtraArg(t *testing.T) {
	input := "tiger tigrus"
	cmd := Command{
		Name:        "add",
		Description: "add an animal to the zoo",
		Arguments: []*Argument{
			{
				Name:        "animal",
				Description: "type of animal",
			},
		},
		OnExecute: func(ns Namespace) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens)
	if err == nil {
		t.Errorf("Should have thrown an error due to too many arguments")
		return
	}
}

func TestCommandCorrectNumArgs(t *testing.T) {
	input := "tiger"
	cmd := Command{
		Name:        "add",
		Description: "add an animal to the zoo",
		Arguments: []*Argument{
			{
				Name:        "animal",
				Description: "type of animal",
			},
		},
		OnExecute: func(ns Namespace) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCommandUnrecognizedOption(t *testing.T) {
	input := "-t tiger"
	cmd := Command{
		Name:        "add",
		Description: "add an animal to the zoo",
		Arguments: []*Argument{
			{
				Name:        "animal",
				Description: "type of animal",
			},
		},
		Options: []*Option{
			{
				Name:        "attribute",
				Description: "animal attribute",
				ShortName:   'a',
				IsArray:     true,
			},
			{
				Name:        "age",
				Description: "age of the animal",
				Type:        Int,
			},
		},
		OnExecute: func(ns Namespace) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens)
	if err == nil {
		t.Errorf("Should have errored for unrecognized option")
		return
	}
}
