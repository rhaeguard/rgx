package rgx

import "fmt"

// generates a dot graph
func dot(s *State, processedStateForDot map[string]bool) {
	for char, states := range s.transitions {
		var label string
		if char == AnyChar {
			label = "any"
		} else if char == EpsilonChar {
			label = "ε"
		} else {
			label = fmt.Sprintf("%c", char)
		}

		processedStateForDot[s.name] = true
		fmt.Printf("%s [label=\"\"]\n", s.name)
		if s.startOfText {
			fmt.Printf("%s [color=red,style=filled]\n", s.name)
		}

		if s.endOfText {
			fmt.Printf("%s [color=blue,style=filled]\n", s.name)
		}

		for _, state := range states {
			fmt.Printf("%s -> %s [label=%s]\n", s.name, state.name, label)
			if _, ok := processedStateForDot[state.name]; !ok {
				dot(state, processedStateForDot)
			}
		}
	}
}
