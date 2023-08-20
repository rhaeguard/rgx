package rgx

import "fmt"

type regexTokenType string

type regexToken struct {
	tokenType regexTokenType
	value     interface{}
}

func (t regexToken) is(_type regexTokenType) bool {
	return t.tokenType == _type
}

type parsingMemory struct {
	pos    int
	tokens []regexToken
}

func (p *parsingMemory) loc() int {
	return p.pos
}

func (p *parsingMemory) adv() int {
	p.pos += 1
	return p.pos
}

func (p *parsingMemory) push(token regexToken) {
	p.tokens = append(p.tokens, token)
}

func (p *parsingMemory) getLast(count int) []regexToken {
	return p.tokens[len(p.tokens)-count:]
}

func (p *parsingMemory) removeLast(count int) {
	p.tokens = append([]regexToken{}, p.tokens[:len(p.tokens)-count]...)
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

func parseRange(regexString string, memory *parsingMemory) {
	for regexString[memory.loc()] != ']' {
		ch := regexString[memory.loc()]

		if ch == '-' {
			prevChar := regexString[memory.loc()-1]
			nextChar := regexString[memory.adv()]
			token := regexToken{
				tokenType: "range",
				value:     fmt.Sprintf("%c-%c", prevChar, nextChar),
			}
			memory.tokens = append(memory.tokens, token)
		}

		memory.adv()
	}
}

func parseGroup(regexString string, memory *parsingMemory) {
	count := len(memory.tokens)
	for regexString[memory.loc()] != ')' {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
		memory.adv()
	}
	elementsCount := len(memory.tokens) - count

	token := regexToken{
		tokenType: "group",
		value:     memory.getLast(elementsCount),
	}
	memory.removeLast(elementsCount)
	memory.push(token)
}

var quantifiers = map[uint8]regexTokenType{
	'*': "none_or_more",
	'+': "one_or_more",
	'?': "optional",
}

func parseQuantifier(ch uint8, memory *parsingMemory) {
	token := regexToken{
		tokenType: quantifiers[ch],
		value:     memory.getLast(1),
	}
	memory.removeLast(1)
	memory.push(token)
}

func processChar(regexString string, memory *parsingMemory, ch uint8) {
	if ch == '(' {
		memory.adv()
		parseGroup(regexString, memory)
	} else if ch == '[' {
		memory.adv()
		parseRange(regexString, memory)
	} else if isQuantifier(ch) {
		parseQuantifier(ch, memory)
	} else if isAlphabetUppercase(ch) || isAlphabetLowercase(ch) || isNumeric(ch) {
		token := regexToken{
			tokenType: "construct",
			value:     ch,
		}
		memory.push(token)
	} else if ch == '|' {
		// OR requires two tokens
		// we process the ch to get the next token
		// and then construct the OR token
		processChar(regexString, memory, regexString[memory.adv()])
		token := regexToken{
			tokenType: "or",
			value:     memory.getLast(2),
		}
		memory.removeLast(2)
		memory.push(token)
	}
}

func regex(regexString string, memory *parsingMemory) {
	for memory.loc() < len(regexString) {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
		memory.adv()
	}
}