package main

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestValidateParentheses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
		pos      int
	}{
		{"Empty", "", true, -1},
		{"Simple parens", "()", true, -1},
		{"All types", "()[]{}<>", true, -1},
		{"Mismatch type", "(]", false, 1},
		{"Mismatch order", "([)]", false, 2},
		{"Nested", "{[]}", true, -1},
		{"Single open", "(", false, 0},
		{"Single close", ")", false, 0},
		{"Multiple nested", "((()))", true, -1},
		{"Incomplete", "((())", false, 0},
		{"Complex balanced", "({[<>]})", true, -1},
		{"Complex mismatch", "({[<]>})", false, 4},
		{"Single square open", "[", false, 0},
		{"Single angle close", ">", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewSliceStack[rune]()
			ok, err := ValidateParentheses(tt.input, s)
			if ok != tt.expected {
				t.Errorf("ValidateParentheses(%q) ok = %v; want %v", tt.input, ok, tt.expected)
			}
			if err != nil {
				var mErr *MismatchError
				if errors.As(err, &mErr) {
					if mErr.Position != tt.pos {
						t.Errorf("ValidateParentheses(%q) pos = %d; want %d", tt.input, mErr.Position, tt.pos)
					}
				}
			}
		})
	}
}

func TestInfixToPostfix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
		err      bool
	}{
		{"Simple addition", "3 + 4", []string{"3", "4", "+"}, false},
		{"Complex expression", "3 + 4 * 2 / (1 - 5) ^ 2", []string{"3", "4", "2", "*", "1", "5", "-", "2", "^", "/", "+"}, false},
		{"Missing closing parenthesis", "(3 + 4", nil, true},
		{"Missing opening parenthesis", "3 + 4 )", nil, true},
		{"Right associativity", "1 + 2 ^ 3 ^ 4", []string{"1", "2", "3", "4", "^", "^", "+"}, false},
		{"Parentheses with multi-digit", "10 * (2 + 3)", []string{"10", "2", "3", "+", "*"}, false},
		{"Empty string", "", nil, false},
		{"Single operand", "123", []string{"123"}, false},
		{"Operator precedence", "1 + 2 * 3", []string{"1", "2", "3", "*", "+"}, false},
		{"Parentheses change precedence", "(1+2)*3", []string{"1", "2", "+", "3", "*"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			operators := NewSliceStack[string]()
			res, err := InfixToPostfix(tt.input, operators)
			if (err != nil) != tt.err {
				t.Errorf("InfixToPostfix(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !tt.err {
				if len(res) != len(tt.expected) {
					t.Errorf("InfixToPostfix(%q) = %v; want %v", tt.input, res, tt.expected)
					return
				}
				for i := range res {
					if res[i] != tt.expected[i] {
						t.Errorf("InfixToPostfix(%q) = %v; want %v", tt.input, res, tt.expected)
						break
					}
				}
			}
		})
	}
}

func TestEvaluatePostfix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []string
		expected float64
		checkErr func(error) bool
	}{
		{"Simple add", []string{"3.3", "4.4", "+"}, 7.7, nil},
		{"Complex expr", []string{"3", "4", "2", "*", "1", "5", "-", "2", "^", "/", "+"}, 3.5, nil},
		{"Div by zero", []string{"10", "0", "/"}, 0, func(err error) bool {
			return errors.As(err, &([]*DivisionByZeroError{nil}[0])) || strings.Contains(err.Error(), "division by zero")
		}},
		{"Power", []string{"2", "3", "^"}, 8, nil},
		{"Missing operand", []string{"+"}, 0, func(err error) bool {
			var sErr *SyntaxError
			return errors.As(err, &sErr)
		}},
		{"Extra operand", []string{"1", "2", "3", "+"}, 0, func(err error) bool {
			var sErr *SyntaxError
			return errors.As(err, &sErr)
		}},
		{"Single value", []string{"5"}, 5, nil},
		{"Multiple multiply", []string{"2", "2", "2", "*", "*"}, 8, nil},
		{"Subtract", []string{"10", "2", "-"}, 8, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewSliceStack[float64]()
			res, err := EvaluatePostfix(tt.input, s)
			if tt.checkErr != nil {
				if err == nil || !tt.checkErr(err) {
					t.Errorf("EvaluatePostfix(%v) error = %v; want specific error", tt.input, err)
				}
			} else if err != nil {
				t.Errorf("EvaluatePostfix(%v) unexpected error: %v", tt.input, err)
			} else if math.Abs(res-tt.expected) > 1e-9 {
				t.Errorf("EvaluatePostfix(%v) = %v; want %v", tt.input, res, tt.expected)
			}
		})
	}
}

func TestCalculateCombined(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected float64
		err      bool
	}{
		{"3 + 4 * 2 / (1 - 5) ^ 2", 3.5, false},
		{"(1 + 2) * 3", 9, false},
		{"1 / 0", 0, true},
		{"{1 + 2)", 0, true},
		{"2 ^ 3", 8, false},
		{"10 - 2 - 3", 5, false},
		{"10 / 2 / 5", 1, false},
		{"(1 + (2 * (3 - 1)))", 5, false},
		{"(1 + 2)", 3, false},
		{"4 * 2.5", 10, false},
		{"sin(0)", 0, false},
		{"cos(0)", 1, false},
		{"sin(3.1415926535 / 2)", 1, false},
		{"cos(3.1415926535)", -1, false},
		{"sin(0) + cos(0)", 1, false},
		{"2 * sin(0) + 3 * cos(0)", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			res, err := Calculate(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("Calculate(%q) unexpected error: %v", tt.input, err)
			}
			if !tt.err && math.Abs(res-tt.expected) > 1e-9 {
				t.Errorf("Calculate(%q) = %v; want %v", tt.input, res, tt.expected)
			}
		})
	}
}

func TestDeeplyNested(t *testing.T) {
	t.Parallel()
	levels := 1000
	var sb strings.Builder
	for i := 0; i < levels; i++ {
		sb.WriteString("(")
	}
	sb.WriteString("1")
	for i := 0; i < levels; i++ {
		sb.WriteString(")")
	}

	res, err := Calculate(sb.String())
	if err != nil {
		t.Errorf("Deeply nested failed: %v", err)
	}
	if res != 1 {
		t.Errorf("Deeply nested = %v; want 1", res)
	}
}

func TestMultipleErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		errCount int
	}{
		{"Multiple mismatched parens", "((1+2", 2},              // Two open parens unmatched
		{"Unknown operators and mismatch", "1 $ 2 @ (3 + 4", 3}, // $, @, and one (
		{"Div by zero and mismatch", "(1/0) + (", 2},            // Div by zero, and one (
		{"Complex multiple errors", "((1 + 2) * $ / 0", 3},      // $, div by zero (but / 0 needs two operands), mismatched (
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Calculate(tt.input)
			if err == nil {
				t.Errorf("Calculate(%q) expected error, got nil", tt.input)
				return
			}
			var mErr *MultiError
			if errors.As(err, &mErr) {
				if len(mErr.Errors) < tt.errCount {
					t.Errorf("Calculate(%q) got %d errors; want at least %d. Errors: %v", tt.input, len(mErr.Errors), tt.errCount, mErr.Errors)
				}
			} else {
				t.Errorf("Calculate(%q) error is not MultiError: %T (%v)", tt.input, err, err)
			}
		})
	}
}

func BenchmarkCalculateSlice(b *testing.B) {
	expr := "3 + 4 * 2 / (1 - 5) ^ 2 + sin(0) + cos(0)"
	for i := 0; i < b.N; i++ {
		_, _ = Calculate(expr)
	}
}

// CalculateWithLinkedList is a helper to run Calculate logic with LinkedListStack
func CalculateWithLinkedList(expr string) (float64, error) {
	allErrors := &MultiError{}

	// Step 1: Validate parentheses
	s := NewLinkedListStack[rune]()
	if ok, err := ValidateParentheses(expr, s); !ok {
		allErrors.Add(err)
	}

	// Step 2: Convert to Postfix
	operators := NewLinkedListStack[string]()
	postfix, err := InfixToPostfix(expr, operators)
	if err != nil {
		allErrors.Add(err)
	}

	// Step 3: Evaluate
	evalStack := NewLinkedListStack[float64]()
	result, err := EvaluatePostfix(postfix, evalStack)
	if err != nil {
		allErrors.Add(err)
	}

	if allErrors.HasErrors() {
		return 0, allErrors
	}

	return result, nil
}

func BenchmarkCalculateLinkedList(b *testing.B) {
	expr := "3 + 4 * 2 / (1 - 5) ^ 2 + sin(0) + cos(0)"
	for i := 0; i < b.N; i++ {
		_, _ = CalculateWithLinkedList(expr)
	}
}

func BenchmarkDeepNestedSlice(b *testing.B) {
	levels := 10000
	var expr string
	for i := 0; i < levels; i++ {
		expr += "("
	}
	expr += "1"
	for i := 0; i < levels; i++ {
		expr += ")"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Calculate(expr)
	}
}

func BenchmarkDeepNestedLinkedList(b *testing.B) {
	levels := 10000
	var expr string
	for i := 0; i < levels; i++ {
		expr += "("
	}
	expr += "1"
	for i := 0; i < levels; i++ {
		expr += ")"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateWithLinkedList(expr)
	}
}


func TestTokenize(t *testing.T) {
	t.Parallel()
	
}