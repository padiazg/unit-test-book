# Unit Testing in Go — 29 Patterns from Real Projects

Full documentation: https://padiazg.github.io/unit-test-book

This repository contains the MkDocs source for the book at the above URL.

**Contents:**

- `doc/docs/` — markdown pages (29 chapter pages + 7 section indexes + home)
- `doc/mkdocs.yml` — site configuration
- `doc/generate.py` — generates mkdocs pages from chapter README files

**Build locally:**

```bash
cd doc
pip install mkdocs-material pymdown-extensions
mkdics build
mkdocs serve
```
