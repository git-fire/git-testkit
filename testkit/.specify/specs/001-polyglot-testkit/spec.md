# Feature Specification: Polyglot git-testkit (Hybrid Option A first)

**Feature Branch**: `001-polyglot-testkit`  
**Created**: 2026-04-07  
**Status**: Implemented (canonical spec-kit baseline)  
**Input**: User description: "reverse-spec existing Go git-testkit API and deliver polyglot reuse with high DevEx and adoption"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reuse Go behavior from other languages (Priority: P1)

As a test author in Python or Java, I can invoke git-testkit operations without reimplementing git semantics, so my tests share the same behavior guarantees as the Go source.

**Why this priority**: Reuse and behavior consistency are the highest-value outcome for fast adoption.

**Independent Test**: Create repos/remotes via wrapper and validate real git state without using Go test APIs directly.

**Acceptance Scenarios**:

1. **Given** a temp directory, **When** Python wrapper requests `create_test_repo`, **Then** a valid git repo path is returned and has a commit.
2. **Given** a local repo and bare remote, **When** wrapper runs remote-add/push flow, **Then** remote branch SHA matches local branch SHA.

---

### User Story 2 - Keep Go API backward-compatible (Priority: P1)

As an existing Go consumer, I continue using current `testing.T` helpers unchanged while the bridge adds non-`testing.T` reusable APIs.

**Why this priority**: Backward compatibility is required to avoid adoption blockers/regressions.

**Independent Test**: Run full existing Go test suite unchanged.

**Acceptance Scenarios**:

1. **Given** current Go tests, **When** code adds bridge support, **Then** all tests remain green.
2. **Given** existing exported Go fixture/scenario/snapshot APIs, **When** consumers compile, **Then** no API break is introduced.

---

### User Story 3 - Prove bridge architecture with smoke conformance (Priority: P2)

As a maintainer, I can run smoke conformance tests per language to verify fixture/snapshot contracts in a repeatable workflow.

**Why this priority**: Confidence and maintainability require executable evidence.

**Independent Test**: Run Python and Java smoke tests via their native test runners.

**Acceptance Scenarios**:

1. **Given** wrapper clients, **When** smoke tests run, **Then** create-repo, push-flow, and snapshot-roundtrip pass.
2. **Given** bridge JSON protocol, **When** operations fail, **Then** wrappers surface deterministic errors.

---

### Edge Cases

- Remote paths containing spaces or literal suffix-like text such as `" (push)"`.
- Snapshot save path parent directory missing.
- Git command with no stdout should still be treated as success.
- Default branch differs (`main` vs `master`).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST preserve existing Go test-facing APIs and semantics.
- **FR-002**: System MUST provide reusable Go error-returning fixture/snapshot helpers callable without `testing.T`.
- **FR-003**: System MUST expose core operations through a JSON CLI bridge process.
- **FR-004**: Python wrapper MUST invoke bridge operations and expose ergonomic methods for fixtures/queries/snapshots.
- **FR-005**: Java wrapper MUST invoke bridge operations and expose ergonomic methods for fixtures/queries/snapshots.
- **FR-006**: Wrapper smoke tests MUST validate create repo, push SHA equivalence, and snapshot restore roundtrip against real git.
- **FR-007**: Bridge responses MUST include structured success/error payloads and deterministic field names.

### Key Entities

- **RepoOptions**: input contract for repository construction (`name`, `dirty`, `files`, `remotes`, `branches`, `initialCommit`).
- **CLI Request**: JSON operation payload containing `op` and operation-specific fields.
- **CLI Response**: JSON operation result containing `ok`, optional `error`, and operation data fields.
- **Snapshot**: name + byte payload representation of archived repository filesystem state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go test ./...` passes on the branch after bridge integration.
- **SC-002**: Python wrapper smoke suite passes via `python3 -m pytest tests/ -v`.
- **SC-003**: Java wrapper smoke suite passes via `mvn test`.
- **SC-004**: Branch includes roadmap and spec artifacts describing Option A now and Option B follow-up.

## Assumptions

- `git` is available on PATH in all test environments.
- Wrappers initially prioritize correctness/reuse over minimizing process-spawn overhead.
- Native Option B ports are intentionally deferred after Option A merge.
