from pathlib import Path

from git_testkit import GitTestKitClient, RepoOptions


def test_python_wrapper_bridge_smoke(tmp_path: Path) -> None:
    client = GitTestKitClient()
    repo = client.create_test_repo(tmp_path, RepoOptions(name="bridge-scenario"))
    assert Path(repo, ".git").exists()
