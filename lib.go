package rgx

func Check(regexString string, inputString string) bool {
	memory := parsingMemory{
		pos:    0,
		tokens: []regexToken{},
	}
	regex(regexString, &memory)
	nfaEntry := toNfa(&memory)

	return nfaEntry.check(inputString, -1)
}
