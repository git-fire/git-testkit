from __future__ import annotations

from pathlib import Path
import tempfile

from git_testkit import GitTestKitClient


def main() -> int:
    client = GitTestKitClient()
    with tempfile.TemporaryDirectory(prefix="git-testkit-py-repo-") as tmp:
        base = Path(tmp)
        remote = client.create_bare_remote(base, "origin")
        repo = client.create_test_repo(base, name="local-sample")

        client.run_git_cmd(repo, "remote", "add", "origin", remote)
        client.run_git_cmd(repo, "checkout", "-b", "feature/sample")
        client.run_git_cmd(repo, "push", "-u", "origin", "feature/sample")

        local_sha = client.get_current_sha(repo)
        remote_sha = client.run_git_cmd(remote, "rev-parse", "feature/sample").strip()
        if local_sha != remote_sha:
            raise RuntimeError(f"sha mismatch local={local_sha} remote={remote_sha}")
        print("python sample repo flow: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
