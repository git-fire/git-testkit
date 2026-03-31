# git-testkit

`git-testkit` provides helpers for writing Go tests that exercise real Git repositories.

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
