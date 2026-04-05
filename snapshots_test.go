package testutil

import (
	"os"
	"path/filepath"
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
