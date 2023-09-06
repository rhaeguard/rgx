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

type State struct {
	name        string
	terminal    bool
	endOfText   bool
	startOfText bool
	transitions map[uint8][]*State
}

func (s *State) makeTerminal() {
	s.terminal = true
}

const (
	StartOfText = 0  // ascii: null char
	EndOfText   = 3  // ascii: end of text
	AnyChar     = 26 // ascii: substitute
	EpsilonChar = 0  // ascii: null char
)

func toNfa(memory *context) *State {
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
		name: "start",
		transitions: map[uint8][]*State{
			EpsilonChar: {startState},
		},
	}

	end := &State{
		name:        "terminal",
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
				EpsilonChar: {start, to},
			},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to, start)

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
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
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
			name:        name(),
			transitions: map[uint8][]*State{},
		}

		from := &State{
			name: name(),
			transitions: map[uint8][]*State{
				EpsilonChar: {start, to},
			},
		}

		end.transitions[EpsilonChar] = append(end.transitions[EpsilonChar], to)

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
					EpsilonChar: {to},
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
	} else if token.is(TextBeginning) {
		state := &State{
			name:        name(),
			startOfText: true,
			transitions: map[uint8][]*State{},
		}

		return state, state
	} else if token.is(TextEnd) {
		state := &State{
			name:        name(),
			endOfText:   true,
			transitions: map[uint8][]*State{},
		}

		return state, state
	}

	panic(fmt.Sprintf("unrecognized token type: %s", token.tokenType))
}

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

// generates a dot graph
func (s *State) dot(processedStateForDot map[string]bool) {
	for char, states := range s.transitions {
		var label string
		if char == AnyChar {
			label = "any"
		} else if char == EpsilonChar {
			label = "ε"
		} else {
			label = fmt.Sprintf("%c", char)
		}

		processedStateForDot[s.name] = true
		fmt.Printf("%s [label=\"\"]\n", s.name)
		if s.startOfText {
			fmt.Printf("%s [color=red,style=filled]\n", s.name)
		}

		if s.endOfText {
			fmt.Printf("%s [color=blue,style=filled]\n", s.name)
		}

		for _, state := range states {
			fmt.Printf("%s -> %s [label=%s]\n", s.name, state.name, label)
			if _, ok := processedStateForDot[state.name]; !ok {
				state.dot(processedStateForDot)
			}
		}
	}
}
