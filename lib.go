package rgx

// Compile compiles the given regex string
func Compile(regexString string) (*State, *RegexError) {
	parseContext := parsingContext{
		pos:            0,
		tokens:         []regexToken{},
		capturedGroups: map[string]bool{},
	}
	if err := parse(regexString, &parseContext); err != nil {
		return nil, err
	}
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
		Matches: result,
		Groups:  groups,
	}
}

func (s *State) FindMatches(inputString string) []Result {
	var results []Result
	start := -1
	for start < len(inputString) {
		checkContext := &regexCheckContext{
			groups: map[string]*capture{},
		}
		result := s.check(inputString, start, s.startOfText, checkContext)
		if !result {
			break
		}
		// prepare the result
		groups := map[string]string{}

		if result {
			// extract strings from the groups
			for groupName, captured := range checkContext.groups {
				groups[groupName] = captured.string(inputString)
				if groupName == "0" {
					start = captured.end + 1
				}
			}
		}

		r := Result{
			Matches: result,
			Groups:  groups,
		}

		results = append(results, r)
	}
	return results
}

// Check compiles the regexString and tests the inputString against it
func Check(regexString string, inputString string) (Result, *RegexError) {
	compiledNfa, err := Compile(regexString)
	if err != nil {
		return Result{}, err
	}
	return compiledNfa.Test(inputString), nil
}
