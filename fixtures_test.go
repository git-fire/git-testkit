package testutil_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	testutil "github.com/git-fire/git-testkit"
)

func TestCreateTestRepo(t *testing.T) {
	// Test creating a basic clean repo
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "test-repo",
	})

	// Verify .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("Expected .git directory to exist at %s", gitDir)
	}

	// Verify repo is clean (not dirty)
	if testutil.IsDirty(t, repoPath) {
		t.Fatal("Expected repo to be clean, but it has uncommitted changes")
	}
}

func TestCreateTestRepo_Dirty(t *testing.T) {
	// Test creating a dirty repo
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name:  "dirty-repo",
		Dirty: true,
	})

	// Verify repo is dirty
	if !testutil.IsDirty(t, repoPath) {
		t.Fatal("Expected repo to be dirty, but it's clean")
	}
}

func TestCreateTestRepo_WithFiles(t *testing.T) {
	// Test creating a repo with custom files
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "files-repo",
		Files: map[string]string{
			"test.txt":       "test content",
			"src/main.go":    "package main",
			"config/app.yml": "port: 8080",
		},
	})

	// Verify files exist
	files := []string{"test.txt", "src/main.go", "config/app.yml"}
	for _, file := range files {
		filePath := filepath.Join(repoPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("Expected file to exist: %s", filePath)
		}
	}
}

func TestCreateTestRepo_WithRemotes(t *testing.T) {
	// Create a bare remote first
	remotePath := testutil.CreateBareRemote(t, "origin")

	// Create repo with remote configured
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "remote-repo",
		Remotes: map[string]string{
			"origin": remotePath,
		},
	})

	// Verify remote is configured
	remotes := testutil.GetRemotes(t, repoPath)
	if _, exists := remotes["origin"]; !exists {
		t.Fatal("Expected 'origin' remote to be configured")
	}
}

func TestCreateTestRepo_WithRemotePathContainingSpaces(t *testing.T) {
	remotePath := testutil.CreateBareRemote(t, "origin with space")

	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "remote-space-repo",
		Remotes: map[string]string{
			"origin": remotePath,
		},
	})

	remotes := testutil.GetRemotes(t, repoPath)
	originURL, exists := remotes["origin"]
	if !exists {
		t.Fatal("Expected 'origin' remote to be configured")
	}
	if originURL != remotePath {
		t.Fatalf("Expected origin URL %q, got %q", remotePath, originURL)
	}
}

func TestCreateTestRepo_WithRemotePathContainingPushSuffix(t *testing.T) {
	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "origin (push)")
	if err := os.MkdirAll(remotePath, 0755); err != nil {
		t.Fatalf("Failed to create bare repo directory: %v", err)
	}
	testutil.RunGitCmd(t, tmpDir, "init", "--bare", remotePath)

	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "remote-push-suffix-repo",
		Remotes: map[string]string{
			"origin": remotePath,
		},
	})

	remotes := testutil.GetRemotes(t, repoPath)
	originURL, exists := remotes["origin"]
	if !exists {
		t.Fatal("Expected 'origin' remote to be configured")
	}
	if originURL != remotePath {
		t.Fatalf("Expected origin URL %q, got %q", remotePath, originURL)
	}
}

func TestCreateBareRemote(t *testing.T) {
	remotePath := testutil.CreateBareRemote(t, "test-remote")

	// Verify it's a bare repo
	gitDir := filepath.Join(remotePath, "config")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("Expected bare repo config to exist at %s", gitDir)
	}
}

func TestSetupFakeFilesystem(t *testing.T) {
	fsRoot := testutil.SetupFakeFilesystem(t)

	// Verify expected directories exist
	dirs := []string{
		"home/testuser/projects",
		"home/testuser/.cache",
		"root/sys",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(fsRoot, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Fatalf("Expected directory to exist: %s", dirPath)
		}
	}
}

func TestQueryHelpersIgnoreGitTraceStderr(t *testing.T) {
	t.Setenv("GIT_TRACE", "1")

	remotePath := testutil.CreateBareRemote(t, "origin")
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "trace-safe-repo",
		Remotes: map[string]string{
			"origin": remotePath,
		},
	})

	dirty, err := testutil.IsDirtyE(repoPath)
	if err != nil {
		t.Fatalf("IsDirtyE failed: %v", err)
	}
	if dirty {
		t.Fatal("expected clean repo to remain clean when git writes trace to stderr")
	}

	sha, err := testutil.GetCurrentSHAE(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentSHAE failed: %v", err)
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	expectedSHABytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get expected sha: %v", err)
	}
	expectedSHA := strings.TrimSpace(string(expectedSHABytes))
	if sha != expectedSHA {
		t.Fatalf("expected sha %q, got %q", expectedSHA, sha)
	}

	remotes, err := testutil.GetRemotesE(repoPath)
	if err != nil {
		t.Fatalf("GetRemotesE failed: %v", err)
	}
	if got := remotes["origin"]; got != remotePath {
		t.Fatalf("expected origin remote %q, got %q", remotePath, got)
	}
}

func TestCreateTestRepoInDir_InvalidName(t *testing.T) {
	tmp := t.TempDir()

	if _, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{Name: ""}); err == nil {
		t.Fatal("expected error for empty repo name")
	}
	if _, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{Name: "../escape"}); err == nil {
		t.Fatal("expected error for traversal repo name")
	}
	if _, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{Name: "nested/repo"}); err == nil {
		t.Fatal("expected error for nested repo name")
	}
	if _, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{Name: `nested\repo`}); err == nil {
		t.Fatal("expected error for separator in repo name")
	}

	absoluteName := "/tmp/abs-repo"
	if runtime.GOOS == "windows" {
		absoluteName = `C:\abs-repo`
	}
	if _, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{Name: absoluteName}); err == nil {
		t.Fatal("expected error for absolute repo name")
	}
}

func TestCreateBareRemoteInDir_InvalidName(t *testing.T) {
	tmp := t.TempDir()

	if _, err := testutil.CreateBareRemoteInDir(tmp, ""); err == nil {
		t.Fatal("expected error for empty remote name")
	}
	if _, err := testutil.CreateBareRemoteInDir(tmp, "../escape"); err == nil {
		t.Fatal("expected error for traversal remote name")
	}
	if _, err := testutil.CreateBareRemoteInDir(tmp, "nested/remote"); err == nil {
		t.Fatal("expected error for nested remote name")
	}
	if _, err := testutil.CreateBareRemoteInDir(tmp, `nested\remote`); err == nil {
		t.Fatal("expected error for separator in remote name")
	}
}

func TestCreateTestRepoInDir_RestoresOriginalBranch(t *testing.T) {
	tmp := t.TempDir()
	repoPath, err := testutil.CreateTestRepoInDir(tmp, testutil.RepoOptions{
		Name:     "branch-restore",
		Branches: []string{"feature-a", "feature-b"},
	})
	if err != nil {
		t.Fatalf("CreateTestRepoInDir failed: %v", err)
	}

	currentBranch, err := testutil.RunGitCmdE(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatalf("failed to read current branch: %v", err)
	}
	if currentBranch == "feature-a" || currentBranch == "feature-b" {
		t.Fatalf("expected repo to restore original branch, got branch %q", currentBranch)
	}
	if _, err := testutil.RunGitCmdE(repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+currentBranch); err != nil {
		t.Fatalf("expected current branch %q to exist: %v", currentBranch, err)
	}
}
