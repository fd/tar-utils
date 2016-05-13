package tarbuild

import (
	"reflect"
	"testing"
)

func Test_parseConf(t *testing.T) {
	spec := []byte(`

# This is a comment
# another comment

COMMAND arg1 arg2 arg3
COMMAND ["arg1", "arg2", "arg3"]
# Multi line commands
COMMAND \
  arg1 \
	arg2 \
	arg3
COMMAND [
	"arg1",
	"arg2",
	"arg3"
]

	`)

	actual, err := parseConf(spec)
	if err != nil {
		t.Fatal(err)
	}

	expected := &tarSpec{
		Commands: []tarOp{
			{Name: "COMMAND", Args: []string{"arg1", "arg2", "arg3"}},
			{Name: "COMMAND", Args: []string{"arg1", "arg2", "arg3"}},
			{Name: "COMMAND", Args: []string{"arg1", "arg2", "arg3"}},
			{Name: "COMMAND", Args: []string{"arg1", "arg2", "arg3"}},
		},
	}

	t.Logf("expected: %v", expected)
	t.Logf("actual:   %v", actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("did not match")
	}
}
