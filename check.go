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

// get the next state given the 'ch' as an input
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

// checks if the inputString is accepted by this NFA
// pos - starting position in the string
// started - have we started matching characters? it's useful when we need to skip characters before starting to match
// ctx - the context for this particular check, the groups, etc.
func (s *State) check(inputString string, pos int, started bool, ctx *regexCheckContext) bool {
	if s.groups != nil {
		// if this state has groups associated with it
		// go through each group
		for _, capturedGroup := range s.groups {
			// if it's a start of a group
			if capturedGroup.start {
				c := &capture{
					start: pos,
					end:   -1,
				}
				// for each name of the captured group
				// add the capture object
				// a group can have 2 different names: numeric (\1) and user-set (\k<animal>)
				for _, groupName := range capturedGroup.names {
					ctx.groups[groupName] = c
				}
			}
		}
	}

	if s.groups != nil {
		// if this state has groups associated with it
		// go through each group
		for _, capturedGroup := range s.groups {
			// if the group ends
			if capturedGroup.end {
				// for each name of the captured group
				// set the end to the current position
				// a group can have 2 different names: numeric (\1) and user-set (\k<animal>)
				for _, groupName := range capturedGroup.names {
					ctx.groups[groupName].end = pos
				}
			}
		}
	}

	currentChar := getChar(inputString, pos)

	// if it needs to be the end of the text, and it isn't
	// or if it needs to be the start of the text and it isn't
	if (s.endOfText && currentChar != EndOfText) || (s.startOfText && currentChar != StartOfText) {
		return false
	}

	if s.terminal {
		return true
	}

	// if there's a backreference transition
	if s.backreference != nil {
		// get the captured reference
		captured, found := ctx.groups[s.backreference.name]
		if !found {
			return false
		}
		// get the string value of it
		capturedString := captured.string(inputString)
		size := len(capturedString)
		backreferenceCheckFailed := false
		for i := 0; i < size; i++ {
			// see if matches with the next set of characters
			if pos >= len(inputString) || inputString[pos+i] != capturedString[i] {
				backreferenceCheckFailed = true
				break
			}
		}
		if !backreferenceCheckFailed {
			return s.backreference.target.check(inputString, pos+size, true, ctx)
		}
		// backreference check failed, let's see if
		// there are any other transitions we can use
	}

	nextState := s.nextStateWith(currentChar)
	// if there are no transitions for the current char as is
	// then see if there's a transition for any char, i.e. dot (.) sign
	if nextState == nil && currentChar != EndOfText {
		nextState = s.nextStateWith(AnyChar)
	}

	result := nextState != nil && nextState.check(inputString, pos+1, true, ctx)
	for _, state := range s.transitions[EpsilonChar] {
		// we need to evaluate all the epsilon transitions
		// because there's a chance that we'll finish early
		// while there's still more to process
		result = state.check(inputString, pos, true, ctx) || result
		result = (currentChar == StartOfText && state.check(inputString, pos+1, true, ctx)) || result
	}

	if result {
		return true
	}

	// if we haven't started matching,
	// then we need to move on to the next character
	// while staying in the same state
	if !started && pos+1 < len(inputString) {
		return s.check(inputString, pos+1, false, ctx)
	}

	return false
}

type Result struct {
	matches bool
	groups  map[string]string
}

type capture struct {
	start int
	end   int
}

func (c *capture) string(inputString string) string {
	s := c.start
	e := c.end

	if s < 0 {
		s = 0
	}

	if e > len(inputString) || e == -1 {
		e = len(inputString)
	}

	return inputString[s:e]
}

type regexCheckContext struct {
	groups map[string]*capture
}
