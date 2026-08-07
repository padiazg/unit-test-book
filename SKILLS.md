# AI Skills

Skills in `go-testgen/skills/` are auto-generated from these chapters. Each skill describes a testing pattern, maps it to a `go-testgen` style (`table`, `check`, `simple`, `manual`), and includes the Code/Test blocks from the chapter README.

## Generate skills

```bash
make skills                    # generate stubs + router + manifests
make skills-validate           # validate all gates pass
python3 doc/generate_skill_bodies.py  # fill each skill body from READMEs
```

## Skill pipeline

```
chapters/chapter-NN-slug/README.md   ← source (chapter description, code, test)
doc/skills-manifest.yaml             ← routing table (style, scenario, route_keys)
doc/generate_skills.py               ← generates stubs, router, plugin manifests, skills.sh
doc/generate_skill_bodies.py         ← fills skill bodies from chapter READMEs
go-testgen/skills/pattern-NN-slug/   ← generated skill (copied to ../go-testgen/skills/)
```

Each generated skill has:

- **When to use** — paraphrased from the chapter Description
- **go-testgen style** — `table` | `check` | `simple` | `manual`
- **Pattern** — Code block + Test block from the README
- **Fill the cases** — Testing Approach as bullet points
- **Cross-references** — link back to the chapter

The manifest (`doc/skills-manifest.yaml`) also controls:

- `style` — determines go-testgen scaffolding behavior
- `demand_style: true` — missing rows cause build errors
- `require_route_keys: true` — each chapter must have routing keys
- `version` — synced across plugin manifests (`.claude-plugin/`, `.cursor-plugin/`, `gemini-extension.json`)

For the full maintenance workflow, see [MAINTAIN-SKILLS.md](MAINTAIN-SKILLS.md).
