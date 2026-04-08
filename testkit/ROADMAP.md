# Polyglot testkit roadmap

This roadmap adopts a hybrid strategy:

- **Option A (now):** single reusable Go core with a JSON CLI bridge.
- **Option B (later):** native Python and Java implementations validated against shared behavior tests.

## Goals

- Maximize reuse by keeping one behavior source-of-truth first.
- Maximize DevEx by exposing thin language wrappers that feel simple to call.
- Prove adoption path with smoke tests and executable examples.

## Phase 1 (Option A): Go core + CLI bridge

Deliverables:

1. Keep existing Go `testing.T` APIs for backward compatibility.
2. Add reusable error-returning Go APIs that do not depend on `testing.T`.
3. Add CLI binary (`cmd/git-testkit-cli`) with JSON request/response protocol.
4. Add Python and Java thin wrappers that shell out to the CLI.
5. Add smoke tests proving fixture -> scenario-like flow -> snapshot round-trip.
6. Add spec-kit artifact set under `.specify/`:
   - constitution (`.specify/memory/constitution.md`)
   - feature spec (`.specify/specs/001-polyglot-testkit/spec.md`)
   - implementation plan (`.specify/specs/001-polyglot-testkit/plan.md`)
   - tasks (`.specify/specs/001-polyglot-testkit/tasks.md`)
   - protocol contract and quality checklist

Success criteria:

- Existing Go tests stay green.
- CLI handles core fixture and snapshot operations.
- Python and Java smoke tests pass against real `git`.
- `.specify` artifacts remain executable and aligned with test commands in `tasks.md`.

## Phase 2 (Option B): Native ports

Deliverables:

1. Native Python implementation (fixtures/scenarios/snapshots).
2. Native Java implementation (fixtures/scenarios/snapshots).
3. Cross-language conformance tests generated from `GIT_TESTKIT_SPEC.md`.
4. Optional dual mode in wrappers:
   - `mode=cli` (Go bridge)
   - `mode=native` (language-native)

Success criteria:

- Native implementations pass conformance tests and smoke tests.
- CLI mode remains available as stable fallback.

## Adoption strategy

- Start users on thin wrapper + CLI mode for reliability.
- Keep wrapper API stable while internals evolve.
- Introduce native mode only when parity and maintenance plan are ready.
