from __future__ import annotations

import json
import os
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

_CLI_TIMEOUT_SECONDS = 60


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def _cli_cmd() -> list[str]:
    cli = os.environ.get("GIT_TESTKIT_CLI", "").strip()
    if cli:
        cli_path = Path(cli)
        if not cli_path.is_absolute():
            cli_path = _repo_root() / cli_path
        return [str(cli_path)]
    return ["go", "run", "./cmd/git-testkit-cli"]


def _call(op: str, **payload: Any) -> dict[str, Any]:
    request = {"op": op, **payload}
    try:
        proc = subprocess.run(
            _cli_cmd(),
            cwd=_repo_root(),
            input=json.dumps(request),
            text=True,
            capture_output=True,
            check=False,
            timeout=_CLI_TIMEOUT_SECONDS,
        )
    except subprocess.TimeoutExpired as exc:
        raise RuntimeError(
            f"git-testkit-cli timed out after {_CLI_TIMEOUT_SECONDS}s (op={op})"
        ) from exc
    stdout = (proc.stdout or "").strip()
    stderr = (proc.stderr or "").strip()
    if proc.returncode != 0:
        if stdout:
            try:
                response = json.loads(stdout)
            except json.JSONDecodeError:
                response = {}
            if not response.get("ok", True) and response.get("error"):
                raise RuntimeError(str(response["error"]))
        raise RuntimeError(
            f"git-testkit-cli exited {proc.returncode}: {stderr}; stdout: {stdout}"
        )

    try:
        response = json.loads(stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(
            f"invalid JSON from git-testkit-cli: {stdout!r}; stderr: {stderr}"
        ) from exc
    if not response.get("ok", False):
        raise RuntimeError(response.get("error", "unknown git-testkit-cli error"))
    return response


@dataclass(slots=True)
class RepoOptions:
    name: str
    dirty: bool = False
    files: dict[str, str] = field(default_factory=dict)
    remotes: dict[str, str] = field(default_factory=dict)
    branches: list[str] = field(default_factory=list)
    initial_commit: str = ""

    def to_payload(self) -> dict[str, Any]:
        return {
            "name": self.name,
            "dirty": self.dirty,
            "files": self.files,
            "remotes": self.remotes,
            "branches": self.branches,
            "initialCommit": self.initial_commit,
        }


class GitTestKitClient:
    def create_test_repo(
        self,
        base_dir: Path | str,
        options: RepoOptions | None = None,
        **kwargs: Any,
    ) -> str:
        if options is None:
            options = RepoOptions(**kwargs)
        res = _call("create_test_repo", baseDir=str(base_dir), options=options.to_payload())
        return str(res["repoPath"])

    def create_bare_remote(self, base_dir: Path | str, name: str) -> str:
        res = _call(
            "create_bare_remote",
            baseDir=str(base_dir),
            options={"name": name},
        )
        return str(res["remotePath"])

    def setup_fake_filesystem(self, base_dir: Path | str) -> str:
        res = _call("setup_fake_filesystem", baseDir=str(base_dir))
        return str(res["fsRoot"])

    def run_git_cmd(self, repo_path: str, *args: str) -> str:
        res = _call("run_git_cmd", repoPath=repo_path, args=list(args))
        return str(res.get("output", ""))

    def is_dirty(self, repo_path: str) -> bool:
        res = _call("is_dirty", repoPath=repo_path)
        return bool(res["dirty"])

    def get_remotes(self, repo_path: str) -> dict[str, str]:
        res = _call("get_remotes", repoPath=repo_path)
        return dict(res.get("remotes", {}))

    def get_current_sha(self, repo_path: str) -> str:
        res = _call("get_current_sha", repoPath=repo_path)
        return str(res["sha"])

    def get_branches(self, repo_path: str) -> list[str]:
        res = _call("get_branches", repoPath=repo_path)
        return [str(b) for b in res.get("branches", [])]

    def save_snapshot(self, repo_path: str, snapshot_path: Path | str) -> tuple[str, int]:
        res = _call("snapshot_save", repoPath=repo_path, snapshotPath=str(snapshot_path))
        return str(res["snapshotName"]), int(res["snapshotSize"])

    def load_restore_snapshot(self, snapshot_path: Path | str, base_dir: Path | str) -> str:
        res = _call(
            "snapshot_load_restore",
            snapshotPath=str(snapshot_path),
            baseDir=str(base_dir),
        )
        return str(res["restorePath"])

    # Backward-compatible aliases for smoke tests/docs.
    def snapshot_save(self, repo_path: str, snapshot_path: Path | str) -> dict[str, Any]:
        name, size = self.save_snapshot(repo_path, snapshot_path)
        return {"snapshot_name": name, "snapshot_size": size}

    def snapshot_load_restore(self, snapshot_path: Path | str, base_dir: Path | str) -> str:
        return self.load_restore_snapshot(snapshot_path, base_dir)

