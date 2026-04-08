package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	// Add symlink entry that restore does not support.
	if err := tw.WriteHeader(&tar.Header{
		Name:     "repo/link",
		Typeflag: tar.TypeSymlink,
		Linkname: "target",
		Mode:     0777,
	}); err != nil {
		t.Fatalf("failed to write symlink header: %v", err)
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
