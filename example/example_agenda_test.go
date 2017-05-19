/*

These are data-snapshot-driven tests

*/
package example

import (
	"encoding/json"
	"testing"

	"github.com/iafan/agenda"
)

// Sum

type TestSum struct {
	input struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	output struct {
		Result int `json:"result"`
	}
}

func (t *TestSum) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestSum) Run() error {
	t.output.Result = Sum(t.input.A, t.input.B)
	return nil
}

func (t *TestSum) MarshalOutput() ([]byte, error) {
	return json.Marshal(t.output)
}

func TestSumWithAgenda(t *testing.T) {
	agenda.Run(t, "testdata/sum", &TestSum{})
}

// Mul

type TestMul struct {
	input struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	output struct {
		Result int `json:"result"`
	}
}

func (t *TestMul) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestMul) Run() error {
	t.output.Result = Mul(t.input.A, t.input.B)
	return nil
}

func (t *TestMul) MarshalOutput() ([]byte, error) {
	return json.Marshal(t.output)
}

func TestMulWithAgenda(t *testing.T) {
	agenda.Run(t, "testdata/mul", &TestMul{})
}

// Div

type TestDiv struct {
	input struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	output struct {
		Result int         `json:"result"`
		Error  interface{} `json:"error"`
	}
}

func (t *TestDiv) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestDiv) Run() error {
	var err error
	t.output.Result, err = Div(t.input.A, t.input.B)
	t.output.Error = agenda.SerializableError(err)

	return nil
}

func (t *TestDiv) MarshalOutput() ([]byte, error) {
	return json.Marshal(t.output)
}

func TestDivWithAgenda(t *testing.T) {
	agenda.Run(t, "testdata/div", &TestDiv{})
}
