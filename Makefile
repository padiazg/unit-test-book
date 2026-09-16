.PHONY: test lint fmt clean help docs skills skills-validate

test:
	go test -race -count=1 ./chapters/chapter-*/...

lint:
	find . -name go.mod -execdir golangci-lint run ./... \;

fix:
	find . -name go.mod -execdir go fix \;

fmt:
	gofmt -s -w .

docs:
	python3 doc/generate.py
	cd doc && mkdocs build --strict

skills:
	python3 doc/generate_skills.py

skills-validate:
	python3 doc/skills_validate.py

clean:
	rm -f coverage.out mutation*.json

help:
	@echo "test    - run all chapter tests with race detector"
	@echo "lint    - run golangci-lint"
	@echo "fmt     - format source code"
	@echo "docs    - generate MkDocs site"
	@echo "skills  - regenerate skills from chapters + manifest"
	@echo "validate - run skills validation gates"
	@echo "clean   - remove artifacts"
