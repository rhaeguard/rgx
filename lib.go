package rgx

// Compile compiles the given regex string
func Compile(regexString string) *State {
	parseContext := parsingContext{
		pos:            0,
		tokens:         []regexToken{},
		capturedGroups: map[string]bool{},
	}
	regex(regexString, &parseContext)
	return toNfa(&parseContext)
}

// Test checks if the given input string conforms to this NFA
func (s *State) Test(inputString string) Result {
	checkContext := &regexCheckContext{
		groups: map[string]*capture{},
	}

	result := s.check(inputString, -1, s.startOfText, checkContext)

	// prepare the result
	groups := map[string]string{}

	if result {
		// extract strings from the groups
		for groupName, captured := range checkContext.groups {
			groups[groupName] = captured.string(inputString)
		}
	}

	return Result{
		matches: result,
		groups:  groups,
	}
}

// Check compiles the regexString and tests the inputString against it
func Check(regexString string, inputString string) Result {
	return Compile(regexString).Test(inputString)
}
