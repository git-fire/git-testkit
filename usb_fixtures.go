package testutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// validateFixtureLayoutDir reports whether layoutDir may be joined under a fixture root.
// Empty layoutDir is allowed (caller may default it).
func validateFixtureLayoutDir(layoutDir string) error {
	if layoutDir == "" {
		return nil
	}
	clean := filepath.Clean(layoutDir)
	if filepath.IsAbs(clean) {
		return fmt.Errorf("must be relative to fixture root: %q", layoutDir)
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("must be relative to fixture root: %q", layoutDir)
	}
	return nil
}

type USBVolumeOptions struct {
	LayoutDir string
	Strategy  string
	CreateReposDir bool
}

type USBVolumeConfig struct {
	SchemaVersion int
	LayoutDir     string
	Strategy      string
	CreatedAt     time.Time
}

func mustRelativeLayoutDir(t *testing.T, layoutDir string) string {
	t.Helper()
	if layoutDir == "" {
		return "repos"
	}
	if err := validateFixtureLayoutDir(layoutDir); err != nil {
		t.Fatalf("layout_dir %v", err)
	}
	return filepath.Clean(layoutDir)
}

func MustUSBVolumeRoot(t *testing.T, opts USBVolumeOptions) string {
	t.Helper()
	root := t.TempDir()
	cfg := USBVolumeConfig{
		SchemaVersion: 1,
		LayoutDir:     mustRelativeLayoutDir(t, opts.LayoutDir),
		Strategy:      opts.Strategy,
		CreatedAt:     time.Now().UTC(),
	}
	if cfg.Strategy == "" {
		cfg.Strategy = "git-mirror"
	}
	WriteUSBVolumeConfig(t, root, cfg)
	if opts.CreateReposDir {
		if err := os.MkdirAll(filepath.Join(root, cfg.LayoutDir), 0o755); err != nil {
			t.Fatalf("failed creating repos dir: %v", err)
		}
	}
	return root
}

func WriteUSBVolumeConfig(t *testing.T, root string, cfg USBVolumeConfig) {
	t.Helper()
	if cfg.SchemaVersion <= 0 {
		cfg.SchemaVersion = 1
	}
	cfg.LayoutDir = mustRelativeLayoutDir(t, cfg.LayoutDir)
	if cfg.Strategy == "" {
		cfg.Strategy = "git-mirror"
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now().UTC()
	}
	content := fmt.Sprintf(
		"schema_version = %d\nlayout_dir = %q\nstrategy = %q\ncreated_at = %q\n",
		cfg.SchemaVersion,
		cfg.LayoutDir,
		cfg.Strategy,
		cfg.CreatedAt.Format(time.RFC3339),
	)
	if err := os.WriteFile(filepath.Join(root, ".git-fire"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing .git-fire: %v", err)
	}
}

func readUSBVolumeConfigBytes(data []byte) (USBVolumeConfig, error) {
	cfg := USBVolumeConfig{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), "\"")
		switch key {
		case "schema_version":
			n, err := strconv.Atoi(val)
			if err != nil {
				return cfg, fmt.Errorf("invalid schema_version %q: %w", val, err)
			}
			cfg.SchemaVersion = n
		case "layout_dir":
			if err := validateFixtureLayoutDir(val); err != nil {
				return cfg, fmt.Errorf("layout_dir: %w", err)
			}
			if val == "" {
				cfg.LayoutDir = ""
			} else {
				cfg.LayoutDir = filepath.Clean(val)
			}
		case "strategy":
			cfg.Strategy = val
		case "created_at":
			if val == "" {
				return cfg, fmt.Errorf("created_at: empty value")
			}
			ts, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return cfg, fmt.Errorf("invalid created_at %q: %w", val, err)
			}
			cfg.CreatedAt = ts
		}
	}
	return cfg, nil
}

func ReadUSBVolumeConfig(t *testing.T, root string) USBVolumeConfig {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".git-fire"))
	if err != nil {
		t.Fatalf("failed reading .git-fire: %v", err)
	}
	cfg, err := readUSBVolumeConfigBytes(data)
	if err != nil {
		t.Fatalf("parse .git-fire: %v", err)
	}
	return cfg
}

func AssertGitDirAt(t *testing.T, path string, wantBare bool) {
	t.Helper()
	if wantBare {
		if _, err := os.Stat(filepath.Join(path, "HEAD")); err != nil {
			t.Fatalf("expected bare repo at %s: %v", path, err)
		}
		return
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		t.Fatalf("expected non-bare repo at %s: %v", path, err)
	}
}

func FileURLForPath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to make abs path: %v", err)
	}
	uPath := filepath.ToSlash(abs)
	if filepath.VolumeName(abs) != "" && !strings.HasPrefix(uPath, "/") {
		uPath = "/" + uPath
	}
	u := &url.URL{Scheme: "file", Path: uPath}
	return u.String()
}
