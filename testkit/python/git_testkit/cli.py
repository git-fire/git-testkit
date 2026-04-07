from __future__ import annotations

import json
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def _cli_cmd() -> list[str]:
    return ["go", "run", "./cmd/git-testkit-cli"]


def _call(op: str, **payload: Any) -> dict[str, Any]:
    request = {"op": op, **payload}
    proc = subprocess.run(
        _cli_cmd(),
        cwd=_repo_root(),
        input=json.dumps(request),
        text=True,
        capture_output=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(f"git-testkit-cli exited {proc.returncode}: {proc.stderr.strip()}")

    response = json.loads(proc.stdout)
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


GitTestKitCLI = GitTestKitClient
