# MAINTAIN-SKILLS.md — Regenerate pattern skills from unit-test-book

Run this when you've added, updated, or removed a chapter in the book.
It regenerates the 35 pattern SKILL.md bodies from the chapter READMEs.

---

## Prerequisites

1. `doc/skills-manifest.yaml` — must have a row for every chapter.
   If a chapter is missing, add it with `style` and `route_keys` then re-run.
2. `doc/generate_skills.py` — already ran; stub SKILL.md files exist.

## Steps

### 1. Validate the manifest

```bash
cd doc
python3 generate_skills.py
```

This validates: every chapter has a manifest row, style is valid,
route_keys non-empty, no duplicates.

If it errors, fix `skills-manifest.yaml` and retry.

### 2. For each changed chapter, update its pattern SKILL.md

For each chapter that was added, updated, or removed:

1. Read `chapters/chapter-NN-slug/README.md`
2. Read the stub: `go-testgen/skills/pattern-NN-slug/SKILL.md`
3. Overwrite the SKILL.md body with this structure:

```markdown
# <SkillName>

> unit-test-book chapter NN — style: <go-testgen-style>

## When to use

<Paraphrase the chapter's Description section in 3-5 sentences.
Focus on what kind of function/signature triggers this pattern.>

## go-testgen style

<state: style is the go-testgen style from the manifest.>

<!-- If style is manual: -->
<!-- This pattern does NOT map to `go-testgen gen`. Write the test by hand. -->

<!-- If style is table/check/simple: -->
<!-- Run `go-testgen gen --file <source.go> --style <style>` to scaffold. -->
<!-- The scaffolding includes the test runner — do NOT modify it. -->

## Pattern

<!-- Extract the Code and Test sections from the README.
Keep the most relevant snippets. Omit boilerplate that the scaffolding
already provides. Focus on what the developer needs to SEE to write
their own test. -->

<Code block>

<Test block>

## Fill the cases

<!-- Extract the Testing Approach section. Summarize the testing strategy:
- How test cases are structured
- What assertions are used
- How mocks/fixtures are set up (before hooks)
- Edge cases to cover -->

<Summarized testing strategy>

## Cross-references

- Unit-test-book: chapter NN
- [Source code](https://github.com/padiazg/unit-test-book/tree/main/chapters/chapter-NN-slug/)
- Related skill: pattern-NN-other (if applicable)
```

### 3. Verify frontmatter

Each SKILL.md must have:
- `name` = directory name
- `version` = matches manifest version
- `go-testgen-style` = the style from manifest
- `chapter` = the chapter number
- `style` in frontmatter must match `go-testgen-style` in metadata

### 4. Remove stale skills

If a chapter was removed from the book:
1. Remove the corresponding directory from `go-testgen/skills/`
2. Remove the row from `skills-manifest.yaml`

### 5. Final validation

```bash
python3 generate_skills.py --dry-run
```

Check there are no orphan skills or manifest mismatches.

---

## Tips

- **Read the README fully** before writing. The Description, Code, Test,
  and Testing Approach sections together paint the full picture.
- **Keep it concise.** Aim for ~100-150 lines per skill. Skip redundant
  boilerplate that's already in closure-check-tests or gen-test-cases.
- **Be specific about assertions.** When the chapter uses `assert.Equal`,
  say so. When it uses `assert.InDelta`, note the tolerance.
- **Note the go-testgen style clearly.** The AI that reads this skill
  needs to know whether to run `go-testgen gen --style <X>` or write by hand.
