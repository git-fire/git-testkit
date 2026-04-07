from pathlib import Path
import subprocess

import pytest

from git_testkit import GitTestKitClient, RepoOptions


def test_create_test_repo_clean(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    repo = cli.create_test_repo(tmp_path, RepoOptions(name="subject"))
    assert Path(repo, ".git").exists()
    assert not cli.is_dirty(repo)
    assert cli.get_branches(repo) != []
    assert cli.get_current_sha(repo)


def test_create_test_repo_dirty(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    repo = cli.create_test_repo(tmp_path, RepoOptions(name="dirty", dirty=True))
    assert cli.is_dirty(repo)
    assert Path(repo, "uncommitted.txt").exists()


def test_create_test_repo_with_files_and_branches(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    repo = cli.create_test_repo(
        tmp_path,
        RepoOptions(
            name="files",
            files={"src/main.py": "print('ok')\n", "config/app.yml": "port: 8080\n"},
            branches=["feature-a", "feature-b"],
        ),
    )
    assert Path(repo, "src/main.py").exists()
    assert Path(repo, "config/app.yml").exists()
    branches = cli.get_branches(repo)
    assert "feature-a" in branches
    assert "feature-b" in branches


def test_create_bare_remote(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    bare = cli.create_bare_remote(tmp_path, "origin")
    assert Path(bare, "HEAD").exists()
    assert Path(bare, "config").exists()


def test_get_remotes_handles_path_with_spaces(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    bare = cli.create_bare_remote(tmp_path, "origin with space")
    repo = cli.create_test_repo(tmp_path, RepoOptions(name="local", remotes={"origin": bare}))
    remotes = cli.get_remotes(repo)
    assert remotes["origin"] == bare


def test_get_remotes_handles_path_with_push_suffix_literal(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    weird_remote = tmp_path / "origin (push)"
    weird_remote.mkdir(parents=True)
    subprocess.run(
        ["git", "init", "--bare", str(weird_remote)],
        check=True,
        capture_output=True,
        text=True,
    )
    repo = cli.create_test_repo(tmp_path, RepoOptions(name="local", remotes={"origin": str(weird_remote)}))
    remotes = cli.get_remotes(repo)
    assert remotes["origin"] == str(weird_remote)


def test_run_git_cmd_failure_raises(tmp_path: Path) -> None:
    cli = GitTestKitClient()
    repo = cli.create_test_repo(tmp_path, RepoOptions(name="r"))
    with pytest.raises(RuntimeError):
        cli.run_git_cmd(repo, "nonexistent-command")
