## Reverse spec-kit integration

This repository now includes a reverse-engineered spec-kit style workspace under `testkit/.specify`.

### Structure

- `testkit/.specify/memory/constitution.md`
- `testkit/.specify/specs/001-polyglot-testkit/spec.md`
- `testkit/.specify/specs/001-polyglot-testkit/plan.md`
- `testkit/.specify/specs/001-polyglot-testkit/tasks.md`
- `testkit/.specify/specs/001-polyglot-testkit/contracts/cli-protocol.json`
- `testkit/.specify/specs/001-polyglot-testkit/checklists/quality.md`

### Why this exists

- Preserves one source-of-truth specification flow in spec-kit format.
- Keeps current implementation strategy hybrid:
  - Option A now: Go core + JSON CLI + thin Python/Java wrappers
  - Option B later: native Python/Java implementations validated against these artifacts

### Conformance execution

Use existing smoke tests as the executable conformance path:

- Python: `cd testkit/python && python3 -m pytest tests/ -v`
- Java: `cd testkit/java && mvn test`
- Go regression: `cd /workspace && go test ./...`
