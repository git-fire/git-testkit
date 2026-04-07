package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CreateTestRepo creates a real git repository in a temporary directory
// with optional configuration for testing different scenarios
type RepoOptions struct {
	// Name of the repo (used for directory name)
	Name string

	// Add uncommitted files (makes repo "dirty")
	Dirty bool

	// Files to create and commit
	Files map[string]string

	// Remotes to configure (name -> URL)
	Remotes map[string]string

	// Branches to create
	Branches []string

	// Initial commit message
	InitialCommit string
}

// CreateTestRepo creates a test git repository
func CreateTestRepo(t *testing.T, opts RepoOptions) string {
	t.Helper()

	repoPath, err := CreateTestRepoInDir(t.TempDir(), opts)
	if err != nil {
		t.Fatalf("Failed to create test repo: %v", err)
	}
	return repoPath
}

// CreateTestRepoInDir creates a test repository under the provided base directory.
func CreateTestRepoInDir(baseDir string, opts RepoOptions) (string, error) {
	repoName, err := validateSimpleName(opts.Name)
	if err != nil {
		return "", fmt.Errorf("invalid repo name %q: %w", opts.Name, err)
	}
	repoPath := filepath.Join(baseDir, repoName)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create repo directory: %w", err)
	}

	if _, err := RunGitCmdE(repoPath, "init"); err != nil {
		return "", err
	}
	if _, err := RunGitCmdE(repoPath, "config", "user.email", "test@example.com"); err != nil {
		return "", err
	}
	if _, err := RunGitCmdE(repoPath, "config", "user.name", "Test User"); err != nil {
		return "", err
	}

	commitMsg := opts.InitialCommit
	if commitMsg == "" {
		commitMsg = "Initial commit"
	}
	initialFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(initialFile, []byte("# Test Repo\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to create README: %w", err)
	}
	if _, err := RunGitCmdE(repoPath, "add", "README.md"); err != nil {
		return "", err
	}
	if _, err := RunGitCmdE(repoPath, "commit", "-m", commitMsg); err != nil {
		return "", err
	}
	originalBranch, err := RunGitCmdE(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	for filename, content := range opts.Files {
		filePath := filepath.Join(repoPath, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		if _, err := RunGitCmdE(repoPath, "add", filename); err != nil {
			return "", err
		}
		if _, err := RunGitCmdE(repoPath, "commit", "-m", "Add "+filename); err != nil {
			return "", err
		}
	}

	for name, url := range opts.Remotes {
		if _, err := RunGitCmdE(repoPath, "remote", "add", name, url); err != nil {
			return "", err
		}
	}

	for _, branch := range opts.Branches {
		if branch == originalBranch {
			continue
		}
		if _, err := RunGitCmdE(repoPath, "checkout", "-b", branch); err != nil {
			return "", err
		}
	}

	if len(opts.Branches) > 0 {
		if _, err := RunGitCmdE(repoPath, "checkout", originalBranch); err != nil {
			return "", err
		}
	}

	if opts.Dirty {
		dirtyFile := filepath.Join(repoPath, "uncommitted.txt")
		if err := os.WriteFile(dirtyFile, []byte("uncommitted changes\n"), 0644); err != nil {
			return "", fmt.Errorf("failed to create dirty file: %w", err)
		}
	}

	return repoPath, nil
}

// CreateBareRemote creates a bare git repository to use as a remote
func CreateBareRemote(t *testing.T, name string) string {
	t.Helper()

	remotePath, err := CreateBareRemoteInDir(t.TempDir(), name)
	if err != nil {
		t.Fatalf("Failed to create bare remote: %v", err)
	}
	return remotePath
}

// CreateBareRemoteInDir creates a bare remote repository under the provided base directory.
func CreateBareRemoteInDir(baseDir, name string) (string, error) {
	remoteName, err := validateSimpleName(name)
	if err != nil {
		return "", fmt.Errorf("invalid remote name %q: %w", name, err)
	}
	remotePath := filepath.Join(baseDir, remoteName+".git")
	if err := os.MkdirAll(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create bare repo directory: %w", err)
	}
	if _, err := RunGitCmdE(remotePath, "init", "--bare"); err != nil {
		return "", err
	}
	return remotePath, nil
}

// SetupFakeFilesystem creates a fake filesystem structure for scanning tests
func SetupFakeFilesystem(t *testing.T) string {
	t.Helper()

	root, err := SetupFakeFilesystemInDir(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to setup fake filesystem: %v", err)
	}
	return root
}

// SetupFakeFilesystemInDir creates a deterministic fake filesystem tree under baseDir.
func SetupFakeFilesystemInDir(baseDir string) (string, error) {
	dirs := []string{
		"home/testuser/projects",
		"home/testuser/src",
		"home/testuser/.cache",
		"home/testuser/node_modules",
		"root/sys",
		"root/proc",
	}
	for _, dir := range dirs {
		path := filepath.Join(baseDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return baseDir, nil
}

// runGit is a helper to run git commands in a specific directory
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	_, err := RunGitCmdE(dir, args...)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// IsDirty checks if a git repo has uncommitted changes
func IsDirty(t *testing.T, repoPath string) bool {
	t.Helper()

	dirty, err := IsDirtyE(repoPath)
	if err != nil {
		t.Fatalf("Failed to check git status: %v", err)
	}
	return dirty
}

// GetRemotes returns the configured remotes for a repo
func GetRemotes(t *testing.T, repoPath string) map[string]string {
	t.Helper()

	remotes, err := GetRemotesE(repoPath)
	if err != nil {
		t.Fatalf("Failed to get remotes: %v", err)
	}
	return remotes
}

// RunGitCmd runs a git command and fails the test if it errors
// Exported version of runGit for use in other test packages
func RunGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	runGit(t, dir, args...)
}

// RunGitCmdE runs git command in dir and returns trimmed command output.
func RunGitCmdE(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf(
			"git command failed: git %v\nStdout: %s\nStderr: %s\nError: %w",
			args,
			strings.TrimSpace(string(output)),
			strings.TrimSpace(stderr.String()),
			err,
		)
	}
	return strings.TrimSpace(string(output)), nil
}

func validateSimpleName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	if filepath.IsAbs(trimmed) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	if trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("relative traversal segments are not allowed")
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return "", fmt.Errorf("path separators are not allowed")
	}
	return trimmed, nil
}

// GetCurrentSHA returns the current commit SHA
func GetCurrentSHA(t *testing.T, repoPath string) string {
	t.Helper()

	sha, err := GetCurrentSHAE(repoPath)
	if err != nil {
		t.Fatalf("Failed to get current SHA: %v", err)
	}
	return sha
}

// GetBranches returns all branches in the repo
func GetBranches(t *testing.T, repoPath string) []string {
	t.Helper()

	branches, err := GetBranchesE(repoPath)
	if err != nil {
		t.Fatalf("Failed to get branches: %v", err)
	}
	return branches
}

// IsDirtyE checks if a git repo has uncommitted changes.
func IsDirtyE(repoPath string) (bool, error) {
	output, err := RunGitCmdE(repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

// GetRemotesE returns configured remotes for a repo.
func GetRemotesE(repoPath string) (map[string]string, error) {
	output, err := RunGitCmdE(repoPath, "remote", "-v")
	if err != nil {
		return nil, err
	}
	return parseRemotesOutput(output), nil
}

func parseRemotesOutput(output string) map[string]string {
	remotes := make(map[string]string)
	lines := strings.TrimSpace(output)
	if lines == "" {
		return remotes
	}
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, remainder, ok := strings.Cut(line, "\t")
		if !ok {
			idx := strings.IndexAny(line, " \t")
			if idx == -1 {
				continue
			}
			name = strings.TrimSpace(line[:idx])
			remainder = strings.TrimSpace(line[idx+1:])
		} else {
			name = strings.TrimSpace(name)
			remainder = strings.TrimSpace(remainder)
		}

		if strings.HasSuffix(remainder, " (fetch)") {
			remainder = strings.TrimSuffix(remainder, " (fetch)")
		} else if strings.HasSuffix(remainder, " (push)") {
			remainder = strings.TrimSuffix(remainder, " (push)")
		}

		if name != "" && remainder != "" {
			remotes[name] = remainder
		}
	}
	return remotes
}

// GetCurrentSHAE returns the current commit SHA.
func GetCurrentSHAE(repoPath string) (string, error) {
	return RunGitCmdE(repoPath, "rev-parse", "HEAD")
}

// GetBranchesE returns all branches in a repo.
func GetBranchesE(repoPath string) ([]string, error) {
	output, err := RunGitCmdE(repoPath, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	branches := strings.Split(strings.TrimSpace(output), "\n")

	// Filter out empty lines
	var result []string
	for _, b := range branches {
		if b != "" {
			result = append(result, b)
		}
	}

	return result, nil
}
