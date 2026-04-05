# AGENTS.md

## Cursor Cloud specific instructions

This is a **pure Go test-utility library** with zero external dependencies beyond the Go toolchain and `git` CLI. There are no services, databases, or containers to start.

### Prerequisites

- Go 1.22+ and `git` must be on `PATH` (both are pre-installed in the Cloud Agent VM).

### Key commands

All commands are run from the repo root (`/workspace`):

| Task | Command |
|------|---------|
| Lint (vet) | `go vet ./...` |
| Format check | `gofmt -l *.go` |
| Format fix | `gofmt -w *.go` |
| Run tests | `go test ./...` |
| Verbose tests | `go test -v ./...` |

See `DEVELOPER_GUIDE.md` for full contributor workflow and PR checklist.

### Caveats

- The full test suite (`go test ./...`) takes ~90 seconds due to `TestCreateMultiBranchScenario`, `TestSnapshotAndRestore`, and `TestCreateTestRepo_Dirty` which create many git commits. Use `go test -short ./...` for faster iteration if the tests under change don't need the large-repo scenarios.
- `gofmt` is the only formatting tool used; there is no `golangci-lint` config.
