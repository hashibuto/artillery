package artillery

import "testing"

func TestTokenizer(t *testing.T) {
	cmd := "  this  is  \"a test of \" some tokens  "
	tokens, openQuote := tokenize(cmd)
	if openQuote {
		t.Errorf("Reported open quote when none was present")
		return
	}

	if len(tokens) != 5 {
		t.Errorf("Expected 5 tokens but got %d\n%v", len(tokens), tokens)
		return
	}

	expected := []string{"this", "is", "a test of ", "some", "tokens"}
	for idx, token := range tokens {
		if expected[idx] != token {
			t.Errorf("At position %d, expected\n%s\ngot\n%s\n", idx, expected[idx], token)
			return
		}
	}
}

func TestRunOnQuote(t *testing.T) {
	cmd := "this is a 'run-on quote without ending"
	_, openQuote := tokenize(cmd)
	if !openQuote {
		t.Errorf("Expected tokenizer to catch un-terminated quotation mark")
	}
}

func TestQuoteMismatch(t *testing.T) {
	cmd := "this is a 'quote mismatch\""
	_, openQuote := tokenize(cmd)
	if !openQuote {
		t.Errorf("Expected tokenizer to catch quote mismatch")
	}
}

func TestParser(t *testing.T) {
	cmd := "-a -b hello world"
	all, err := parse(cmd)
	if err != nil {
		t.Error(err)
		return
	}
	opts, args, err := group(all)
	if err != nil {
		t.Error(err)
		return
	}

	if len(opts) != 2 {
		t.Errorf("Expected 2 options, got %d", len(opts))
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(args))
	}

	expected := []*OptionInput{
		{
			Name:  "a",
			Value: "",
		},
		{
			Name:  "b",
			Value: "",
		},
	}

	for idx, optInput := range opts {
		exp := expected[idx]
		if optInput.Name != exp.Name || optInput.Value != exp.Value {
			t.Errorf("Expected does not match actual for item %d\nExpected\n%v\nActual\n%v\n", idx, exp, optInput)
			return
		}
	}

	expArgs := []string{"hello", "world"}
	for idx, arg := range args {
		exp := expArgs[idx]
		if exp != arg {
			t.Errorf("Arg %d did not match:  Expected %s, got %s", idx, exp, arg)
			return
		}
	}
}

func TestParser2(t *testing.T) {
	cmd := "-abc hello world"
	all, err := parse(cmd)
	if err != nil {
		t.Error(err)
		return
	}
	opts, args, err := group(all)
	if err != nil {
		t.Error(err)
		return
	}

	if len(opts) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts))
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(args))
	}

	expected := []*OptionInput{
		{
			Name:  "a",
			Value: "",
		},
		{
			Name:  "b",
			Value: "",
		},
		{
			Name:  "c",
			Value: "",
		},
	}

	for idx, optInput := range opts {
		exp := expected[idx]
		if optInput.Name != exp.Name || optInput.Value != exp.Value {
			t.Errorf("Expected does not match actual for item %d\nExpected\n%v\nActual\n%v\n", idx, exp, optInput)
			return
		}
	}

	expArgs := []string{"hello", "world"}
	for idx, arg := range args {
		exp := expArgs[idx]
		if exp != arg {
			t.Errorf("Arg %d did not match:  Expected %s, got %s", idx, exp, arg)
			return
		}
	}
}

func TestParser3(t *testing.T) {
	cmd := "-a=123 -b=344 hello world"
	all, err := parse(cmd)
	if err != nil {
		t.Error(err)
		return
	}
	opts, args, err := group(all)
	if err != nil {
		t.Error(err)
		return
	}

	if len(opts) != 2 {
		t.Errorf("Expected 3 options, got %d", len(opts))
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(args))
	}

	expected := []*OptionInput{
		{
			Name:  "a",
			Value: "123",
		},
		{
			Name:  "b",
			Value: "344",
		},
		{
			Name:  "c",
			Value: "",
		},
	}

	for idx, optInput := range opts {
		exp := expected[idx]
		if optInput.Name != exp.Name || optInput.Value != exp.Value {
			t.Errorf("Expected does not match actual for item %d\nExpected\n%v\nActual\n%v\n", idx, exp, optInput)
			return
		}
	}

	expArgs := []string{"hello", "world"}
	for idx, arg := range args {
		exp := expArgs[idx]
		if exp != arg {
			t.Errorf("Arg %d did not match:  Expected %s, got %s", idx, exp, arg)
			return
		}
	}
}

func TestParser4(t *testing.T) {
	cmd := "--hello=world -a=3"
	all, err := parse(cmd)
	if err != nil {
		t.Error(err)
		return
	}
	opts, args, err := group(all)
	if err != nil {
		t.Error(err)
		return
	}

	if len(opts) != 2 {
		t.Errorf("Expected 2 options, got %d", len(opts))
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}

	expected := []*OptionInput{
		{
			Name:  "hello",
			Value: "world",
		},
		{
			Name:  "a",
			Value: "3",
		},
	}

	for idx, optInput := range opts {
		exp := expected[idx]
		if optInput.Name != exp.Name || optInput.Value != exp.Value {
			t.Errorf("Expected does not match actual for item %d\nExpected\n%v\nActual\n%v\n", idx, exp, optInput)
			return
		}
	}
}

func TestParserQuotes(t *testing.T) {
	cmd := "\"--hello=world hello\" -a=3"
	all, err := parse(cmd)
	if err != nil {
		t.Error(err)
		return
	}
	opts, args, err := group(all)
	if err != nil {
		t.Error(err)
		return
	}

	if len(opts) != 2 {
		t.Errorf("Expected 2 options, got %d", len(opts))
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}

	expected := []*OptionInput{
		{
			Name:  "hello",
			Value: "world hello",
		},
		{
			Name:  "a",
			Value: "3",
		},
	}

	for idx, optInput := range opts {
		exp := expected[idx]
		if optInput.Name != exp.Name || optInput.Value != exp.Value {
			t.Errorf("Expected does not match actual for item %d\nExpected\n%v\nActual\n%v\n", idx, exp, optInput)
			return
		}
	}
}
