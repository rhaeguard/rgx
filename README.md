# rgx

![](https://github.com/rhaeguard/rgx/actions/workflows/go.yml/badge.svg)

a simple regex engine written in go. 

Hopefully will support basic regular syntax from POSIX standard for regex.

## todo

- [x] `^` beginning of the string
- [x] `$` end of the string
- [x] `.` any single character
- [x] `[ ]` bracket notation, ranges and all that
- [x] `[^ ]` bracket negation notation
- [x] `*` none or more times; Kleene's star
- [x] `( )` capturing group or subexpression
- [ ] `\n` backreference, e.g, `(dog)\1` where `n` is in `[0, 9]`
- [ ] `{m,n}` more than or equal to `m` and less than equal to `n` times
- [ ] `\` escape character 