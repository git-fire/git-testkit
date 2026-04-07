from pathlib import Path

from git_testkit import GitTestKitClient


def test_specify_contract_snapshot_smoke(tmp_path: Path) -> None:
    client = GitTestKitClient()
    repo = client.create_test_repo(tmp_path, name="specify-contract")
    snapshot_path = tmp_path / "snapshots" / "specify-contract.tar.gz"
    snapshot_name, snapshot_size = client.save_snapshot(repo, snapshot_path)
    restored = client.load_restore_snapshot(snapshot_path, tmp_path)

    assert snapshot_name == "specify-contract"
    assert snapshot_size > 0
    assert Path(restored).exists()
    assert client.get_current_sha(restored) == client.get_current_sha(repo)
