package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRestoreSnapshotRejectsUnsafeSnapshotNames(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		joinArg string
	}{
		{
			name:    "path traversal name",
			base:    filepath.Join("tmp", "root"),
			joinArg: "../escape",
		},
		{
			name:    "absolute name",
			base:    filepath.Join("tmp", "root"),
			joinArg: string(filepath.Separator) + "tmp/escape",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if _, err := safeJoin(tt.base, tt.joinArg); err == nil {
				t.Fatalf("expected safeJoin to reject %q", tt.joinArg)
			}
		})
	}
}

func TestRestoreSnapshotRejectsUnsafeArchivePaths(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		joinArg string
	}{
		{
			name:    "relative traversal path",
			base:    filepath.Join("tmp", "root"),
			joinArg: "../escape.txt",
		},
		{
			name:    "absolute path",
			base:    filepath.Join("tmp", "root"),
			joinArg: string(filepath.Separator) + "tmp/escape.txt",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if _, err := safeJoin(tt.base, tt.joinArg); err == nil {
				t.Fatalf("expected safeJoin to reject %q", tt.joinArg)
			}
		})
	}
}

func TestSnapshotRepoNormalizesTrailingDotPath(t *testing.T) {
	_, repo := CreateCleanRepoScenario(t)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(repo.path); err != nil {
		t.Fatal(err)
	}
	snap := SnapshotRepo(t, ".")
	if got, want := snap.Name(), "snapshot"; got != want {
		t.Fatalf("expected snapshot name %q, got %q", want, got)
	}
	_ = RestoreSnapshot(t, snap)
}

func TestNormalizeSnapshotNameHandlesDotDot(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "..", want: "snapshot"},
		{input: filepath.Join("foo", ".."), want: "snapshot"},
	}

	for _, tt := range tests {
		if got := normalizeSnapshotName(tt.input); got != tt.want {
			t.Fatalf("normalizeSnapshotName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoadSnapshotFromDiskUsesBaseName(t *testing.T) {
	_, repo := CreateCleanRepoScenario(t)
	snapshot := SnapshotRepo(t, repo.path)

	snapshotPath := filepath.Join(t.TempDir(), "nested", "snapshot.tar.gz")

	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		t.Fatalf("failed to create snapshot dir: %v", err)
	}
	SaveSnapshotToDisk(t, snapshot, snapshotPath)

	loaded := LoadSnapshotFromDisk(t, snapshotPath)
	if got, want := loaded.Name(), filepath.Base(snapshotPath); got != want {
		t.Fatalf("expected loaded snapshot name %q, got %q", want, got)
	}

	restoredPath := RestoreSnapshot(t, loaded)
	if got, want := filepath.Base(restoredPath), loaded.Name(); got != want {
		t.Fatalf("expected restore dir base %q, got %q", want, got)
	}
}

func TestRestoreSnapshotRejectsUnsupportedEntryTypes(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add directory root entry for restore target.
	if err := tw.WriteHeader(&tar.Header{
		Name:     "repo",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}); err != nil {
		t.Fatalf("failed to write dir header: %v", err)
	}

	// Add character device entry that restore does not support.
	if err := tw.WriteHeader(&tar.Header{
		Name:     "repo/chardev",
		Typeflag: tar.TypeChar,
		Mode:     0600,
	}); err != nil {
		t.Fatalf("failed to write unsupported header: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	snapshot := &Snapshot{name: "repo", tarball: buf.Bytes()}
	_, err := RestoreSnapshotToDir(snapshot, t.TempDir())
	if err == nil {
		t.Fatal("expected RestoreSnapshotToDir to fail on unsupported tar entry")
	}
	if got := err.Error(); !strings.Contains(got, "unsupported snapshot entry type") {
		t.Fatalf("expected unsupported entry error, got: %v", err)
	}
}

func TestSnapshotRoundtripRestoresSymlinkEntries(t *testing.T) {
	if _, err := os.Stat("/"); err != nil {
		t.Skip("symlink test requires filesystem support")
	}

	root := t.TempDir()
	repoPath, err := CreateTestRepoInDir(root, RepoOptions{Name: "symlink-repo"})
	if err != nil {
		t.Fatalf("failed creating repo: %v", err)
	}

	targetFile := filepath.Join(repoPath, "target.txt")
	if err := os.WriteFile(targetFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed writing target file: %v", err)
	}
	linkPath := filepath.Join(repoPath, "link.txt")
	if err := os.Symlink("target.txt", linkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	snapshot := SnapshotRepo(t, repoPath)
	restorePath := RestoreSnapshot(t, snapshot)
	restoredLink := filepath.Join(restorePath, "link.txt")

	info, err := os.Lstat(restoredLink)
	if err != nil {
		t.Fatalf("expected symlink to exist after restore: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to be a symlink", restoredLink)
	}
	destination, err := os.Readlink(restoredLink)
	if err != nil {
		t.Fatalf("failed to read restored symlink: %v", err)
	}
	if destination != "target.txt" {
		t.Fatalf("expected symlink target %q, got %q", "target.txt", destination)
	}
}

type stubFileInfo struct {
	mode os.FileMode
}

func (s stubFileInfo) Name() string       { return "stub" }
func (s stubFileInfo) Size() int64        { return 0 }
func (s stubFileInfo) Mode() os.FileMode  { return s.mode }
func (s stubFileInfo) ModTime() time.Time { return time.Time{} }
func (s stubFileInfo) IsDir() bool        { return s.mode.IsDir() }
func (s stubFileInfo) Sys() any           { return nil }

func TestSupportsSnapshotEntry(t *testing.T) {
	tests := []struct {
		name string
		mode os.FileMode
		want bool
	}{
		{name: "regular file", mode: 0644, want: true},
		{name: "directory", mode: os.ModeDir | 0755, want: true},
		{name: "symlink", mode: os.ModeSymlink, want: true},
		{name: "named pipe", mode: os.ModeNamedPipe, want: false},
		{name: "character device", mode: os.ModeCharDevice, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := supportsSnapshotEntry(stubFileInfo{mode: tt.mode})
			if got != tt.want {
				t.Fatalf("supportsSnapshotEntry(%v) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}
