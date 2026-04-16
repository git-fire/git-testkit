package testutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	testutil "github.com/git-fire/git-testkit"
)

func TestCreateRewriteValidationFixtureAndCoverageDetection(t *testing.T) {
	fixture := testutil.CreateRewriteValidationFixture(t, testutil.RewriteValidationFixtureOptions{
		Name:          "rewrite-detection",
		BlockedString: "blockedtoken",
	})

	matches := testutil.AssertBlockedStringCoverageDetected(t, fixture.RepoPath, fixture.BlockedString)

	if len(matches.FileContents) == 0 {
		t.Fatal("expected blocked string in file contents")
	}
	if len(matches.CommitMessages) == 0 {
		t.Fatal("expected blocked string in commit messages")
	}
	if len(matches.Refs) == 0 {
		t.Fatal("expected blocked string in refs")
	}
	if len(matches.Paths) == 0 {
		t.Fatal("expected blocked string in paths")
	}

	if !containsSubstring(matches.Refs, fixture.BranchWithBlocked) {
		t.Fatalf("expected blocked branch ref %q in scan results", fixture.BranchWithBlocked)
	}
	if !containsSubstring(matches.Refs, fixture.TagWithBlocked) {
		t.Fatalf("expected blocked tag ref %q in scan results", fixture.TagWithBlocked)
	}
	if !containsString(matches.Paths, fixture.PathWithBlocked) {
		t.Fatalf("expected blocked path %q in scan results", fixture.PathWithBlocked)
	}
}

func TestAssertBlockedStringCoverageRemovedAfterHistoryRewrite(t *testing.T) {
	fixture := testutil.CreateRewriteValidationFixture(t, testutil.RewriteValidationFixtureOptions{
		Name:          "rewrite-clean",
		BlockedString: "leaktoken",
	})

	repoPath := fixture.RepoPath

	testutil.RunGitCmd(t, repoPath, "checkout", "--orphan", "sanitized-history")
	testutil.RunGitCmd(t, repoPath, "rm", "-rf", "--", ".")

	cleanFile := filepath.Join(repoPath, "sanitized.txt")
	if err := os.WriteFile(cleanFile, []byte("sanitized fixture\n"), 0644); err != nil {
		t.Fatalf("failed writing clean file: %v", err)
	}
	testutil.RunGitCmd(t, repoPath, "add", "--", "sanitized.txt")
	testutil.RunGitCmd(t, repoPath, "commit", "-m", "sanitize fixture history")

	testutil.RunGitCmd(t, repoPath, "branch", "-D", fixture.DefaultBranch)
	testutil.RunGitCmd(t, repoPath, "branch", "-m", fixture.DefaultBranch)
	testutil.RunGitCmd(t, repoPath, "branch", "-D", fixture.BranchWithBlocked)
	testutil.RunGitCmd(t, repoPath, "tag", "-d", fixture.TagWithBlocked)

	testutil.AssertBlockedStringCoverageRemoved(t, repoPath, fixture.BlockedString)

	if err := testutil.ValidateBlockedStringCoverageRemovedE(repoPath, fixture.BlockedString); err != nil {
		t.Fatalf("expected removed coverage validation to pass: %v", err)
	}
}

func TestCreateRewriteValidationFixtureInDir_RejectsInvalidBlockedString(t *testing.T) {
	_, err := testutil.CreateRewriteValidationFixtureInDir(t.TempDir(), testutil.RewriteValidationFixtureOptions{
		Name:          "bad-blocked",
		BlockedString: "blocked token with spaces",
	})
	if err == nil {
		t.Fatal("expected invalid blocked string to be rejected")
	}
}

func TestCreateRewriteValidationFixtureInDir_RejectsLeadingHyphen(t *testing.T) {
	_, err := testutil.CreateRewriteValidationFixtureInDir(t.TempDir(), testutil.RewriteValidationFixtureOptions{
		Name:          "hyphen-blocked",
		BlockedString: "-blockedtoken",
	})
	if err == nil {
		t.Fatal("expected leading-hyphen blocked string to be rejected")
	}
}

func TestValidateBlockedStringCoverageDetectedE_FailsWhenCoverageMissing(t *testing.T) {
	repoPath := testutil.CreateTestRepo(t, testutil.RepoOptions{Name: "missing-coverage"})

	err := testutil.ValidateBlockedStringCoverageDetectedE(repoPath, "blockedtoken")
	if err == nil {
		t.Fatal("expected missing coverage validation to fail")
	}
	if !strings.Contains(err.Error(), "missing expected surfaces") {
		t.Fatalf("expected missing coverage error, got: %v", err)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsSubstring(values []string, target string) bool {
	for _, value := range values {
		if strings.Contains(value, target) {
			return true
		}
	}
	return false
}
