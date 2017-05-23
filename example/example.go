package example

import "errors"

// Sum calculates a sum of two integers
func Sum(a int, b int) int {
	return a + b
}

// Mul multiplies two integers
func Mul(a int, b int) int {
	return a * b
}

// Div performs an integer division of two integers
func Div(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("Division by zero")
	}
	return a / b, nil
}
