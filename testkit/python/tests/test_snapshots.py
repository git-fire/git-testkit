from pathlib import Path

from git_testkit import GitTestKitClient, RepoOptions


def test_snapshot_roundtrip(tmp_path: Path) -> None:
    client = GitTestKitClient()
    repo = client.create_test_repo(tmp_path, RepoOptions(name="snap"))
    snapshot_path = tmp_path / "snapshots" / "snap.tar.gz"
    name, size = client.save_snapshot(repo, snapshot_path)
    restored = client.load_restore_snapshot(snapshot_path, tmp_path)

    assert name == "snap"
    assert size > 0
    assert Path(restored).exists()
    assert client.get_current_sha(restored) == client.get_current_sha(repo)
    assert set(client.get_branches(restored)) == set(client.get_branches(repo))


def test_smoke_remote_push_and_sha(tmp_path: Path) -> None:
    client = GitTestKitClient()
    remote = client.create_bare_remote(tmp_path, "origin")
    local = client.create_test_repo(tmp_path, RepoOptions(name="local"))
    client.run_git_cmd(local, "remote", "add", "origin", remote)
    client.run_git_cmd(local, "checkout", "-b", "feature")
    readme = Path(local) / "README.md"
    readme.write_text(readme.read_text(encoding="utf-8") + "feature\n", encoding="utf-8")
    client.run_git_cmd(local, "add", "README.md")
    client.run_git_cmd(local, "commit", "-m", "feature commit")
    client.run_git_cmd(local, "push", "origin", "feature")

    local_sha = client.get_current_sha(local)
    remote_sha = client.run_git_cmd(remote, "rev-parse", "feature")
    assert local_sha == remote_sha

