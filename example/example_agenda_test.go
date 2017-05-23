/*

These are agenda-driven tests

*/
package example

import (
	"encoding/json"
	"testing"

	"github.com/iafan/agenda"
)

func TestSum(t *testing.T) {
	agenda.Run(t, "testdata/sum", func(path string, data []byte) ([]byte, error) {
		in := struct {
			A int `json:"a"`
			B int `json:"b"`
		}{}

		out := struct {
			Result int `json:"result"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		out.Result = Sum(in.A, in.B)

		return json.Marshal(out)
	})
}

func TestMul(t *testing.T) {
	agenda.Run(t, "testdata/mul", func(path string, data []byte) ([]byte, error) {
		in := struct {
			A int `json:"a"`
			B int `json:"b"`
		}{}

		out := struct {
			Result int `json:"result"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		out.Result = Mul(in.A, in.B)

		return json.Marshal(out)
	})
}

func TestDiv(t *testing.T) {
	agenda.Run(t, "testdata/div", func(path string, data []byte) ([]byte, error) {
		in := struct {
			A int `json:"a"`
			B int `json:"b"`
		}{}

		out := struct {
			Result int         `json:"result"`
			Error  interface{} `json:"error"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var err error
		out.Result, err = Div(in.A, in.B)
		out.Error = agenda.SerializableError(err)

		return json.Marshal(out)
	})
}
