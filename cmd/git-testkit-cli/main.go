package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	testutil "github.com/git-fire/git-testkit"
)

type request struct {
	Op       string            `json:"op"`
	BaseDir  string            `json:"baseDir,omitempty"`
	RepoPath string            `json:"repoPath,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Options  *repoOptionsInput `json:"options,omitempty"`

	SnapshotPath string `json:"snapshotPath,omitempty"`
}

type repoOptionsInput struct {
	Name          string            `json:"name"`
	Dirty         bool              `json:"dirty,omitempty"`
	Files         map[string]string `json:"files,omitempty"`
	Remotes       map[string]string `json:"remotes,omitempty"`
	Branches      []string          `json:"branches,omitempty"`
	InitialCommit string            `json:"initialCommit,omitempty"`
}

type response struct {
	OK bool `json:"ok"`

	Error string `json:"error,omitempty"`

	RepoPath     string            `json:"repoPath,omitempty"`
	RemotePath   string            `json:"remotePath,omitempty"`
	FSRoot       string            `json:"fsRoot,omitempty"`
	Output       string            `json:"output,omitempty"`
	Dirty        *bool             `json:"dirty,omitempty"`
	Remotes      map[string]string `json:"remotes,omitempty"`
	SHA          string            `json:"sha,omitempty"`
	Branches     []string          `json:"branches,omitempty"`
	SnapshotName string            `json:"snapshotName,omitempty"`
	SnapshotSize *int              `json:"snapshotSize,omitempty"`
	RestorePath  string            `json:"restorePath,omitempty"`
}

func main() {
	req, err := parseRequest()
	if err != nil {
		writeResponse(response{OK: false, Error: err.Error()})
		os.Exit(1)
	}

	res, err := handle(req)
	if err != nil {
		writeResponse(response{OK: false, Error: err.Error()})
		os.Exit(1)
	}
	writeResponse(res)
}

func parseRequest() (request, error) {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return request{}, fmt.Errorf("invalid JSON request: %w", err)
	}
	if strings.TrimSpace(req.Op) == "" {
		return request{}, fmt.Errorf("missing required field: op")
	}
	return req, nil
}

func handle(req request) (response, error) {
	switch req.Op {
	case "create_test_repo":
		base, err := ensureBaseDir(req.BaseDir)
		if err != nil {
			return response{}, err
		}
		if req.Options == nil {
			return response{}, fmt.Errorf("missing options")
		}
		repoPath, err := testutil.CreateTestRepoInDir(base, testutil.RepoOptions{
			Name:          req.Options.Name,
			Dirty:         req.Options.Dirty,
			Files:         req.Options.Files,
			Remotes:       req.Options.Remotes,
			Branches:      req.Options.Branches,
			InitialCommit: req.Options.InitialCommit,
		})
		if err != nil {
			return response{}, err
		}
		return response{OK: true, RepoPath: repoPath}, nil

	case "create_bare_remote":
		base, err := ensureBaseDir(req.BaseDir)
		if err != nil {
			return response{}, err
		}
		if req.Options == nil || req.Options.Name == "" {
			return response{}, fmt.Errorf("missing options.name")
		}
		remotePath, err := testutil.CreateBareRemoteInDir(base, req.Options.Name)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, RemotePath: remotePath}, nil

	case "setup_fake_filesystem":
		base, err := ensureBaseDir(req.BaseDir)
		if err != nil {
			return response{}, err
		}
		root, err := testutil.SetupFakeFilesystemInDir(base)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, FSRoot: root}, nil

	case "run_git_cmd":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		output, err := testutil.RunGitCmdE(req.RepoPath, req.Args...)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Output: output}, nil

	case "is_dirty":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		dirty, err := testutil.IsDirtyE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Dirty: &dirty}, nil

	case "get_remotes":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		remotes, err := testutil.GetRemotesE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Remotes: remotes}, nil

	case "get_current_sha":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		sha, err := testutil.GetCurrentSHAE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, SHA: sha}, nil

	case "get_branches":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		branches, err := testutil.GetBranchesE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Branches: branches}, nil

	case "snapshot_repo":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		snapshot, err := testutil.SnapshotRepoE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{
			OK:           true,
			SnapshotName: snapshot.Name(),
			SnapshotSize: intPtr(snapshot.Size()),
		}, nil

	case "snapshot_save":
		if req.RepoPath == "" || req.SnapshotPath == "" {
			return response{}, fmt.Errorf("missing repoPath or snapshotPath")
		}
		snapshot, err := testutil.SnapshotRepoE(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		if err := testutil.SaveSnapshotToDiskE(snapshot, req.SnapshotPath); err != nil {
			return response{}, err
		}
		return response{
			OK:           true,
			SnapshotName: snapshot.Name(),
			SnapshotSize: intPtr(snapshot.Size()),
		}, nil

	case "snapshot_load_restore":
		if req.SnapshotPath == "" {
			return response{}, fmt.Errorf("missing snapshotPath")
		}
		base, err := ensureBaseDir(req.BaseDir)
		if err != nil {
			return response{}, err
		}
		snapshot, err := testutil.LoadSnapshotFromDiskE(req.SnapshotPath)
		if err != nil {
			return response{}, err
		}
		restorePath, err := testutil.RestoreSnapshotToDir(snapshot, base)
		if err != nil {
			return response{}, err
		}
		return response{
			OK:           true,
			RestorePath:  restorePath,
			SnapshotName: snapshot.Name(),
			SnapshotSize: intPtr(snapshot.Size()),
		}, nil

	default:
		return response{}, fmt.Errorf("unsupported op: %s", req.Op)
	}
}

func ensureBaseDir(baseDir string) (string, error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", fmt.Errorf("missing baseDir")
	}
	clean := filepath.Clean(baseDir)
	if err := os.MkdirAll(clean, 0755); err != nil {
		return "", err
	}
	return clean, nil
}

func writeResponse(res response) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(res); err != nil {
		fallback := response{
			OK:    false,
			Error: fmt.Sprintf("failed writing response: %s", err.Error()),
		}
		stderrEnc := json.NewEncoder(os.Stderr)
		stderrEnc.SetEscapeHTML(false)
		if encodeErr := stderrEnc.Encode(fallback); encodeErr != nil {
			fmt.Fprintf(os.Stderr, "failed writing fallback response: %v\n", encodeErr)
		}
	}
}

func intPtr(v int) *int {
	return &v
}
