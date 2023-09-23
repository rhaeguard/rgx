# rgx

![](https://github.com/rhaeguard/rgx/actions/workflows/go.yml/badge.svg)

A very simple regex engine written in go. This library is experimental, use it at your own risk!

### to add the dependency:

```shell
go get github.com/rhaeguard/rgx
```

### how to use: 
```go
import "github.com/rhaeguard/rgx"

pattern, err := rgx.Compile(regexString)
if err != nil {
	// error handling
}
results := pattern.FindMatches(content)

if results.Matches {
	groupMatchString := results.Groups["group-name"]
}
```

### todo

- [x] `^` beginning of the string
- [x] `$` end of the string
- [x] `.` any single character/wildcard
- [x] bracket notation
  - [x] `[ ]` bracket notation/ranges
  - [x] `[^ ]` bracket negation notation
  - [x] better handling of the bracket expressions: e.g., `[ab-exy12]`
  - [x] special characters in the bracket
    - [x] support escape character
- [x] quantifiers
  - [x] `*` none or more times
  - [x] `+` one or more times
  - [x] `?` optional
  - [x] `{m,n}` more than or equal to `m` and less than equal to `n` times
- [x] capturing group
  - [x] `( )` capturing group or subexpression
  - [x] `\n` backreference, e.g, `(dog)\1` where `n` is in `[0, 9]`
  - [x] `\k<name>` named backreference, e.g, `(?<animal>dog)\k<animal>`
  - [x] extracting the string that matches with the regex
- [x] `\` escape character
  - [x] support special characters - context dependant
- [x] better error handling in the API
- [x] ability to work on multi-line strings (tested on [Alice in Wonderland](./lib_testdata) text corpus)
  - [x] `.` should not match the newline - `\n`
  - [x] `$` should match the newline - `\n`
  - [x] multiple full matches

## notes

- `\` escape turns any next character into a literal, no special combinations such as `\d` for digits, `\b` for backspace, etc. are allowed
- numeric groups `\n` only support single digit references, so `\10` will be interpreted as the first capture group followed by a literal `0`

## credits
- [Alice in Wonderland, Lewis Carroll, Project Guttenberg](https://www.gutenberg.org/ebooks/11)