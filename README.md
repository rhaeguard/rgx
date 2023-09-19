# rgx

![](https://github.com/rhaeguard/rgx/actions/workflows/go.yml/badge.svg)

a very simple regex engine written in go.

## todo

- [x] `^` beginning of the string
- [x] `$` end of the string
- [x] `.` any single character/wildcard
- [ ] bracket notation
  - [x] `[ ]` bracket notation/ranges
  - [x] `[^ ]` bracket negation notation
  - [x] better handling of the bracket expressions: e.g., `[ab-exy12]`
  - [x] special characters in the bracket
    - [ ] support escape character
- [x] quantifiers
  - [x] `*` none or more times
  - [x] `+` one or more times
  - [x] `?` optional
  - [x] `{m,n}` more than or equal to `m` and less than equal to `n` times
- [x] capturing group
  - [x] `( )` capturing group or subexpression
  - [x] `\n` backreference, e.g, `(dog)\1` where `n` is in `[0, 9]`
  - [x] `\k<name>` named backreference, e.g, `(?<animal>dog)\k<animal>`
- [ ] `\` escape character
  - [ ] support special characters - context dependant
- [ ] better error handling in the API
- [ ] ability to work on multi-line strings
  - [ ] `.` should not match the newline - `\n`
  - [ ] `$` should match the newline - `\n`
  - [ ] multiple full matches
