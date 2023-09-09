package rgx

import "fmt"

func name(s *State) string {
	return fmt.Sprintf("a%p", s)
}

func DumpDotGraphForRegex(regexString string) {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
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
		fmt.Printf("%s [label=\"\"]\n", thisStateName)
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
			label = fmt.Sprintf("%c", char)
		}

		for _, state := range states {
			thatStateName := name(state)
			fmt.Printf("%s -> %s [label=%s]\n", thisStateName, thatStateName, label)
			if _, ok := processedStateForDot[thatStateName]; !ok {
				dot(state, processedStateForDot)
			}
		}
	}
}
