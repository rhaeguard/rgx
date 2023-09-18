package rgx

import (
	"fmt"
	"strconv"
	"strings"
)

type regexTokenType uint8

const (
	Literal         regexTokenType = iota // any literal character, e.g., a, b, 1, 2, etc.
	Or                             = iota // |
	Bracket                        = iota // []
	BracketNot                     = iota // [^]
	Group                          = iota // ()
	GroupUncaptured                = iota // logical group
	Wildcard                       = iota // .
	TextBeginning                  = iota // ^
	TextEnd                        = iota // $
	Backreference                  = iota // $
	Quantifier                     = iota // {m,n} or {m,}, {m}
)

type regexToken struct {
	tokenType regexTokenType
	value     interface{}
}

type quantifier struct {
	min   int
	max   int
	value interface{}
}

type parsingContext struct {
	pos            int
	tokens         []regexToken
	groupCounter   uint8
	capturedGroups map[string]bool
}

func (p *parsingContext) loc() int {
	return p.pos
}

func (p *parsingContext) nextGroup() uint8 {
	p.groupCounter++
	return p.groupCounter
}

func (p *parsingContext) adv() int {
	p.pos += 1
	return p.pos
}

func (p *parsingContext) advTo(pos int) {
	p.pos = pos
}

func (p *parsingContext) push(token regexToken) {
	p.tokens = append(p.tokens, token)
}

func (p *parsingContext) getLast(count int) []regexToken {
	return p.tokens[len(p.tokens)-count:]
}

func (p *parsingContext) removeLast(count int) {
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

var specialChars = map[uint8]bool{
	'&': true,
	'*': true,
	' ': true,
	//'{': true,
	//'}': true,
	'[': true,
	']': true,
	'(': true,
	')': true,
	',': true,
	'=': true,
	'-': true,
	//'.':  true,
	'+': true,
	';': true,
	//'\'': true,
	'/': true,
}

func isSpecialChar(ch uint8) bool {
	_, ok := specialChars[ch]
	return ok
}

func isLiteral(ch uint8) bool {
	return isAlphabetUppercase(ch) ||
		isAlphabetLowercase(ch) ||
		isNumeric(ch) ||
		isSpecialChar(ch)
}

func isDot(ch uint8) bool {
	return ch == '.'
}

const QuantifierInfinity = -1

var quantifiers = map[uint8][]int{
	'*': {0, QuantifierInfinity},
	'+': {1, QuantifierInfinity},
	'?': {0, 1},
}

func isQuantifier(ch uint8) bool {
	_, ok := quantifiers[ch]
	return ok
}

func parseBracket(regexString string, memory *parsingContext) {
	var pieces []string
	var tokenType regexTokenType

	if regexString[memory.loc()] == '^' {
		tokenType = BracketNot
		memory.adv()
	} else {
		tokenType = Bracket
	}

	for regexString[memory.loc()] != ']' {
		ch := regexString[memory.loc()]

		if ch == '-' {
			prevChar := pieces[len(pieces)-1][0] // TODO: maybe do smth better?
			nextChar := regexString[memory.adv()]
			bothNumeric := isNumeric(prevChar) && isNumeric(nextChar)
			bothLowercase := isAlphabetLowercase(prevChar) && isAlphabetLowercase(nextChar)
			bothUppercase := isAlphabetUppercase(prevChar) && isAlphabetUppercase(nextChar)
			if bothNumeric || bothLowercase || bothUppercase {
				pieces[len(pieces)-1] = fmt.Sprintf("%c%c", prevChar, nextChar)
			} else {
				panic(fmt.Sprintf("'%c-%c' range is invalid", prevChar, nextChar))
			}
		} else {
			pieces = append(pieces, fmt.Sprintf("%c", ch))
		}

		memory.adv()
	}
	var uniqueCharacterPieces []string
	for _, piece := range pieces {
		if !sliceContains(uniqueCharacterPieces, piece) {
			uniqueCharacterPieces = append(uniqueCharacterPieces, piece)
		}
	}

	var finalTokens []regexToken
	for _, piece := range uniqueCharacterPieces {
		if len(piece) == 1 {
			finalTokens = append(finalTokens, regexToken{
				tokenType: Literal,
				value:     piece[0],
			})
		} else if len(piece) == 2 {
			for s := piece[0]; s <= piece[1]; s++ {
				finalTokens = append(finalTokens, regexToken{
					tokenType: Literal,
					value:     s,
				})
			}
		} else {
			panic("piece must have max 2 characters")
		}
	}

	token := regexToken{
		tokenType: tokenType,
		value:     finalTokens,
	}
	memory.tokens = append(memory.tokens, token)
}

func parseGroup(regexString string, memory *parsingContext) {
	groupContext := parsingContext{
		pos:    memory.loc(),
		tokens: []regexToken{},
	}

	groupName := ""
	if regexString[groupContext.loc()] == '?' {
		if regexString[groupContext.adv()] == '<' {
			for regexString[groupContext.adv()] != '>' {
				ch := regexString[groupContext.loc()]
				groupName += fmt.Sprintf("%c", ch)
			}
		} else {
			panic("incomplete group structure")
		}

		groupContext.adv()
	}

	for regexString[groupContext.loc()] != ')' {
		ch := regexString[groupContext.loc()]
		processChar(regexString, &groupContext, ch)
		groupContext.adv()
	}

	token := regexToken{
		tokenType: Group,
		value:     []interface{}{groupContext.tokens, groupName},
	}
	memory.push(token)
	memory.advTo(groupContext.loc())
}

func parseGroupUncaptured(regexString string, memory *parsingContext) {
	groupContext := parsingContext{
		pos:    memory.loc(),
		tokens: []regexToken{},
	}

	for groupContext.loc() < len(regexString) && regexString[groupContext.loc()] != ')' {
		ch := regexString[groupContext.loc()]
		processChar(regexString, &groupContext, ch)
		groupContext.adv()
	}

	token := regexToken{
		tokenType: GroupUncaptured,
		value:     groupContext.tokens,
	}
	memory.push(token)

	if groupContext.loc() >= len(regexString) {
		memory.advTo(groupContext.loc())
	} else if regexString[groupContext.loc()] == ')' {
		memory.advTo(groupContext.loc() - 1) // advance but do not consume the closing parenthesis
	}
}

func parseQuantifier(ch uint8, memory *parsingContext) {
	bounds := quantifiers[ch]
	token := regexToken{
		tokenType: Quantifier,
		value: quantifier{
			min:   bounds[0],
			max:   bounds[1],
			value: memory.getLast(1),
		},
	}
	memory.removeLast(1)
	memory.push(token)
}

func parseLiteral(ch uint8, memory *parsingContext) {
	token := regexToken{
		tokenType: Literal,
		value:     ch,
	}
	memory.push(token)
}

func processChar(regexString string, memory *parsingContext, ch uint8) {
	if ch == '(' {
		memory.adv()
		parseGroup(regexString, memory)
	} else if ch == '[' {
		memory.adv()
		parseBracket(regexString, memory)
	} else if isQuantifier(ch) {
		parseQuantifier(ch, memory)
	} else if ch == '{' {
		startPos := memory.adv()
		var endPos = memory.loc()
		for regexString[endPos] != '}' {
			endPos++
		}
		memory.advTo(endPos)
		expr := regexString[startPos:endPos]
		pieces := strings.Split(expr, ",")

		var start int
		var end int

		if len(pieces) == 1 {
			start, _ = strconv.Atoi(pieces[0])
			end = start
		} else if len(pieces) == 2 {
			start, _ = strconv.Atoi(pieces[0])
			if pieces[1] == "" {
				end = QuantifierInfinity
			} else {
				end, _ = strconv.Atoi(pieces[1])
			}
		}

		token := regexToken{
			tokenType: Quantifier,
			value: quantifier{
				min:   start,
				max:   end,
				value: memory.getLast(1),
			},
		}
		memory.removeLast(1)
		memory.push(token)
	} else if ch == '\\' { // backslash
		parseBackslash(regexString, memory)
	} else if isLiteral(ch) {
		parseLiteral(ch, memory)
	} else if isDot(ch) {
		token := regexToken{
			tokenType: Wildcard,
			value:     ch,
		}
		memory.push(token)
	} else if ch == '|' {
		// everything to the left of the pipe in this specific "parsingContext"
		// is considered as the left side of the OR
		left := regexToken{
			tokenType: GroupUncaptured,
			value:     memory.getLast(len(memory.tokens)),
		}

		memory.adv() // to not get stuck in the pipe char
		parseGroupUncaptured(regexString, memory)
		right := memory.getLast(1)[0] // TODO: better error handling?

		// clear the memory as we do not need
		// any of these tokens anymore
		memory.removeLast(len(memory.tokens))

		token := regexToken{
			tokenType: Or,
			value:     []regexToken{left, right},
		}
		memory.push(token)
	} else if ch == '^' || ch == '$' {
		var tokenType = regexTokenType(TextBeginning)

		if ch == '$' {
			tokenType = TextEnd
		}

		token := regexToken{
			tokenType: tokenType,
			value:     ch,
		}
		memory.push(token)
	}
}

func parseBackslash(regexString string, memory *parsingContext) {
	nextChar := regexString[memory.loc()+1]
	if isNumeric(nextChar) { // \0 should be illegal
		token := regexToken{
			tokenType: Backreference,
			value:     fmt.Sprintf("%c", nextChar),
		}
		memory.push(token)
		memory.adv()
	} else if nextChar == 'k' {
		memory.adv()
		if regexString[memory.adv()] == '<' {
			groupName := ""
			for regexString[memory.adv()] != '>' {
				nextChar = regexString[memory.loc()]
				groupName += fmt.Sprintf("%c", nextChar)
			}
			token := regexToken{
				tokenType: Backreference,
				value:     groupName,
			}
			memory.push(token)
			memory.adv()
		} else {
			panic("invalid backreference syntax")
		}
	} else {
		panic("cannot process escape chars yet")
	}
}

func regex(regexString string, memory *parsingContext) {
	for memory.loc() < len(regexString) {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
		memory.adv()
	}
}
