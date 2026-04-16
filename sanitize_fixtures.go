package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const defaultRewriteValidationFixtureName = "rewrite-validation-fixture"

var blockedRefPathTokenPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// RewriteValidationFixtureOptions configures creation of a sanitize/rewrite fixture repository.
type RewriteValidationFixtureOptions struct {
	// Name controls the repository directory name under the fixture root.
	// Defaults to "rewrite-validation-fixture".
	Name string

	// BlockedString is the token intentionally seeded across file contents,
	// commit messages, refs, and object paths.
	//
	// The value must be ref/path safe (letters, numbers, dot, underscore, dash)
	// so it can be embedded in branch/tag names and file paths.
	BlockedString string
}

// RewriteValidationFixture captures important fixture metadata for rewrite tests.
type RewriteValidationFixture struct {
	RepoPath          string
	DefaultBranch     string
	BlockedString     string
	ContentFile       string
	PathWithBlocked   string
	BranchWithBlocked string
	TagWithBlocked    string
}

// BlockedStringSurfaceMatches reports where a blocked string was found.
type BlockedStringSurfaceMatches struct {
	BlockedString  string
	FileContents   []string
	CommitMessages []string
	Refs           []string
	Paths          []string
}

// IsClean reports whether no blocked-string matches were found.
func (m BlockedStringSurfaceMatches) IsClean() bool {
	return len(m.FileContents) == 0 &&
		len(m.CommitMessages) == 0 &&
		len(m.Refs) == 0 &&
		len(m.Paths) == 0
}

// MissingSurfaces reports which required surfaces did not contain a match.
func (m BlockedStringSurfaceMatches) MissingSurfaces() []string {
	var missing []string
	if len(m.FileContents) == 0 {
		missing = append(missing, "file-contents")
	}
	if len(m.CommitMessages) == 0 {
		missing = append(missing, "commit-messages")
	}
	if len(m.Refs) == 0 {
		missing = append(missing, "refs")
	}
	if len(m.Paths) == 0 {
		missing = append(missing, "paths")
	}
	return missing
}

// CreateRewriteValidationFixture creates a repository seeded for sanitize/rewrite validation.
func CreateRewriteValidationFixture(t *testing.T, opts RewriteValidationFixtureOptions) RewriteValidationFixture {
	t.Helper()

	fixture, err := CreateRewriteValidationFixtureInDir(t.TempDir(), opts)
	if err != nil {
		t.Fatalf("Failed to create rewrite validation fixture: %v", err)
	}
	return fixture
}

// CreateRewriteValidationFixtureInDir creates a repository seeded for sanitize/rewrite validation.
func CreateRewriteValidationFixtureInDir(baseDir string, opts RewriteValidationFixtureOptions) (RewriteValidationFixture, error) {
	blocked := strings.TrimSpace(opts.BlockedString)
	if err := validateBlockedRefPathToken(blocked); err != nil {
		return RewriteValidationFixture{}, fmt.Errorf("invalid blocked string %q: %w", opts.BlockedString, err)
	}

	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = defaultRewriteValidationFixtureName
	}

	repoPath, err := CreateTestRepoInDir(baseDir, RepoOptions{Name: name})
	if err != nil {
		return RewriteValidationFixture{}, err
	}

	defaultBranch, err := RunGitCmdE(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return RewriteValidationFixture{}, err
	}

	contentFile := filepath.ToSlash(filepath.Join("fixtures", "blocked-content.txt"))
	pathWithBlocked := filepath.ToSlash(filepath.Join("fixtures", blocked+"-path.txt"))

	if err := writeFixtureFile(repoPath, contentFile, "fixture blocked content: "+blocked+"\n"); err != nil {
		return RewriteValidationFixture{}, err
	}
	if err := writeFixtureFile(repoPath, pathWithBlocked, "fixture path marker\n"); err != nil {
		return RewriteValidationFixture{}, err
	}

	if _, err := RunGitCmdE(repoPath, "add", "--", contentFile, pathWithBlocked); err != nil {
		return RewriteValidationFixture{}, err
	}
	if _, err := RunGitCmdE(repoPath, "commit", "-m", "seed blocked string "+blocked+" for rewrite coverage"); err != nil {
		return RewriteValidationFixture{}, err
	}

	branchWithBlocked := "rewrite/" + blocked + "-branch"
	tagWithBlocked := "rewrite-" + blocked + "-tag"
	if _, err := RunGitCmdE(repoPath, "branch", branchWithBlocked); err != nil {
		return RewriteValidationFixture{}, err
	}
	if _, err := RunGitCmdE(repoPath, "tag", tagWithBlocked); err != nil {
		return RewriteValidationFixture{}, err
	}

	return RewriteValidationFixture{
		RepoPath:          repoPath,
		DefaultBranch:     defaultBranch,
		BlockedString:     blocked,
		ContentFile:       contentFile,
		PathWithBlocked:   pathWithBlocked,
		BranchWithBlocked: branchWithBlocked,
		TagWithBlocked:    tagWithBlocked,
	}, nil
}

// ScanBlockedStringSurfaces scans file contents, commit messages, refs, and object paths.
func ScanBlockedStringSurfaces(t *testing.T, repoPath, blockedString string) BlockedStringSurfaceMatches {
	t.Helper()

	matches, err := ScanBlockedStringSurfacesE(repoPath, blockedString)
	if err != nil {
		t.Fatalf("Failed to scan blocked-string surfaces: %v", err)
	}
	return matches
}

// ScanBlockedStringSurfacesE scans file contents, commit messages, refs, and object paths.
func ScanBlockedStringSurfacesE(repoPath, blockedString string) (BlockedStringSurfaceMatches, error) {
	blocked := strings.TrimSpace(blockedString)
	if blocked == "" {
		return BlockedStringSurfaceMatches{}, fmt.Errorf("blocked string cannot be empty")
	}

	commits, err := listAllCommits(repoPath)
	if err != nil {
		return BlockedStringSurfaceMatches{}, err
	}

	fileMatches, err := collectBlockedFileContentMatches(repoPath, commits, blocked)
	if err != nil {
		return BlockedStringSurfaceMatches{}, err
	}
	commitMatches, err := collectBlockedCommitMessageMatches(repoPath, blocked)
	if err != nil {
		return BlockedStringSurfaceMatches{}, err
	}
	refMatches, err := collectBlockedRefMatches(repoPath, blocked)
	if err != nil {
		return BlockedStringSurfaceMatches{}, err
	}
	pathMatches, err := collectBlockedPathMatches(repoPath, blocked)
	if err != nil {
		return BlockedStringSurfaceMatches{}, err
	}

	sort.Strings(fileMatches)
	sort.Strings(commitMatches)
	sort.Strings(refMatches)
	sort.Strings(pathMatches)

	return BlockedStringSurfaceMatches{
		BlockedString:  blocked,
		FileContents:   fileMatches,
		CommitMessages: commitMatches,
		Refs:           refMatches,
		Paths:          pathMatches,
	}, nil
}

// ValidateBlockedStringCoverageDetectedE verifies blocked-string detection covers all surfaces.
func ValidateBlockedStringCoverageDetectedE(repoPath, blockedString string) error {
	matches, err := ScanBlockedStringSurfacesE(repoPath, blockedString)
	if err != nil {
		return err
	}
	missing := matches.MissingSurfaces()
	if len(missing) > 0 {
		return fmt.Errorf(
			"blocked string %q missing expected surfaces: %s",
			blockedString,
			strings.Join(missing, ", "),
		)
	}
	return nil
}

// ValidateBlockedStringCoverageRemovedE verifies blocked-string matches are gone from all surfaces.
func ValidateBlockedStringCoverageRemovedE(repoPath, blockedString string) error {
	matches, err := ScanBlockedStringSurfacesE(repoPath, blockedString)
	if err != nil {
		return err
	}
	if matches.IsClean() {
		return nil
	}
	return fmt.Errorf("blocked string %q still present: %s", blockedString, summarizeBlockedMatches(matches))
}

// AssertBlockedStringCoverageDetected fails the test when any required surface is missing.
func AssertBlockedStringCoverageDetected(t *testing.T, repoPath, blockedString string) BlockedStringSurfaceMatches {
	t.Helper()

	matches, err := ScanBlockedStringSurfacesE(repoPath, blockedString)
	if err != nil {
		t.Fatalf("scan blocked-string surfaces: %v", err)
	}
	if missing := matches.MissingSurfaces(); len(missing) > 0 {
		t.Fatalf("blocked string %q missing expected surfaces: %s", blockedString, strings.Join(missing, ", "))
	}
	return matches
}

// AssertBlockedStringCoverageRemoved fails the test when any blocked-string matches remain.
func AssertBlockedStringCoverageRemoved(t *testing.T, repoPath, blockedString string) {
	t.Helper()

	matches, err := ScanBlockedStringSurfacesE(repoPath, blockedString)
	if err != nil {
		t.Fatalf("scan blocked-string surfaces: %v", err)
	}
	if !matches.IsClean() {
		t.Fatalf("blocked string %q still present: %s", blockedString, summarizeBlockedMatches(matches))
	}
}

func validateBlockedRefPathToken(token string) error {
	if token == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if !blockedRefPathTokenPattern.MatchString(token) {
		return fmt.Errorf("must contain only letters, numbers, dot, underscore, or dash")
	}
	if strings.HasPrefix(token, ".") || strings.HasSuffix(token, ".") {
		return fmt.Errorf("cannot start or end with dot")
	}
	if strings.HasSuffix(token, ".lock") {
		return fmt.Errorf("cannot end with .lock")
	}
	if strings.Contains(token, "..") {
		return fmt.Errorf("cannot contain consecutive dots")
	}
	return nil
}

func writeFixtureFile(repoPath, relPath, content string) error {
	absPath := filepath.Join(repoPath, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", relPath, err)
	}
	return nil
}

func listAllCommits(repoPath string) ([]string, error) {
	output, err := RunGitCmdE(repoPath, "rev-list", "--all")
	if err != nil {
		return nil, err
	}
	return splitNonEmptyLines(output), nil
}

func collectBlockedFileContentMatches(repoPath string, commits []string, blocked string) ([]string, error) {
	matchSet := make(map[string]struct{})
	for _, commit := range commits {
		lines, err := runGitGrepForCommit(repoPath, commit, blocked)
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			matchSet[commit+":"+line] = struct{}{}
		}
	}
	return sortedKeys(matchSet), nil
}

func runGitGrepForCommit(repoPath, commit, blocked string) ([]string, error) {
	cmd := exec.Command("git", "grep", "-n", "-I", "-F", blocked, commit, "--")
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"git command failed: git %v\nStdout: %s\nStderr: %s\nError: %w",
			[]string{"grep", "-n", "-I", "-F", blocked, commit, "--"},
			strings.TrimSpace(string(output)),
			strings.TrimSpace(stderr.String()),
			err,
		)
	}
	return splitNonEmptyLines(string(output)), nil
}

func collectBlockedCommitMessageMatches(repoPath, blocked string) ([]string, error) {
	output, err := RunGitCmdE(repoPath, "log", "--all", "--format=%H%x09%B%x00")
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, entry := range strings.Split(output, "\x00") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		sha, message, ok := strings.Cut(entry, "\t")
		if !ok {
			continue
		}
		if strings.Contains(message, blocked) {
			firstLine := firstNonEmptyLine(message)
			if firstLine == "" {
				firstLine = "<empty-message>"
			}
			matches = append(matches, sha+"\t"+firstLine)
		}
	}
	return matches, nil
}

func collectBlockedRefMatches(repoPath, blocked string) ([]string, error) {
	output, err := RunGitCmdE(repoPath, "for-each-ref", "--format=%(refname)")
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, line := range splitNonEmptyLines(output) {
		if strings.Contains(line, blocked) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}

func collectBlockedPathMatches(repoPath, blocked string) ([]string, error) {
	output, err := RunGitCmdE(repoPath, "rev-list", "--all", "--objects")
	if err != nil {
		return nil, err
	}

	matchSet := make(map[string]struct{})
	for _, line := range splitNonEmptyLines(output) {
		_, path, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if strings.Contains(path, blocked) {
			matchSet[path] = struct{}{}
		}
	}

	return sortedKeys(matchSet), nil
}

func summarizeBlockedMatches(matches BlockedStringSurfaceMatches) string {
	var parts []string
	if len(matches.FileContents) > 0 {
		parts = append(parts, fmt.Sprintf("file-contents=%d [%s]", len(matches.FileContents), sampleEntries(matches.FileContents)))
	}
	if len(matches.CommitMessages) > 0 {
		parts = append(parts, fmt.Sprintf("commit-messages=%d [%s]", len(matches.CommitMessages), sampleEntries(matches.CommitMessages)))
	}
	if len(matches.Refs) > 0 {
		parts = append(parts, fmt.Sprintf("refs=%d [%s]", len(matches.Refs), sampleEntries(matches.Refs)))
	}
	if len(matches.Paths) > 0 {
		parts = append(parts, fmt.Sprintf("paths=%d [%s]", len(matches.Paths), sampleEntries(matches.Paths)))
	}
	return strings.Join(parts, "; ")
}

func sampleEntries(values []string) string {
	const max = 3
	if len(values) == 0 {
		return ""
	}
	limit := max
	if len(values) < limit {
		limit = len(values)
	}
	sample := strings.Join(values[:limit], ", ")
	if len(values) > limit {
		return sample + ", ..."
	}
	return sample
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func splitNonEmptyLines(output string) []string {
	rawLines := strings.Split(output, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func sortedKeys(set map[string]struct{}) []string {
	items := make([]string, 0, len(set))
	for key := range set {
		items = append(items, key)
	}
	sort.Strings(items)
	return items
}
