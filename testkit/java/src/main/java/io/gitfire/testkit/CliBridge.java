package io.gitfire.testkit;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public final class CliBridge {
  private static final Pattern ERROR_PATTERN = Pattern.compile("\"error\"\\s*:\\s*\"([^\"]*)\"");
  private static final Pattern REPO_PATH_PATTERN = Pattern.compile("\"repoPath\"\\s*:\\s*\"([^\"]+)\"");
  private static final Pattern REMOTE_PATH_PATTERN = Pattern.compile("\"remotePath\"\\s*:\\s*\"([^\"]+)\"");
  private static final Pattern OUTPUT_PATTERN = Pattern.compile("\"output\"\\s*:\\s*\"([^\"]*)\"");
  private static final Pattern SHA_PATTERN = Pattern.compile("\"sha\"\\s*:\\s*\"([^\"]+)\"");
  private static final Pattern DIRTY_PATTERN = Pattern.compile("\"dirty\"\\s*:\\s*(true|false)");
  private static final Pattern BRANCHES_PATTERN = Pattern.compile("\"branches\"\\s*:\\s*\\[(.*?)]");
  private static final Pattern REMOTES_PATTERN = Pattern.compile("\"remotes\"\\s*:\\s*\\{(.*?)}");
  private static final Pattern SNAPSHOT_NAME_PATTERN = Pattern.compile("\"snapshotName\"\\s*:\\s*\"([^\"]+)\"");
  private static final Pattern SNAPSHOT_SIZE_PATTERN = Pattern.compile("\"snapshotSize\"\\s*:\\s*([0-9]+)");
  private static final Pattern RESTORE_PATH_PATTERN = Pattern.compile("\"restorePath\"\\s*:\\s*\"([^\"]+)\"");

  private static final class CliResult {
    private final String stdout;
    private final String stderr;
    private final int code;

    private CliResult(String stdout, String stderr, int code) {
      this.stdout = stdout;
      this.stderr = stderr;
      this.code = code;
    }
  }

  public record SnapshotInfo(String name, int size) {}

  public record RestoredSnapshot(String path, String name, int size) {}

  private final String cliCommand;
  private final Path workspaceRoot;

  public CliBridge(Path workspaceRoot) {
    this(workspaceRoot, "go run ./cmd/git-testkit-cli");
  }

  public CliBridge() {
    this(Path.of(System.getProperty("user.dir")));
  }

  public CliBridge(Path workspaceRoot, String cliCommand) {
    this.workspaceRoot = workspaceRoot;
    this.cliCommand = cliCommand;
  }

  public String createTestRepo(Path baseDir, String name) {
    String payload =
        "{\"op\":\"create_test_repo\",\"baseDir\":\""
            + escape(baseDir.toString())
            + "\",\"options\":{\"name\":\""
            + escape(name)
            + "\"}}";
    return extractRequired(invoke(payload), REPO_PATH_PATTERN, "repoPath");
  }

  public String createBareRemote(Path baseDir, String name) {
    String payload =
        "{\"op\":\"create_bare_remote\",\"baseDir\":\""
            + escape(baseDir.toString())
            + "\",\"options\":{\"name\":\""
            + escape(name)
            + "\"}}";
    return extractRequired(invoke(payload), REMOTE_PATH_PATTERN, "remotePath");
  }

  public String runGitCmd(String repoPath, String... args) {
    StringBuilder payload =
        new StringBuilder(
            "{\"op\":\"run_git_cmd\",\"repoPath\":\"" + escape(repoPath) + "\",\"args\":[");
    for (int i = 0; i < args.length; i++) {
      if (i > 0) {
        payload.append(',');
      }
      payload.append('"').append(escape(args[i])).append('"');
    }
    payload.append("]}");
    String json = invoke(payload.toString());
    return extractRequired(json, OUTPUT_PATTERN, "output");
  }

  public boolean isDirty(String repoPath) {
    String payload = "{\"op\":\"is_dirty\",\"repoPath\":\"" + escape(repoPath) + "\"}";
    String json = invoke(payload);
    String dirty = extractRequired(json, DIRTY_PATTERN, "dirty");
    return "true".equals(dirty);
  }

  public Map<String, String> getRemotes(String repoPath) {
    String payload = "{\"op\":\"get_remotes\",\"repoPath\":\"" + escape(repoPath) + "\"}";
    String json = invoke(payload);
    Matcher matcher = REMOTES_PATTERN.matcher(json);
    Map<String, String> remotes = new LinkedHashMap<>();
    if (!matcher.find()) {
      return remotes;
    }
    String body = matcher.group(1).trim();
    if (body.isEmpty()) {
      return remotes;
    }
    String[] pairs = body.split(",");
    for (String pair : pairs) {
      String[] kv = pair.split(":", 2);
      if (kv.length != 2) {
        continue;
      }
      String key = unquote(kv[0].trim());
      String value = unquote(kv[1].trim());
      remotes.put(key, value);
    }
    return remotes;
  }

  public String getCurrentSha(String repoPath) {
    String payload = "{\"op\":\"get_current_sha\",\"repoPath\":\"" + escape(repoPath) + "\"}";
    return extractRequired(invoke(payload), SHA_PATTERN, "sha");
  }

  public List<String> getBranches(String repoPath) {
    String payload = "{\"op\":\"get_branches\",\"repoPath\":\"" + escape(repoPath) + "\"}";
    String json = invoke(payload);
    Matcher matcher = BRANCHES_PATTERN.matcher(json);
    List<String> branches = new ArrayList<>();
    if (!matcher.find()) {
      return branches;
    }
    String body = matcher.group(1).trim();
    if (body.isEmpty()) {
      return branches;
    }
    String[] items = body.split(",");
    for (String item : items) {
      branches.add(unquote(item.trim()));
    }
    return branches;
  }

  public SnapshotInfo snapshotRepo(String repoPath) {
    String payload = "{\"op\":\"snapshot_repo\",\"repoPath\":\"" + escape(repoPath) + "\"}";
    String json = invoke(payload);
    String name = extractRequired(json, SNAPSHOT_NAME_PATTERN, "snapshotName");
    int size = Integer.parseInt(extractRequired(json, SNAPSHOT_SIZE_PATTERN, "snapshotSize"));
    return new SnapshotInfo(name, size);
  }

  public SnapshotInfo snapshotSave(String repoPath, String snapshotPath) {
    String payload =
        "{\"op\":\"snapshot_save\",\"repoPath\":\""
            + escape(repoPath)
            + "\",\"snapshotPath\":\""
            + escape(snapshotPath)
            + "\"}";
    String json = invoke(payload);
    String name = extractRequired(json, SNAPSHOT_NAME_PATTERN, "snapshotName");
    int size = Integer.parseInt(extractRequired(json, SNAPSHOT_SIZE_PATTERN, "snapshotSize"));
    return new SnapshotInfo(name, size);
  }

  public RestoredSnapshot snapshotLoadRestore(Path baseDir, String snapshotPath) {
    String payload =
        "{\"op\":\"snapshot_load_restore\",\"baseDir\":\""
            + escape(baseDir.toString())
            + "\",\"snapshotPath\":\""
            + escape(snapshotPath)
            + "\"}";
    String json = invoke(payload);
    String restorePath = extractRequired(json, RESTORE_PATH_PATTERN, "restorePath");
    String name = extractRequired(json, SNAPSHOT_NAME_PATTERN, "snapshotName");
    int size = Integer.parseInt(extractRequired(json, SNAPSHOT_SIZE_PATTERN, "snapshotSize"));
    return new RestoredSnapshot(restorePath, name, size);
  }

  private String invoke(String payload) {
    CliResult result = runCli(payload);
    if (result.code != 0) {
      String stderr = result.stderr == null ? "" : result.stderr;
      throw new RuntimeException("CLI failed with code " + result.code + ": " + stderr);
    }
    if (result.stdout == null || result.stdout.isBlank()) {
      throw new RuntimeException("CLI returned empty response");
    }
    String stdout = result.stdout.trim();
    if (!stdout.contains("\"ok\":true")) {
      String error = extractOptional(stdout, ERROR_PATTERN);
      throw new RuntimeException(error.isEmpty() ? "CLI returned failure response" : error);
    }
    return stdout;
  }

  private CliResult runCli(String payload) {
    try {
      ProcessBuilder pb = new ProcessBuilder("bash", "-lc", cliCommand);
      pb.directory(workspaceRoot.toFile());
      Process process = pb.start();
      process.getOutputStream().write(payload.getBytes(StandardCharsets.UTF_8));
      process.getOutputStream().close();
      int code = process.waitFor();
      String stdout = new String(process.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
      String stderr = new String(process.getErrorStream().readAllBytes(), StandardCharsets.UTF_8);
      return new CliResult(stdout, stderr, code);
    } catch (IOException ex) {
      throw new RuntimeException("failed to invoke CLI", ex);
    } catch (InterruptedException ex) {
      Thread.currentThread().interrupt();
      throw new RuntimeException("interrupted while invoking CLI", ex);
    }
  }

  private static String extractRequired(String json, Pattern pattern, String fieldName) {
    Matcher matcher = pattern.matcher(json);
    if (!matcher.find()) {
      throw new RuntimeException("missing field " + fieldName + " in response: " + json);
    }
    return matcher.group(1);
  }

  private static String extractOptional(String json, Pattern pattern) {
    Matcher matcher = pattern.matcher(json);
    return matcher.find() ? matcher.group(1) : "";
  }

  private static String unquote(String value) {
    String out = value;
    if (out.startsWith("\"") && out.endsWith("\"") && out.length() >= 2) {
      out = out.substring(1, out.length() - 1);
    }
    return out.replace("\\\"", "\"").replace("\\\\", "\\");
  }

  private static String escape(String value) {
    return value.replace("\\", "\\\\").replace("\"", "\\\"");
  }
}
