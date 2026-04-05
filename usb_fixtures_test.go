package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMustUSBVolumeRoot(t *testing.T) {
	root := MustUSBVolumeRoot(t, USBVolumeOptions{
		LayoutDir:      "repos",
		Strategy:       "git-mirror",
		CreateReposDir: true,
	})
	if _, err := os.Stat(filepath.Join(root, ".git-fire")); err != nil {
		t.Fatalf("expected .git-fire marker: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "repos")); err != nil {
		t.Fatalf("expected repos dir: %v", err)
	}
}

func TestReadWriteUSBVolumeConfig(t *testing.T) {
	root := t.TempDir()
	WriteUSBVolumeConfig(t, root, USBVolumeConfig{
		SchemaVersion: 2,
		LayoutDir:     "custom",
		Strategy:      "git-clone",
	})
	cfg := ReadUSBVolumeConfig(t, root)
	if cfg.SchemaVersion != 2 {
		t.Fatalf("schema mismatch: %d", cfg.SchemaVersion)
	}
	if cfg.LayoutDir != "custom" {
		t.Fatalf("layout mismatch: %s", cfg.LayoutDir)
	}
	if cfg.Strategy != "git-clone" {
		t.Fatalf("strategy mismatch: %s", cfg.Strategy)
	}
}

func TestReadWriteUSBVolumeConfigEscapedValues(t *testing.T) {
	root := t.TempDir()
	WriteUSBVolumeConfig(t, root, USBVolumeConfig{
		SchemaVersion: 3,
		LayoutDir:     `repos\windows\"quoted"`,
		Strategy:      `git\mirror`,
	})

	cfg := ReadUSBVolumeConfig(t, root)
	if cfg.LayoutDir != `repos\windows\"quoted"` {
		t.Fatalf("layout mismatch: %q", cfg.LayoutDir)
	}
	if cfg.Strategy != `git\mirror` {
		t.Fatalf("strategy mismatch: %q", cfg.Strategy)
	}
}

func TestFileURLForPath(t *testing.T) {
	root := t.TempDir()
	got := FileURLForPath(t, root)
	if len(got) < 7 || got[:7] != "file://" {
		t.Fatalf("expected file:// URL, got %s", got)
	}
}

func TestAssertGitDirAt(t *testing.T) {
	nonBare := t.TempDir()
	runGit(t, nonBare, "init")
	AssertGitDirAt(t, nonBare, false)

	parent := t.TempDir()
	bare := filepath.Join(parent, "remote.git")
	runGit(t, parent, "init", "--bare", bare)
	AssertGitDirAt(t, bare, true)
}
