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
	group         *group
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
		group: &group{
			names: []string{"0"},
			start: true,
			end:   false,
		},
	}

	end := &State{
		transitions: map[uint8][]*State{},
		terminal:    true,
		group: &group{
			names: []string{"0"},
			start: false,
			end:   true,
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
	case Wildcard:
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.transitions[AnyChar] = append(startFrom.transitions[AnyChar], to)

		return startFrom, to
	case NoneOrMore:
		value := token.value.([]regexToken)[0]
		_, end := tokenToNfa(value, memory, startFrom)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to, startFrom)

		return startFrom, to
	case OneOrMore:
		value := token.value.([]regexToken)[0]
		_, end := tokenToNfa(value, memory, startFrom)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to, startFrom)

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
		givenGroupName := v[1].(string)

		groupName := fmt.Sprintf("%d", memory.nextGroup())
		memory.capturedGroups[groupName] = true
		memory.capturedGroups[givenGroupName] = true

		start, end := tokenToNfa(values[0], memory, &State{
			transitions: map[uint8][]*State{},
		})

		for i := 1; i < len(values); i++ {
			_, endNext := tokenToNfa(values[i], memory, end)
			end = endNext
		}

		from := &State{
			transitions: map[uint8][]*State{
				EpsilonChar: {start},
			},
			group: &group{
				names: []string{groupName, givenGroupName},
				start: true,
			},
		}

		to := &State{
			transitions: map[uint8][]*State{},
			group: &group{
				names: []string{groupName, givenGroupName},
				end:   true,
			},
		}

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], from)
		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to)
		return startFrom, to
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
	case Optional:
		value := token.value.([]regexToken)[0]
		_, end := tokenToNfa(value, memory, startFrom)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.transitions[EpsilonChar] = append(startFrom.transitions[EpsilonChar], to)
		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to)

		return startFrom, to
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
