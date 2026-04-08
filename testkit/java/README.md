## testkit/java

Thin Java wrapper over `git-testkit-cli` (Option A bridge).

### Sample smoke implementations

Two executable sample smoke implementations verify the wrapper API:

- `SampleRepoFlowSmoke` validates repository + remote + push flow.
- `SampleSnapshotFlowSmoke` validates snapshot save/load/restore flow.

Run them from `testkit/java`:

- `mvn -Dtest=SampleRepoFlowSample test`
- `mvn -Dtest=SampleSnapshotFlowSmoke test`

Or run all Java wrapper tests:

- `mvn test`
