#!/usr/bin/env python3
"""Generate bodies for all 35 pattern SKILL.md files from chapter READMEs.

Reads: doc/skills-manifest.yaml + chapters/chapter-NN-slug/README.md
Writes: go-testgen/skills/pattern-NN-slug/SKILL.md (both book and real repo)
"""

import os
import re
import sys
import yaml

# ── Paths ──────────────────────────────────────────────────────────────────

BOOK_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CHAPTERS_DIR = os.path.join(BOOK_ROOT, "chapters")
MANIFEST_PATH = os.path.join(BOOK_ROOT, "doc", "skills-manifest.yaml")
SKILLS_DIR_BOOK = os.path.join(BOOK_ROOT, "go-testgen", "skills")
SKILLS_DIR_REAL = "/home/pato/go/src/github.com/padiazg/go-testgen/skills"

# ── Helpers from chapterlib ────────────────────────────────────────────────

def _slug_from_dir(dirname):
    return dirname


def chapter_number(dirname):
    m = re.search(r"chapter-(\d+)-", dirname)
    return int(m.group(1)) if m else None


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


def find_chapter_dirs():
    entries = sorted(os.listdir(CHAPTERS_DIR))
    dirs = []
    for e in entries:
        path = os.path.join(CHAPTERS_DIR, e)
        if os.path.isdir(path):
            n = chapter_number(e)
            if n and 1 <= n <= 99:
                dirs.append((n, e))
    return dirs


# ── Manifest loader ────────────────────────────────────────────────────────

def load_manifest():
    with open(MANIFEST_PATH) as f:
        data = yaml.safe_load(f)
    return data


# ── Chapter slug resolution ───────────────────────────────────────────────

# Chapter manifest numbers → directory name mapping
# Some dirs have different slugs than the manifest entry name
DIR_NAME_MAP = {
    1: "chapter-01-classic-table-driven",
    2: "chapter-02-value-assertions",
    3: "chapter-03-fields-struct-inputs",
    4: "chapter-04-subtest-naming",
    5: "chapter-05-typed-check-functions",
    6: "chapter-06-check-collection-builder",
    7: "chapter-07-check-factory-closures",
    8: "chapter-08-error-message-verification",
    9: "chapter-09-output-string-inspection",
    10: "chapter-10-before-hook-pattern",
    11: "chapter-11-http-client-interface-mock",
    12: "chapter-12-roundtripper-mock",
    13: "chapter-13-testify-mock-interfaces",
    14: "chapter-14-function-variable-injection",
    15: "chapter-15-package-var-swap",
    16: "chapter-16-httptest-server",
    17: "chapter-17-httptest-response-recorder",
    18: "chapter-18-error-readers",
    19: "chapter-19-temp-files-and-parsing",
    20: "chapter-20-channel-delivery-tests",
    21: "chapter-21-panic-recovery",
    22: "chapter-22-goroutine-run-loops",
    23: "chapter-23-goroutine-leak-detection",
    24: "chapter-24-ast-parsing-complexity",
    25: "chapter-25-benchmark-tests",
    26: "chapter-26-parallel-tests",
    27: "chapter-27-service-layer-mocked-ports",
    28: "chapter-28-json-format-verification",
    29: "chapter-29-setup-teardown-fixtures",
    30: "chapter-30-composable-check-navigation",
    31: "chapter-31-inline-check-closures",
    32: "chapter-32-interface-extraction-3rd-party",
    33: "chapter-33-event-driven-run-loop",
    34: "chapter-34-embedded-infrastructure",
    35: "chapter-35-capturing-stdout-stderr",
}


def find_readme(chapter_num):
    """Find README for a chapter number, trying mapped dir name."""
    dirname = DIR_NAME_MAP.get(chapter_num)
    if dirname:
        path = os.path.join(CHAPTERS_DIR, dirname, "README.md")
        if os.path.exists(path):
            return path
    # fallback: try all dirs with this number
    for _, d in find_chapter_dirs():
        if chapter_number(d) == chapter_num:
            path = os.path.join(CHAPTERS_DIR, d, "README.md")
            if os.path.exists(path):
                return path
    return None


# ── Body generation ───────────────────────────────────────────────────────

def _extract_code_block(section_text):
    """Extract the first Go code block from a section."""
    if not section_text:
        return ""
    start = section_text.find("```go")
    if start < 0:
        return ""
    start += 5
    end = section_text.find("```", start)
    if end < 0:
        return section_text[start:].strip()
    return section_text[start:end].strip()


def _generate_when_to_use(desc_section):
    """Paraphrase the Description section into 'When to use'."""
    if not desc_section:
        return "No description available."
    # Clean up the description text
    text = desc_section.strip()
    # Remove markdown list items from real-world examples section
    # Take only the first paragraph(s) before "Real-world examples:"
    lines = text.split("\n")
    paragraphs = []
    current = []
    for line in lines:
        if line.startswith("Real-world examples") or line.startswith("- `"):
            break
        if line.strip() == "":
            if current:
                paragraphs.append(" ".join(current).strip())
                current = []
            continue
        current.append(line.strip())
    if current:
        paragraphs.append(" ".join(current).strip())

    if len(paragraphs) >= 3:
        return "\n\n".join(paragraphs[:3])
    elif len(paragraphs) >= 1:
        return paragraphs[0]
    return text[:200]


def _generate_pattern(desc_section, code_section, test_section):
    """Generate the Pattern section with code and test snippets."""
    parts = []
    code = _extract_code_block(code_section) if code_section else ""
    test = _extract_code_block(test_section) if test_section else ""

    # Add code block if available
    if code:
        parts.append("```go\n" + code + "\n```")

    # Add test block if available
    if test:
        if parts:
            parts.append("")
        parts.append("```go\n" + test + "\n```")

    return "\n".join(parts) if parts else "(no code snippets)"


def _generate_fill_cases(testing_approach_section):
    """Summarize the Testing Approach section."""
    if not testing_approach_section:
        return "No testing strategy documented for this chapter."

    text = testing_approach_section.strip()
    lines = text.split("\n")
    bullets = []
    current_bullet = None

    for line in lines:
        # Match numbered lists like "1. **Bold text** — explanation"
        m = re.match(r"(\d+)\.\s+((?:\*\*)?.+?)(?:\*\*)?\s*—\s*(.+)$", line)
        if m:
            if current_bullet and current_bullet.strip():
                bullets.append(current_bullet.strip())
            bold = m.group(2).replace("**", "")
            detail = m.group(3).strip()
            current_bullet = f"- **{bold}**: {detail}"
        elif line.strip().startswith("- ") and not re.match(r"\d+\.", line):
            bullets.append(line.strip())
        elif line.strip() and not line.startswith("###"):
            # continuation of current bullet
            if current_bullet:
                current_bullet += " " + line.strip()

    if current_bullet and current_bullet.strip():
        bullets.append(current_bullet.strip())

    if bullets:
        return "\n".join(bullets[:10])
    return text[:300]


def _generate_skill_body(chapter_num, style, desc_section, code_section, test_section, testing_section):
    """Generate the full body of a SKILL.md file."""
    when = _generate_when_to_use(desc_section)

    style_note = ""
    if style == "manual":
        style_note = "This pattern does NOT map to `go-testgen gen`. Write the test by hand."
    elif style in ("table", "check", "simple"):
        style_note = f"Run `go-testgen gen --file <source.go> --style {style}` to scaffold.\nThe scaffolding includes the test runner — do NOT modify it."

    pattern = _generate_pattern(desc_section, code_section, test_section)
    fill = _generate_fill_cases(testing_section)

    body = f"""## When to use

{when}

## go-testgen style

**{style}**

{style_note}

## Pattern

{pattern}

## Fill the cases

{fill}

## Cross-references

- Unit-test-book: chapter {chapter_num:02d}
- [Source code](https://github.com/padiazg/unit-test-book/tree/main/chapters/{_dir_name(chapter_num)}/)
"""
    return body


# ── Directory name resolver ────────────────────────────────────────────────

def _dir_name(chapter_num):
    d = DIR_NAME_MAP.get(chapter_num)
    if d:
        return d
    return f"chapter-{chapter_num:02d}-unknown"


def _pattern_dir(chapter_num):
    """pattern-01-table-validation from manifest entry."""
    return f"pattern-{chapter_num:02d}"


def _read_stub(path):
    """Read an existing stub SKILL.md and extract the frontmatter + header."""
    with open(path) as f:
        content = f.read()
    # Frontmatter is between first "---" markers
    first = content.find("---")
    if first < 0:
        return None, content
    second = content.find("---", first + 3)
    if second < 0:
        return None, content
    fm = content[:second + 3].strip()
    rest = content[second + 3:].strip()
    return fm, rest


def _write_skill(chapter_num, style, pattern_name, fm, body, target_dir):
    """Write a complete SKILL.md to target_dir."""
    skill_dir = os.path.join(target_dir, pattern_name)
    os.makedirs(skill_dir, exist_ok=True)
    path = os.path.join(skill_dir, "SKILL.md")

    content = f"""{fm}
{body.lstrip()}"""

    with open(path, "w") as f:
        f.write(content)
    return path


# ── Main ───────────────────────────────────────────────────────────────────

def main():
    manifest = load_manifest()
    chapters = manifest.get("chapters", {})

    # Build a lookup: chapter_num → manifest entry
    manifest_by_num = {}
    for num_str, entry in chapters.items():
        manifest_by_num[int(num_str)] = entry

    total = 0
    generated = 0
    errors = 0

    for chapter_num in sorted(manifest_by_num.keys()):
        entry = manifest_by_num[chapter_num]
        style = entry.get("style", "check")
        desc_section = None
        code_section = None
        test_section = None
        testing_section = None

        # Find and read the README
        readme_path = find_readme(chapter_num)
        if readme_path:
            title, sections = parse_readme(readme_path)
            desc_section = sections.get("Description", "")
            code_section = sections.get("Code", "")
            test_section = sections.get("Test", "")
            testing_section = sections.get("Testing Approach", "")
        else:
            print(f"WARNING: No README found for chapter {chapter_num}")
            errors += 1
            continue

        total += 1

        # Generate body
        body = _generate_skill_body(chapter_num, style, desc_section, code_section, test_section, testing_section)

        # Read stub and write full file to both locations
        pattern_name = entry.get("name", f"pattern-{chapter_num:02d}")
        for target_dir in [SKILLS_DIR_BOOK, SKILLS_DIR_REAL]:
            pattern_dir = pattern_name
            stub_path = os.path.join(target_dir, pattern_dir, "SKILL.md")
            if os.path.exists(stub_path):
                fm, _ = _read_stub(stub_path)
            else:
                fm = None
            if fm is None:
                print(f"  SKIP {chapter_num}: no stub at {stub_path}")
                errors += 1
                continue
            path = _write_skill(chapter_num, style, pattern_name, fm, body, target_dir)
            print(f"  ✓ {path}")
            generated += 1

    print(f"\nGenerated {generated} skill bodies out of {total} chapters. Errors: {errors}")
    if errors > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
