package rgx

import "fmt"

type ParseErrorCode string

const (
	SyntaxError      ParseErrorCode = "SyntaxError"
	CompilationError                = "CompilationError"
)

type RegexError struct {
	Code    ParseErrorCode
	Message string
	Pos     int
}

func (p *RegexError) Error() string {
	return fmt.Sprintf("code=%s, message=%s, pos=%d", p.Code, p.Message, p.Pos)
}
