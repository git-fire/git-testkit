# Testkit Constitution

## Core Principles

### I. Real Git Only
All implementations and tests MUST run against the real `git` binary on `PATH`.
No mocks or fake git output are permitted for conformance validation.

### II. Single Behavior Source
Option A (Go core + bridge) is the authoritative behavior source for polyglot consumers.
Language wrappers MUST delegate behavior to the Go bridge unless explicitly running an approved native mode.

### III. Backward Compatibility
Existing Go test APIs that accept `*testing.T` MUST remain stable and compatible.
New reusable APIs SHOULD be additive and return errors rather than aborting.

### IV. Deterministic Fixtures and Snapshots
Fixture creation and snapshot restoration MUST be deterministic and bounded to temporary roots.
Path traversal and unsafe archive extraction MUST be rejected.

### V. Test-First and Smoke Proof
Every new cross-language integration step MUST include executable smoke proof:
fixture creation, remote push flow, and snapshot roundtrip at minimum.

### VI. Simplicity Before Native Reimplementation
Polyglot adoption SHOULD start with thin wrappers over the bridge.
Native ports (Option B) are allowed only after conformance contracts and parity tests exist.
