# git-testkit

`git-testkit` provides helpers for writing Go tests that exercise real Git repositories.

## Install

```bash
go get github.com/git-fire/git-testkit
```

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
