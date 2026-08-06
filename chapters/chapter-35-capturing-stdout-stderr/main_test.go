package capturing_stdout_stderr

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureStdOut(f func()) string {
	var (
		orig    = os.Stdout
		r, w, _ = os.Pipe()
	)

	os.Stdout = w
	f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func captureOutput(f func()) string {
	custReader, custWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	origStdout := os.Stdout
	origStderr := os.Stderr
	origLog := log.Writer()

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		log.SetOutput(origLog)
	}()

	os.Stdout, os.Stderr = custWriter, custWriter
	log.SetOutput(custWriter)

	out := make(chan string, 1)
	go func() {
		var buff bytes.Buffer
		io.Copy(&buff, custReader)
		out <- buff.String()
	}()

	f()

	_ = custWriter.Close()
	return <-out
}

type checkOutputFn func(*testing.T, string)

var checkOutput = func(fns ...checkOutputFn) []checkOutputFn { return fns }

var checkContains = func(want string) checkOutputFn {
	return func(t *testing.T, got string) {
		t.Helper()
		assert.Containsf(t, got, want, "output should contain %q", want)
	}
}

var checkNotContains = func(want string) checkOutputFn {
	return func(t *testing.T, got string) {
		t.Helper()
		assert.NotContainsf(t, got, want, "output should not contain %q", want)
	}
}

var checkEqual = func(want string) checkOutputFn {
	return func(t *testing.T, got string) {
		t.Helper()
		assert.Equalf(t, want, got, "output mismatch")
	}
}

var checkEmpty = func() checkOutputFn {
	return func(t *testing.T, got string) {
		t.Helper()
		assert.Emptyf(t, got, "output should be empty")
	}
}

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name   string
		rows   []Row
		checks []checkOutputFn
	}{
		{
			name: "empty_rows",
			rows: nil,
			checks: checkOutput(
				checkEqual("(empty)\n"),
				checkNotContains("LABEL"),
			),
		},
		{
			name: "single_row",
			rows: []Row{{Label: "CPU", Value: 45, Unit: "%"}},
			checks: checkOutput(
				checkContains("CPU"),
				checkContains("45"),
				checkContains("%"),
				checkContains("LABEL"),
				checkNotContains("MEM"),
			),
		},
		{
			name: "multiple_rows",
			rows: []Row{
				{Label: "CPU", Value: 45, Unit: "%"},
				{Label: "MEM", Value: 2048, Unit: "MB"},
			},
			checks: checkOutput(
				checkContains("CPU"),
				checkContains("MEM"),
				checkContains("45"),
				checkContains("2048"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdOut(func() {
				PrintTable(tt.rows)
			})
			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}

func TestReportErrors(t *testing.T) {
	tests := []struct {
		name   string
		errs   []error
		checks []checkOutputFn
	}{
		{
			name: "no_errors_no_output",
			errs: nil,
			checks: checkOutput(
				checkEmpty(),
			),
		},
		{
			name: "single_error_logged",
			errs: []error{errors.New("boom")},
			checks: checkOutput(
				checkContains("boom"),
			),
		},
		{
			name: "multiple_errors_logged",
			errs: []error{errors.New("err1"), errors.New("err2")},
			checks: checkOutput(
				checkContains("err1"),
				checkContains("err2"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(func() {
				ReportErrors(tt.errs)
			})
			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}

func TestWriteDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		checks []checkOutputFn
	}{
		{
			name: "writes_to_stderr",
			msg:  "disk full",
			checks: checkOutput(
				checkContains("disk full"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(func() {
				WriteDiagnostics(tt.msg)
			})
			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}
