package io.gitfire.testkit;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class CliBridgeTest {
  @TempDir Path tmp;

  @Test
  void createTestRepoProducesCleanRepoAndBranches() {
    CliBridge bridge = new CliBridge(Path.of("../..").toAbsolutePath().normalize());
    String repo = bridge.createTestRepo(tmp, "subject");

    assertTrue(Files.exists(Path.of(repo, ".git")));
    assertFalse(bridge.isDirty(repo));
    assertFalse(bridge.getBranches(repo).isEmpty());
  }

  @Test
  void createBareRemoteAndPushSmokeFlow() throws Exception {
    CliBridge bridge = new CliBridge(Path.of("../..").toAbsolutePath().normalize());
    String remote = bridge.createBareRemote(tmp, "origin");
    String local = bridge.createTestRepo(tmp, "local");

    bridge.runGitCmd(local, "remote", "add", "origin", remote);
    bridge.runGitCmd(local, "checkout", "-b", "feature");
    Files.writeString(Path.of(local, "README.md"), "feature\n", StandardOpenOption.APPEND);
    bridge.runGitCmd(local, "add", "README.md");
    bridge.runGitCmd(local, "commit", "-m", "update readme");
    bridge.runGitCmd(local, "push", "origin", "feature");

    String localSha = bridge.getCurrentSha(local);
    String remoteSha = bridge.runGitCmd(remote, "rev-parse", "feature").trim();
    assertEquals(localSha, remoteSha);
  }

  @Test
  void snapshotRoundtripSmoke() {
    CliBridge bridge = new CliBridge(Path.of("../..").toAbsolutePath().normalize());
    String repo = bridge.createTestRepo(tmp, "snap");
    Path snapshotPath = tmp.resolve("snapshots").resolve("snap.tar.gz");
    CliBridge.SnapshotInfo info = bridge.snapshotSave(repo, snapshotPath.toString());
    CliBridge.RestoredSnapshot restored = bridge.snapshotLoadRestore(tmp, snapshotPath.toString());

    assertTrue(info.size() > 0);
    assertTrue(Files.exists(Path.of(restored.path())));
    assertEquals(bridge.getCurrentSha(repo), bridge.getCurrentSha(restored.path()));
  }
}
