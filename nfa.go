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
	EpsilonChar = 0
	StartOfText = 1
	EndOfText   = 2
	AnyChar     = 3
)

func toNfa(memory *parsingContext) (*State, *RegexError) {
	startFrom := 0
	endAt := len(memory.tokens) - 1

	token := memory.tokens[startFrom]
	startState, endState, err := tokenToNfa(token, memory, &State{
		transitions: map[uint8][]*State{},
	})

	if err != nil {
		return nil, err
	}

	for i := startFrom + 1; i <= endAt; i++ {
		_, endNext, err := tokenToNfa(memory.tokens[i], memory, endState)
		if err != nil {
			return nil, err
		}
		endState = endNext
	}

	start := &State{
		start: true,
		transitions: map[uint8][]*State{
			EpsilonChar: {startState},
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

	endState.transitions[EpsilonChar] = append(endState.transitions[EpsilonChar], end)

	return start, nil
}

func tokenToNfa(token regexToken, memory *parsingContext, startFrom *State) (*State, *State, *RegexError) {
	switch token.tokenType {
	case Literal:
		value := token.value.(uint8)
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[value] = []*State{to}
		return startFrom, to, nil
	case Quantifier:
		return handleQuantifierToToken(token, memory, startFrom)
	case Wildcard:
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.transitions[AnyChar] = []*State{to}

		return startFrom, to, nil
	case Or:
		values := token.value.([]regexToken)
		_, end1, err := tokenToNfa(values[0], memory, startFrom)
		if err != nil {
			return nil, nil, err
		}
		_, end2, err := tokenToNfa(values[1], memory, startFrom)
		if err != nil {
			return nil, nil, err
		}

		to := &State{
			transitions: map[uint8][]*State{},
		}

		end1.transitions[EpsilonChar] = append(end1.transitions[EpsilonChar], to)
		end2.transitions[EpsilonChar] = append(end2.transitions[EpsilonChar], to)

		return startFrom, to, nil
	case Group:
		v := token.value.(groupTokenPayload)

		// concatenate all the elements in the group
		start, end, err := tokenToNfa(v.tokens[0], memory, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		for i := 1; i < len(v.tokens); i++ {
			_, endNext, err := tokenToNfa(v.tokens[i], memory, end)
			if err != nil {
				return nil, nil, err
			}
			end = endNext
		}
		// concatenation ends

		groupNameNumeric := fmt.Sprintf("%d", memory.nextGroup())
		groupNameUserSet := v.name

		groupNames := []string{groupNameNumeric}
		memory.capturedGroups[groupNameNumeric] = true
		if groupNameUserSet != "" {
			groupNames = append(groupNames, groupNameUserSet)
			memory.capturedGroups[groupNameUserSet] = true
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

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], start)
		return startFrom, end, nil
	case GroupUncaptured:
		values := token.value.([]regexToken)

		start, end, err := tokenToNfa(values[0], memory, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		for i := 1; i < len(values); i++ {
			_, endNext, err := tokenToNfa(values[i], memory, end)

			if err != nil {
				return nil, nil, err
			}

			end = endNext
		}

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], start)
		return startFrom, end, nil
	case Bracket:
		constructTokens := token.value.([]regexToken)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			startFrom.transitions[ch] = []*State{to}
		}

		return startFrom, to, nil
	case BracketNot:
		constructTokens := token.value.([]regexToken)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		deadEnd := &State{
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			startFrom.transitions[ch] = []*State{deadEnd}
		}
		startFrom.transitions[AnyChar] = []*State{to}

		return startFrom, to, nil
	case TextBeginning:
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.startOfText = true
		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
		return startFrom, to, nil
	case TextEnd:
		startFrom.endOfText = true
		return startFrom, startFrom, nil
	case Backreference:
		groupName := token.value.(string)
		if _, ok := memory.capturedGroups[groupName]; !ok {
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

func handleQuantifierToToken(token regexToken, memory *parsingContext, startFrom *State) (*State, *State, *RegexError) {
	payload := token.value.(quantifier)
	// the minimum amount of time the NFA needs to repeat
	min := payload.min
	// the maximum amount of time the NFA needs to repeat
	max := payload.max

	to := &State{
		transitions: map[uint8][]*State{},
	}

	if min == 0 {
		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
	}

	// how many times should the NFA be generated in the bigger state machine
	var total int

	if max != QuantifierInfinity {
		total = max
	} else {
		if min == 0 {
			total = 1 // we need to at least create this NFA once, even if we require it 0 times
		} else {
			total = min
		}
	}
	var value regexToken
	if token.tokenType == Quantifier {
		value = token.value.(quantifier).value.([]regexToken)[0]
	} else {
		value = token.value.([]regexToken)[0]
	}
	previousStart, previousEnd, err := tokenToNfa(value, memory, &State{
		transitions: map[uint8][]*State{},
	})

	if err != nil {
		return nil, nil, err
	}

	startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], previousStart)

	// starting from 2, because the one above is the first one
	for i := 2; i <= total; i++ {
		// the same NFA needs to be generated 'total' times
		start, end, err := tokenToNfa(value, memory, &State{
			transitions: map[uint8][]*State{},
		})

		if err != nil {
			return nil, nil, err
		}

		// connect the end of the previous one to the start of this one
		previousEnd.transitions[EpsilonChar] = append(previousEnd.transitions[EpsilonChar], start)

		// keep track of the previous NFA's entry and exit states
		previousStart = start
		previousEnd = end

		// after the minimum required amount of repetitions
		// the rest must be optional, thus we add an epsilon transition
		// to the start of each NFA so that we can skip them if needed
		if i > min {
			start.transitions[EpsilonChar] = append(start.transitions[EpsilonChar], to)
		}
	}

	previousEnd.transitions[EpsilonChar] = append(previousEnd.transitions[EpsilonChar], to)
	if max == QuantifierInfinity {
		to.transitions[EpsilonChar] = append(to.transitions[EpsilonChar], previousStart)
	}
	return startFrom, to, nil
}
