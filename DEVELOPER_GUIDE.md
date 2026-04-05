# Developer Guide

This guide covers how to use `git-testkit`, contribute safely, run checks locally, and maintain releases.

## Audience

- **Users** writing tests that need real git repositories.
- **Contributors** making code/docs changes.
- **Maintainers/admins** preparing releases and keeping CI healthy.

## Local setup

Requirements:

- Go 1.22+
- `git` on `PATH`

Clone and verify:

```bash
# HTTPS (no SSH keys required)
git clone https://github.com/git-fire/git-testkit.git

# or SSH
# git clone git@github.com:git-fire/git-testkit.git

cd git-testkit
go test ./...
```

## Project structure

- `fixtures.go`: base repo creation and command helpers.
- `scenarios.go`: fluent scenario builder and predefined multi-repo scenarios.
- `snapshots.go`: snapshot/restore utilities for expensive test setup reuse.
- `fixtures_test.go`: external package tests (`testutil_test`) that validate public API usage.
- `scenarios_test.go`: package-internal tests for scenario/snapshot behavior.

## Design principles

- Prefer real git commands over mocks for behavior confidence.
- Keep helper APIs composable and minimal.
- Fail fast in setup helpers (`t.Fatalf`) so fixture errors are obvious.
- Keep tests isolated using `t.TempDir()`.

## Testing and quality checks

Run before opening a PR:

```bash
gofmt -w *.go
go vet ./...
go test ./...
```

Optional:

```bash
go test -short ./...
```

CI mirrors the required checks (`go vet` and `go test`) on pull requests.

## Usage guidance for library consumers

- Prefer scenario builders for integration-style tests with remotes/worktrees.
- Prefer base fixtures for small unit/integration tests that only need one repo.
- Keep assertions close to setup; helper failures already include command output.
- Use snapshots for expensive setups that are reused across subtests.

Common usage split:

- `CreateTestRepo`: fast setup for single-repo state.
- `NewScenario`: multi-repo topologies and fluent setup.
- `SnapshotRepo`/`RestoreSnapshot`: performance optimization for repeated expensive setups.

## Adding new helpers

When adding new exported helpers:

- Add or update tests in `fixtures_test.go` and/or `scenarios_test.go`.
- Add usage notes/examples to `README.md` when the helper is user-facing.
- Keep function names explicit and test-focused (avoid generic utility names).

## Snapshot safety expectations

- Snapshot restore must never write outside the restore root.
- Avoid introducing behavior that can restore unsupported file types silently.
- Keep snapshot behavior deterministic for repeatable tests.

## Pull request guidance

- Keep PRs focused and small when possible.
- Include a clear "why" in the commit/PR description.
- Include a short test plan with commands you ran locally.
- Prefer additive, backward-compatible changes for existing exported helpers.

Suggested PR checklist:

- [ ] public API changes documented in `README.md`
- [ ] tests added/updated for behavior changes
- [ ] `gofmt`, `go vet`, and `go test` all pass locally
- [ ] no unrelated refactors mixed into functional changes

## Maintainer/admin workflow

### Branch and PR flow

1. Create a feature branch from `main`.
2. Commit focused changes with clear commit messages.
3. Open a PR with summary and test plan.
4. Merge only after CI is green.

### Release flow

1. Verify `main` has passed CI.
2. Confirm `README.md` and this guide reflect current API/behavior.
3. Pull release notes from merged PRs (user-visible changes first).
4. Create and push a version tag.
5. Publish a GitHub release with:
   - notable changes
   - compatibility notes (for example, minimum Go version)
   - migration notes when behavior changed

## Release notes checklist

Before cutting a release:

- Ensure CI is green on `main`.
- Summarize user-visible changes (new helpers, behavior changes, compatibility notes).
- Call out any minimum Go version changes.
- Verify examples in `README.md` still compile conceptually with current API.
