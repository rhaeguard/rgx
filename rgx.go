package rgx

import "fmt"

func toNfa(memory *parsingMemory) *State {
	token := memory.tokens[0]
	start, end := tokenToNfa(token)

	for i := range memory.tokens {
		if i == 0 {
			continue
		}

		startNext, endNext := tokenToNfa(memory.tokens[i])
		end.transitions[0] = append(end.transitions[0], startNext)

		end = endNext
	}

	end.makeTerminal()

	return start
}

func tokenToNfa(token regexToken) (*State, *State) {
	if token.is("construct") {
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
	} else if token.is("none_or_more") {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				0: {start, to},
			},
		}

		end.transitions[0] = append(end.transitions[0], to, start)

		return from, to
	} else if token.is("or") {
		values := token.value.([]regexToken)
		start1, end1 := tokenToNfa(values[0])
		start2, end2 := tokenToNfa(values[1])

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				0: {start1, start2},
			},
		}

		end1.transitions[0] = append(end1.transitions[0], to)
		end2.transitions[0] = append(end2.transitions[0], to)

		return from, to
	} else if token.is("group") {
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
	} else if token.is("optional") {
		value := token.value.([]regexToken)[0]
		start, end := tokenToNfa(value)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		from := &State{
			transitions: map[uint8][]*State{
				0: {start, to},
			},
		}

		end.transitions[0] = append(end.transitions[0], to)

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
	StartOfText = 0
	EndOfText   = 3
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
	ch := getChar(regex, pos)

	if ch == EndOfText && s.terminal {
		return true
	}

	realTransitions := s.transitions[ch]
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