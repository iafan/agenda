/*

Standard table-driven tests (provided for comparison)

*/
package example

import (
	"testing"
)

func TestSumWithTable(t *testing.T) {
	var tests = []struct {
		a      int
		b      int
		result int
	}{
		{1, 2, 3},
		{2, 3, 5},
		{3, 4, 7},
		{-1, 1, 0},
	}

	for _, test := range tests {
		t.Logf("%d + %d", test.a, test.b)
		if result := Sum(test.a, test.b); result != test.result {
			t.Errorf("Expected %d, got %d", test.result, result)
		}
	}
}

func TestMulWithTable(t *testing.T) {
	var tests = []struct {
		a      int
		b      int
		result int
	}{
		{1, 2, 2},
		{2, 3, 6},
		{3, 4, 12},
		{-1, 1, -1},
	}

	for _, test := range tests {
		t.Logf("%d * %d", test.a, test.b)
		if result := Mul(test.a, test.b); result != test.result {
			t.Errorf("Expected %d, got %d", test.result, result)
		}
	}
}

func TestDivWithTable(t *testing.T) {
	var tests = []struct {
		a      int
		b      int
		result int
		errStr string
	}{
		{2, 1, 2, ""},
		{5, 2, 2, ""},
		{10, 2, 5, ""},
		{-1, 1, -1, ""},
		{1, 0, 0, "Division by zero"},
	}

	for _, test := range tests {
		t.Logf("%d * %d", test.a, test.b)
		result, err := Div(test.a, test.b)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		if errStr != test.errStr {
			t.Errorf("Expected error '%s', got '%s'", test.errStr, errStr)
		} else if result != test.result {
			t.Errorf("Expected %d, got %d", test.result, result)
		}
	}
}
