package rgx

import "fmt"

func getChar(input string, pos int) uint8 {
	if pos >= 0 && pos < len(input) {
		return input[pos]
	}

	if pos >= len(input) {
		return EndOfText
	}

	return StartOfText
}

func (s *State) check(inputString string, pos int, started bool) bool {
	current := getChar(inputString, pos)

	if s.endOfText && current != EndOfText {
		return false
	}

	if s.startOfText && current != StartOfText {
		return false
	}

	if s.terminal {
		return true
	}

	realTransitions := s.transitions[current]

	// if there are no transitions for the current char as is
	// then see if there's a transition for any char, i.e. dot (.) sign
	if len(realTransitions) == 0 && current != EndOfText {
		realTransitions = s.transitions[AnyChar]
	}

	for i := range realTransitions {
		state := realTransitions[i]
		if state.check(inputString, pos+1, true) {
			return true
		}
	}

	epsilonTransitions := s.transitions[EpsilonChar]
	for i := range epsilonTransitions {
		state := epsilonTransitions[i]
		if state.check(inputString, pos, true) {
			return true
		}
	}

	if !started && pos+1 < len(inputString) {
		return s.check(inputString, pos+1, false)
	}

	return false
}

func Check(regexString string, inputString string) bool {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)
	fmt.Printf("%+v\n", nfaEntry)
	return nfaEntry.check(inputString, -1, nfaEntry.startOfText)
}
