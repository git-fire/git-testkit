# git-testkit

[![CI](https://github.com/git-fire/git-testkit/actions/workflows/ci.yml/badge.svg)](https://github.com/git-fire/git-testkit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/git-fire/git-testkit.svg)](https://pkg.go.dev/github.com/git-fire/git-testkit)
[![Go 1.22+](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/dl/)
[![Latest Release](https://img.shields.io/github/v/release/git-fire/git-testkit)](https://github.com/git-fire/git-testkit/releases/latest)
[![MIT License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`git-testkit` provides helpers for writing Go tests that exercise real Git repositories.

## Why use this

- Exercise real git behavior instead of mocking command output.
- Build common repo states quickly (dirty trees, detached HEAD, diverged remotes, worktrees).
- Reuse expensive setups across tests with in-memory snapshots.

## Install

```bash
go get github.com/git-fire/git-testkit
```

## Requirements

- `git` must be installed and available on `PATH`
- Go 1.22+

## What it includes

- Repository fixtures (`CreateTestRepo`, `CreateBareRemote`, `RunGitCmd`)
- Scenario builders for common multi-repo states (`NewScenario`, conflict/worktree helpers)
- Snapshot helpers for capturing and restoring repository state in tests

## API overview

- `CreateTestRepo(t, RepoOptions)` creates a real repository with optional files/remotes/branches.
- `CreateBareRemote(t, name)` creates a bare repository for remote testing.
- `NewScenario(t)` returns a fluent builder for multi-repo test topologies.
- `SnapshotRepo(t, path)` and `RestoreSnapshot(t, snap)` speed up repeated fixture setup.
- `IsDirty`, `GetCurrentSHA`, `GetBranches`, `GetRemotes` provide common assertions/helpers.

## Quickstart for using in tests

1. Create fixture repos with `CreateTestRepo` or `NewScenario`.
2. Apply setup operations with fluent helpers (`AddFile`, `Commit`, `WithRemote`, `Push`).
3. Run your code under test against the real repository paths.
4. Assert with helper methods (`IsDirty`, `GetCurrentSHA`, `GetBranches`).

Minimal test flow:

```go
func TestMyGitBehavior(t *testing.T) {
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{Name: "subject"})
	testutil.RunGitCmd(t, repoPath, "checkout", "-b", "feature")
	// call your package functions here
	if testutil.IsDirty(t, repoPath) {
		t.Fatal("repo should be clean")
	}
}
```

## Example

```go
package mypkg_test

import (
	"testing"

	testutil "github.com/git-fire/git-testkit"
)

func TestWithRepo(t *testing.T) {
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "my-repo",
	})
	if repoPath == "" {
		t.Fatal("expected a repo path")
	}
}
```

## Cleanup behavior

- All helper-created repositories use `t.TempDir()`.
- Repos/worktrees are automatically removed by Go's test framework at test completion.
- As with any temp directories, force-killed test processes may leave files behind.

## Common patterns

### Build a conflict scenario

```go
func TestConflictFlow(t *testing.T) {
	_, local, _ := testutil.CreateConflictScenario(t)

	// Exercise your logic against a real diverged local clone.
	testutil.RunGitCmd(t, local.Path(), "status")
}
```

### Snapshot expensive setup

```go
func TestUsingSnapshot(t *testing.T) {
	_, repo := testutil.CreateLargeRepoScenario(t, 20, 10)

	snap := testutil.SnapshotRepo(t, repo.Path())
	clonePath := testutil.RestoreSnapshot(t, snap)

	// Use clonePath in assertions without rebuilding the fixture each time.
	testutil.RunGitCmd(t, clonePath, "status")
}
```

## Notes

- Snapshots are intended for deterministic test fixtures and only restore regular files/directories.
- Helpers fail tests immediately (`t.Fatalf`) when git commands fail, so errors surface close to setup code.
- When tests build large repo graphs repeatedly, prefer snapshot/restore to reduce total runtime.

## Origin

`git-testkit` was extracted from [git-fire](https://github.com/git-fire/git-fire) before its public launch and published as a standalone Go library. The test helpers had accumulated enough utility to be useful on their own, so they were split out to let other projects automate real Git repositories without depending on git-fire itself.

## Developer docs

- See `DEVELOPER_GUIDE.md` for:
  - testing and quality gates
  - usage guidance for library consumers
  - administration/maintenance and release workflow
  - contribution process and PR expectations
