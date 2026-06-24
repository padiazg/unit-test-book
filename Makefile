.PHONY: test lint fmt clean help

test:
	go test -race -count=1 ./chapter-*/...

lint:
	find . -name go.mod -execdir golangci-lint run ./... \;

fix:
	find . -name go.mod -execdir go fix \;

fmt:
	gofmt -s -w .

clean:
	rm -f coverage.out mutation*.json

help:
	@echo "test  - run all chapter tests with race detector"
	@echo "lint  - run golangci-lint"
	@echo "fmt   - format source code"
	@echo "clean - remove artifacts"
