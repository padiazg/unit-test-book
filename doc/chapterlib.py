#!/usr/bin/env python3
"""Shared helpers for doc generators — read from generate.py."""

import os
import re

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CHAPTERS_DIR = os.path.join(PROJECT_ROOT, "chapters")
GITHUB_REPO = "https://github.com/padiazg/unit-test-book"

CHAPTER_ONE_LINERS = {
    1: "Classic Table-Driven Tests — struct fields define test cases with `wantErr bool` for expected pass/fail outcomes",
    2: "Value Assertions — compare production output against a `want T` value using `assert.Equal` or `assert.InDelta`",
    3: "Fields Struct for Inputs — group related test inputs into a dedicated struct for reusable test vectors",
    4: "Subtest Naming Strategies — use descriptive constants or natural language as subtest names",
    5: "Typed Check Functions — `type checkFn func(t, result, error)` enables composable, reusable assertion blocks",
    6: "Check Collection Builder — `var check = func(fns...)` collects multiple check functions into a single slice",
    7: "Check Factory Closures — `checkStatus(want)` returns a closure that captures the expected value",
    8: "Error Message Verification — `assert.Contains(err.Error(), \"substring\")` for error message inspection",
    9: "Output String Inspection — `assert.Contains/NotContains` on string output for content verification",
    10: "The `before` Hook Pattern — a typed fixture function returns fresh test state for each case",
    11: "HTTP Client Interface Mock — define an `HTTPClient{ Do() }` interface and stub its single method",
    12: "RoundTripper Mock — implement `http.RoundTripper` to mock at the transport layer without changing production types",
    13: "testify/mock for Interfaces — embed `mock.Mock`, use `On().Return()` for interface mock expectations",
    14: "Function Variable Injection — store `json.Marshal`, `http.NewRequest` as struct fields for test seams",
    15: "Package-Level Var Swap — override a package variable and restore with `defer` for minimal seam injection",
    16: "httptest.Server Integration — start a real HTTP server on a random port for full-stack HTTP tests",
    17: "httptest.ResponseRecorder — capture handler output without starting a server for pure handler unit tests",
    18: "Error Readers — implement `io.Reader` that returns errors on demand to test I/O failure paths",
    19: "Temporary Files & Parsing — use `t.TempDir()` for isolated file I/O with automatic cleanup",
    20: "Channel Delivery Tests — fan-out work to workers via buffered channels with graceful channel-close shutdown",
    21: "Panic Recovery in Tests — `defer recover()` converts panics to errors for safe testing of edge cases",
    22: "Goroutine Run Loops — select-based event loops with context cancellation and started-channel synchronization",
    23: "Goroutine Leak Detection — `goleak.VerifyNone(t)` catches leaked goroutines after each test completes",
    24: "AST Parsing for Cyclomatic Complexity — walk the Go AST with `ast.Inspect` to compute complexity metrics",
    25: "Benchmark Tests — compare implementations with `b.N` loops, sub-benchmarks, and `b.ResetTimer()`",
    26: "Parallel Tests — `t.Parallel()` with safe (mutex/atomic) vs unsafe concurrent access patterns",
    27: "Service Layer with Mocked Ports — mock repository and email interfaces to test business logic in isolation",
    28: "JSON Format Verification — `assert.JSONEq`, `json.MarshalIndent`, and round-trip serialization tests",
    29: "Setup/Teardown Fixtures — struct-based fixtures with Setup/Teardown for isolated, repeatable test state",
    30: "Composable Check Navigation — navigator factories `checkReportEntry(i, ...sub)` descend into nested output, delegating assertions to sub-checks",
    31: "Inline Check Closures — define assertions inline when a check is used once; no factory extraction needed",
    32: "Interface Extraction from Third-Party Deps — wrap a concrete library behind a small interface, inject through the constructor, swap with testify/mock in tests",
    33: "Event-Driven Run Loop Tests — `Run()` returns an event channel, `before`/`after` lifecycle with testify.Mock, `select`-with-timeout event capture",
    34: "Embedded Infrastructure Testing — use in-memory replacements like `miniredis.RunT(t)` for realistic protocol coverage without external processes",
    35: "Capturing Stdout/Stderr — swap `os.Stdout`/`os.Stderr` with an `os.Pipe()` writer to test functions that print directly to the console",
}


def chapter_number(dirname):
    m = re.search(r"chapter-(\d+)-", dirname)
    return int(m.group(1)) if m else None


def find_chapter_dirs():
    """Return sorted list of (number, dirname) for valid chapter dirs."""
    entries = sorted(os.listdir(CHAPTERS_DIR))
    dirs = []
    for e in entries:
        path = os.path.join(CHAPTERS_DIR, e)
        if os.path.isdir(path):
            n = chapter_number(e)
            if n and 1 <= n <= 99:
                dirs.append((n, e))
    return dirs


def parse_readme(path):
    """Parse a chapter README into (title, sections_dict)."""
    with open(path) as f:
        lines = f.readlines()

    sections = {}
    current_section = None
    current_lines = []
    in_code = False
    title = None

    for line in lines:
        if line.startswith("```"):
            in_code = not in_code
            current_lines.append(line)
            continue

        if not in_code and line.startswith("# Chapter "):
            title = line.strip()
            continue

        if not in_code and line.startswith("## "):
            if current_section:
                sections[current_section] = "".join(current_lines).strip()
            current_section = line[3:].strip()
            current_lines = []
            continue

        current_lines.append(line)

    if current_section:
        sections[current_section] = "".join(current_lines).strip()

    if title:
        m = re.match(r"# Chapter \d+: (.+)$", title)
        if m:
            title = m.group(1)

    return title, sections


def extract_code_block(section_text):
    """Extract the first Go code block from a section, or empty string."""
    if not section_text:
        return ""
    start = section_text.find("```go")
    if start < 0:
        return ""
    start += 5  # past "```go"
    end = section_text.find("```", start)
    if end < 0:
        return section_text[start:].strip()
    return section_text[start:end].strip()


def slug_from_dir(dirname):
    """Convert chapter-01-classic-table-driven → chapter-01-classic-table-driven"""
    return dirname
