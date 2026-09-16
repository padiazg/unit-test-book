#!/usr/bin/env python3
"""Deterministic generator for go-testgen skill pipeline.

Reads doc/skills-manifest.yaml + chapters/*/README.md and generates:
  - shared/scenario-map.md  (lookup table for router)
  - testgen-scenario/SKILL.md  (router — static body + injected table)
  - Plugin manifests  (skill lists in .claude-plugin, .cursor-plugin, gemini)
  - doc/docs/skills.sh  (loop install)
  - go-testgen AGENTS.md  (skill listing)

Does NOT generate per-chapter pattern SKILL.md bodies (that's the AI
maintenance skill's job — see MAINTAIN-SKILLS.md).  It DOES create
stub files so the directory structure exists.
"""

import json
import os
import re
import shutil
import sys

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
BOOK_ROOT = os.path.dirname(SCRIPT_DIR)
CHAPTERS_DIR = os.path.join(BOOK_ROOT, "chapters")
MANIFEST_PATH = os.path.join(SCRIPT_DIR, "skills-manifest.yaml")

# Import shared helpers from chapterlib
sys.path.insert(0, SCRIPT_DIR)
import chapterlib  # noqa: E402

# ---------------------------------------------------------------------------
# Minimal YAML parser (no external deps — only the subset we need)
# ---------------------------------------------------------------------------

def parse_manifest(path):
    """Parse skills-manifest.yaml into a dict. Handles scalars, lists, nested maps."""
    with open(path) as f:
        raw = f.read()

    top = {}
    chapters = {}
    current_key = None
    current_num = None
    current_map = None
    in_chapters = False

    for line in raw.splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue

        # Top-level keys (indent 0)
        if not line.startswith(" ") and ":" in stripped:
            # Finish previous chapter map
            if current_num is not None and current_map is not None:
                chapters[current_num] = current_map
            in_chapters = False
            key, _, val = stripped.partition(":")
            key = key.strip()
            raw_val = val.strip()
            if key == "chapters":
                in_chapters = True
                top[key] = {}
            elif raw_val.startswith("[") and raw_val.endswith("]"):
                # Top-level inline list
                items = raw_val[1:-1].split(",")
                top[key] = [
                    item.strip().strip('"').strip("'")
                    for item in items
                    if item.strip()
                ]
            elif raw_val.lower() == "true":
                top[key] = True
            elif raw_val.lower() == "false":
                top[key] = False
            elif raw_val:
                top[key] = raw_val.strip('"').strip("'")
            continue

        if not in_chapters:
            continue

        # Chapter number line: `  42:`
        num_m = re.match(r"^  +(\d+):", line)
        if num_m:
            # Save previous
            if current_num is not None and current_map is not None:
                chapters[current_num] = current_map
            current_num = int(num_m.group(1))
            current_map = {}
            current_key = None
            continue

        # Inside a chapter map
        if current_map is None:
            continue

        # Key: value or Key: with multiline list
        kv_m = re.match(r"^    +(\w[\w-]*):\s*(.*)", line)
        if kv_m:
            key = kv_m.group(1)
            val = kv_m.group(2).strip()
            if val == "":
                # Start of a list or multiline value
                current_map[key] = []
                current_key = key
            elif val.startswith("[") and val.endswith("]"):
                # Inline list
                items = val[1:-1].split(",")
                current_map[key] = [
                    item.strip().strip('"').strip("'")
                    for item in items
                    if item.strip()
                ]
            else:
                current_map[key] = val.strip('"').strip("'")
                current_key = key
            continue

        # Continuation line (more indentation)
        cont_m = re.match(r"^      +(.+)", line)
        if cont_m and current_key:
            item = cont_m.group(1).strip().strip('"').strip("'")
            if isinstance(current_map.get(current_key), list):
                current_map[current_key].append(item)
            else:
                current_map[current_key] += "\n" + item

    # Save last chapter
    if current_num is not None and current_map is not None:
        chapters[current_num] = current_map

    top["chapters"] = chapters
    return top


# ---------------------------------------------------------------------------
# Generation functions
# ---------------------------------------------------------------------------

def validate(manifest, chapters_dir):
    """Run validation gates. Returns list of error strings."""
    errors = []
    demand = manifest.get("demand_style", True)
    require_keys = manifest.get("require_route_keys", True)
    chapters = manifest.get("chapters", {})
    exempt = manifest.get("exempt_skills", [])
    valid_styles = {"table", "check", "simple", "manual"}

    # Discover actual chapter dirs
    actual = set(n for n, _ in chapterlib.find_chapter_dirs())
    manifest_keys = set(chapters.keys())

    if demand:
        missing = actual - manifest_keys
        if missing:
            errors.append(
                f"demand_style: true — missing manifest rows for chapters: {sorted(missing)}"
            )
        extra = manifest_keys - actual
        if extra:
            errors.append(
                f"manifest rows for chapters that no longer exist: {sorted(extra)}"
            )

    for num, row in sorted(chapters.items()):
        name = row.get("name", "")
        style = row.get("style", "")
        keys = row.get("route_keys", [])

        if style not in valid_styles:
            errors.append(f"ch.{num}: invalid style {style!r}")

        if require_keys and not keys:
            errors.append(f"ch.{num}: route_keys is empty (require_route_keys=true)")

        if not name:
            errors.append(f"ch.{num}: name is missing")

    # Check for duplicate route_keys across chapters
    all_keys = {}
    for num, row in sorted(chapters.items()):
        for k in (row.get("route_keys") or []):
            all_keys.setdefault(k, []).append(num)
    dupes = {k: nums for k, nums in all_keys.items() if len(nums) > 1}
    if dupes:
        errors.append(
            f"route_keys appear in multiple chapters: {dict(dupes)}"
        )

    return errors


def generate_scenario_map(chapters):
    """Generate shared/scenario-map.md from manifest chapters."""
    lines = [
        "# Scenario Map",
        "",
        "Auto-generated from skills-manifest.yaml — lookup table for the router.",
        "",
        "| Chapter | Skill | Style | Scenario |",
        "|---------|-------|-------|----------|",
    ]
    for num in sorted(chapters.keys()):
        row = chapters[num]
        name = row.get("name", f"pattern-{num:02d}-unknown")
        style = row.get("style", "check")
        scenario = row.get("scenario", "")
        lines.append(f"| {num:02d} | `{name}` | {style} | {scenario} |")

    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## Route Key Index")
    lines.append("")
    lines.append("Flat key → chapter → style (for router matching).")
    lines.append("")

    # Build reverse index: route_key → (chapter, style)
    key_index = {}
    for num in sorted(chapters.keys()):
        row = chapters[num]
        style = row.get("style", "check")
        for k in (row.get("route_keys") or []):
            key_index.setdefault(k, []).append(f"{num:02d} ({style})")

    lines.append("| Key | Chapters |")
    lines.append("|-----|----------|")
    for k in sorted(key_index.keys()):
        chs = ", ".join(key_index[k])
        lines.append(f"| `{k}` | {chs} |")
    lines.append("")
    return "\n".join(lines) + "\n"


def generate_router_skill(chapters, version="1.0.0", emoji="🧭"):
    """Generate testgen-scenario/SKILL.md (router)."""
    # Build detection table
    detection_rows = []
    for num in sorted(chapters.keys()):
        row = chapters[num]
        style = row.get("style", "check")
        name = row.get("name", f"pattern-{num:02d}")
        scenario = row.get("scenario", "")
        keys = row.get("route_keys", [])
        keys_str = ", ".join(f"`{k}`" for k in keys[:3])
        detection_rows.append(
            f"| `{num:02d}` | {style:6} | {scenario[:80]} | {keys_str} |"
        )

    detection_table = "\n".join([
        "| Ch | Style | Scenario | Route Keys |",
        "|----|-------|----------|------------|",
        *detection_rows,
    ])

    content = f"""---
name: testgen-scenario
description: >
  Route any Go function/method to the right test pattern from unit-test-book.
  Detect the scenario from the function signature and behavior, look up the
  matching pattern-NN-* skill, and delegate to it for style selection
  (table | check | simple | manual) and case-filling guidance.
license: MIT
compatibility: Designed for Claude Code or similar AI coding agents.
user-invocable: false
metadata:
  author: padiazg
  version: "{version}"
  openclaw:
    emoji: "{emoji}"
    homepage: https://github.com/padiazg/go-testgen
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Write Glob Grep Bash(go:*) Bash(go-testgen:*) Bash(git:*)
---

# Scenario Router

Detect the pattern. Route to the right skill. Fill the cases.

## Decision flow

1. **Read the source function.** Signature, return types, dependencies.
2. **Match against the scenario table below.** First match wins.
3. **Load the matching** `pattern-NN-*` **skill.** It carries the go-testgen
   style and case-filling guidance.
4. **Run** `go-testgen gen --style <style>` (or write by hand for `manual`).
5. **Fill the test cases** using the loaded skill's guidance.

{detection_table}

## Quick detection cues

| Signature cue | → Load |
|---|---|
| `(T, error)` constructor | ch01 table, ch08 error-msg |
| Pure value return | ch02 table |
| Many input params | ch03 table |
| `≥2` return values | ch05 check |
| Interface method + mock | ch11/13 check |
| Function field injection | ch14 check |
| Channel/chan + select | ch20 check |
| goroutine + context | ch22 check |
| Panic/defer recover | ch21 check |
| t.Parallel / race | ch26 manual |
| b.N / benchmark | ch25 manual |
| goleak | ch23 manual |
| httptest.Server | ch16 manual |
| miniredis / embedded | ch34 manual |
| os.Pipe / stdout | ch35 manual |

## Workflow

```bash
# 1. Detect → load the matching pattern skill
# 2. Generate scaffolding
go-testgen gen --file <source.go> --style <style>

# 3. Open the generated _test.go
# → Fill cases using the loaded skill's guidance

# 4. Run tests
go test ./... -v
```

## Cross-references

- `gen-test-cases` — scaffolding workflow (always read first)
- `closure-check-tests` — check-pattern mechanics (always read first)
- Book chapters: https://github.com/padiazg/unit-test-book/chapters/
"""
    return content


def generate_plugin_manifests(skill_names, exempt, version="1.0.0"):
    """Generate plugin.json content for .claude-plugin, .cursor-plugin, gemini."""
    skills_list = exempt + sorted(skill_names)
    skills_str = json.dumps(skills_list, indent=2)

    base = {
        "name": "go-testgen",
        "displayName": "Go Test Generator Skills",
        "description": "AI Agent Skills for scaffolding and writing Go tests with go-testgen",
        "version": version,
        "author": {"name": "padiazg", "email": "padiazg@gmail.com"},
        "homepage": "https://github.com/padiazg/go-testgen",
        "repository": "https://github.com/padiazg/go-testgen",
        "license": "MIT",
        "keywords": ["plugin", "go", "golang", "testing", "ai-agent", "skills"],
        "skills": "./skills/",
    }

    plugin_json = json.dumps(base, indent=4) + "\n"
    # Inject skills list (skills field is last, so no trailing comma)
    plugin_json = plugin_json.replace(
        '"skills": "./skills/"',
        '"skills": "./skills/",\n    "skillsList": ' + skills_str
    )

    # Gemini is simpler
    gemini = {
        "name": "go-testgen",
        "version": version,
        "description": "AI Agent Skills for scaffolding and writing Go tests with go-testgen",
        "skills": skills_list,
    }
    gemini_json = json.dumps(gemini, indent=4) + "\n"

    return plugin_json, gemini_json


def generate_skills_sh(skill_names, exempt):
    """Generate skills.sh that loops over all skills."""
    all_skills = sorted(set(list(skill_names) + exempt))
    lines = [
        "#!/bin/sh",
        "set -eu",
        "",
        "# skills.sh — Auto-installs all go-testgen AI agent skills",
        "# Generated from skills-manifest.yaml — do not edit by hand.",
        "",
        "REPO=padiazg/go-testgen",
        "BRANCH=${BRANCH:-main}",
        "TARGET=${TARGET:-$HOME/.agents/skills}",
        'BASE="https://raw.githubusercontent.com/$REPO/$BRANCH/skills"',
        "",
        'echo "Installing go-testgen AI agent skills to $TARGET"',
        "",
    ]
    for skill in all_skills:
        lines.append(f'mkdir -p "$TARGET/{skill}"')
        lines.append(f'curl -fsSL "$BASE/{skill}/SKILL.md" -o "$TARGET/{skill}/SKILL.md"')
        lines.append(f'echo "  ✔ {skill}"')
        lines.append("")

    lines.append('mkdir -p "$TARGET/go-testgen"')
    lines.append('curl -fsSL "$BASE/AGENTS.md" -o "$TARGET/go-testgen/AGENTS.md"')
    lines.append('echo "  ✔ AGENTS.md"')
    lines.append("")
    lines.append('echo "Done. Skills installed to $TARGET"')
    return "\n".join(lines) + "\n"


def generate_agents_md(chapters, exempt):
    """Generate skills/AGENTS.md with the full skill listing."""
    router_name = "testgen-scenario"
    lines = [
        "# go-testgen — AI agent instructions",
        "",
        "## Workflow",
        "",
        "```",
        "go-testgen gen <package> <FuncSpec>   →   _test.go scaffolding",
        "IA lee los skills + código fuente     →   completa los test cases",
        "go test ./...                         →   verde",
        "```",
        "",
        "No hay paso intermedio. No hay archivos de spec. La IA genera los cases",
        "directamente a partir del código fuente aplicando los skills.",
        "",
        "## Skills disponibles",
        "",
        "Están en `skills/`.",
        "",
        "| Skill | Cuándo leerlo |",
        "|---|---|",
        f"| `{router_name}` | Siempre — detecta el patrón y enruta al skill correcto |",
        "| `closure-check-tests` | Siempre — enseña la mecánica del patrón |",
        "| `gen-test-cases` | Siempre — enseña qué casos generar |",
        "",
        "### Pattern skills (uno por capítulo del libro)",
        "",
        "| Skill | Capítulo | Estilo |",
        "|---|---|---|",
    ]

    for num in sorted(chapters.keys()):
        row = chapters[num]
        name = row.get("name", f"pattern-{num:02d}")
        style = row.get("style", "check")
        scenario = row.get("scenario", "")[:60]
        lines.append(f"| `{name}` | {num:02d} | {style} | {scenario} |")

    lines.extend([
        "",
        "## Qué genera `go-testgen gen`",
        "",
        "El scaffolding ya incluye:",
        "- El tipo `checkXxxFn` con los parámetros correctos según la firma de la función",
        "- El composer `var checkXxx = func(fns ...checkXxxFn) []checkXxxFn { return fns }`",
        "- El struct del slice `tests` con los campos `name`, `before`, `after` (si aplica), `checks`",
        "- El test runner completo (el `for` loop con `t.Run`, la construcción del sujeto, la llamada a la función, la ejecución de checks)",
        "- El slice `tests` vacío",
        "",
        "**No modificar** el test runner. Solo agregar entries al slice `tests` y",
        "los check functions y fixtures que van encima.",
        "",
        "## Cómo pedir a la IA que genere los cases",
        "",
        "Prompt mínimo:",
        "```",
        "Lee skills/testgen-scenario/SKILL.md, skills/gen-test-cases/SKILL.md y skills/closure-check-tests/SKILL.md.",
        "Lee el archivo fuente. Detecta el patrón y genera los test cases.",
        "```",
        "",
        "Con contexto DDD:",
        "```",
        "Lee skills/testgen-scenario/SKILL.md, skills/gen-test-cases/SKILL.md y skills/closure-check-tests/SKILL.md.",
        "Contexto del dominio: [descripción de entidades, invariantes, puertos].",
        "Lee el archivo fuente. Genera los test cases.",
        "```",
        "",
        "## En el pipeline hexago",
        "",
        "```",
        "hexago generate [prompt DDD]",
        "  └─► estructura hexagonal + interfaces",
        "      └─► go-testgen gen [cada adapter/usecase/entity]",
        "          └─► IA aplica skills + contexto DDD → test cases completos",
        "```",
        "",
    ])
    return "\n".join(lines) + "\n"


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) > 1 and sys.argv[1] == "--help":
        print("Usage: generate_skills.py [--dry-run]")
        print("  Reads doc/skills-manifest.yaml + chapters/*/README.md")
        print("  Generates skills/ and plugin manifests.")
        return

    dry_run = "--dry-run" in sys.argv

    # Load manifest
    manifest = parse_manifest(MANIFEST_PATH)
    chapters = manifest.get("chapters", {})
    if not chapters:
        print("ERROR: no chapters in manifest")
        sys.exit(1)

    target_dir = os.path.join(
        SCRIPT_DIR,
        manifest.get("target_dir", "../go-testgen/skills"),
    )
    plugin_dir = os.path.join(
        SCRIPT_DIR,
        manifest.get("plugin_dir", "../../.agents/skills/go-testgen"),
    )

    # Resolve to absolute
    target_dir = os.path.abspath(target_dir)
    plugin_dir = os.path.abspath(plugin_dir)

    # Validation
    errors = validate(manifest, CHAPTERS_DIR)
    if errors:
        print("VALIDATION ERRORS:")
        for e in errors:
            print(f"  ✗ {e}")
        sys.exit(1)
    print("✓ Validation passed")

    # Build skill name set (excluding exempt)
    exempt = manifest.get("exempt_skills", [])
    skill_names = set()
    for num, row in sorted(chapters.items()):
        skill_names.add(row.get("name", ""))

    version = manifest.get("version", "1.0.0")
    router_emoji = manifest.get("router_emoji", "🧭")

    # 1. Generate scenario-map.md
    scenario_map = generate_scenario_map(chapters)
    shared_dir = os.path.join(target_dir, manifest.get("shared_dir", "shared"))
    os.makedirs(shared_dir, exist_ok=True)
    map_path = os.path.join(shared_dir, "scenario-map.md")
    if not dry_run:
        with open(map_path, "w") as f:
            f.write(scenario_map)
        print(f"  ✓ {map_path}")

    # 2. Generate router SKILL.md
    router_skill = generate_router_skill(chapters, version, router_emoji)
    router_name = manifest.get("router_name", "testgen-scenario")
    router_dir = os.path.join(target_dir, router_name)
    os.makedirs(router_dir, exist_ok=True)
    router_path = os.path.join(router_dir, "SKILL.md")
    if not dry_run:
        with open(router_path, "w") as f:
            f.write(router_skill)
        print(f"  ✓ {router_path}")

    # 3. Generate stub SKILL.md for each pattern-* (body to be filled by AI)
    #    Each stub has frontmatter + "TODO" body — AI maintenance skill replaces it.
    pattern_dir = os.path.join(target_dir)
    for num in sorted(chapters.keys()):
        row = chapters[num]
        name = row.get("name", f"pattern-{num:02d}")
        style = row.get("style", "check")
        scenario = row.get("scenario", "")
        keys = row.get("route_keys", [])

        bins = "        - go\n        - go-testgen" if style != "manual" else "        - go"
        stub = f"""---
name: {name}
description: >
  {scenario}. Read this skill to learn the pattern, go-testgen style, and how
  to fill test cases. Derived from unit-test-book chapter {num:02d}.
license: MIT
compatibility: Designed for Claude Code or similar AI coding agents.
user-invocable: false
metadata:
  author: padiazg
  version: "{version}"
  go-testgen-style: {style}
  chapter: {num:02d}
  openclaw:
    emoji: "🧪"
    homepage: https://github.com/padiazg/go-testgen
    requires:
      bins:
{bins}
    install: []
allowed-tools: Read Write Glob Grep Bash(go:*) Bash(go-testgen:*) Bash(git:*)
---

# {name}

> unit-test-book chapter {num:02d} — style: **{style}**
"""
        stub_path = os.path.join(pattern_dir, f"{name}", "SKILL.md")
        if not dry_run:
            os.makedirs(os.path.dirname(stub_path), exist_ok=True)
            with open(stub_path, "w") as f:
                f.write(stub)
            print(f"  ✓ {stub_path} (stub — fill body via MAINTAIN-SKILLS.md)")

    # 4. Generate plugin manifests
    plugin_json, gemini_json = generate_plugin_manifests(
        skill_names, exempt, version
    )
    for dname in [".claude-plugin", ".cursor-plugin"]:
        dpath = os.path.join(plugin_dir, dname)
        os.makedirs(dpath, exist_ok=True)
        fpath = os.path.join(dpath, "plugin.json")
        if not dry_run:
            with open(fpath, "w") as f:
                f.write(plugin_json)
            print(f"  ✓ {fpath}")

    gemini_path = os.path.join(plugin_dir, "gemini-extension.json")
    if not dry_run:
        with open(gemini_path, "w") as f:
            f.write(gemini_json)
        print(f"  ✓ {gemini_path}")

    # 5. Generate skills.sh (write locally; copy to go-testgen/doc/docs/)
    skills_sh = generate_skills_sh(skill_names, exempt)
    skills_sh_path = os.path.join(SCRIPT_DIR, "skills.sh")
    if not dry_run:
        with open(skills_sh_path, "w") as f:
            f.write(skills_sh)
        os.chmod(skills_sh_path, 0o755)
        print(f"  ✓ {skills_sh_path} (copy to go-testgen/doc/docs/skills.sh)")

    # 6. Generate AGENTS.md for go-testgen/skills/
    agents_md = generate_agents_md(chapters, exempt)
    agents_path = os.path.join(target_dir, "AGENTS.md")
    if not dry_run:
        with open(agents_path, "w") as f:
            f.write(agents_md)
        print(f"  ✓ {agents_path}")

    print(f"\nDone. Generated {len(chapters)} pattern stubs + router + manifests.")
    print("Next: run MAINTAIN-SKILLS.md to fill each pattern SKILL.md body from README.")


if __name__ == "__main__":
    main()
