package rgx

import (
	"fmt"
)

type group struct {
	names []string
	start bool
	end   bool
}

type backreference struct {
	name   string
	target *State
}

type State struct {
	start         bool
	terminal      bool
	endOfText     bool
	startOfText   bool
	transitions   map[uint8][]*State
	groups        []*group
	backreference *backreference
}

const (
	epsilonChar = 0
	startOfText = 1
	endOfText   = 2
	anyChar     = 3
	newline     = 10
)

func toNfa(parseCtx *parsingContext) (*State, *RegexError) {
	startFrom := 0
	endAt := len(parseCtx.tokens) - 1

	token := parseCtx.tokens[startFrom]
	startState, endState, err := tokenToNfa(token, parseCtx, &State{
		transitions: map[uint8][]*State{},
	})

	if err != nil {
		return nil, err
	}

	for i := startFrom + 1; i <= endAt; i++ {
		_, endNext, err := tokenToNfa(parseCtx.tokens[i], parseCtx, endState)
		if err != nil {
			return nil, err
		}
		endState = endNext
	}

	start := &State{
		start: true,
		transitions: map[uint8][]*State{
			epsilonChar: {startState},
		},
		groups: []*group{{
			names: []string{"0"},
			start: true,
			end:   false,
		}},
	}

	end := &State{
		transitions: map[uint8][]*State{},
		terminal:    true,
		groups: []*group{
			{
				names: []string{"0"},
				start: false,
				end:   true,
			},
		},
	}

	endState.transitions[epsilonChar] = append(endState.transitions[epsilonChar], end)

	return start, nil
}

func tokenToNfa(token regexToken, parseCtx *parsingContext, startFrom *State) (*State, *State, *RegexError) {
	switch token.tokenType {
	case literal:
		value := token.value.(uint8)
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[value] = []*State{to}
		return startFrom, to, nil
	case quantifier:
		return handleQuantifierToToken(token, parseCtx, startFrom)
	case wildcard:
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[anyChar] = []*State{to}
		return startFrom, to, nil
	case or:
		values := token.value.([]regexToken)
		_, end1, err := tokenToNfa(values[0], parseCtx, startFrom)
		if err != nil {
			return nil, nil, err
		}
		_, end2, err := tokenToNfa(values[1], parseCtx, startFrom)
		if err != nil {
			return nil, nil, err
		}

		to := &State{
			transitions: map[uint8][]*State{},
		}

		end1.transitions[epsilonChar] = append(end1.transitions[epsilonChar], to)
		end2.transitions[epsilonChar] = append(end2.transitions[epsilonChar], to)

		return startFrom, to, nil
	case groupCaptured:
		v := token.value.(groupTokenPayload)

		// concatenate all the elements in the group
		start, end, err := tokenToNfa(v.tokens[0], parseCtx, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		for i := 1; i < len(v.tokens); i++ {
			_, endNext, err := tokenToNfa(v.tokens[i], parseCtx, end)
			if err != nil {
				return nil, nil, err
			}
			end = endNext
		}
		// concatenation ends

		groupNameNumeric := fmt.Sprintf("%d", parseCtx.nextGroup())
		groupNameUserSet := v.name

		groupNames := []string{groupNameNumeric}
		parseCtx.capturedGroups[groupNameNumeric] = true
		if groupNameUserSet != "" {
			groupNames = append(groupNames, groupNameUserSet)
			parseCtx.capturedGroups[groupNameUserSet] = true
		}

		if startFrom.groups != nil {
			startFrom.groups = append(startFrom.groups, &group{
				names: groupNames,
				start: true,
			})
		} else {
			startFrom.groups = []*group{{
				names: groupNames,
				start: true,
			}}
		}

		if end.groups != nil {
			end.groups = append(end.groups, &group{
				names: groupNames,
				end:   true,
			})
		} else {
			end.groups = []*group{{
				names: groupNames,
				end:   true,
			}}
		}

		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], start)
		return startFrom, end, nil
	case groupUncaptured:
		values := token.value.([]regexToken)

		if len(values) == 0 {
			end := &State{
				transitions: map[uint8][]*State{},
			}

			startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], startFrom)
			return startFrom, end, nil
		}

		start, end, err := tokenToNfa(values[0], parseCtx, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		for i := 1; i < len(values); i++ {
			_, endNext, err := tokenToNfa(values[i], parseCtx, end)

			if err != nil {
				return nil, nil, err
			}

			end = endNext
		}

		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], start)
		return startFrom, end, nil
	case bracket:
		to := &State{
			transitions: map[uint8][]*State{},
		}

		constructTokens := token.value.(map[uint8]bool)
		for ch := range constructTokens {
			startFrom.transitions[ch] = []*State{to}
		}

		return startFrom, to, nil
	case bracketNot:
		to := &State{
			transitions: map[uint8][]*State{},
		}

		deadEnd := &State{
			transitions: map[uint8][]*State{},
		}

		constructTokens := token.value.(map[uint8]bool)
		for ch := range constructTokens {
			startFrom.transitions[ch] = []*State{deadEnd}
		}
		startFrom.transitions[anyChar] = []*State{to}

		return startFrom, to, nil
	case textBeginning:
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.startOfText = true
		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], to)
		return startFrom, to, nil
	case textEnd:
		startFrom.endOfText = true
		return startFrom, startFrom, nil
	case backReference:
		groupName := token.value.(string)
		if _, ok := parseCtx.capturedGroups[groupName]; !ok {
			return nil, nil, &RegexError{
				Code:    CompilationError,
				Message: fmt.Sprintf("Group (%s) does not exist", groupName),
			}
		}
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.backreference = &backreference{
			name:   groupName,
			target: to,
		}

		return startFrom, to, nil
	default:
		return nil, nil, &RegexError{
			Code:    CompilationError,
			Message: fmt.Sprintf("unrecognized token: %+v", token),
		}
	}
}

func handleQuantifierToToken(token regexToken, parseCtx *parsingContext, startFrom *State) (*State, *State, *RegexError) {
	payload := token.value.(quantifierPayload)
	// the minimum amount of time the NFA needs to repeat
	min := payload.min
	// the maximum amount of time the NFA needs to repeat
	max := payload.max

	to := &State{
		transitions: map[uint8][]*State{},
	}

	if min == 0 {
		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], to)
	}

	// how many times should the NFA be generated in the bigger state machine
	var total int

	if max != quantifierInfinity {
		total = max
	} else {
		if min == 0 {
			total = 1 // we need to at least create this NFA once, even if we require it 0 times
		} else {
			total = min
		}
	}
	var value regexToken
	if token.tokenType == quantifier {
		value = token.value.(quantifierPayload).value.([]regexToken)[0]
	} else {
		value = token.value.([]regexToken)[0]
	}
	previousStart, previousEnd, err := tokenToNfa(value, parseCtx, &State{
		transitions: map[uint8][]*State{},
	})

	if err != nil {
		return nil, nil, err
	}

	startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], previousStart)

	// starting from 2, because the one above is the first one
	for i := 2; i <= total; i++ {
		// the same NFA needs to be generated 'total' times
		start, end, err := tokenToNfa(value, parseCtx, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		// connect the end of the previous one to the start of this one
		previousEnd.transitions[epsilonChar] = append(previousEnd.transitions[epsilonChar], start)

		// keep track of the previous NFA's entry and exit states
		previousStart = start
		previousEnd = end

		// after the minimum required amount of repetitions
		// the rest must be optional, thus we add an epsilon transition
		// to the start of each NFA so that we can skip them if needed
		if i > min {
			start.transitions[epsilonChar] = append(start.transitions[epsilonChar], to)
		}
	}

	previousEnd.transitions[epsilonChar] = append(previousEnd.transitions[epsilonChar], to)
	if max == quantifierInfinity {
		to.transitions[epsilonChar] = append(to.transitions[epsilonChar], previousStart)
	}
	return startFrom, to, nil
}
