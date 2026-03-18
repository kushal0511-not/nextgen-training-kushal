package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Custom error types
type MismatchError struct {
	Position int
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("invalid parentheses at position %d", e.Position)
}

type SyntaxError struct {
	Message string
}

func (e *SyntaxError) Error() string {
	return e.Message
}

type DivisionByZeroError struct{}

func (e *DivisionByZeroError) Error() string {
	return "division by zero"
}

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *MultiError) Unwrap() []error {
	return e.Errors
}

// Feature (a): Bracket Matching
func ValidateParentheses(str string, s Stack[rune]) (bool, error) {
	brackets := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
		'>': '<',
	}
	openings := map[rune]bool{
		'(': true, '[': true, '{': true, '<': true,
	}
	posStack := NewSliceStack[int]()
	multiErr := &MultiError{}

	for i, char := range str {
		if openings[char] {
			s.Push(char)
			posStack.Push(i)
		} else if opening, ok := brackets[char]; ok {
			top, popped := s.Pop()
			if !popped || top != opening {
				multiErr.Add(&MismatchError{Position: i})
			} else {
				posStack.Pop()
			}
		}
	}

	for !posStack.IsEmpty() {
		pos, _ := posStack.Pop()
		multiErr.Add(&MismatchError{Position: pos})
	}

	if multiErr.HasErrors() {
		return false, multiErr
	}
	return true, nil
}

// Feature (b): Infix to Postfix (Shunting-Yard)
func InfixToPostfix(expr string, operators Stack[string]) ([]string, error) {
	var output []string
	precedence := map[string]int{
		"+": 1, "-": 1,
		"*": 2, "/": 2,
		"^": 3,
	}
	associativity := map[string]int{
		"^": 1, // 1 for right, 0 for left
	}

	multiErr := &MultiError{}
	tokens := tokenize(expr)
	isFunction := map[string]bool{"sin": true, "cos": true}

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if isFunction[token] {
			operators.Push(token)
		} else if token == "(" {
			operators.Push(token)
		} else if token == ")" {
			found := false
			for !operators.IsEmpty() {
				op, _ := operators.Pop()
				if op == "(" {
					found = true
					break
				}
				output = append(output, op)
			}
			if !found {
				multiErr.Add(&SyntaxError{Message: "mismatched parentheses (extra ')')"})
			} else if !operators.IsEmpty() {
				if top, _ := operators.Peek(); isFunction[top] {
					fun, _ := operators.Pop()
					output = append(output, fun)
				}
			}
		} else if p1, ok := precedence[token]; ok {
			for !operators.IsEmpty() {
				top, _ := operators.Peek()
				if top == "(" {
					break
				}
				p2 := precedence[top]
				if p2 > p1 || (p2 == p1 && associativity[token] == 0) {
					op, _ := operators.Pop()
					output = append(output, op)
				} else {
					break
				}
			}
			operators.Push(token)
		} else {
			multiErr.Add(&SyntaxError{Message: fmt.Sprintf("unknown operator: %s", token)})
		}
	}

	for !operators.IsEmpty() {
		op, _ := operators.Pop()
		if op == "(" {
			multiErr.Add(&SyntaxError{Message: "mismatched parentheses (extra '(')"})
			continue
		}
		output = append(output, op)
	}

	if multiErr.HasErrors() {
		return output, multiErr
	}
	return output, nil
}

// Feature (c): Postfix Evaluation
func EvaluatePostfix(tokens []string, stack Stack[float64]) (float64, error) {
	multiErr := &MultiError{}

	for _, token := range tokens {
		if isNumber(token) {
			val, _ := strconv.ParseFloat(token, 64)
			stack.Push(val)
		} else if token == "sin" || token == "cos" {
			val, ok := stack.Pop()
			if !ok {
				multiErr.Add(&SyntaxError{Message: fmt.Sprintf("invalid postfix expression: insufficient operands for %s", token)})
				continue
			}
			var result float64
			if token == "sin" {
				result = math.Sin(val)
			} else {
				result = math.Cos(val)
			}
			stack.Push(result)
		} else {
			b, ok1 := stack.Pop()
			a, ok2 := stack.Pop()
			if !ok1 || !ok2 {
				multiErr.Add(&SyntaxError{Message: fmt.Sprintf("invalid postfix expression: insufficient operands for %s", token)})
				continue
			}

			var result float64
			switch token {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					multiErr.Add(&DivisionByZeroError{})
					result = 0 // continue with 0
				} else {
					result = a / b
				}
			case "^":
				result = math.Pow(a, b)
			default:
				multiErr.Add(&SyntaxError{Message: fmt.Sprintf("unknown operator: %s", token)})
				continue
			}
			stack.Push(result)
		}
	}

	if multiErr.HasErrors() {
		return 0, multiErr
	}

	res, ok := stack.Pop()
	if !ok || !stack.IsEmpty() {
		return 0, &SyntaxError{Message: "invalid postfix expression (too many values or empty)"}
	}
	return res, nil
}

// Helper methods
func tokenize(expr string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range expr {
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}
		if unicode.IsDigit(r) || r == '.' {
			current.WriteRune(r)
		} else if unicode.IsLetter(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// Feature (d): Combined
func Calculate(expr string) (float64, error) {
	allErrors := &MultiError{}

	// Step 1: Validate parentheses
	s := NewSliceStack[rune]()
	if ok, err := ValidateParentheses(expr, s); !ok {
		allErrors.Add(err)
	}
	// Step 2: Convert to Postfix
	operators := NewSliceStack[string]()
	postfix, err := InfixToPostfix(expr, operators)
	if err != nil {
		allErrors.Add(err)
	}

	// Step 3: Evaluate
	evalStack := NewSliceStack[float64]()
	result, err := EvaluatePostfix(postfix, evalStack)
	if err != nil {
		allErrors.Add(err)
	}

	if allErrors.HasErrors() {
		return 0, allErrors
	}

	return result, nil
}
