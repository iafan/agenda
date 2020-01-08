/*

Package agenda provides an ability to run [Auto-GENerated DAta]-driven tests.

Agenda testing is an approach where you store your tests in external data files
(usually JSON), and the same test infrastructure can be used to generate your
reference output data files (in initialization mode), or to compare previously
created reference files with current test results (in regular mode).

Agenda allows you to focus on business logic of your tests (making sure you
run your business code with all required combination of input data), and spare
yourself from writing value/type/structure comparison logic. In some sense,
the core code for agenda-based tests somewhat resembles functional programming approach,
as your tests just take the input data, do the computation, and return
their artifacts. Agenda package takes care of data storage, retrieval and comparison.

While all this might sound complicated, Agenda is a very small package
with more documentation (and tests) than code.

Agenda works on top of standard 'testing' package and can be mixed together
with traditionally written unit tests; the test directory structure and file naming
are configurable as well. You can choose any file formats to store your input data,
and use any serialization format of the output data.

See https://github.com/iafan/agenda for more information and examples.

*/
package agenda

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Strum355/go-difflib/difflib"
)

// Test defines the callback function of an agenda test, which takes raw bytes
// (the contents of the test data file), de-serializes the input data
// and runs the test against it, then serializes the output and returns it, along with
// the error in case the supporting test code encountered an unexpected behavior.
type Test func(path string, data []byte) ([]byte, error)

// StringSerializerFunc defines the callback function that is used to serialize
// raw file byte data into a string suitable for diff-ing
type StringSerializerFunc func(data []byte) (string, error)

// optionSet is an internal structure that contains all the
// computed options before the tests are run with Run().
// The structure is not created or modified directly;
// use available OptionFunc options to modify individual options.
type optionSet struct {
	fileSuffix    string
	resultSuffix  string
	initMode      bool
	serializeFunc StringSerializerFunc
}

func serializeUTF8Bytes(data []byte) (string, error) {
	return string(data), nil
}

func serializeBinaryData(data []byte) (string, error) {
	return hex.Dump(data), nil
}

// option is a type of the function that can modify
// one or more of the options in the optionSet structure.
type option func(options *optionSet)

// FileSuffix allows you to set the suffix of the main test
// files (all files with this suffix will be gathered at Run() time).
//
// Default: ".json"
//
// Example:
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.FileSuffix(".rawdata"))
func FileSuffix(suffix string) option {
	return func(o *optionSet) {
		o.fileSuffix = suffix
	}
}

// ResultSuffix is appended to the file path of the main test file
// to get the file name of the generated result file.
//
// Default: ".result"
//
// Example:
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.ResultSuffix(".out"))
func ResultSuffix(suffix string) option {
	return func(o *optionSet) {
		o.resultSuffix = suffix
	}
}

// InitMode allows you to manually control the mode the test is run
// (initialization or regular). By default, the mode is determined
// by the presence of the "init" argument:
//
//     flag.Arg(0) == "init"
//
// This means that you can run `go test -args init` to initialize
// your agenda tests, and `go test` to tun the tests in regular mode.
//
// Example:
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.InitMode(os.Getenv("INIT_TEST") != ""))
func InitMode(enabled bool) option {
	return func(o *optionSet) {
		o.initMode = enabled
	}
}

// Serializer allows you to specify the callback function
// to serialize file contents into a string for diff-ing purposes.
// Serialization is used only for reporting purposes to highlight changes
// between the reference and actual data.
//
//     flag.Arg(0) == "init"
//
// This means that you can run `go test -args init` to initialize
// your agenda tests, and `go test` to tun the tests in regular mode.
//
// Example:
//
// function renderFile(data []byte) (string, error) {
//     // render data into a string structure
//     // ...
// }
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.Serializer(renderFile))
func Serializer(f StringSerializerFunc) option {
	return func(o *optionSet) {
		o.serializeFunc = f
	}
}

// BinarySerializer is a shortcut option that sets the binary data
// serializer function to render diffs for binary files
//
// Example:
//
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.BinarySerializer())
func BinarySerializer() option {
	return Serializer(serializeBinaryData)
}

// UTF8Serializer is a shortcut option that sets the UTF8 string
// serializer function to render diffs for plain-text files.
// It is used by default, and provided for completeness.
//
// Example:
//
// agenda.Run(t, "./testdata/mytest", testFunc, agenda.UTF8Serializer())
// // which is equivalent to:
// agenda.Run(t, "./testdata/mytest", testFunc)
func UTF8Serializer() option {
	return Serializer(serializeUTF8Bytes)
}

// Run executes an agenda test function (`test`) against all input data files
// in the specified directory `dir`. Directory can be relative to the directory
// you run the tests from. One or more `option`s allow you to control the behavior
// of the tests.
//
// Example:
//
//		agenda.Run(t, "testdata/sum", func(path string, data []byte) ([]byte, error) {
//			in := struct {
//				A int `json:"a"`
//				B int `json:"b"`
//			}{}
//
//			out := struct {
//				Result int `json:"result"`
//			}{}
//
//			if err := json.Unmarshal(data, &in); err != nil {
//				return nil, err
//			}
//
//			out.Result = in.A + in.B
//
//			return json.Marshal(out)
//		})
//
// When the test is run, it will scan "testdata/sum" directory
// for .json files, and run the test against each of them.
// Each test file has input data. Assume we have a test file 01.json
// with the following content:
//    {"a":1,"b":2}
// If we run tests in initialization mode (`go test -args init`),
// this test will produce the corresponding result file (01.json.result):
//    {"result":3}
// Next time the test is run in regular mode (`go test`), Agenda will
// read the 01.json.result file and compare it with the current test output.
func Run(t *testing.T, dir string, test Test, options ...option) {
	if test == nil {
		panic("test function is nil")
	}

	opt := &optionSet{
		fileSuffix:    ".json",
		resultSuffix:  ".result",
		initMode:      flag.Arg(0) == "init",
		serializeFunc: serializeUTF8Bytes,
	}

	for _, f := range options {
		f(opt)
	}

	if opt.initMode {
		t.Logf("Initializing snapshots for %s directory", dir)
	} else {
		t.Logf("Running snapshot-based tests for %s directory", dir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if opt.initMode {
			t.Logf("Creating directory '%s'", dir)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				t.Fatalf("Can't create the snapshot directory: %v", err)
			}
		} else {
			t.Fatalf("Snapshot directory '%s' doesn't exist (try initializing snapshots with 'go test -args init')", dir)
		}
	}

	// Process the files in the directory

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("Can't read the directory contents: %v", err)
	}

	found := false
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), opt.fileSuffix) {
			found = true
			processFile(t, filepath.Join(dir, f.Name()), test, opt)
		}
	}

	if !found && !opt.initMode {
		t.Fatalf("No files ending with '%s' found in '%s' directory", opt.fileSuffix, dir)
	}
}

// processFile is an internal function that deals with one source test file at a time
func processFile(t *testing.T, path string, test Test, opt *optionSet) {
	var referenceOutput []byte

	var resultPath = path + opt.resultSuffix

	// read JSON with test data

	t.Log(path)
	input, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Can't read the file: %v", err)
	}

	if !opt.initMode {
		// test mode: read reference results

		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			t.Fatalf("File '%s' doesn't exist (try initializing snapshots with 'go test -args init')", resultPath)
		}

		referenceOutput, err = ioutil.ReadFile(resultPath)
		if err != nil {
			t.Fatalf("Can't read the '%s' file: %v", resultPath, err)
		}
	}

	// perform the actual test computation

	output, err := test(path, input)
	if err != nil {
		t.Errorf("Error during test() call: %v", err)
	}

	// marshal the result of the computation

	if !opt.initMode {
		// test mode: compare result with the reference data
		// and print the diff when the test fails

		if !bytes.Equal(output, referenceOutput) {
			mainErrText := fmt.Sprintf("Reference %s contents don't match the generated output.", resultPath)

			if opt.serializeFunc == nil {
				t.Errorf("%s Also, no data serialization function provided; can't render a diff.", mainErrText)
				return
			}

			refStr, refErr := opt.serializeFunc(referenceOutput)
			if refErr != nil {
				t.Errorf("%s Also, serializing reference output data failed: %v",
					mainErrText, refErr)
				return
			}

			outStr, outErr := opt.serializeFunc(output)
			if outErr != nil {
				t.Errorf("%s Also, serializing generated output data failed: %v",
					mainErrText, outErr)
				return
			}

			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(refStr),
				B:        difflib.SplitLines(outStr),
				FromFile: resultPath + " (reference)",
				ToFile:   resultPath + " (generated)",
				Context:  3,
				Colored:  true,
			}
			text, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				t.Errorf("%s Also, generating the diff failed: %v",
					mainErrText, err)
				return
			}

			t.Errorf("%s Here's the diff:\n\n%s\n", mainErrText, text)
		}
	} else {
		// init mode: save reference data

		t.Logf("Writing file '%s'", resultPath)
		err = ioutil.WriteFile(resultPath, output, 0644)
		if err != nil {
			t.Fatalf("Can't save file: %v", err)
		}
	}
}
