package rgx

import (
	"fmt"
)

type State struct {
	start       bool
	terminal    bool
	endOfText   bool
	startOfText bool
	transitions map[uint8][]*State
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
	startState, endState := tokenToNfa(token)

	for i := startFrom + 1; i <= endAt; i++ {
		startNext, endNext := tokenToNfa(memory.tokens[i])
		endState.transitions[EpsilonChar] = append(endState.transitions[EpsilonChar], startNext)

		endState = endNext
	}

	start := &State{
		start: true,
		transitions: map[uint8][]*State{
			EpsilonChar: {startState},
		},
	}

	end := &State{
		transitions: map[uint8][]*State{},
		terminal:    true,
	}

	endState.transitions[EpsilonChar] = append(endState.transitions[EpsilonChar], end)

	return start
}

func tokenToNfa(token regexToken) (*State, *State) {
	if token.is(Literal) {
		value := token.value.(uint8)
		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				value: {to},
			},
		}

		return from, to
	} else if token.is(Wildcard) {
		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				AnyChar: {to},
			},
		}

		return from, to
	} else if token.is(NoneOrMore) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				EpsilonChar: {start, to},
			},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to, start)

		return from, to
	} else if token.is(OneOrMore) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				EpsilonChar: {start},
			},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to, start)

		return from, to
	} else if token.is(Or) {
		values := token.value.([]regexToken)
		start1, end1 := tokenToNfa(values[0])
		start2, end2 := tokenToNfa(values[1])

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				EpsilonChar: {start1, start2},
			},
		}

		end1.transitions[EpsilonChar] = append(end1.transitions[EpsilonChar], to)
		end2.transitions[EpsilonChar] = append(end2.transitions[EpsilonChar], to)

		return from, to
	} else if token.is(Group) || token.is(GroupUncaptured) {
		values := token.value.([]regexToken)
		start, end := tokenToNfa(values[0])

		i := 1
		for i < len(values) {
			startNext, endNext := tokenToNfa(values[i])
			end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], startNext)

			end = endNext
			i++
		}

		return start, end
	} else if token.is(Optional) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				EpsilonChar: {start, to},
			},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to)

		return from, to
	} else if token.is(Bracket) {
		constructTokens := token.value.([]regexToken)

		from := &State{
			transitions: map[uint8][]*State{},
		}

		to := &State{
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			start := &State{
				transitions: map[uint8][]*State{
					EpsilonChar: {to},
				},
			}
			from.transitions[ch] = []*State{start}
		}

		return from, to
	} else if token.is(BracketNot) {
		constructTokens := token.value.([]regexToken)

		from := &State{
			transitions: map[uint8][]*State{},
		}

		to := &State{
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			start := &State{
				transitions: map[uint8][]*State{},
			}
			from.transitions[ch] = []*State{start}
		}
		from.transitions[AnyChar] = []*State{to}

		return from, to
	} else if token.is(TextBeginning) {
		state := &State{
			startOfText: true,
			transitions: map[uint8][]*State{},
		}

		return state, state
	} else if token.is(TextEnd) {
		state := &State{
			endOfText:   true,
			transitions: map[uint8][]*State{},
		}

		return state, state
	}

	panic(fmt.Sprintf("unrecognized token: %+v", token))
}
