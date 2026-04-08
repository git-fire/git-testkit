# testkit/python

Thin Python wrapper over `git-testkit-cli` (Option A bridge).

The wrapper prioritizes:

- a clean Pythonic API surface,
- zero drift from Go behavior by delegating execution to the Go core,
- easy migration to optional native implementation in a later phase.

## Runnable smoke samples

Two sample implementations exercise and verify the wrapper end-to-end:

- `samples/smoke_repo_flow.py` validates repo creation, remote wiring, push, and SHA parity.
- `samples/smoke_snapshot_flow.py` validates snapshot save/load/restore and SHA parity.

Run from repository root:

- `python3 testkit/python/samples/smoke_repo_flow.py`
- `python3 testkit/python/samples/smoke_snapshot_flow.py`
