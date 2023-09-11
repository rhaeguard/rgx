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

func (s *State) nextStateWith(ch uint8) *State {
	states := s.transitions[ch]

	size := len(states)

	if size == 0 {
		return nil
	} else if size == 1 {
		return states[0]
	}

	panic(fmt.Sprintf("There must be at most 1 transition, found %d", size))
}

func (s *State) check(inputString string, pos int, started bool) bool {
	currentChar := getChar(inputString, pos)

	if s.endOfText && currentChar != EndOfText {
		return false
	}

	if s.startOfText && currentChar != StartOfText {
		return false
	}

	if s.terminal {
		return true
	}

	nextState := s.nextStateWith(currentChar)
	// if there are no transitions for the current char as is
	// then see if there's a transition for any char, i.e. dot (.) sign
	if nextState == nil && currentChar != EndOfText {
		nextState = s.nextStateWith(AnyChar)
	}

	if nextState != nil && nextState.check(inputString, pos+1, true) {
		return true
	}

	epsilonTransitions := s.transitions[EpsilonChar]
	for i := range epsilonTransitions {
		state := epsilonTransitions[i]
		if state.check(inputString, pos, true) {
			return true
		}
		// if we're at the start of the text, we should try progressing
		if currentChar == StartOfText && state.check(inputString, pos+1, true) {
			return true
		}
	}

	if !started && pos+1 < len(inputString) {
		return s.check(inputString, pos+1, false)
	}

	return false
}

func Check(regexString string, inputString string) bool {
	memory := parsingContext{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)
	return nfaEntry.check(inputString, -1, nfaEntry.startOfText)
}
