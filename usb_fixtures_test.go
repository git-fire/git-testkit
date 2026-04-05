package testutil

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMustUSBVolumeRoot(t *testing.T) {
	root := MustUSBVolumeRoot(t, USBVolumeOptions{
		LayoutDir:       "repos",
		Strategy:        "git-mirror",
		CreateReposDir:  true,
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

func TestValidateFixtureLayoutDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	absUnderRoot := filepath.Join(root, "abs-layout")
	if err := os.MkdirAll(absUnderRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []string{
		"..",
		"../escape",
		"nested/../../../escape",
		absUnderRoot,
	}
	for _, dir := range bad {
		if err := validateFixtureLayoutDir(dir); err == nil {
			t.Errorf("validateFixtureLayoutDir(%q): want error, got nil", dir)
		}
	}
	good := []string{"", "repos", "nested", "nested/../repos"}
	for _, dir := range good {
		if err := validateFixtureLayoutDir(dir); err != nil {
			t.Errorf("validateFixtureLayoutDir(%q): %v", dir, err)
		}
	}
}

func TestReadUSBVolumeConfigBytes_roundTrip(t *testing.T) {
	t.Parallel()
	input := "schema_version = 2\nlayout_dir = \"custom\"\nstrategy = \"git-clone\"\ncreated_at = \"2020-01-02T15:04:05Z\"\n"
	cfg, err := readUSBVolumeConfigBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SchemaVersion != 2 {
		t.Fatalf("schema: got %d", cfg.SchemaVersion)
	}
	if cfg.LayoutDir != "custom" {
		t.Fatalf("layout: got %q", cfg.LayoutDir)
	}
	if cfg.Strategy != "git-clone" {
		t.Fatalf("strategy: got %q", cfg.Strategy)
	}
	if cfg.CreatedAt.IsZero() {
		t.Fatal("created_at: zero")
	}
}

func TestReadUSBVolumeConfigBytes_errors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, content, wantSubstring string
	}{
		{
			name:           "invalid_schema_version",
			content:        "schema_version = notint\n",
			wantSubstring:  "schema_version",
		},
		{
			name:           "layout_dir_escape",
			content:        "layout_dir = ../x\n",
			wantSubstring:  "layout_dir",
		},
		{
			name:           "invalid_created_at",
			content:        "created_at = not-a-date\n",
			wantSubstring:  "created_at",
		},
		{
			name:           "empty_created_at",
			content:        "created_at = \n",
			wantSubstring:  "created_at",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := readUSBVolumeConfigBytes([]byte(tc.content))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantSubstring) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSubstring)
			}
		})
	}
}

func TestFileURLForPath(t *testing.T) {
	root := t.TempDir()
	got := FileURLForPath(t, root)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	if parsed.Scheme != "file" {
		t.Fatalf("scheme %q, want file", parsed.Scheme)
	}
	if parsed.Path == "" || parsed.Path[0] != '/' {
		t.Fatalf("expected absolute path in URL, got path=%q for %q", parsed.Path, got)
	}
	if !strings.HasPrefix(got, "file:///") {
		t.Fatalf("expected canonical file URL with empty authority (file:///...), got %q", got)
	}
}
