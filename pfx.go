package main

import "fmt"

type OperatorStack struct {
	ops []uint8
}

func (os *OperatorStack) pop() uint8 {
	x := os.ops[len(os.ops)-1]
	os.ops = append([]uint8{}, os.ops[:len(os.ops)-1]...)
	return x
}

func (os *OperatorStack) top() uint8 {
	if os.hasElements() {
		return os.ops[len(os.ops)-1]
	}
	return '0'
}

func (os *OperatorStack) hasElements() bool {
	return len(os.ops) > 0
}

func (os *OperatorStack) push(ch uint8) {
	os.ops = append(os.ops, ch)
}

var operators = map[uint8]uint8{
	'|': 0,
	'?': 2,
	'*': 2,
	'+': 2,
}

func toPostfix(regexString string) {
	i := 0
	out := ""
	stack := OperatorStack{}
	for i < len(regexString) {
		ch := regexString[i]
		if ch == '(' {
			stack.push(ch)
		} else if ch == ')' {
			for stack.top() != '(' {
				out += fmt.Sprintf("%c", stack.pop())
			}
			stack.pop()
		} else if tokenPrecedence, ok := operators[ch]; ok {
			for stackPrecedence, ok := operators[stack.top()]; ok && (stackPrecedence >= tokenPrecedence && stack.top() != '('); {
				out += fmt.Sprintf("%c", stack.pop())
				stackPrecedence, ok = operators[stack.top()]
			}
			stack.push(ch)
		} else {
			out += fmt.Sprintf("%c", ch)
		}
		i += 1
	}

	for stack.hasElements() {
		out += fmt.Sprintf("%c", stack.pop())
	}

	fmt.Println(out)
}
