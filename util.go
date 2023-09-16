package rgx

import (
	"fmt"
	"strings"
)

func name(s *State) string {
	return fmt.Sprintf("a%p", s)
}

func sliceContains(slice []string, element string) bool {
	for _, el := range slice {
		if el == element {
			return true
		}
	}
	return false
}

func DumpDotGraphForRegex(regexString string) {
	memory := parsingContext{
		pos:            0,
		tokens:         []regexToken{},
		capturedGroups: map[string]bool{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)

	fmt.Printf("digraph G {\n")
	dot(nfaEntry, map[string]bool{})
	fmt.Printf("}\n")
}

// generates a dot graph
func dot(s *State, processedStateForDot map[string]bool) {
	thisStateName := name(s)
	processedStateForDot[thisStateName] = true

	if !s.start || !s.terminal {
		if s.group != nil {
			if s.group.start {
				fmt.Printf("%s [label=\"[%s\"]\n", thisStateName, strings.Join(s.group.names[:], ":"))
			} else {
				fmt.Printf("%s [label=\"%s]\"]\n", thisStateName, strings.Join(s.group.names[:], ":"))
			}
		} else {
			fmt.Printf("%s [label=\"\"]\n", thisStateName)
		}
	}

	if s.start {
		fmt.Printf("%s [label=start]\n", thisStateName)
	}

	if s.terminal {
		fmt.Printf("%s [peripheries=2]\n", thisStateName)
	}

	if s.startOfText {
		fmt.Printf("%s [color=red,style=filled]\n", thisStateName)
	}

	if s.endOfText {
		fmt.Printf("%s [color=blue,style=filled]\n", thisStateName)
	}

	for char, states := range s.transitions {
		var label string
		if char == AnyChar {
			label = "any"
		} else if char == EpsilonChar {
			label = "ε"
		} else {
			label = fmt.Sprintf("\"%c\"", char)
		}

		for _, state := range states {
			thatStateName := name(state)
			fmt.Printf("%s -> %s [label=%s]\n", thisStateName, thatStateName, label)
			if _, ok := processedStateForDot[thatStateName]; !ok {
				dot(state, processedStateForDot)
			}
		}
	}

	if s.backreference != nil {
		thatStateName := name(s.backreference.target)
		fmt.Printf("%s -> %s [label=\"g%s\"]\n", thisStateName, thatStateName, s.backreference.name)
		if _, ok := processedStateForDot[thatStateName]; !ok {
			dot(s.backreference.target, processedStateForDot)
		}
	}
}
