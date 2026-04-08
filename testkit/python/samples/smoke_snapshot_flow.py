from __future__ import annotations

import tempfile
from pathlib import Path

from git_testkit import GitTestKitClient


def main() -> None:
    client = GitTestKitClient()
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        repo = client.create_test_repo(root, name="smoke-snapshot")

        snapshot_path = root / "snapshots" / "smoke-snapshot.tar.gz"
        snapshot_path.parent.mkdir(parents=True, exist_ok=True)
        snapshot_name, snapshot_size = client.save_snapshot(repo, snapshot_path)
        restored_path = client.load_restore_snapshot(snapshot_path, root)

        assert snapshot_name == Path(repo).name, "unexpected snapshot name"
        assert snapshot_size > 0, "expected non-empty snapshot"
        assert Path(restored_path).exists(), "restored path must exist"
        assert client.get_current_sha(restored_path) == client.get_current_sha(repo), (
            "snapshot restore must preserve HEAD SHA"
        )

    print("python sample smoke_snapshot_flow: OK")


if __name__ == "__main__":
    main()
