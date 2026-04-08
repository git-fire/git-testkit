## testkit/java

Thin Java wrapper over `git-testkit-cli` (Option A bridge).

### API ergonomics

`CliBridge` now supports typed repo creation options, similar to Python:

- `CliBridge.RepoOptions.builder("repo-name")`
  - `.dirty(true)`
  - `.putFile("src/main.go", "package main\n")`
  - `.putRemote("origin", "/tmp/origin.git")`
  - `.addBranch("feature/demo")`
  - `.initialCommit("Initial commit")`

Use with:

- `bridge.createTestRepo(baseDir, options)`
- `bridge.setupFakeFilesystem(baseDir)`

### Sample smoke implementations

Two executable sample smoke implementations verify the wrapper API:

- `SampleRepoFlowSmoke` validates repository + remote + push flow.
- `SampleSnapshotFlowSmoke` validates snapshot save/load/restore flow.

Run them from `testkit/java`:

- `mvn -Dtest=SampleRepoFlowSmoke test`
- `mvn -Dtest=SampleSnapshotFlowSmoke test`

Or run all Java wrapper tests:

- `mvn test`
