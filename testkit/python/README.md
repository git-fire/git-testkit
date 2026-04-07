# testkit/python

Thin Python wrapper over `git-testkit-cli` (Option A bridge).

The wrapper prioritizes:

- a clean Pythonic API surface,
- zero drift from Go behavior by delegating execution to the Go core,
- easy migration to optional native implementation in a later phase.
