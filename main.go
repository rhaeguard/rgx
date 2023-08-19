package main

import "fmt"

type TokenType string

type Token struct {
	tokenType TokenType
	value     interface{}
}

func (t Token) is(_type TokenType) bool {
	return t.tokenType == _type
}

type Memory struct {
	pos    int
	tokens []Token
}

func (p *Memory) loc() int {
	return p.pos
}

func (p *Memory) adv() int {
	p.pos += 1
	return p.pos
}

func isAlphabetUppercase(ch uint8) bool {
	return ch >= 'A' && ch <= 'Z'
}

func isAlphabetLowercase(ch uint8) bool {
	return ch >= 'a' && ch <= 'z'
}

func isNumeric(ch uint8) bool {
	return ch >= '0' && ch <= '9'
}

func isQuantifier(ch uint8) bool {
	return ch == '*' || ch == '?' || ch == '+'
}

func parseRange(regexString string, memory *Memory) {
	for regexString[memory.loc()] != ']' {
		ch := regexString[memory.loc()]

		if ch == '-' {
			prevChar := regexString[memory.loc()-1]
			nextChar := regexString[memory.adv()]
			token := Token{
				tokenType: "range",
				value:     fmt.Sprintf("%c-%c", prevChar, nextChar),
			}
			memory.tokens = append(memory.tokens, token)
		}

		memory.adv()
	}
}

func parseGroup(regexString string, memory *Memory) {
	count := len(memory.tokens)
	for regexString[memory.loc()] != ')' {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
	}
	elementsCount := len(memory.tokens) - count

	token := Token{
		tokenType: "group",
		value:     memory.tokens[len(memory.tokens)-elementsCount:],
	}
	memory.tokens = append([]Token{}, memory.tokens[:len(memory.tokens)-elementsCount]...)
	memory.tokens = append(memory.tokens, token)
}

var quantifiers = map[uint8]TokenType{
	'*': "none_or_more",
	'+': "one_or_more",
	'?': "optional",
}

func parseQuantifier(ch uint8, memory *Memory) {
	lastToken := memory.tokens[len(memory.tokens)-1]
	token := Token{
		tokenType: quantifiers[ch],
		value:     lastToken,
	}
	memory.tokens = append([]Token{}, memory.tokens[:len(memory.tokens)-1]...)
	memory.tokens = append(memory.tokens, token)
}

func parseOr(regexString string, memory *Memory) {
	token := Token{
		tokenType: "or",
	}
	memory.tokens = append(memory.tokens, token)
}

func processChar(regexString string, memory *Memory, ch uint8) {
	if ch == '(' {
		memory.adv()
		parseGroup(regexString, memory)
	} else if ch == '[' {
		memory.adv()
		parseRange(regexString, memory)
	} else if isQuantifier(ch) {
		parseQuantifier(ch, memory)
	} else if isAlphabetUppercase(ch) || isAlphabetLowercase(ch) || isNumeric(ch) {
		token := Token{
			tokenType: "construct",
			value:     fmt.Sprintf("%c", ch),
		}
		memory.tokens = append(memory.tokens, token)
	} else if ch == '|' {
		parseOr(regexString, memory)
	}
	memory.adv()
}

func regex(regexString string, memory *Memory) {
	for memory.loc() < len(regexString) {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
	}
}

func main() {
	input := "(a|b|c)?cd*"
	memory := Memory{
		pos:    0,
		tokens: []Token{},
	}
	regex(input, &memory)

	for i := range memory.tokens {
		fmt.Printf("%+v\n", memory.tokens[i])
	}
}
