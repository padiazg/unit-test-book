#!/usr/bin/env python3
"""Validation gates for the skill pipeline.

Checks:
  1. Every chapter has a manifest row (demand_style)
  2. No manifest rows for removed chapters
  3. style ∈ {table, check, simple, manual}
  4. route_keys non-empty
  5. route_keys unique across chapters (no duplicates)
  6. generated SKILL.md files exist for each chapter
  7. plugin manifests are in sync
  8. router + scenario-map exist
"""

import json
import os
import sys

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
BOOK_ROOT = os.path.dirname(SCRIPT_DIR)
CHAPTERS_DIR = os.path.join(BOOK_ROOT, "chapters")
MANIFEST_PATH = os.path.join(SCRIPT_DIR, "skills-manifest.yaml")
TARGET_DIR = os.path.abspath(os.path.join(SCRIPT_DIR, "../go-testgen/skills"))

# Also check the real go-testgen repo for exempt skills
GOTESTGEN_SKILLS = os.path.abspath(os.path.join(BOOK_ROOT, "../go-testgen/skills"))

sys.path.insert(0, SCRIPT_DIR)
import chapterlib  # noqa: E402
from generate_skills import parse_manifest  # noqa: E402

VALID_STYLES = {"table", "check", "simple", "manual"}


def get_plugin_dir():
    """Resolve plugin_dir from manifest."""
    m = parse_manifest(MANIFEST_PATH)
    raw = m.get("plugin_dir", "")
    if os.path.isabs(raw):
        return raw
    return os.path.abspath(os.path.join(SCRIPT_DIR, raw))


def load_manifest():
    return parse_manifest(MANIFEST_PATH)


def check_manifest_chapters(manifest):
    """Check chapters in manifest match discovered dirs."""
    errors = []
    demand = manifest.get("demand_style", True)
    chapters = manifest.get("chapters", {})
    actual = set(n for n, _ in chapterlib.find_chapter_dirs())
    manifest_keys = set(chapters.keys())

    if demand:
        missing = sorted(actual - manifest_keys)
        if missing:
            errors.append(
                f"demand_style: true — missing manifest rows for chapters: {missing}"
            )
        extra = sorted(manifest_keys - actual)
        if extra:
            errors.append(
                f"manifest rows for chapters that no longer exist: {extra}"
            )

    # Check individual chapter rows
    for num in sorted(chapters.keys()):
        row = chapters[num]
        name = row.get("name", "")
        style = row.get("style", "")
        keys = row.get("route_keys", [])

        if style not in VALID_STYLES:
            errors.append(f"ch.{num:02d}: invalid style {style!r}")

        if not keys:
            errors.append(f"ch.{num:02d}: route_keys is empty")

        if not name:
            errors.append(f"ch.{num:02d}: name is missing")

    # Check for duplicate route_keys
    all_keys = {}
    for num, row in sorted(chapters.items()):
        for k in (row.get("route_keys") or []):
            all_keys.setdefault(k, []).append(num)
    dupes = {k: nums for k, nums in all_keys.items() if len(nums) > 1}
    if dupes:
        errors.append(f"duplicate route_keys: {dict(dupes)}")

    return errors


def check_generated_files(manifest):
    """Check generated SKILL.md files exist."""
    errors = []
    chapters = manifest.get("chapters", {})
    exempt = manifest.get("exempt_skills", [])
    router_name = manifest.get("router_name", "testgen-scenario")

    # Check router exists
    router_path = os.path.join(TARGET_DIR, router_name, "SKILL.md")
    if not os.path.exists(router_path):
        errors.append(f"router skill missing: {router_path}")

    # Check scenario-map exists
    map_path = os.path.join(TARGET_DIR, "shared", "scenario-map.md")
    if not os.path.exists(map_path):
        errors.append(f"scenario-map missing: {map_path}")

    # Check each pattern skill stub exists
    for num, row in sorted(chapters.items()):
        name = row.get("name", f"pattern-{num:02d}")
        stub_path = os.path.join(TARGET_DIR, name, "SKILL.md")
        if not os.path.exists(stub_path):
            errors.append(f"pattern skill missing: {stub_path}")

    # Check exempt skills in staging area OR real go-testgen repo
    for skill in exempt:
        staging = os.path.join(TARGET_DIR, skill, "SKILL.md")
        real = os.path.join(GOTESTGEN_SKILLS, skill, "SKILL.md")
        if not os.path.exists(staging) and not os.path.exists(real):
            errors.append(f"exempt skill missing: {skill} (checked staging + {GOTESTGEN_SKILLS})")

    return errors


def check_plugin_manifests(plugin_dir):
    """Check plugin manifests exist and have matching versions."""
    errors = []

    if not os.path.isdir(plugin_dir):
        errors.append(f"plugin_dir not found: {plugin_dir}")
        return errors

    versions = {}
    for dname in [".claude-plugin", ".cursor-plugin"]:
        pjson = os.path.join(plugin_dir, dname, "plugin.json")
        if os.path.exists(pjson):
            with open(pjson) as f:
                data = json.load(f)
            versions[dname] = data.get("version", "")
        else:
            errors.append(f"plugin manifest missing: {pjson}")

    gemini = os.path.join(plugin_dir, "gemini-extension.json")
    if os.path.exists(gemini):
        with open(gemini) as f:
            data = json.load(f)
        versions["gemini"] = data.get("version", "")
    else:
        errors.append(f"gemini manifest missing: {gemini}")

    # Check version sync
    unique = set(versions.values())
    if len(unique) > 1:
        errors.append(f"plugin manifests have different versions: {versions}")

    return errors


def main():
    print("=== Skills Pipeline Validation ===\n")

    manifest = load_manifest()
    chapters = manifest.get("chapters", {})

    # 1. Manifest chapter validation
    print("1. Manifest chapters...")
    merrors = check_manifest_chapters(manifest)
    if merrors:
        print("   FAIL:")
        for e in merrors:
            print(f"     ✗ {e}")
    else:
        print("   ✓ OK")

    # 2. Generated files check
    print("2. Generated files...")
    gerrors = check_generated_files(manifest)
    if gerrors:
        print("   FAIL:")
        for e in gerrors:
            print(f"     ✗ {e}")
    else:
        print("   ✓ OK")

    # 3. Plugin manifest sync
    print("3. Plugin manifests...")
    plugin_dir = get_plugin_dir()
    perrors = check_plugin_manifests(plugin_dir)
    if perrors:
        print("   FAIL:")
        for e in perrors:
            print(f"     ✗ {e}")
    else:
        print("   ✓ OK")

    # 4. Summary
    total = len(chapters)
    print(f"\n{'=' * 40}")
    print(f"Chapters: {total}  |  Status: {'PASS' if not (merrors or gerrors or perrors) else 'FAIL'}")

    if not merrors and not gerrors and not perrors:
        print("\nAll checks passed.")
        return 0
    else:
        print(f"\n{len(merrors) + len(gerrors) + len(perrors)} error(s) found.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
