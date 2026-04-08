package io.gitfire.testkit;

import java.nio.file.Files;
import java.nio.file.Path;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

public class SampleSnapshotFlowSmoke {
  @TempDir Path tmp;

  @Test
  void sampleSnapshotRoundtrip() {
    CliBridge bridge = new CliBridge(Path.of("../..").toAbsolutePath().normalize());
    String repo = bridge.createTestRepo(tmp, "sample-snapshot");
    Path snapshotPath = tmp.resolve("snapshots").resolve("sample-snapshot.tar.gz");
    CliBridge.SnapshotInfo info = bridge.snapshotSave(repo, snapshotPath.toString());
    CliBridge.RestoredSnapshot restored = bridge.snapshotLoadRestore(tmp, snapshotPath.toString());

    if (info.size() <= 0) {
      throw new IllegalStateException("expected snapshot size > 0");
    }
    if (!Files.exists(Path.of(restored.path()))) {
      throw new IllegalStateException("expected restored repo path to exist");
    }

    String originalSha = bridge.getCurrentSha(repo);
    String restoredSha = bridge.getCurrentSha(restored.path());
    if (!originalSha.equals(restoredSha)) {
      throw new IllegalStateException(
          "snapshot roundtrip SHA mismatch: " + originalSha + " vs " + restoredSha);
    }
  }
}
