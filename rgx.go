package rgx

import (
	"fmt"
	"math/rand"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var charset = "abcdefghijklmnopqrstuvwxyz"

// generates random name
func name() string {
	b := make([]byte, 4)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset)-1)]
	}
	return string(b)
}

func toNfa(memory *context) *State {
	token := memory.tokens[0]
	startState, endState := tokenToNfa(token)

	for i := range memory.tokens {
		if i == 0 {
			continue
		}
		startNext, endNext := tokenToNfa(memory.tokens[i])
		endState.transitions[0] = append(endState.transitions[0], startNext)

		endState = endNext
	}

	start := &State{
		name: "start",
		transitions: map[uint8][]*State{
			0: {startState},
		},
	}

	end := &State{
		name:        "terminal",
		transitions: map[uint8][]*State{},
		terminal:    true,
	}

	endState.transitions[0] = append(endState.transitions[0], end)

	return start
}

func tokenToNfa(token regexToken) (*State, *State) {
	if token.is(Construct) {
		value := token.value.(uint8)
		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				value: {to},
			},
		}

		return from, to
	} else if token.is(Wildcard) {
		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				AnyChar: {to},
			},
		}

		return from, to
	} else if token.is(NoneOrMore) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				0: {start, to},
			},
		}

		end.transitions[0] = append(end.transitions[0], to, start)

		return from, to
	} else if token.is(OneOrMore) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				0: {start},
			},
		}

		end.transitions[0] = append(end.transitions[0], to, start)

		return from, to
	} else if token.is(Or) {
		values := token.value.([]regexToken)
		start1, end1 := tokenToNfa(values[0])
		start2, end2 := tokenToNfa(values[1])

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				0: {start1, start2},
			},
		}

		end1.transitions[0] = append(end1.transitions[0], to)
		end2.transitions[0] = append(end2.transitions[0], to)

		return from, to
	} else if token.is(Group) || token.is(GroupUncaptured) {
		values := token.value.([]regexToken)
		start, end := tokenToNfa(values[0])

		i := 1
		for i < len(values) {
			startNext, endNext := tokenToNfa(values[i])
			end.transitions[0] = append(end.transitions[0], startNext)

			end = endNext
			i++
		}

		return start, end
	} else if token.is(Optional) {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				0: {start, to},
			},
		}

		end.transitions[0] = append(end.transitions[0], to)

		return from, to
	} else if token.is(Bracket) {
		constructTokens := token.value.([]regexToken)

		from := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			start := &State{
				name: name(),
				transitions: map[uint8][]*State{
					0: {to},
				},
			}
			from.transitions[ch] = []*State{start}
		}

		return from, to
	} else if token.is(BracketNot) {
		constructTokens := token.value.([]regexToken)

		from := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		to := &State{
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		for _, construct := range constructTokens {
			ch := construct.value.(uint8)
			start := &State{
				name:        name(),
				transitions: map[uint8][]*State{},
			}
			from.transitions[ch] = []*State{start}
		}
		from.transitions[AnyChar] = []*State{to}

		return from, to
	}

	panic(fmt.Sprintf("unrecognized token type: %s", token.tokenType))
}

type State struct {
	name        string
	terminal    bool
	transitions map[uint8][]*State
}

func (s *State) makeTerminal() {
	s.terminal = true
}

const (
	StartOfText = 0  // ascii: null char
	EndOfText   = 3  // ascii: end of text
	AnyChar     = 26 // ascii: substitute
)

func getChar(input string, pos int) uint8 {
	if pos >= 0 && pos < len(input) {
		return input[pos]
	}

	if pos >= len(input) {
		return EndOfText
	}

	return StartOfText
}

func (s *State) check(regex string, pos int) bool {
	current := getChar(regex, pos)

	if current == EndOfText && s.terminal {
		return true
	}

	realTransitions := s.transitions[current]

	if len(realTransitions) == 0 && current != EndOfText {
		realTransitions = s.transitions[AnyChar]
	}

	for i := range realTransitions {
		state := realTransitions[i]
		if state.check(regex, pos+1) {
			return true
		}
	}

	epsilonTransitions := s.transitions[0]
	for i := range epsilonTransitions {
		state := epsilonTransitions[i]
		if state.check(regex, pos) {
			return true
		}
	}

	return false
}

// generates a dot graph
func (s *State) dot(processedStateForDot map[string]bool) {
	for char, states := range s.transitions {
		var label string
		if char == AnyChar {
			label = "any"
		} else if char == 0 {
			label = "ε"
		} else {
			label = fmt.Sprintf("%c", char)
		}
		processedStateForDot[s.name] = true
		for _, state := range states {
			fmt.Printf("%s -> %s [label=%s]\n", s.name, state.name, label)
			if _, ok := processedStateForDot[state.name]; !ok {
				state.dot(processedStateForDot)
			}
		}
	}
}
