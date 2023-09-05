package rgx

func Check(regexString string, inputString string) bool {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)
	return nfaEntry.check(inputString, -1)
}

func DumpDotGraphForRegex(regexString string) {
	memory := context{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)
	nfaEntry.dot(map[string]bool{})
}
