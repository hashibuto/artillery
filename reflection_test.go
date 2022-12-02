package artillery

import "testing"

type personInput struct {
	Name    string
	Age     int
	Friends []string
}

func TestReflect(t *testing.T) {
	ns := Namespace{
		"name":    "hello",
		"age":     55,
		"friends": []any{"donnie", "willy", "barnard"},
	}
	person := &personInput{}

	err := Reflect(ns, person)
	if err != nil {
		t.Error(err)
		return
	}

	if person.Name != "hello" {
		t.Errorf("Got incorrect name")
	}

	if person.Age != 55 {
		t.Errorf("Got incorrect age")
	}

	if len(person.Friends) != 3 {
		t.Errorf("Got incorrect friends")
	}
}
