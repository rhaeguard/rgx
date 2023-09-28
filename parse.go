package rgx

import (
	"fmt"
	"strconv"
	"strings"
)

type regexTokenType uint8

const (
	literal         regexTokenType = iota // any literal character, e.g., a, b, 1, 2, etc.
	or                             = iota // |
	bracket                        = iota // []
	bracketNot                     = iota // [^]
	groupCaptured                  = iota // ()
	groupUncaptured                = iota // logical group
	wildcard                       = iota // .
	textBeginning                  = iota // ^
	textEnd                        = iota // $
	backReference                  = iota // $
	quantifier                     = iota // {m,n} or {m,}, {m}
)

type regexToken struct {
	tokenType regexTokenType
	value     interface{}
}

type quantifierPayload struct {
	min   int
	max   int
	value interface{}
}

type groupTokenPayload struct {
	tokens []regexToken
	name   string
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

// removeLast pops the last count number of elements and returns the popped elements
func (p *parsingContext) removeLast(count int) []regexToken {
	elementsToBeRemoved := p.tokens[len(p.tokens)-count:]
	p.tokens = append([]regexToken{}, p.tokens[:len(p.tokens)-count]...)
	return elementsToBeRemoved
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
	'&':  true,
	'*':  true,
	' ':  true,
	'{':  true,
	'}':  true,
	'[':  true,
	']':  true,
	'(':  true,
	')':  true,
	',':  true,
	'=':  true,
	'-':  true,
	'.':  true,
	'+':  true,
	';':  true,
	'\\': true,
	'/':  true,
}

var mustBeEscapedCharacters = map[uint8]bool{
	'[':  true,
	'\\': true,
	'^':  true,
	'$':  true,
	'.':  true,
	'|':  true,
	'?':  true,
	'*':  true,
	'+':  true,
	'(':  true,
	')':  true,
	'{':  true,
	'}':  true,
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

func isWildcard(ch uint8) bool {
	return ch == '.'
}

const quantifierInfinity = -1

var quantifiersWithBounds = map[uint8][]int{
	'*': {0, quantifierInfinity},
	'+': {1, quantifierInfinity},
	'?': {0, 1},
}

func isQuantifier(ch uint8) bool {
	_, ok := quantifiersWithBounds[ch]
	return ok
}

func parseBracket(regexString string, context *parsingContext) *RegexError {
	var pieces []string
	var tokenType regexTokenType

	// do we need to negate the bracket?
	if regexString[context.loc()] == '^' {
		tokenType = bracketNot
		context.adv()
	} else {
		tokenType = bracket
	}

	for context.loc() < len(regexString) && regexString[context.loc()] != ']' {
		ch := regexString[context.loc()]

		if ch == '-' && context.loc()+1 < len(regexString) {
			nextChar := regexString[context.loc()+1]
			// if - is the first character OR is the last character, it's a literal
			if len(pieces) == 0 || nextChar == ']' {
				pieces = append(pieces, fmt.Sprintf("%c", ch))
			} else {
				context.adv() // to process the nextChar's position
				piece := pieces[len(pieces)-1]
				if len(piece) == 1 {
					prevChar := piece[0]
					if prevChar <= nextChar {
						pieces[len(pieces)-1] = fmt.Sprintf("%c%c", prevChar, nextChar)
					} else {
						return &RegexError{
							Code:    SyntaxError,
							Message: fmt.Sprintf("'%c-%c' range is invalid", prevChar, nextChar),
							Pos:     context.loc(),
						}
					}
				} else {
					pieces = append(pieces, fmt.Sprintf("%c", ch))
				}
			}
		} else if ch == '\\' && context.loc()+1 < len(regexString) {
			nextChar := regexString[context.adv()]
			// TODO: some characters are special: \a does not just mean a, it means alarm ascii char etc.
			// TODO: maybe in future, I'll implement that as well
			// TODO: for now, all the escaped characters will be treated as literals
			pieces = append(pieces, fmt.Sprintf("%c", nextChar))
		} else {
			pieces = append(pieces, fmt.Sprintf("%c", ch))
		}
		context.adv()
	}

	if regexString[context.loc()] != ']' {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Bracket is not closed properly",
			Pos:     context.loc(),
		}
	}

	if len(pieces) == 0 {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Bracket should not be empty",
			Pos:     context.loc(),
		}
	}

	uniqueCharacterPieces := map[uint8]bool{}
	for _, piece := range pieces {
		for s := piece[0]; s <= piece[len(piece)-1]; s++ {
			uniqueCharacterPieces[s] = true
		}
	}

	token := regexToken{
		tokenType: tokenType,
		value:     uniqueCharacterPieces,
	}
	context.tokens = append(context.tokens, token)

	return nil
}

func parseGroup(regexString string, context *parsingContext) *RegexError {
	groupContext := parsingContext{
		pos:    context.loc(),
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
			return &RegexError{
				Code:    SyntaxError,
				Message: "Group name syntax is incorrect",
				Pos:     groupContext.loc(),
			}
		}

		groupContext.adv()
	}

	for groupContext.loc() < len(regexString) && regexString[groupContext.loc()] != ')' {
		ch := regexString[groupContext.loc()]
		if err := processChar(regexString, &groupContext, ch); err != nil {
			return err
		}
		groupContext.adv()
	}

	if regexString[groupContext.loc()] != ')' {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Group has not been properly closed",
			Pos:     groupContext.loc(),
		}
	}

	token := regexToken{
		tokenType: groupCaptured,
		value: groupTokenPayload{
			tokens: groupContext.tokens,
			name:   groupName,
		},
	}
	context.push(token)
	context.advTo(groupContext.loc())
	return nil
}

func parseGroupUncaptured(regexString string, context *parsingContext) *RegexError {
	groupContext := parsingContext{
		pos:    context.loc(),
		tokens: []regexToken{},
	}

	for groupContext.loc() < len(regexString) && regexString[groupContext.loc()] != ')' {
		ch := regexString[groupContext.loc()]
		if err := processChar(regexString, &groupContext, ch); err != nil {
			return err
		}
		groupContext.adv()
	}

	if regexString[groupContext.loc()] != ')' {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Group has not been properly closed",
			Pos:     groupContext.loc(),
		}
	}

	token := regexToken{
		tokenType: groupUncaptured,
		value:     groupContext.tokens,
	}
	context.push(token)

	if groupContext.loc() >= len(regexString) {
		context.advTo(groupContext.loc())
	} else if regexString[groupContext.loc()] == ')' {
		context.advTo(groupContext.loc() - 1) // advance but do not consume the closing parenthesis
	}

	return nil
}

func processChar(regexString string, context *parsingContext, ch uint8) *RegexError {
	if ch == '(' {
		context.adv()
		if err := parseGroup(regexString, context); err != nil {
			return err
		}
	} else if ch == '[' {
		context.adv()
		if err := parseBracket(regexString, context); err != nil {
			return err
		}
	} else if isQuantifier(ch) {
		bounds := quantifiersWithBounds[ch]
		token := regexToken{
			tokenType: quantifier,
			value: quantifierPayload{
				min:   bounds[0],
				max:   bounds[1],
				value: context.removeLast(1),
			},
		}
		context.push(token)
	} else if ch == '{' {
		if err := parseBoundedQuantifier(regexString, context); err != nil {
			return err
		}
	} else if ch == '\\' { // escaped backslash
		if err := parseBackslash(regexString, context); err != nil {
			return err
		}
	} else if isWildcard(ch) {
		token := regexToken{
			tokenType: wildcard,
			value:     ch,
		}
		context.push(token)
	} else if isLiteral(ch) {
		token := regexToken{
			tokenType: literal,
			value:     ch,
		}
		context.push(token)
	} else if ch == '|' {
		// everything to the left of the pipe in this specific "parsingContext"
		// is considered as the left side of the OR
		left := regexToken{
			tokenType: groupUncaptured,
			value:     context.removeLast(len(context.tokens)),
		}

		context.adv() // to not get stuck in the pipe char
		if err := parseGroupUncaptured(regexString, context); err != nil {
			return err
		}
		right := context.removeLast(1)[0] // TODO: better error handling?

		token := regexToken{
			tokenType: or,
			value:     []regexToken{left, right},
		}
		context.push(token)
	} else if ch == '^' || ch == '$' { // anchors
		var tokenType = regexTokenType(textBeginning)

		if ch == '$' {
			tokenType = textEnd
		}

		token := regexToken{
			tokenType: tokenType,
			value:     ch,
		}
		context.push(token)
	}
	return nil
}

func parseBoundedQuantifier(regexString string, context *parsingContext) *RegexError {
	startPos := context.adv()
	var endPos = context.loc()
	for regexString[endPos] != '}' {
		endPos++
	}
	context.advTo(endPos)
	expr := regexString[startPos:endPos]
	pieces := strings.Split(expr, ",")

	if len(pieces) == 0 {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Quantifier must have at least one bound",
			Pos:     startPos,
		}
	}

	var start int
	var end int
	var err error
	if len(pieces) == 1 {
		start, err = strconv.Atoi(pieces[0])
		if err != nil {
			return &RegexError{
				Code:    SyntaxError,
				Message: err.Error(),
				Pos:     startPos,
			}
		}
		end = start
	} else if len(pieces) == 2 {
		start, err = strconv.Atoi(pieces[0])
		if err != nil {
			return &RegexError{
				Code:    SyntaxError,
				Message: err.Error(),
				Pos:     startPos,
			}
		}
		if pieces[1] == "" {
			end = quantifierInfinity
		} else {
			end, err = strconv.Atoi(pieces[1])
			if err != nil {
				return &RegexError{
					Code:    SyntaxError,
					Message: err.Error(),
					Pos:     startPos,
				}
			}
		}
	}

	token := regexToken{
		tokenType: quantifier,
		value: quantifierPayload{
			min:   start,
			max:   end,
			value: context.removeLast(1),
		},
	}
	context.push(token)

	return nil
}

func parseBackslash(regexString string, context *parsingContext) *RegexError {
	nextChar := regexString[context.loc()+1]
	if isNumeric(nextChar) { // cares about the next single digit
		token := regexToken{
			tokenType: backReference,
			value:     fmt.Sprintf("%c", nextChar),
		}
		context.push(token)
		context.adv()
	} else if nextChar == 'k' { // \k<name> reference
		context.adv()
		if regexString[context.adv()] == '<' {
			groupName := ""
			for regexString[context.adv()] != '>' {
				nextChar = regexString[context.loc()]
				groupName += fmt.Sprintf("%c", nextChar)
			}
			token := regexToken{
				tokenType: backReference,
				value:     groupName,
			}
			context.push(token)
			context.adv()
		} else {
			return &RegexError{
				Code:    SyntaxError,
				Message: "Invalid backreference syntax",
				Pos:     context.loc(),
			}
		}
	} else if _, canBeEscaped := mustBeEscapedCharacters[nextChar]; canBeEscaped {
		token := regexToken{
			tokenType: literal,
			value:     nextChar,
		}
		context.push(token)
		context.adv()
	} else {
		// we're treating newline and tab as special characters
		// and not the rest
		if nextChar == 'n' {
			nextChar = '\n'
		} else if nextChar == 't' {
			nextChar = '\t'
		}
		token := regexToken{
			tokenType: literal,
			value:     nextChar,
		}
		context.push(token)
		context.adv()
	}

	return nil
}

func parse(regexString string, context *parsingContext) *RegexError {
	for context.loc() < len(regexString) {
		ch := regexString[context.loc()]
		if err := processChar(regexString, context, ch); err != nil {
			return err
		}
		context.adv()
	}
	return nil
}
