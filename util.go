package rgx

import (
	"fmt"
	"strings"
)

func name(s *State) string {
	return fmt.Sprintf("a%p", s)
}

func dumpDotGraphForRegex(regexString string) {
	memory := parsingContext{
		pos:            0,
		tokens:         []regexToken{},
		capturedGroups: map[string]bool{},
	}
	regexError := parse(regexString, &memory)
	if regexError != nil {
		panic(regexError.Error())
	}

	nfaEntry, regexError := toNfa(&memory)
	if regexError != nil {
		panic(regexError.Error())
	}

	fmt.Printf("digraph G {\n")
	dot(nfaEntry, map[string]bool{})
	fmt.Printf("}\n")
}

// generates a dot graph
func dot(s *State, processedStateForDot map[string]bool) {
	thisStateName := name(s)
	processedStateForDot[thisStateName] = true

	if !s.start || !s.terminal {
		if s.groups != nil {
			startGroups := ""
			endGroups := ""
			for _, capturedGroup := range s.groups {
				if capturedGroup.start {
					startGroups += "[" + strings.Join(capturedGroup.names[:], ":") + "]"
				} else {
					endGroups += "[" + strings.Join(capturedGroup.names[:], ":") + "]"
				}
				fmt.Printf("%s [label=\"%s--%s\"]\n", thisStateName, startGroups, endGroups)
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
		if char == anyChar {
			label = "any"
		} else if char == epsilonChar {
			label = "ε"
		} else if char == '\\' {
			label = "backslash"
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
