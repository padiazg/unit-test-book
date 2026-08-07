---
section: http-io
---

# Chapter 35: Capturing Stdout/Stderr in Tests

## Description

Some functions print directly to `os.Stdout` or `os.Stderr` — CLI tools, HackerRank-style challenge solutions, and code that logs to the console. To test them, swap the global `os.Stdout`/`os.Stderr` with an `os.Pipe()` writer, run the function, restore the globals, and read back what was written.

Two capture implementations exist in the wild. A simple one swaps only `os.Stdout`, reads the pipe after the function returns, and restores manually. An elaborated one swaps both streams plus the `log` package output, restores with `defer` (panic-safe), and reads the pipe concurrently from a goroutine — avoiding the pipe-buffer deadlock when a function writes a lot of output.

Real-world examples:

- `hackerrank/problem-solving/easy/020-bon-appetit/main_test.go` and `hackerrank/problem-solving/hard/001-matrix-rotation-algo/main_test.go` — the simple `captureStdOut`, used to test solutions that print their answer to stdout
- `go-crap/internal/scan/scan_test.go:803` — the elaborated `captureOutput` (adapted from [rednafi.com/go/capture_console_output](https://rednafi.com/go/capture_console_output/)), capturing stdout, stderr, and `log` output for `logCoverageErrors`

## Code

```go
type Row struct {
	Label string
	Unit  string
	Value int
}

func PrintTable(rows []Row) {
	if len(rows) == 0 {
		fmt.Println("(empty)")
		return
	}

	fmt.Println("LABEL       | VALUE | UNIT")
	fmt.Println("------------+-------+------")
	for _, r := range rows {
		fmt.Printf("%-10s | %5d | %s\n", r.Label, r.Value, r.Unit)
	}
}

func ReportErrors(errs []error) {
	for _, err := range errs {
		log.Print(err)
	}
}

func WriteDiagnostics(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
```

`PrintTable` writes to stdout with `fmt.Println`/`fmt.Printf`. `ReportErrors` logs through the standard `log` package (stderr by default). `WriteDiagnostics` writes directly to `os.Stderr`. None of them return a string — the output is the side effect, so the test must capture the stream.

## Test

```go
// Simple version — swap os.Stdout, run, restore, read.
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

// Elaborated version — capture stdout + stderr + log, defer-restore,
// read concurrently to avoid pipe-buffer deadlock.
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
```

## Scaffold

```shell
go-testgen report . --format table
go-testgen gen . PrintTable
go-testgen gen . ReportErrors
go-testgen gen . WriteDiagnostics
```

## Testing Approach

Stream capture testing:

1. **Swap, run, restore, read** — the simple `captureStdOut` replaces the `os.Stdout` global with the write end of an `os.Pipe()`, invokes the function, restores the global, closes the writer, and reads the pipe. This is the classic HackerRank testing setup, where challenge functions print their answer rather than returning it.

2. **Elaborated version: deadlock-proof concurrent read** — `os.Pipe` has a kernel buffer (64KB on Linux). If a function writes more than the buffer holds and nothing reads until the writer closes, the write blocks forever. `captureOutput` starts a goroutine that drains the pipe into a `bytes.Buffer` via `io.Copy` while the function runs, then delivers the result on a channel. This is safe for any output size. The go-crap source additionally synchronizes goroutine startup with a `sync.WaitGroup`; the channel receive at the end is the only synchronization that really matters.

3. **Capture stderr and `log` output too** — `captureOutput` redirects `os.Stdout`, `os.Stderr`, and `log.SetOutput` all to the same pipe. This catches code that prints errors to stderr (`WriteDiagnostics`) and code that logs through the standard logger (`ReportErrors`). Note that both streams share one pipe here, so the captured string mixes them; separate pipes per stream would let you assert on each independently.

4. **`defer` restores are panic-safe** — the elaborated version restores `os.Stdout`, `os.Stderr`, and the `log` writer in a `defer`, so a panic inside `f()` cannot leak a redirected global (and, in the go-crap original, an unrestored `log` writer pointed at a closed pipe). The simple version restores only `os.Stdout`, manually, after `f()` returns.

5. **Not parallel-safe** — `os.Stdout` is a package global. Swapping it while another test runs concurrently (via `t.Parallel()`) would make tests race and corrupt each other's output. Capture tests must run sequentially, unlike the `t.Parallel()` patterns in chapter 26.

6. **Capture vs returning strings** — this pattern is only necessary when code prints directly to the streams (legacy code, CLI tools, challenge solutions you cannot change). When you own the API, prefer returning a string and asserting on it directly, as in chapter 09 (`assert.Contains` on the returned value). When you must capture, `Contains`/`NotContains` checks keep assertions resilient to formatting changes, reusing the go-crap check factories from chapters 6–7.
