package main

import (
	"strings"
	"testing"
)

func TestPreprocessArgs(t *testing.T) {
	tests := []struct {
		data1 string
		e1    string
	}{
		// Special case --disable at end of line.
		{"-disable", "-disable"},
		// Simple message.
		{"-disable message", `-disable --disable-message "message"`},
		// Message is not the end of the command.
		{"-disable message -flag", `-disable --disable-message "message" -flag`},
	}
	for i, test := range tests {
		got := preprocessArgs(strings.Split(test.data1, " "))
		result := strings.Join(got, " ")
		if result != test.e1 {
			t.Errorf("%v: expected (%v) got (%v)", i, test.e1, result)
		}
	}
}
