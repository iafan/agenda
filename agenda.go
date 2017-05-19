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
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test defines the interface of an agenda test. It has three methods:
//
// - UnmarshalInput() takes raw bytes and populates internal data structure
// with input parameters for the test; it should return an error if there was a problem
// unmarshaling the data.
//
// - Run() runs the actual test against input data and stores result internally;
// it should return an error only if there was a problem running the supporting
// test code; if the code you're testing returns an error, and you want to capture
// and test such errors, they need to be placed in the internal fields and then
// marshaled in MarshalOutput().
//
// - MarshalOutput() serializes internal result data into raw bytes;
// it should return an error in case marshaling fails.
type Test interface {
	UnmarshalInput(data []byte) error
	Run() error
	MarshalOutput() ([]byte, error)
}

// optionSet is an internal structure that contains all the
// computed options before the tests are run with Run().
// The structure is not created or modified directly;
// use available OptionFunc options to modify individual options.
type optionSet struct {
	fileSuffix   string
	resultSuffix string
	initMode     bool
}

// option is a type of the function that can modify
// one or more of the options in the Options structure.
type option func(options *optionSet)

// FileSuffix allows you to set the suffix of the main test
// files (all files with this suffix will be gathered at Run() time).
//
// Default: ".json"
//
// Example:
// agenda.Run(t, "./testdata/mytest", agenda.FileSuffix(".rawdata"))
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
// agenda.Run(t, "./testdata/mytest", agenda.ResultSuffix(".out"))
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
// agenda.Run(t, "./testdata/mytest", agenda.InitMode(os.Getenv("INIT_TEST") != ""))
func InitMode(enabled bool) option {
	return func(o *optionSet) {
		o.initMode = enabled
	}
}

// Run executes an agenda test (`test`) against all input data files
// in the specified directory `dir`. Directory can be relative to the directory
// you run the tests from. One or more `option`s allow you to control the behavior
// of the tests.
//
// Example:
//
//     // TestSum satisfies `agenda.Test` interface and stores input/output data
//
//     type TestSum struct {
//         input struct {
//             A int `json:"a"`
//             B int `json:"b"`
//         }
//         output struct {
//             Result int `json:"result"`
//         }
//     }
//
//     func (t *TestSum) UnmarshalInput(data []byte) error {
//         return json.Unmarshal(data, &t.input)
//     }
//
//     func (t *TestSum) Run() error {
//         t.output.Result = t.input.A + t.input.B
//         return nil
//     }
//
//     func (t *TestSum) MarshalOutput() ([]byte, error) {
//         return json.Marshal(t.output)
//     }
//
//     // Run the test
//
//     func TestSumWithAgenda(t *testing.T) {
//         agenda.Run(t, "testdata/sum", &TestSum{})
//     }
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
		panic("test is nil")
	}

	opt := &optionSet{
		fileSuffix:   ".json",
		resultSuffix: ".result",
		initMode:     flag.Arg(0) == "init",
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
func processFile(t *testing.T, path string, test Test, opt *Options) {
	var referenceOutput []byte

	var resultPath = path + opt.resultSuffix

	// read JSON with test data

	t.Log(path)
	input, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Can't read the file: %v", err)
	}

	err = test.UnmarshalInput(input)
	if err != nil {
		t.Fatalf("Can't unmarshal file: %v", err)
	}

	if !opt.initMode {
		// test mode: read JSON with reference results

		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			t.Fatalf("File '%s' doesn't exist (try initializing snapshots with 'go test -args init')", resultPath)
		}

		referenceOutput, err = ioutil.ReadFile(resultPath)
		if err != nil {
			t.Fatalf("Can't read the '%s' file: %v", resultPath, err)
		}
	}

	// perform the actual test computation

	err = test.Run()
	if err != nil {
		t.Errorf("Error during Run(): %v", err)
	}

	// marshal the result of the computation

	output, err := test.MarshalOutput()
	if err != nil {
		t.Fatalf("Can't marshal output data: %v", err)
	}

	if !opt.initMode {
		// test mode: compare result with the reference data

		if !bytes.Equal(output, referenceOutput) {
			t.Errorf("%s contents don't match the actual output", resultPath)
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
