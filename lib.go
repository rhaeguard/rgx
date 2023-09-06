package rgx

import "fmt"

func Check(regexString string, inputString string) bool {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)
	fmt.Printf("%+v\n", nfaEntry)
	return nfaEntry.check(inputString, -1, nfaEntry.startOfText)
}

func DumpDotGraphForRegex(regexString string) {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)

	for _, x := range memory.tokens {
		fmt.Printf("%+v\n", x)
	}

	nfaEntry := toNfa(&memory)
	nfaEntry.dot(map[string]bool{})
}
