# rgx

![](https://github.com/rhaeguard/rgx/actions/workflows/go.yml/badge.svg)

a very simple regex engine written in go.

## todo

- [x] `^` beginning of the string
- [x] `$` end of the string
- [x] `.` any single character/wildcard
- [x] `[ ]` bracket notation/ranges
- [x] `[^ ]` bracket negation notation
- [ ] Quantifiers
  - [x] `*` none or more times
  - [x] `+` one or more times
  - [x] `?` optional
  - [ ] `{m,n}` more than or equal to `m` and less than equal to `n` times
- [ ] Capturing Group
  - [x] `( )` capturing group or subexpression
  - [x] `\n` backreference, e.g, `(dog)\1` where `n` is in `[0, 9]`
  - [x] `\k<name>` named backreference, e.g, `(?<animal>dog)\k<animal>`
- [ ] `\` escape character
