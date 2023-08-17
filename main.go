package main

import (
	"fmt"
)

type DfaNode struct {
	id          int
	transitions map[uint8]*DfaNode
}

type Dfa struct {
	start *DfaNode
}

func (node *DfaNode) end() bool {
	return len(node.transitions) == 0
}

func (node *DfaNode) print() string {
	stuff := ""
	for token, n := range node.transitions {
		stuff += fmt.Sprintf("[%d] -> '%c' -> (%s)\n", node.id, token, n.print())
	}

	return stuff
}

func (node *DfaNode) check(input string, pos int) bool {
	if pos == len(input) && node.end() {
		return true
	}

	if pos >= len(input) {
		return false
	}

	ch := input[pos]
	result := node.transitions[ch]
	if result != nil {
		return result.check(input, pos+1)
	}

	epsilon := node.transitions[0]
	if epsilon != nil {
		return epsilon.check(input, pos)
	}

	return false
}

func (dfa Dfa) print() {
	fmt.Println(dfa.start.print())
}

func (dfa Dfa) check(input string) {
	fmt.Println(dfa.start.check(input, 0))
}

func toDfa(regexString string, pos int, startNode *DfaNode) {
	if pos >= len(regexString) {
		return
	}

	optionalChar := false
	if pos < len(regexString)-1 && regexString[pos+1] == '?' {
		epsilon := DfaNode{id: pos + 100, transitions: map[uint8]*DfaNode{}}
		(*startNode).transitions[0] = &epsilon
		toDfa(regexString, pos+2, &epsilon)
		optionalChar = true
	}

	ch := regexString[pos]
	node := DfaNode{id: pos, transitions: map[uint8]*DfaNode{}}
	(*startNode).transitions[ch] = &node

	if optionalChar {
		toDfa(regexString, pos+2, &node)
	} else {
		toDfa(regexString, pos+1, &node)
	}
}

func ToDfa(regexString string) Dfa {
	startNode := DfaNode{id: -1, transitions: map[uint8]*DfaNode{}}
	toDfa(regexString, 0, &startNode)

	return Dfa{
		start: &startNode,
	}
}

func main() {
	//ToDfa("ab?c").check("ac")
	//ToDfa("ab?c").check("abc")
	//ToDfa("ab?c").check("abcd")
	ToDfa("a?b?c?").check("abc")
	ToDfa("a?b?c?").check("a")
	ToDfa("a?b?c?").print()
}
