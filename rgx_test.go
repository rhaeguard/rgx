package rgx

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	var data = []struct {
		regexString, input string
		expected           bool
	}{
		// optionals
		{"a?b?c?", "abc", true},
		{"a?b?c?", "cd", false},
		{"a?b?c?", "cdddd", false},
		{"a?b?c?", "c", true},
		{"a?b?c?", "bc", true},
		{"a?b?c?", "", true},
		{"colou?r", "color", true},
		{"colou?r", "colour", true},
		// basic groups
		{"gr(a|e)y", "grey", true},
		{"gr(a|e)y", "gray", true},
		{"gr(a|e)y", "gruy", false},
		// quantifiers
		{"hel+o", "helo", true},
		{"hel+o", "hellllllo", true},
		{"hel+o", "helllllloooooo", false},
		{"hel+o", "heo", false},
		{"hel*o", "helo", true},
		{"hel*o", "hellllllo", true},
		{"hel*o", "helllllloooooo", false},
		{"hel*o", "heo", true},
		// quantifiers with groups
		{"he(ya)*o", "heo", true},
		{"he(ya)*o", "heyao", true},
		{"he(ya)*o", "heyayao", true},
		{"he(ya)*o", "heyayayo", false},
		{"he(ya)*o", "heyayaya", false},
		{"he(ya)+o", "heo", false},
		{"he(ya)+o", "heyao", true},
		{"he(ya)+o", "heyayao", true},
		{"he(ya)+o", "heyayayo", false},
		{"he(ya)+o", "heyayaya", false},
		// wildcard
		{"h.i", "hxi", true},
		{"h.i", "hxxxi", false},
		// wildcard and quantifiers
		{"h.+i", "hxxxi", true},
		{"h.*i", "hi", true},
		{"hi.*", "hi", true},
		{"hi.*", "hixxxx", true},
		{"hi.*k", "hixxxz", false},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%t", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			if test.expected != Check(test.regexString, test.input) {
				fmt.Errorf("test %s failed", testName)
				t.Fail()
			}
		})
	}
}
