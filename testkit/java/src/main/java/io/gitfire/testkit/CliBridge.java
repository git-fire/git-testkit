package io.gitfire.testkit;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public final class CliBridge {
  private static final Pattern ERROR_PATTERN =
      Pattern.compile("\"error\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern REPO_PATH_PATTERN =
      Pattern.compile("\"repoPath\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern REMOTE_PATH_PATTERN =
      Pattern.compile("\"remotePath\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern OUTPUT_PATTERN =
      Pattern.compile("\"output\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern SHA_PATTERN =
      Pattern.compile("\"sha\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern DIRTY_PATTERN = Pattern.compile("\"dirty\"\\s*:\\s*(true|false)");
  private static final Pattern JSON_STRING_ITEM_PATTERN =
      Pattern.compile("\\s*\"((?:\\\\.|[^\\\\\"])*)\"\\s*(?:,|$)");
  private static final Pattern JSON_STRING_PAIR_PATTERN =
      Pattern.compile("\\s*\"((?:\\\\.|[^\\\\\"])*)\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"\\s*(?:,|$)");
  private static final Pattern SNAPSHOT_NAME_PATTERN =
      Pattern.compile("\"snapshotName\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");
  private static final Pattern SNAPSHOT_SIZE_PATTERN = Pattern.compile("\"snapshotSize\"\\s*:\\s*([0-9]+)");
  private static final Pattern RESTORE_PATH_PATTERN =
      Pattern.compile("\"restorePath\"\\s*:\\s*\"((?:\\\\.|[^\\\\\"])*)\"");

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
    if (!json.contains("\"output\"")) {
      return "";
    }
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
    Map<String, String> remotes = new LinkedHashMap<>();
    String body = extractContainerBody(json, "remotes", '{', '}').trim();
    if (body.isEmpty()) {
      return remotes;
    }
    int index = 0;
    while (index < body.length()) {
      Matcher pairMatcher = JSON_STRING_PAIR_PATTERN.matcher(body);
      pairMatcher.region(index, body.length());
      if (!pairMatcher.lookingAt()) {
        throw new RuntimeException("invalid remotes payload: " + json);
      }
      remotes.put(unquote(pairMatcher.group(1)), unquote(pairMatcher.group(2)));
      index = pairMatcher.end();
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
    List<String> branches = new ArrayList<>();
    String body = extractContainerBody(json, "branches", '[', ']').trim();
    if (body.isEmpty()) {
      return branches;
    }
    int index = 0;
    while (index < body.length()) {
      Matcher itemMatcher = JSON_STRING_ITEM_PATTERN.matcher(body);
      itemMatcher.region(index, body.length());
      if (!itemMatcher.lookingAt()) {
        throw new RuntimeException("invalid branches payload: " + json);
      }
      branches.add(unquote(itemMatcher.group(1)));
      index = itemMatcher.end();
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
    String stdout = result.stdout == null ? "" : result.stdout.trim();
    String stderr = result.stderr == null ? "" : result.stderr.trim();
    if (stdout.isBlank() && result.code != 0) {
      throw new RuntimeException("CLI failed with code " + result.code + ": " + stderr);
    }
    if (stdout.isBlank()) {
      throw new RuntimeException("CLI returned empty response");
    }
    if (result.code != 0 || !stdout.contains("\"ok\":true")) {
      String error = extractOptional(stdout, ERROR_PATTERN);
      throw new RuntimeException(
          error.isEmpty() ? "CLI failed with code " + result.code + ": " + stderr : error);
    }
    return stdout;
  }

  private CliResult runCli(String payload) {
    ExecutorService streamReaderPool = null;
    try {
      ProcessBuilder pb = new ProcessBuilder("bash", "-lc", cliCommand);
      pb.directory(workspaceRoot.toFile());
      Process process = pb.start();
      streamReaderPool = Executors.newFixedThreadPool(2);
      Future<String> stdoutFuture =
          streamReaderPool.submit(
              () -> new String(process.getInputStream().readAllBytes(), StandardCharsets.UTF_8));
      Future<String> stderrFuture =
          streamReaderPool.submit(
              () -> new String(process.getErrorStream().readAllBytes(), StandardCharsets.UTF_8));
      process.getOutputStream().write(payload.getBytes(StandardCharsets.UTF_8));
      process.getOutputStream().close();
      int code = process.waitFor();
      String stdout = stdoutFuture.get();
      String stderr = stderrFuture.get();
      return new CliResult(stdout, stderr, code);
    } catch (IOException ex) {
      throw new RuntimeException("failed to invoke CLI", ex);
    } catch (ExecutionException ex) {
      throw new RuntimeException("failed to read CLI output", ex);
    } catch (InterruptedException ex) {
      Thread.currentThread().interrupt();
      throw new RuntimeException("interrupted while invoking CLI", ex);
    } finally {
      if (streamReaderPool != null) {
        streamReaderPool.shutdownNow();
      }
    }
  }

  private static String extractRequired(String json, Pattern pattern, String fieldName) {
    Matcher matcher = pattern.matcher(json);
    if (!matcher.find()) {
      throw new RuntimeException("missing field " + fieldName + " in response: " + json);
    }
    return unquote(matcher.group(1));
  }

  private static String extractOptional(String json, Pattern pattern) {
    Matcher matcher = pattern.matcher(json);
    return matcher.find() ? unquote(matcher.group(1)) : "";
  }

  private static String extractContainerBody(String json, String fieldName, char open, char close) {
    String fieldToken = "\"" + fieldName + "\"";
    int fieldStart = json.indexOf(fieldToken);
    if (fieldStart < 0) {
      return "";
    }
    int colon = json.indexOf(':', fieldStart + fieldToken.length());
    if (colon < 0) {
      throw new RuntimeException("invalid JSON response for field " + fieldName + ": " + json);
    }

    int valueStart = colon + 1;
    while (valueStart < json.length() && Character.isWhitespace(json.charAt(valueStart))) {
      valueStart++;
    }
    if (valueStart >= json.length() || json.charAt(valueStart) != open) {
      throw new RuntimeException("field " + fieldName + " has unexpected JSON type: " + json);
    }

    boolean inString = false;
    boolean escaping = false;
    int depth = 1;
    for (int i = valueStart + 1; i < json.length(); i++) {
      char ch = json.charAt(i);
      if (inString) {
        if (escaping) {
          escaping = false;
        } else if (ch == '\\') {
          escaping = true;
        } else if (ch == '"') {
          inString = false;
        }
        continue;
      }
      if (ch == '"') {
        inString = true;
      } else if (ch == open) {
        depth++;
      } else if (ch == close) {
        depth--;
        if (depth == 0) {
          return json.substring(valueStart + 1, i);
        }
      }
    }
    throw new RuntimeException("unterminated field " + fieldName + " in response: " + json);
  }

  private static String unquote(String value) {
    String out = value;
    if (out.startsWith("\"") && out.endsWith("\"") && out.length() >= 2) {
      out = out.substring(1, out.length() - 1);
    }

    StringBuilder sb = new StringBuilder(out.length());
    for (int i = 0; i < out.length(); i++) {
      char ch = out.charAt(i);
      if (ch != '\\') {
        sb.append(ch);
        continue;
      }
      if (i + 1 >= out.length()) {
        sb.append('\\');
        break;
      }
      char next = out.charAt(++i);
      switch (next) {
        case '"':
          sb.append('"');
          break;
        case '\\':
          sb.append('\\');
          break;
        case '/':
          sb.append('/');
          break;
        case 'b':
          sb.append('\b');
          break;
        case 'f':
          sb.append('\f');
          break;
        case 'n':
          sb.append('\n');
          break;
        case 'r':
          sb.append('\r');
          break;
        case 't':
          sb.append('\t');
          break;
        case 'u':
          if (i + 4 >= out.length()) {
            throw new RuntimeException("invalid unicode escape in JSON string: " + value);
          }
          String hex = out.substring(i + 1, i + 5);
          try {
            sb.append((char) Integer.parseInt(hex, 16));
          } catch (NumberFormatException ex) {
            throw new RuntimeException("invalid unicode escape in JSON string: " + value, ex);
          }
          i += 4;
          break;
        default:
          sb.append(next);
      }
    }
    return sb.toString();
  }

  private static String escape(String value) {
    StringBuilder sb = new StringBuilder(value.length());
    for (int i = 0; i < value.length(); i++) {
      char ch = value.charAt(i);
      switch (ch) {
        case '"':
          sb.append("\\\"");
          break;
        case '\\':
          sb.append("\\\\");
          break;
        case '\b':
          sb.append("\\b");
          break;
        case '\f':
          sb.append("\\f");
          break;
        case '\n':
          sb.append("\\n");
          break;
        case '\r':
          sb.append("\\r");
          break;
        case '\t':
          sb.append("\\t");
          break;
        default:
          if (ch < 0x20) {
            sb.append(String.format("\\u%04x", (int) ch));
          } else {
            sb.append(ch);
          }
      }
    }
    return sb.toString();
  }
}
