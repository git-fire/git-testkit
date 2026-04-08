package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func normalizeSnapshotName(path string) string {
	name := filepath.Base(path)
	if name == "." || name == ".." || name == string(filepath.Separator) {
		return "snapshot"
	}
	return name
}

// Snapshot represents a saved state of a git repository
type Snapshot struct {
	name    string
	tarball []byte // Compressed repository state in memory
}

// SnapshotRepo creates an in-memory snapshot of a repository
// This allows fast restoration of expensive test setups
func SnapshotRepo(t *testing.T, repoPath string) *Snapshot {
	t.Helper()
	snapshot, err := SnapshotRepoE(repoPath)
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}
	return snapshot
}

// SnapshotRepoE creates an in-memory snapshot of a repository and returns errors.
func SnapshotRepoE(repoPath string) (*Snapshot, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	// Walk the repository directory and add all files to tarball
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// Set relative path in tarball
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		if relPath == "." {
			return nil
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// Write file content (if regular file)
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			if _, err := io.Copy(tarWriter, file); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file %s to tar: %w", path, err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("failed to close file %s: %w", path, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return &Snapshot{
		name:    normalizeSnapshotName(repoPath),
		tarball: buf.Bytes(),
	}, nil
}

// RestoreSnapshot restores a snapshot to a new temporary directory
// Returns the path to the restored repository
func RestoreSnapshot(t *testing.T, snapshot *Snapshot) string {
	t.Helper()

	restorePath, err := RestoreSnapshotToDir(snapshot, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to restore snapshot: %v", err)
	}
	return restorePath
}

// RestoreSnapshotToDir restores a snapshot under baseDir and returns restore path.
func RestoreSnapshotToDir(snapshot *Snapshot, baseDir string) (string, error) {
	restorePath, err := safeJoin(baseDir, snapshot.name)
	if err != nil {
		return "", fmt.Errorf("invalid snapshot name %q: %w", snapshot.name, err)
	}

	if err := os.MkdirAll(restorePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create restore directory: %w", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(snapshot.tarball))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath, err := safeJoin(restorePath, header.Name)
		if err != nil {
			return "", fmt.Errorf("invalid snapshot path %q: %w", header.Name, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return "", fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(targetPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
			}
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return "", fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			if err := file.Close(); err != nil {
				return "", fmt.Errorf("failed closing file %s: %w", targetPath, err)
			}
		default:
			return "", fmt.Errorf(
				"unsupported snapshot entry type %d for %q",
				header.Typeflag,
				header.Name,
			)
		}
	}

	return restorePath, nil
}

func safeJoin(base, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || cleanName == string(filepath.Separator) {
		return "", fmt.Errorf("empty or root path is not allowed")
	}
	if filepath.IsAbs(cleanName) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	if cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal is not allowed")
	}

	target := filepath.Join(base, cleanName)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("resolved path escapes restore directory")
	}

	return target, nil
}

// SnapshotSize returns the size of the snapshot in bytes
func (s *Snapshot) Size() int {
	return len(s.tarball)
}

// SnapshotName returns the name of the snapshot
func (s *Snapshot) Name() string {
	return s.name
}

// SaveSnapshotToDisk saves a snapshot to a file (for debugging or caching)
func SaveSnapshotToDisk(t *testing.T, snapshot *Snapshot, filepath string) {
	t.Helper()
	if err := SaveSnapshotToDiskE(snapshot, filepath); err != nil {
		t.Fatalf("Failed to save snapshot to disk: %v", err)
	}
}

// LoadSnapshotFromDisk loads a snapshot from a file
func LoadSnapshotFromDisk(t *testing.T, filePath string) *Snapshot {
	t.Helper()
	snapshot, err := LoadSnapshotFromDiskE(filePath)
	if err != nil {
		t.Fatalf("Failed to load snapshot from disk: %v", err)
	}
	return snapshot
}

// SaveSnapshotToDiskE saves a snapshot to disk and returns errors.
func SaveSnapshotToDiskE(snapshot *Snapshot, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, snapshot.tarball, 0644)
}

// LoadSnapshotFromDiskE loads a snapshot from disk and returns errors.
func LoadSnapshotFromDiskE(filePath string) (*Snapshot, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		name:    normalizeSnapshotName(filePath),
		tarball: data,
	}, nil
}

// Example usage in tests:
//
// Expensive setup (run once):
//   repo := CreateLargeRepoScenario(t, 100, 50)
//   snapshot := SnapshotRepo(t, repo.Path())
//
// Fast restoration (run many times):
//   repoPath := RestoreSnapshot(t, snapshot)
//   // Run test on repoPath (10-100x faster than recreating)
