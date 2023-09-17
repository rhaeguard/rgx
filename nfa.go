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
	StartOfText = 1 // ascii: null char
	EndOfText   = 2 // ascii: end of text
	AnyChar     = 3 // ascii: substitute
	EpsilonChar = 0 // ascii: null char
)

func toNfa(memory *parsingContext) *State {
	startFrom := 0
	endAt := len(memory.tokens) - 1

	token := memory.tokens[startFrom]
	startState, endState := tokenToNfa(token, memory, &State{
		transitions: map[uint8][]*State{},
	})

	for i := startFrom + 1; i <= endAt; i++ {
		_, endNext := tokenToNfa(memory.tokens[i], memory, endState)
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

	return start
}

func tokenToNfa(token regexToken, memory *parsingContext, startFrom *State) (*State, *State) {
	switch token.tokenType {
	case Literal:
		value := token.value.(uint8)
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[value] = append(startFrom.transitions[value], to)
		return startFrom, to
	case OneOrMore, NoneOrMore, Optional:
		return handleQuantifierToToken(token, memory, startFrom)
	case Wildcard:
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.transitions[AnyChar] = append(startFrom.transitions[AnyChar], to)

		return startFrom, to
	case Or:
		values := token.value.([]regexToken)
		_, end1 := tokenToNfa(values[0], memory, startFrom)
		_, end2 := tokenToNfa(values[1], memory, startFrom)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		end1.transitions[EpsilonChar] = append(end1.transitions[EpsilonChar], to)
		end2.transitions[EpsilonChar] = append(end2.transitions[EpsilonChar], to)

		return startFrom, to
	case Group:
		v := token.value.([]interface{})
		values := v[0].([]regexToken)

		// concatenate all the elements in the group
		start, end := tokenToNfa(values[0], memory, &State{
			transitions: map[uint8][]*State{},
		})

		for i := 1; i < len(values); i++ {
			_, endNext := tokenToNfa(values[i], memory, end)
			end = endNext
		}
		// concatenation ends

		groupNameNumeric := fmt.Sprintf("%d", memory.nextGroup())
		groupNameUserSet := v[1].(string)

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
		return startFrom, end
	case GroupUncaptured:
		values := token.value.([]regexToken)

		start, end := tokenToNfa(values[0], memory, &State{
			transitions: map[uint8][]*State{},
		})

		for i := 1; i < len(values); i++ {
			_, endNext := tokenToNfa(values[i], memory, end)
			end = endNext
		}

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], start)
		return startFrom, end
	case Bracket:
		constructTokens := token.value.([]regexToken)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			startFrom.transitions[ch] = []*State{to}
		}

		return startFrom, to
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

		return startFrom, to
	case TextBeginning:
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.startOfText = true
		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
		return startFrom, to
	case TextEnd:
		startFrom.endOfText = true
		return startFrom, startFrom
	case Backreference:
		groupName := token.value.(string)
		if _, ok := memory.capturedGroups[groupName]; !ok {
			panic(fmt.Sprintf("Group (%s) does not exist", groupName))
		}
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.backreference = &backreference{
			name:   groupName,
			target: to,
		}

		return startFrom, to
	default:
		panic(fmt.Sprintf("unrecognized token: %+v", token))
	}
}

func handleQuantifierToToken(token regexToken, memory *parsingContext, startFrom *State) (*State, *State) {
	// the minimum amount of time the NFA needs to repeat
	var min int
	// the maximum amount of time the NFA needs to repeat
	var max int

	const Infinity = -1

	switch token.tokenType {
	case OneOrMore:
		min = 1
		max = Infinity
	case NoneOrMore:
		min = 0
		max = Infinity
	case Optional:
		min = 0
		max = 1
	}

	to := &State{
		transitions: map[uint8][]*State{},
	}

	if min == 0 {
		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
	}

	var total int

	if max != Infinity {
		total = max
	} else {
		if min == 0 {
			total = 1 // we need to at least create this NFA once, even if we require it 0 times
		} else {
			total = min
		}
	}

	value := token.value.([]regexToken)[0]
	previousStart, previousEnd := tokenToNfa(value, memory, &State{
		transitions: map[uint8][]*State{},
	})
	startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], previousStart)

	// starting from 2, because the one above is the first one
	for i := 2; i <= total; i++ {
		// the same NFA needs to be generated 'total' times
		start, end := tokenToNfa(value, memory, &State{
			transitions: map[uint8][]*State{},
		})
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
	if max == Infinity {
		to.transitions[EpsilonChar] = append(to.transitions[EpsilonChar], previousStart)
	}
	return startFrom, to
}
