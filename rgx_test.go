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
		{"a?b?c?", "abc", true},
		{"a?b?c?", "cd", false},
		{"a?b?c?", "cdddd", false},
		{"a?b?c?", "c", true},
		{"a?b?c?", "bc", true},
		{"a?b?c?", "", true},
		{"colou?r", "color", true},
		{"colou?r", "colour", true},
		{"gr(a|e)y", "grey", true},
		{"gr(a|e)y", "gray", true},
		{"gr(a|e)y", "gruy", false},
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
