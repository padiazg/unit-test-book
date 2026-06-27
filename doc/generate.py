#!/usr/bin/env python3
"""Generate MkDocs pages from chapter README.md files."""

import os
import re
import shutil
import sys

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DOCS_OUT = os.path.join(PROJECT_ROOT, "doc", "docs")
CHAPTER_DIR = PROJECT_ROOT
CHAPTERS_DIR = os.path.join(PROJECT_ROOT, "chapters")
GITHUB_REPO = "https://github.com/padiazg/unit-test-book"

SECTION_MAP = [
    ("foundations", "Foundations", (1, 4),
     "Four classic table-driven testing patterns. Each chapter builds on the previous, from basic `wantErr` assertions through value inspection and subtest naming conventions."),
    ("closure-check", "Closure-Check", (5, 9),
     "Seven chapters constructing a composable assertion system. Starting from typed check functions, the pattern evolves through collection builders, factory closures, inline checks, and composable navigators for nested output."),
    ("mocking", "Mocking", (10, 15),
     "Six techniques for isolating external dependencies. Covers HTTP client interfaces, RoundTripper transport mocking, testify/mock, function variable injection, and package-level var swap."),
    ("http-io", "HTTP / I/O", (16, 19),
     "Four patterns for testing HTTP handlers and I/O operations. Uses httptest.Server for integration tests, ResponseRecorder for handler unit tests, error readers for failure paths, and temp files for filesystem tests."),
    ("concurrency", "Concurrency", (20, 22),
     "Three chapters on goroutine safety and lifecycle. Covers channel-based worker pools, panic recovery with defer/recover, and select-based run loops with context cancellation."),
    ("advanced-go", "Advanced Go", (23, 26),
     "Four advanced testing topics: goroutine leak detection with goleak, AST-based cyclomatic complexity analysis, benchmark-driven performance comparison, and parallel test safety with -race."),
    ("integration", "Integration", (27, 29),
     "Three integration patterns: service-layer port mocking with testify/mock, JSON format verification and round-trip testing, and struct-based fixture setup/teardown for isolated test state."),
]

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
}


def chapter_number(dirname):
    m = re.search(r"chapter-(\d+)-", dirname)
    return int(m.group(1)) if m else None


def find_chapter_dirs():
    entries = sorted(os.listdir(CHAPTERS_DIR))
    dirs = []
    for e in entries:
        if os.path.isdir(os.path.join(CHAPTERS_DIR, e)):
            n = chapter_number(e)
            if n and 1 <= n <= 32:
                dirs.append((n, e))
    return dirs


def get_section(num):
    if num in (30, 31):
        return "closure-check", "Closure-Check", "Extends the closure-check pattern with composable navigator checks that select sub-elements of nested output and delegate assertions to sub-check functions."
    if num == 32:
        return "mocking", "Mocking", "Extends the mocking section with interface extraction from third-party dependencies — wrapping a concrete library behind a small interface and swapping it with testify/mock in tests."
    for slug, title, (lo, hi), desc in SECTION_MAP:
        if lo <= num <= hi:
            return slug, title, desc
    return None, None, None


def slug_from_dir(dirname):
    return dirname


def parse_readme(path):
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


def strip_realworld(text):
    if not text:
        return text
    lines = text.split("\n")
    filtered = [l for l in lines if not l.strip().startswith("Real-world example:")]
    return "\n".join(filtered).strip()


def write_chapter_page(out_dir, num, dirname, title, sections):
    slug = slug_from_dir(dirname)
    out_path = os.path.join(out_dir, f"{slug}.md")
    description = strip_realworld(sections.get("Description", ""))
    code = sections.get("Code", "")
    test = sections.get("Test", "")
    approach = sections.get("Testing Approach", "")

    parts = [f"# Chapter {num:02d}: {title}"]
    if description:
        parts.extend(["", "## Description", "", description])
    if code:
        parts.extend(["", "## Code", "", code])
    if test:
        parts.extend(["", "## Test", "", test])
    if approach:
        parts.extend(["", "## Testing Approach", "", approach])

    parts.extend(["", "---", "", f"[View source code]({GITHUB_REPO}/tree/master/chapters/{dirname}/) on GitHub"])

    with open(out_path, "w") as f:
        f.write("\n".join(parts) + "\n")
    return slug


def generate():
    # Clean slate — orphaned files from previous runs would cause
    # `mkdocs build --strict` warnings.
    if os.path.exists(DOCS_OUT):
        shutil.rmtree(DOCS_OUT)
    os.makedirs(DOCS_OUT, exist_ok=True)

    dirs = find_chapter_dirs()
    if not dirs:
        print("ERROR: no chapter directories found")
        sys.exit(1)

    section_pages = {slug: [] for slug, _, _, _ in SECTION_MAP}

    for num, dirname in dirs:
        readme_path = os.path.join(CHAPTERS_DIR, dirname, "README.md")
        if not os.path.exists(readme_path):
            print(f"WARN: {readme_path} not found")
            continue

        title, sections = parse_readme(readme_path)
        if not title:
            print(f"WARN: could not parse title from {readme_path}")
            continue

        section_slug, section_title, section_desc = get_section(num)
        if not section_slug:
            print(f"WARN: no section for chapter {num}")
            continue

        out_dir = os.path.join(DOCS_OUT, section_slug)
        os.makedirs(out_dir, exist_ok=True)

        slug = write_chapter_page(out_dir, num, dirname, title, sections)
        one_liner = CHAPTER_ONE_LINERS.get(num, "")
        section_pages[section_slug].append((num, slug, one_liner))
        print(f"  {dirname}/README.md -> {section_slug}/{slug}.md")

    for slug, title, (lo, hi), desc in SECTION_MAP:
        pages = section_pages.get(slug, [])
        out_dir = os.path.join(DOCS_OUT, slug)
        os.makedirs(out_dir, exist_ok=True)
        index_path = os.path.join(out_dir, "index.md")

        lines = [f"# {title}", "", desc, "", "## Chapters"]
        for num, p_slug, one_liner in pages:
            source = f"chapters/{p_slug}"
            lines.append(f"- [{p_slug}.md]({p_slug}.md) — {one_liner}  \n  Source: `{source}/`")
        lines.extend(["", "## Running the code", "",
                       "Each chapter is a standalone Go module. To run tests for a chapter:",
                       "", "```bash", "cd <source-directory>", "go test -v ./...", "```"])
        lines.append("")

        with open(index_path, "w") as f:
            f.write("\n".join(lines) + "\n")
        print(f"  index.md -> {slug}/index.md")

    root_readme = os.path.join(CHAPTER_DIR, "README.md")
    home_out = os.path.join(DOCS_OUT, "index.md")
    with open(root_readme) as f:
        content = f.read()
    for num, dirname in dirs:
        section_slug, _, _ = get_section(num)
        dir_re = os.path.join("chapters", dirname, "README.md")
        mkdocs_link = os.path.join(section_slug, f"{dirname}.md")
        content = content.replace(f"]({dir_re})", f"]({mkdocs_link})")
    with open(home_out, "w") as f:
        f.write(content)
    print(f"  {root_readme} -> index.md")


if __name__ == "__main__":
    generate()
