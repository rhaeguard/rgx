package rgx

import (
	"fmt"
	"os"
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
		{"[0-c-^[_$hello]", "heo", true},
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
		{`([0-9])\1?hi`, "h2hi", true},
		{`([0-9])([a-d](hello))\1`, "bazoo23", false},
		{`(dog)-(cat)-\2-\1`, "nonsensedog-cat-cat-dognonsense", true},
		{`(?<anim>cat)-\k<anim>`, "nonsensedog-cat-cat-dognonsense", true},
		{`(?<letter>[cxv])-[a-z]+-\k<letter>`, "c-abcd-c", true},
		{`(?<letter>[cxv])-[a-z]+-\k<letter>`, "c-abcd-d", false},
		// capturing groups, numeric groups and named groups: example with a phone number format
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1[0-9]{2}`, `123-678-99-32`, true},
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `123 678 99 32`, true},
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `123 678 9932`, true},
		{`[0-9]{3}(-| |)?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `1236789932`, true},
		{`[0-9]{3}(|-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `1236789932`, true},
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
		// escape chars
		{`\\\^\$\.\|\?\*\+\(\)\{\}-hello`, `\^$.|?*+(){}-hello`, true},
		{`[[\]-]+`, `]-[]-[]-[[]]--[]`, true},
		{`[[\]-]+$`, `]-[]-[]-[[]]--[]\`, false},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%t", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			result, err := Check(test.regexString, test.input)
			if err != nil {
				t.Errorf(err.Error())
			}
			if test.expected != result.Matches {
				t.Errorf("test %s failed", testName)
			}
		})
	}
}

func TestFindMatches(t *testing.T) {
	var data = []struct {
		regexString, input string
		expected           []map[string]string
	}{
		{`[0-9]{3}-[0-9]{3}-[0-9]{2}-[0-9]{2}`, `hi 123-678-99-32 is my number, so is 239-987-63-21.`, []map[string]string{
			{"0": "123-678-99-32"},
			{"0": "239-987-63-21"},
		}},
		// multiline extracts
		{`[0-9]{3}-[0-9]{3}-[0-9]{2}-[0-9]{2}$`, "hi 123-678-99-32\n is my number, so is 239-987-63-21", []map[string]string{
			{"0": "123-678-99-32"},
			{"0": "239-987-63-21"},
		}},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%v", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			pattern, err := Compile(test.regexString)
			if err != nil {
				t.Fatalf(err.Error())
			}
			results := pattern.FindMatches(test.input)
			if len(results) != len(test.expected) {
				t.Fatalf("must have expected number of results: expected %d got %d", len(test.expected), len(results))
			}
			for i, expected := range test.expected {
				for k, v := range expected {
					if results[i].Groups[k] != v {
						t.Fatalf("expected '%s' got: '%s'", v, results[i].Groups[k])
					}
				}
			}
		})
	}
}

func TestFindMatchesInTextFile(t *testing.T) {
	bytes, err := os.ReadFile("lib_testdata")
	if err != nil {
		t.Fatalf("could not open the file: %s", err.Error())
	}
	content := string(bytes)

	var data = []struct {
		regexString             string
		expectedOccurrenceCount int
	}{
		{`door`, 33},
		{`door `, 14},
		{`[a-z]+-[a-z]+`, 121},
	}

	for _, test := range data {
		pattern, err := Compile(test.regexString)
		if err != nil {
			t.Fatalf(err.Error())
		}
		testName := fmt.Sprintf("alice in wonderland: regex='%s'", test.regexString)
		t.Run(testName, func(t *testing.T) {
			results := pattern.FindMatches(content)
			if len(results) != test.expectedOccurrenceCount {
				t.Fatalf("expected %d, got: %d", test.expectedOccurrenceCount, len(results))
			}
		})
	}

}

func TestCheckForDev(t *testing.T) {
	var data = []struct {
		regexString, input string
		expected           bool
	}{
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1[0-9]{2}`, `123-678-99-32`, true},
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `123 678 99 32`, true},
		{`[0-9]{3}(-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `123 678 9932`, true},
		{`[0-9]{3}(-| |)?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `1236789932`, true},
		{`[0-9]{3}(|-| )?[0-9]{3}\1[0-9]{2}\1?[0-9]{2}`, `1236789932`, true},
	}

	for _, test := range data {
		testName := fmt.Sprintf("%s-%s-%t", test.regexString, test.input, test.expected)
		t.Run(testName, func(t *testing.T) {
			dumpDotGraphForRegex(test.regexString)
			result, err := Check(test.regexString, test.input)
			if err != nil {
				t.Errorf(err.Error())
			}
			if test.expected != result.Matches {
				_ = fmt.Errorf("test %s failed", testName)
				t.Fail()
			}
		})
	}
}
