package agenda

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Test01 represents a simple test with JSON serialization
type Test01 struct {
	input struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
		C float64 `json:"c"`
	}
	output struct {
		Sum         float64     `json:"sum"`
		Mul         float64     `json:"mul"`
		Div         float64     `json:"div"`
		Error       interface{} `json:"error"`
		Explanation string      `json:"explanation"`
	}
}

func (t *Test01) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *Test01) Run() error {
	// reset the output struct between the runs
	t.output.Div = 0
	t.output.Error = nil

	// fill the output struct with values
	t.output.Sum = t.input.A + t.input.B + t.input.C
	t.output.Mul = t.input.A * t.input.B * t.input.C
	if t.input.B == 0 {
		t.output.Error = SerializableError(errors.New("Can't divide: B is zero"))
	} else if t.input.C == 0 {
		t.output.Error = SerializableError(errors.New("Can't divide: C is zero"))
	} else {
		t.output.Div = float64(t.input.A) / float64(t.input.B) / float64(t.input.C)
	}
	t.output.Explanation = fmt.Sprintf(
		"Input parameters were: [%v, %v, %v]", t.input.A, t.input.B, t.input.C,
	)
	// no supporting code in the test itself produces errors,
	// so we always return nil
	return nil
}

func (t *Test01) MarshalOutput() ([]byte, error) {
	return json.Marshal(t.output)
}

// Test01RunDefault runs agenda tests with the default parameters
// (it will look for files ending with .json)
func Test01RunDefault(t *testing.T) {
	Run(t, "testdata/01/default", &Test01{})
}

// Test01RunWithCustomFileSuffix runs tests with custom file suffix option:
// only files ending with '.custom' will be considered as tests
func Test01RunWithCustomFileSuffix(t *testing.T) {
	Run(t, "testdata/01/custom-file-suffix", &Test01{},
		FileSuffix(".custom"))
}

// Test01RunWithCustomFileAndResultSuffix runs tests with
// custom file suffix anmd custom result file suffix:
// for each file ending with '.in' there will be a '.in.out' file
// with serialized results
func Test01RunWithCustomFileAndResultSuffix(t *testing.T) {
	Run(t, "testdata/01/custom-file-result-suffix", &Test01{},
		FileSuffix(".in"), ResultSuffix(".out"))
}

type FileRec struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// TestDirectorySnapshot represents a test that compares directories
type TestDirectorySnapshot struct {
	input struct {
		Path string `json:"path"`
	}
	output []*FileRec
}

func (t *TestDirectorySnapshot) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestDirectorySnapshot) Run() error {
	t.output = make([]*FileRec, 0)
	err := filepath.Walk(
		t.input.Path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			t.output = append(
				t.output,
				&FileRec{path, info.Size()},
			)
			return nil
		},
	)
	return err
}

func (t *TestDirectorySnapshot) MarshalOutput() ([]byte, error) {
	// Pretty-print the JSON for easier diff-ing
	return json.MarshalIndent(t.output, "", "\t")
}

// TestDirectorySnapshotRun01Default runs agenda tests
// to verify that the contents of the test folders generated
// by previous tests matches the expectations
func TestDirectorySnapshotRun01Default(t *testing.T) {
	Run(t, "testdata/dir-snapshots", &TestDirectorySnapshot{})
}
