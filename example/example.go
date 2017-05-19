package example

import "errors"

func Sum(a int, b int) int {
	return a + b
}

func Mul(a int, b int) int {
	return a * b
}

func Div(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("Division by zero")
	}
	return a / b, nil
}
