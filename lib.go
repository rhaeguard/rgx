package rgx

import "fmt"

func Check(regexString string, inputString string) bool {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	for _, t := range memory.tokens {
		fmt.Printf("%+v\n", t)
	}
	nfaEntry := toNfa(&memory)

	return nfaEntry.check(inputString, -1)
}
