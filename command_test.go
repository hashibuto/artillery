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
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
	if err == nil {
		t.Errorf("Should have thrown an error due to insufficient arguments (missing subtype)")
		return
	}
}

func TestCommandMissingRequiredOption(t *testing.T) {
	input := ""
	cmd := Command{
		Name:        "add",
		Description: "add an animal to the zoo",
		Options: []*Option{
			{
				Name:        "animal",
				ShortName:   'a',
				Description: "type of animal",
				IsRequired:  true,
			},
		},
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
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
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
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
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
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
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
	if err == nil {
		t.Errorf("Should have errored for unrecognized option")
		return
	}
}

func TestCommandGroupArgs(t *testing.T) {
	input := "-n brando --age=27 marlon"
	cmd := Command{
		Name:        "add",
		Description: "add a person to the roster",
		Arguments: []*Argument{
			{
				Name:        "name",
				Description: "person's name",
			},
		},
		Options: []*Option{
			{
				Name:        "nickname",
				Description: "nickname of the person",
				ShortName:   'n',
			},
			{
				Name:        "age",
				Description: "age of the person",
				Type:        Int,
			},
		},
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}
	cmd.Prepare()

	var err error
	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	tokens, err = cmd.CompressTokens(tokens)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCommandStringArrayDefaultOption(t *testing.T) {
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
		Options: []*Option{
			{
				Name:        "attribute",
				Description: "animal attribute",
				ShortName:   'a',
				Default:     []string{"blotchy", "skinny"},
				IsArray:     true,
			},
			{
				Name:        "age",
				Description: "age of the animal",
				Type:        Int,
			},
		},
		OnExecute: func(ns Namespace, processor *Processor) error {
			return nil
		},
	}

	tokens, err := parse(input)
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.Execute(tokens, nil, false)
	if err == nil {
		t.Errorf("Should have errored for unrecognized option")
		return
	}
}
