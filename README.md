Agenda (Auto-GENerated DAta) Testing in Go
==========================================

Agenda testing is an approach where you store your tests in external data files
(usually JSON), and the same test infrastructure can be used to generate
your reference output data files (in initialization mode), or to compare
previously created reference files with current test results (in regular mode).

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

Storing both input data and test computation artifacts as external files,
along with the ability to re-generate all reference files and do e.g. `git diff`
afterwards, can bring better understanding on how your code behaves. As you commit
updated test artifacts along with the corresponding code changes, the diffs
will also help others during code review.

Status
======

This project, while being 'feature-complete' in the sense that it has everything
to be successfully used for testing purposes, is still in its infancy, so its API
and Test interface signature may change in the future. Your feedback is welcome!

Usage
=====

Install the Go package:
```
go get -u github.com/iafan/agenda
```

Now create the `example_test.go` file in any directory:

```go
package example

import "github.com/iafan/agenda"

// TestSum is a agenda.Test-compatible type
// that holds input and output data for the test,
// and does the computation and serialization
type TestSum struct {
    input struct {
        A int `json:"a"`
        B int `json:"b"`
    }
    output struct {
        Result int `json:"result"`
    }
}

// UnmarshalInput is called to unmarshal raw bytes
// from the external test file into test input structure.
func (t *TestSum) UnmarshalInput(data []byte) error {
    return json.Unmarshal(data, &t.input)
}

// Run executes the actual test. Note there's no assertion here;
// one just need to run the business logic
// and store the output internally.
func (t *TestSum) Run() error {
    t.output.Result = t.input.A + t.input.B
    return nil
}

// MarshalOutput serializes output to raw bytes
// suitable for saving into file and for comparison.
func (t *TestSum) MarshalOutput() ([]byte, error) {
    return json.Marshal(t.output)
}

// The actual test executed by `go test`;
// agenda.Run() will scan `testdata/sum` directory
// for .json files, and run test for each of them.
func TestSumWithAgenda(t *testing.T) {
    agenda.Run(t, "testdata/sum", &TestSum{})
}
```

Now run the test in initialization mode:
```
$ go test -args init
```

This will create the test directory structure for you (`testdata`->`sum`).

It's time to create some test data:
```
echo '{"a":1,"b":2}' >testdata/sum/1.json
echo '{"a":2,"b":3}' >testdata/sum/2.json
echo '{"a":-4,"b":5}' >testdata/sum/3.json
```

Run the test in initialization mode again to compute the results and save them as files:
```
$ go test -args init
```

Let's see what we've got:
```
$ cat testdata/sum/1.json.result
{"result":3}

$ cat testdata/sum/2.json.result
{"result":5}

$ cat testdata/sum/3.json.result
{"result":1}
```

Now these results are the artifacts that you can commit to version control
along with test data (*.json files).

Next time you change your code, you can run tests in a regular mode:
```
$ go test
```

In this mode, the freshly computed results of your tests will be
compared with the contents of previously saved .result files, and tests will fail
if they differ.

Now imagine that you changed your business logic, and this also brings
some expected changes to test results. Just run `go test -args init` again
and then analyze the diff using e.g. `git diff` or other favorite diffing tool.
This diff will give you a clear picture of how your new code behaves.
If everything looks good, commit your changed test results along with the change
to the code.

More Examples
=============

[example/](https://github.com/iafan/agenda/tree/master/example) directory contains a very simple project that has both traditional and agenda tests.

Tests for Agenda are written in Agenda as well. See [agenda_test.go](https://github.com/iafan/agenda/blob/master/agenda_test.go).

Pros
====
- Test code and test data are separated. When working in teams, this means one can add tests and analyze their output without modifying the code.
- Output data files help you visualize the data that you work with.
- You can quickly re-generate all reference files and do e.g. `git diff` afterwards to see what exact changes your modified code introduces.
- As you commit reference result files along with the corresponding code changes, the diffs will help others so code reviews.
- The approach works best for complex input and output data structures which are hard to maintain inside table-driven tests.
- Since the entire output snapshot is validated, it ensures that you won't miss some assertions. You automatically test every field of every structure.
- While your test coverage grows, your supporting test code can stay simple.

Cons
====
- You will have more files to work with and to commit to the repo.
- The approach might be an overkill for testing simple functions that accept basic input and output data types, and when the number of edge cases to test doesn't grow over time. In such cases, consider [table-driven tests](https://github.com/golang/go/wiki/TableDrivenTests) approach.
