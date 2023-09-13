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

func (s *State) check(inputString string, pos int, started bool, ctx *regexCheckContext) bool {
	if s.group != nil && s.group.start {
		ctx.groups[s.group.name] = &capture{
			start: pos,
			end:   -1,
		} // start the group
	}

	if s.group != nil && s.group.end {
		ctx.groups[s.group.name].end = pos
	}

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

	if nextState != nil && nextState.check(inputString, pos+1, true, ctx) {
		return true
	}

	epsilonTransitionsResult := false
	for _, state := range s.transitions[EpsilonChar] {
		// we need to evaluate all the epsilon transitions
		// because there's a chance that we'll finish early
		// while there's still more to process
		epsilonTransitionsResult = state.check(inputString, pos, true, ctx) || epsilonTransitionsResult
		epsilonTransitionsResult = (currentChar == StartOfText && state.check(inputString, pos+1, true, ctx)) || epsilonTransitionsResult
	}

	if epsilonTransitionsResult {
		return true
	}

	if !started && pos+1 < len(inputString) {
		return s.check(inputString, pos+1, false, ctx)
	}

	return false
}

func Check(regexString string, inputString string) Result {
	parseContext := parsingContext{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &parseContext)
	nfaEntry := toNfa(&parseContext)

	checkContext := &regexCheckContext{
		groups: map[string]*capture{},
	}
	result := nfaEntry.check(inputString, -1, nfaEntry.startOfText, checkContext)

	// prepare the result
	groups := map[string]string{}

	if result {
		// extract strings from the groups
		for group, captured := range checkContext.groups {
			groups[group] = captured.string(inputString)
		}
	}

	return Result{
		matches: result,
		groups:  groups,
	}
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
