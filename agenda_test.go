package agenda

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// test01 is a sample test callback function
// that uses JSON to store both input and output data
func test01(path string, data []byte) ([]byte, error) {
	in := struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
		C float64 `json:"c"`
	}{}

	out := struct {
		Sum         float64     `json:"sum"`
		Mul         float64     `json:"mul"`
		Div         float64     `json:"div"`
		Error       interface{} `json:"error"`
		Explanation string      `json:"explanation"`
	}{}

	if err := json.Unmarshal(data, &in); err != nil {
		return nil, err
	}

	out.Sum = in.A + in.B + in.C
	out.Mul = in.A * in.B * in.C
	if in.B == 0 {
		out.Error = SerializableError(errors.New("Can't divide: B is zero"))
	} else if in.C == 0 {
		out.Error = SerializableError(errors.New("Can't divide: C is zero"))
	} else {
		out.Div = float64(in.A) / float64(in.B) / float64(in.C)
	}
	out.Explanation = fmt.Sprintf(
		"Input parameters were: [%v, %v, %v]", in.A, in.B, in.C,
	)

	return json.Marshal(out)
}

// Test01RunDefault runs agenda tests with the default parameters
// (it will look for files ending with .json)
func Test01RunDefault(t *testing.T) {
	Run(t, "testdata/01/default", test01)
}

// Test01RunWithBinarySerializer runs agenda tests with a
// custom diff serializer (in this case, binary)
func Test01RunWithBinarySerializer(t *testing.T) {
	Run(t, "testdata/01/custom-serializer", test01,
		BinarySerializer())
}

// Test01RunWithCustomFileSuffix runs tests with custom file suffix option:
// only files ending with '.custom' will be considered as tests
func Test01RunWithCustomFileSuffix(t *testing.T) {
	Run(t, "testdata/01/custom-file-suffix", test01,
		FileSuffix(".custom"))
}

// Test01RunWithCustomFileAndResultSuffix runs tests with
// custom file suffix and custom result file suffix:
// for each file ending with '.in' there will be a '.in.out' file
// with serialized results
func Test01RunWithCustomFileAndResultSuffix(t *testing.T) {
	Run(t, "testdata/01/custom-file-result-suffix", test01,
		FileSuffix(".in"), ResultSuffix(".out"))
}

// TestBinarySerializer runs agenda tests
// to verify that binary serializer produces correct results
func TestBinarySerializer(t *testing.T) {
	Run(t, "testdata/binary-serializer", func(path string, data []byte) ([]byte, error) {
		out, err := serializeBinaryData(data)
		return []byte(out), err
	}, FileSuffix(".result"), ResultSuffix(".serialized"))
}

// TestDirectorySnapshotRun01Default runs agenda tests
// to verify that the contents of the test folders generated
// by previous tests matches the expectations
func TestDirectorySnapshotRun01Default(t *testing.T) {
	Run(t, "testdata/dir-snapshots", func(path string, data []byte) ([]byte, error) {
		type FileRec struct {
			Path string `json:"path"`
			Size int64  `json:"size"`
		}

		in := struct {
			Path string `json:"path"`
		}{}

		out := make([]*FileRec, 0)

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		if err := filepath.Walk(
			in.Path,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				out = append(
					out,
					&FileRec{path, info.Size()},
				)
				return nil
			},
		); err != nil {
			return nil, err
		}

		// Pretty-print the JSON for easier diff-ing
		return json.MarshalIndent(out, "", "\t")
	})
}
