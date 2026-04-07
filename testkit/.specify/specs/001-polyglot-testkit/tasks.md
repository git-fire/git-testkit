# Tasks: Polyglot git-testkit Option A bridge

## Phase 1 - Core bridge plumbing

- [x] T001 Add error-returning fixture APIs in `fixtures.go`
- [x] T002 Add error-returning snapshot APIs in `snapshots.go`
- [x] T003 Add CLI entrypoint `cmd/git-testkit-cli/main.go` with JSON protocol

## Phase 2 - Wrapper clients

- [x] T004 Add Python wrapper client in `testkit/python/git_testkit/cli.py`
- [x] T005 Add Java wrapper client in `testkit/java/src/main/java/io/gitfire/testkit/CliBridge.java`

## Phase 3 - Validation and smoke flows

- [x] T006 Add Python smoke tests for fixture + snapshot + push flow
- [x] T007 Add Java smoke tests for fixture + snapshot + push flow
- [x] T008 Run Go regressions: `go vet ./...`, `go test ./...`
- [x] T009 Run Python smoke tests via `python3 -m pytest tests/ -v`
- [x] T010 Run Java smoke tests via `mvn test`

## Phase 4 - Spec-kit alignment artifacts

- [x] T011 Add `.specify/memory/constitution.md`
- [x] T012 Add spec-kit-style spec in `.specify/specs/001-polyglot-testkit/spec.md`
- [x] T013 Add spec-kit-style plan in `.specify/specs/001-polyglot-testkit/plan.md`
- [x] T014 Add this task ledger in `.specify/specs/001-polyglot-testkit/tasks.md`
- [x] T015 Add spec-kit command workflow doc + shell helper
- [x] T016 Wire CI workflow to enforce spec-kit artifacts + polyglot smoke suites
