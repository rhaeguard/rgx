package rgx

import "fmt"

const (
	Literal         regexTokenType = "literal"
	NoneOrMore                     = "none_or_more"
	OneOrMore                      = "one_or_more"
	Optional                       = "optional"
	Or                             = "or"
	Bracket                        = "range"
	BracketNot                     = "range_not"
	Group                          = "group"
	GroupUncaptured                = "group_uncaptured"
	Wildcard                       = "wildcard"
	TextBeginning                  = "start_of_text"
	TextEnd                        = "end_of_text"
)

type regexTokenType string

type regexToken struct {
	tokenType regexTokenType
	value     interface{}
}

func (t regexToken) is(_type regexTokenType) bool {
	return t.tokenType == _type
}

type context struct {
	pos    int
	tokens []regexToken
}

func (p *context) loc() int {
	return p.pos
}

func (p *context) adv() int {
	p.pos += 1
	return p.pos
}

func (p *context) advTo(pos int) {
	p.pos = pos
}

func (p *context) push(token regexToken) {
	p.tokens = append(p.tokens, token)
}

func (p *context) getLast(count int) []regexToken {
	return p.tokens[len(p.tokens)-count:]
}

func (p *context) removeLast(count int) {
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

func isDot(ch uint8) bool {
	return ch == '.'
}

var quantifiers = map[uint8]regexTokenType{
	'*': NoneOrMore,
	'+': OneOrMore,
	'?': Optional,
}

func isQuantifier(ch uint8) bool {
	_, ok := quantifiers[ch]
	return ok
}

func sliceContains(slice []string, element string) bool {
	for _, el := range slice {
		if el == element {
			return true
		}
	}
	return false
}

func parseBracket(regexString string, memory *context) {
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

func parseGroup(regexString string, memory *context) {
	groupContext := context{
		pos:    memory.loc(),
		tokens: []regexToken{},
	}

	for regexString[groupContext.loc()] != ')' {
		ch := regexString[groupContext.loc()]
		processChar(regexString, &groupContext, ch)
		groupContext.adv()
	}

	token := regexToken{
		tokenType: Group,
		value:     groupContext.tokens,
	}
	memory.push(token)
	memory.advTo(groupContext.loc())
}

func parseGroupUncaptured(regexString string, memory *context) {
	groupContext := context{
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

func parseQuantifier(ch uint8, memory *context) {
	token := regexToken{
		tokenType: quantifiers[ch],
		value:     memory.getLast(1),
	}
	memory.removeLast(1)
	memory.push(token)
}

func parseAlphaNums(ch uint8, memory *context) {
	token := regexToken{
		tokenType: Literal,
		value:     ch,
	}
	memory.push(token)
}

func processChar(regexString string, memory *context, ch uint8) {
	if ch == '(' {
		memory.adv()
		parseGroup(regexString, memory)
	} else if ch == '[' {
		memory.adv()
		parseBracket(regexString, memory)
	} else if isQuantifier(ch) {
		parseQuantifier(ch, memory)
	} else if isAlphabetUppercase(ch) || isAlphabetLowercase(ch) || isNumeric(ch) {
		parseAlphaNums(ch, memory)
	} else if isDot(ch) {
		token := regexToken{
			tokenType: Wildcard,
			value:     ch,
		}
		memory.push(token)
	} else if ch == '|' {
		// everything to the left of the pipe in this specific "context"
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

func regex(regexString string, memory *context) {
	for memory.loc() < len(regexString) {
		ch := regexString[memory.loc()]
		processChar(regexString, memory, ch)
		memory.adv()
	}
}
