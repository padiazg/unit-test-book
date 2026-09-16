#!/bin/sh
set -eu

# skills.sh — Auto-installs all go-testgen AI agent skills
# Generated from skills-manifest.yaml — do not edit by hand.

REPO=padiazg/go-testgen
BRANCH=${BRANCH:-main}
TARGET=${TARGET:-$HOME/.agents/skills}
BASE="https://raw.githubusercontent.com/$REPO/$BRANCH/skills"

echo "Installing go-testgen AI agent skills to $TARGET"

mkdir -p "$TARGET/closure-check-tests"
curl -fsSL "$BASE/closure-check-tests/SKILL.md" -o "$TARGET/closure-check-tests/SKILL.md"
echo "  ✔ closure-check-tests"

mkdir -p "$TARGET/gen-test-cases"
curl -fsSL "$BASE/gen-test-cases/SKILL.md" -o "$TARGET/gen-test-cases/SKILL.md"
echo "  ✔ gen-test-cases"

mkdir -p "$TARGET/pattern-01-table-validation"
curl -fsSL "$BASE/pattern-01-table-validation/SKILL.md" -o "$TARGET/pattern-01-table-validation/SKILL.md"
echo "  ✔ pattern-01-table-validation"

mkdir -p "$TARGET/pattern-02-value-assertions"
curl -fsSL "$BASE/pattern-02-value-assertions/SKILL.md" -o "$TARGET/pattern-02-value-assertions/SKILL.md"
echo "  ✔ pattern-02-value-assertions"

mkdir -p "$TARGET/pattern-03-fields-struct-inputs"
curl -fsSL "$BASE/pattern-03-fields-struct-inputs/SKILL.md" -o "$TARGET/pattern-03-fields-struct-inputs/SKILL.md"
echo "  ✔ pattern-03-fields-struct-inputs"

mkdir -p "$TARGET/pattern-04-subtest-naming"
curl -fsSL "$BASE/pattern-04-subtest-naming/SKILL.md" -o "$TARGET/pattern-04-subtest-naming/SKILL.md"
echo "  ✔ pattern-04-subtest-naming"

mkdir -p "$TARGET/pattern-05-typed-check-functions"
curl -fsSL "$BASE/pattern-05-typed-check-functions/SKILL.md" -o "$TARGET/pattern-05-typed-check-functions/SKILL.md"
echo "  ✔ pattern-05-typed-check-functions"

mkdir -p "$TARGET/pattern-06-check-collection-builder"
curl -fsSL "$BASE/pattern-06-check-collection-builder/SKILL.md" -o "$TARGET/pattern-06-check-collection-builder/SKILL.md"
echo "  ✔ pattern-06-check-collection-builder"

mkdir -p "$TARGET/pattern-07-check-factory-closures"
curl -fsSL "$BASE/pattern-07-check-factory-closures/SKILL.md" -o "$TARGET/pattern-07-check-factory-closures/SKILL.md"
echo "  ✔ pattern-07-check-factory-closures"

mkdir -p "$TARGET/pattern-08-error-message-verification"
curl -fsSL "$BASE/pattern-08-error-message-verification/SKILL.md" -o "$TARGET/pattern-08-error-message-verification/SKILL.md"
echo "  ✔ pattern-08-error-message-verification"

mkdir -p "$TARGET/pattern-09-output-string-inspection"
curl -fsSL "$BASE/pattern-09-output-string-inspection/SKILL.md" -o "$TARGET/pattern-09-output-string-inspection/SKILL.md"
echo "  ✔ pattern-09-output-string-inspection"

mkdir -p "$TARGET/pattern-10-before-hook-pattern"
curl -fsSL "$BASE/pattern-10-before-hook-pattern/SKILL.md" -o "$TARGET/pattern-10-before-hook-pattern/SKILL.md"
echo "  ✔ pattern-10-before-hook-pattern"

mkdir -p "$TARGET/pattern-11-http-client-interface-mock"
curl -fsSL "$BASE/pattern-11-http-client-interface-mock/SKILL.md" -o "$TARGET/pattern-11-http-client-interface-mock/SKILL.md"
echo "  ✔ pattern-11-http-client-interface-mock"

mkdir -p "$TARGET/pattern-12-roundtripper-mock"
curl -fsSL "$BASE/pattern-12-roundtripper-mock/SKILL.md" -o "$TARGET/pattern-12-roundtripper-mock/SKILL.md"
echo "  ✔ pattern-12-roundtripper-mock"

mkdir -p "$TARGET/pattern-13-testify-mock-interfaces"
curl -fsSL "$BASE/pattern-13-testify-mock-interfaces/SKILL.md" -o "$TARGET/pattern-13-testify-mock-interfaces/SKILL.md"
echo "  ✔ pattern-13-testify-mock-interfaces"

mkdir -p "$TARGET/pattern-14-function-variable-injection"
curl -fsSL "$BASE/pattern-14-function-variable-injection/SKILL.md" -o "$TARGET/pattern-14-function-variable-injection/SKILL.md"
echo "  ✔ pattern-14-function-variable-injection"

mkdir -p "$TARGET/pattern-15-package-var-swap"
curl -fsSL "$BASE/pattern-15-package-var-swap/SKILL.md" -o "$TARGET/pattern-15-package-var-swap/SKILL.md"
echo "  ✔ pattern-15-package-var-swap"

mkdir -p "$TARGET/pattern-16-httptest-server"
curl -fsSL "$BASE/pattern-16-httptest-server/SKILL.md" -o "$TARGET/pattern-16-httptest-server/SKILL.md"
echo "  ✔ pattern-16-httptest-server"

mkdir -p "$TARGET/pattern-17-httptest-recorder"
curl -fsSL "$BASE/pattern-17-httptest-recorder/SKILL.md" -o "$TARGET/pattern-17-httptest-recorder/SKILL.md"
echo "  ✔ pattern-17-httptest-recorder"

mkdir -p "$TARGET/pattern-18-error-readers"
curl -fsSL "$BASE/pattern-18-error-readers/SKILL.md" -o "$TARGET/pattern-18-error-readers/SKILL.md"
echo "  ✔ pattern-18-error-readers"

mkdir -p "$TARGET/pattern-19-temp-files-parsing"
curl -fsSL "$BASE/pattern-19-temp-files-parsing/SKILL.md" -o "$TARGET/pattern-19-temp-files-parsing/SKILL.md"
echo "  ✔ pattern-19-temp-files-parsing"

mkdir -p "$TARGET/pattern-20-channel-delivery"
curl -fsSL "$BASE/pattern-20-channel-delivery/SKILL.md" -o "$TARGET/pattern-20-channel-delivery/SKILL.md"
echo "  ✔ pattern-20-channel-delivery"

mkdir -p "$TARGET/pattern-21-panic-recovery"
curl -fsSL "$BASE/pattern-21-panic-recovery/SKILL.md" -o "$TARGET/pattern-21-panic-recovery/SKILL.md"
echo "  ✔ pattern-21-panic-recovery"

mkdir -p "$TARGET/pattern-22-goroutine-run-loops"
curl -fsSL "$BASE/pattern-22-goroutine-run-loops/SKILL.md" -o "$TARGET/pattern-22-goroutine-run-loops/SKILL.md"
echo "  ✔ pattern-22-goroutine-run-loops"

mkdir -p "$TARGET/pattern-23-goroutine-leak"
curl -fsSL "$BASE/pattern-23-goroutine-leak/SKILL.md" -o "$TARGET/pattern-23-goroutine-leak/SKILL.md"
echo "  ✔ pattern-23-goroutine-leak"

mkdir -p "$TARGET/pattern-24-ast-complexity"
curl -fsSL "$BASE/pattern-24-ast-complexity/SKILL.md" -o "$TARGET/pattern-24-ast-complexity/SKILL.md"
echo "  ✔ pattern-24-ast-complexity"

mkdir -p "$TARGET/pattern-25-benchmarks"
curl -fsSL "$BASE/pattern-25-benchmarks/SKILL.md" -o "$TARGET/pattern-25-benchmarks/SKILL.md"
echo "  ✔ pattern-25-benchmarks"

mkdir -p "$TARGET/pattern-26-parallel-tests"
curl -fsSL "$BASE/pattern-26-parallel-tests/SKILL.md" -o "$TARGET/pattern-26-parallel-tests/SKILL.md"
echo "  ✔ pattern-26-parallel-tests"

mkdir -p "$TARGET/pattern-27-service-mocked-ports"
curl -fsSL "$BASE/pattern-27-service-mocked-ports/SKILL.md" -o "$TARGET/pattern-27-service-mocked-ports/SKILL.md"
echo "  ✔ pattern-27-service-mocked-ports"

mkdir -p "$TARGET/pattern-28-json-format"
curl -fsSL "$BASE/pattern-28-json-format/SKILL.md" -o "$TARGET/pattern-28-json-format/SKILL.md"
echo "  ✔ pattern-28-json-format"

mkdir -p "$TARGET/pattern-29-setup-teardown-fixtures"
curl -fsSL "$BASE/pattern-29-setup-teardown-fixtures/SKILL.md" -o "$TARGET/pattern-29-setup-teardown-fixtures/SKILL.md"
echo "  ✔ pattern-29-setup-teardown-fixtures"

mkdir -p "$TARGET/pattern-30-composable-check-navigation"
curl -fsSL "$BASE/pattern-30-composable-check-navigation/SKILL.md" -o "$TARGET/pattern-30-composable-check-navigation/SKILL.md"
echo "  ✔ pattern-30-composable-check-navigation"

mkdir -p "$TARGET/pattern-31-inline-check-closures"
curl -fsSL "$BASE/pattern-31-inline-check-closures/SKILL.md" -o "$TARGET/pattern-31-inline-check-closures/SKILL.md"
echo "  ✔ pattern-31-inline-check-closures"

mkdir -p "$TARGET/pattern-32-interface-extraction"
curl -fsSL "$BASE/pattern-32-interface-extraction/SKILL.md" -o "$TARGET/pattern-32-interface-extraction/SKILL.md"
echo "  ✔ pattern-32-interface-extraction"

mkdir -p "$TARGET/pattern-33-event-driven-run-loop"
curl -fsSL "$BASE/pattern-33-event-driven-run-loop/SKILL.md" -o "$TARGET/pattern-33-event-driven-run-loop/SKILL.md"
echo "  ✔ pattern-33-event-driven-run-loop"

mkdir -p "$TARGET/pattern-34-embedded-infrastructure"
curl -fsSL "$BASE/pattern-34-embedded-infrastructure/SKILL.md" -o "$TARGET/pattern-34-embedded-infrastructure/SKILL.md"
echo "  ✔ pattern-34-embedded-infrastructure"

mkdir -p "$TARGET/pattern-35-capturing-stdout"
curl -fsSL "$BASE/pattern-35-capturing-stdout/SKILL.md" -o "$TARGET/pattern-35-capturing-stdout/SKILL.md"
echo "  ✔ pattern-35-capturing-stdout"

mkdir -p "$TARGET/go-testgen"
curl -fsSL "$BASE/AGENTS.md" -o "$TARGET/go-testgen/AGENTS.md"
echo "  ✔ AGENTS.md"

echo "Done. Skills installed to $TARGET"
