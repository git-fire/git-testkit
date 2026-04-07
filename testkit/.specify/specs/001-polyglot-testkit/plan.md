# Implementation Plan: Polyglot testkit via Go bridge (Option A)

**Feature**: `001-polyglot-testkit`  
**Input**: `testkit/.specify/specs/001-polyglot-testkit/spec.md`  
**Status**: Implemented (baseline)

## Summary

Deliver a reusable polyglot interface to `git-testkit` by exposing Go core behavior through a JSON CLI bridge and thin wrappers in Python and Java. Keep Option B (native ports) as a future phase.

## Technical decisions

1. Preserve existing Go `testing.T` APIs for backward compatibility.
2. Add error-returning Go helpers for non-test callers.
3. Expose fixtures/query/snapshot operations through `cmd/git-testkit-cli`.
4. Wrap CLI in Python (`GitTestKitClient`) and Java (`CliBridge`) with minimal logic.
5. Validate using smoke tests that execute real `git`.

## Artifact map

- Go core updates:
  - `fixtures.go`
  - `snapshots.go`
- Go bridge:
  - `cmd/git-testkit-cli/main.go`
- Python wrapper:
  - `testkit/python/git_testkit/cli.py`
  - `testkit/python/tests/*`
- Java wrapper:
  - `testkit/java/src/main/java/io/gitfire/testkit/CliBridge.java`
  - `testkit/java/src/test/java/io/gitfire/testkit/CliBridgeTest.java`

## Risks and mitigations

- **Risk**: Wrapper drift from Go behavior.
  - **Mitigation**: Keep wrappers thin; assert behavior through smoke tests using real repos.
- **Risk**: Snapshot edge cases (unsafe paths, parent dirs).
  - **Mitigation**: enforce safe joins and explicit snapshot save directory creation.

## Phase 2 placeholder (Option B)

- Add native Python/Java implementations behind same wrapper surface.
- Add conformance matrix that runs in both bridge and native modes.
