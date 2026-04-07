# GIT_TESTKIT_SPEC

Language-agnostic behavioral contract for the polyglot `testkit` implementations.

## 1) Scope and non-goals

This spec defines fixture, scenario, and snapshot behavior implemented against the real `git` executable.

The existing Go module (`github.com/git-fire/git-testkit`) remains the source behavior reference. The first cross-language delivery uses a Go CLI bridge with thin wrappers in Python/Java. Future native ports (Option B) must satisfy the same document.

## 2) Global guarantees

- **Real git only**: all repository operations execute the system `git` binary.
- **No mocking**: tests exercise actual repositories on disk.
- **Failure is fatal**:
  - Go: test-abort semantics (`t.Fatalf`) for legacy test-facing APIs.
  - Bridge APIs: error-returning core + CLI process exit with JSON error.
  - Python/Java wrappers: bridge failures surface as exceptions.
- **Ephemeral filesystem**:
  - All APIs are intended to operate under per-test temporary directories.
  - Callers own temporary root lifecycle (`t.TempDir`, `tmp_path`, `@TempDir`).

## 3) Fixtures contract

Bridge equivalents are exposed as CLI methods with JSON request/response payloads and deterministic field names.

## 3.1 RepoOptions

Logical options:

- `name`: repository directory name.
- `dirty`: if true, leaves uncommitted changes.
- `files`: map of path -> content for additional committed files.
- `remotes`: map of remote name -> URL/path.
- `branches`: list of branches to create.
- `initialCommitMsg`: optional first commit message (default: `"Initial commit"`).

## 3.2 createTestRepo

Creates a non-bare git repository at `<tmp>/<name>` and returns its path.

Behavior:

1. Initializes git repo and configures user identity.
2. Creates `README.md`, stages, and commits initial commit.
3. For each `files` entry:
   - creates parent directories,
   - writes content,
   - stages and commits with message `Add <filename>`.
4. Adds configured remotes.
5. Creates each branch in `branches` via checkout-new branch.
6. If branches were created, returns checkout to `main` if present, otherwise `master`.
7. If `dirty` is true, writes an uncommitted file (unstaged).

Postconditions:

- Returned path exists.
- Repository has at least one commit.
- Clean unless `dirty=true`.

## 3.3 createBareRemote

Creates bare repo at `<tmp>/<name>.git`, returns its path.

Postconditions:

- Returned path exists.
- It is a valid bare git repository (e.g., has `HEAD`/`config`).

## 3.4 setupFakeFilesystem

Creates a deterministic fake directory tree for path-scanning tests and returns the root path.

Minimum directories:

- `home/testuser/projects`
- `home/testuser/src`
- `home/testuser/.cache`
- `home/testuser/node_modules`
- `root/sys`
- `root/proc`

## 3.5 runGitCmd

Runs `git <args>` in a target repo path.

- Returns command stdout when successful.
- On failure throws/aborts immediately.

## 3.6 isDirty

Returns whether `git status --porcelain` is non-empty.

## 3.7 getRemotes

Returns map `remoteName -> remoteURL`.

- Parses `git remote -v`.
- Handles both `(fetch)` and `(push)` suffixes.
- Handles paths containing spaces and literal text like `" (push)"`.

## 3.8 getCurrentSHA

Returns `git rev-parse HEAD` result trimmed.

## 3.9 getBranches

Returns branch short names from `git branch --format=%(refname:short)`.

## 4) Scenario contract

## 4.1 Scenario

Scenario tracks named repos under one test temp root.

Core operations:

- `createRepo(name)` -> `ScenarioRepo`
- `createBareRepo(name)` -> `ScenarioRepo`
- `getRepo(name)` -> `ScenarioRepo`

## 4.2 ScenarioRepo fluent API

Mutating operations are fluent: each returns the same logical repository object (or a worktree repository object for `addWorktree`) and apply side effects immediately.

Methods:

- `withRemote(remoteName, remoteRepo)`
- `withBranch(branchName)`
- `addFile(name, content)` (writes + stages)
- `modifyFile(name, content)` (writes only)
- `stageFile(name)`
- `commit(msg)`
- `push(remote, branch)`
- `checkout(branch)`
- `addWorktree(branch, path)` -> `ScenarioRepo`
- `path()` -> string
- `getDefaultBranch()` -> `"main"` if exists, else `"master"` if exists, else `"main"`

## 4.3 Prebuilt scenarios

Implementations should provide parity helpers:

- `createCleanRepoScenario`
- `createConflictScenario`
- `createDirtyRepoScenario`
- `createDetachedHeadScenario`
- `createMultiRemoteScenario`
- `createMultiBranchScenario`
- `createLargeRepoScenario`
- `createWorktreeScenario`

Each helper must construct real repositories using fixture/scenario primitives.

## 5) Snapshot contract

## 5.1 snapshotRepo

Captures complete repository filesystem state into an in-memory snapshot object.

- Includes `.git` metadata and working tree files.
- Produces deterministic restoration behavior.

## 5.2 restoreSnapshot

Restores snapshot into `<tmp>/<snapshot.name>` and returns restored repo path.

Security/validity:

- Reject absolute or traversal (`..`) snapshot names/entries.

## 5.3 save/load

- `saveSnapshotToDisk(snapshot, filePath)` writes snapshot bytes.
- `loadSnapshotFromDisk(filePath)` loads bytes and creates snapshot object.

## 5.4 Snapshot metadata

Snapshot exposes:

- `name()`: logical snapshot name.
  - For `snapshotRepo(path)`, derived from source directory basename.
  - For `loadSnapshotFromDisk(path)`, derived from file basename.
  - `"."`, `".."`, and root-like basenames normalize to `"snapshot"`.
- `size()`: byte size of serialized snapshot payload.

## 6) Cross-language smoke flow

Both language ports must support the same end-to-end path:

1. Create bare remote.
2. Create local repo.
3. Add remote and commit.
4. Push default branch.
5. Validate SHA/branch/remotes.
