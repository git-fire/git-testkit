package io.gitfire.testkit;

import java.nio.file.Files;
import java.nio.file.Path;

/** Runnable sample that exercises repo/remote push flow. */
public class SampleRepoFlowSmoke {
  @org.junit.jupiter.api.Test
  void sampleRepoFlowRuns() throws Exception {
    Path workspaceRoot = Path.of("../..").toAbsolutePath().normalize();
    CliBridge bridge = new CliBridge(workspaceRoot);
    Path tmp = Files.createTempDirectory("git-testkit-java-sample-repo");

    String remote = bridge.createBareRemote(tmp, "origin");
    String local = bridge.createTestRepo(tmp, "local");

    bridge.runGitCmd(local, "remote", "add", "origin", remote);
    bridge.runGitCmd(local, "checkout", "-b", "feature");
    Files.writeString(Path.of(local, "README.md"), "sample update\n", java.nio.file.StandardOpenOption.APPEND);
    bridge.runGitCmd(local, "add", "README.md");
    bridge.runGitCmd(local, "commit", "-m", "sample update");
    bridge.runGitCmd(local, "push", "origin", "feature");

    String localSha = bridge.getCurrentSha(local);
    String remoteSha = bridge.runGitCmd(remote, "rev-parse", "feature").trim();
    if (!localSha.equals(remoteSha)) {
      throw new IllegalStateException("SHA mismatch between local and remote feature branch");
    }
  }
}
