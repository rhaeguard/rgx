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
		{"a?b?c?$", "abc", true},
		{"a?b?c?$", "cd", false},
		{"a?b?c?$", "cdddd", false},
		{"a?b?c?$", "c", true},
		{"a?b?c?$", "bc", true},
		{"a?b?c?$", "", true},
		{"^a?b?c?", "", true},
		{"colou?r", "color", true},
		{"colou?r", "colour", true},
		// basic groups
		{"gr(a|e)y", "grey", true},
		{"gr(a|e)y", "gray", true},
		{"gr(a|e)y", "gruy", false},
		// quantifiers
		{"hel+o", "helo", true},
		{"hel+o", "hellllllo", true},
		{"hel+o$", "helllllloooooo", false},
		{"hel+o", "heo", false},
		{"hel*o", "helo", true},
		{"hel*o", "hellllllo", true},
		{"hel*o$", "helllllloooooo", false},
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
		// brackets and ranges
		{"h[ae-ux]llo", "hello", true},
		{"h[ae-ux]llo", "hallo", true},
		{"h[ae-ux]llo", "hmllo", true},
		{"h[ae-ux]llo", "hullo", true},
		{"h[ae-ux]llo", "hxllo", true},
		{"h[ae-ux]llo", "hwllo", false},
		{"199[0-3]", "1990", true},
		{"199[0-3]", "1991", true},
		{"199[0-3]", "1992", true},
		{"199[0-3]", "1993", true},
		// brackets and ranges with negation
		{"h[^ae-ux]llo", "hello", false},
		{"h[^ae-ux]llo", "hallo", false},
		{"h[^ae-ux]llo", "hmllo", false},
		{"h[^ae-ux]llo", "hullo", false},
		{"h[^ae-ux]llo", "hxllo", false},
		{"h[^ae-ux]llo", "hwllo", true},
		{"h[^ae-ux]llo", "hzllo", true},
		{"h[^ae-ux]llo", "h.llo", true},
		{"h[^ae-ux]llo", "h@llo", true},
		{"h[^ae-ux]llo", "hllo", false},
		{"199[^0-3]", "1990", false},
		{"199[^0-3]", "1991", false},
		{"199[^0-3]", "1992", false},
		{"199[^0-3]", "1993", false},
		// brackets and ranges + optionals
		{"199[0-3]?", "1993", true},
		{"199[0-3]?", "199", true},
		// or/alternative
		{"(gray|grey)", "gray", true},
		{"(gray|grey)", "grey", true},
		{"(gray|grey)", "gryy", false},
		{"((gray|gruy)|grey)", "grey", true},
		{"((gray|gruy)|grey)", "gray", true},
		{"((gray|gruy)|grey)", "gruy", true},
		{"((gray|gruy)|grey)", "gryy", false},
		{"(gray|gruy|grey)", "gruy", true},
		{"(gray|gruy|grey)", "gray", true},
		{"(gray|gruy|grey)", "grey", true},
		{"(gray|gruy|grey)", "greyish", true},
		// start and end of string
		{"(ha$|^hi)", "aha", true},
		{"(ha$|^hi)", "hill", true},
		{"(ha$|^hi)", "ahaa", false},
		{"(ha$|^hi)", "ahii", false},
		// capturing groups, numeric groups and named groups
		{"([0-9])\\1?hi", "h2hi", true},
		{"([0-9])([a-d](hello))\\1", "bazoo23", false},
		{"(dog)-(cat)-\\2-\\1", "nonsensedog-cat-cat-dognonsense", true},
		{"(?<anim>cat)-\\k<anim>", "nonsensedog-cat-cat-dognonsense", true},
		{"(?<letter>[cxv])-[a-z]+-\\k<letter>", "c-abcd-c", true},
		{"(?<letter>[cxv])-[a-z]+-\\k<letter>", "c-abcd-d", false},
		// quantifiers
		{"(hi){2,3}", "hi hihi hihi", true},
		{`ab{0,}bc`, `abbbbc`, true},
		{`ab{1,}bc`, `abq`, false},
		{`ab{1,}bc`, `abbbbc`, true},
		{`ab{1,3}bc`, `abbbbc`, true},
		{`ab{3,4}bc`, `abbbbc`, true},
		{`ab{4,5}bc`, `abbbbc`, false},
		{`ab{0,1}bc`, `abc`, true},
		{`ab{0,1}c`, `abc`, true},
		{`a{1,}b{1,}c`, `aabbabc`, true},
		{`(a+|b){0,}`, `ab`, true},
		{`(a+|b){1,}`, `ab`, true},
		{`(a+|b){0,1}`, `ab`, true},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%t", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			if test.expected != Check(test.regexString, test.input).matches {
				_ = fmt.Errorf("test %s failed", testName)
				t.Fail()
			}
		})
	}
}

func TestCheckForDev(t *testing.T) {
	var data = []struct {
		regexString, input string
		expected           bool
	}{
		{"he(ya)*o", "heo", true},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%t", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			dumpDotGraphForRegex(test.regexString)
			result := Check(test.regexString, test.input)
			if test.expected != result.matches {
				_ = fmt.Errorf("test %s failed", testName)
				t.Fail()
			}
		})
	}
}
