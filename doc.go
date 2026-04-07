// Package testutil provides helpers for writing Go tests that exercise real
// Git repositories. Instead of mocking git command output, tests create actual
// repositories on disk, apply operations through fluent builders, and assert
// against real git state.
//
// The three main entry points are:
//
//   - [CreateTestRepo] for quickly spinning up a single repository with files,
//     branches, and remotes already configured.
//   - [NewScenario] for building multi-repository topologies (local + bare
//     remote, diverged clones, worktrees, conflict states).
//   - [SnapshotRepo] / [RestoreSnapshot] for capturing expensive fixture state
//     once and restoring it cheaply across subtests.
//
// All helpers integrate with Go's testing.T: repositories are created inside
// t.TempDir() and cleaned up automatically, and setup failures call t.Fatalf
// so errors surface close to the fixture code rather than deep in assertions.
//
// # Quickstart
//
//	func TestMyFeature(t *testing.T) {
//		repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{Name: "subject"})
//		testutil.RunGitCmd(t, repoPath, "checkout", "-b", "feature")
//		// exercise your code against repoPath
//		if testutil.IsDirty(t, repoPath) {
//			t.Fatal("expected clean repo")
//		}
//	}
package testutil
