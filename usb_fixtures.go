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

type USBVolumeOptions struct {
	LayoutDir      string
	Strategy       string
	CreateReposDir bool
}

type USBVolumeConfig struct {
	SchemaVersion int
	LayoutDir     string
	Strategy      string
	CreatedAt     time.Time
}

func MustUSBVolumeRoot(t *testing.T, opts USBVolumeOptions) string {
	t.Helper()
	root := t.TempDir()
	cfg := USBVolumeConfig{
		SchemaVersion: 1,
		LayoutDir:     opts.LayoutDir,
		Strategy:      opts.Strategy,
		CreatedAt:     time.Now().UTC(),
	}
	if cfg.LayoutDir == "" {
		cfg.LayoutDir = "repos"
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
	if cfg.LayoutDir == "" {
		cfg.LayoutDir = "repos"
	}
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

func ReadUSBVolumeConfig(t *testing.T, root string) USBVolumeConfig {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".git-fire"))
	if err != nil {
		t.Fatalf("failed reading .git-fire: %v", err)
	}
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
		val = strings.TrimSpace(val)
		if unquoted, err := strconv.Unquote(val); err == nil {
			val = unquoted
		} else {
			val = strings.Trim(val, "\"")
		}
		switch key {
		case "schema_version":
			n, _ := strconv.Atoi(val)
			cfg.SchemaVersion = n
		case "layout_dir":
			cfg.LayoutDir = val
		case "strategy":
			cfg.Strategy = val
		case "created_at":
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				cfg.CreatedAt = ts
			}
		}
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
	filePath := filepath.ToSlash(abs)
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	u := &url.URL{Scheme: "file", Path: filePath}
	return u.String()
}
