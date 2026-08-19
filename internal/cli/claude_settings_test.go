package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureClaudeSessionStartHookCreatesFile(t *testing.T) {
	projectDir := t.TempDir()

	changed, path, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		t.Fatalf("ensureClaudeSessionStartHook returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true for a freshly created file")
	}
	wantPath := filepath.Join(projectDir, ".claude", "settings.local.json")
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}
	if !sessionStartHookPresentInFile(t, path) {
		t.Fatalf("expected SessionStart hook command in %s", path)
	}
}

func TestEnsureClaudeSessionStartHookIsIdempotent(t *testing.T) {
	projectDir := t.TempDir()

	if _, _, err := ensureClaudeSessionStartHook(projectDir); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	firstContent := readFile(t, claudeSettingsPath(projectDir))

	changed, _, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		t.Fatalf("second call returned error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false on the second call")
	}
	secondContent := readFile(t, claudeSettingsPath(projectDir))
	if firstContent != secondContent {
		t.Fatalf("settings file changed on idempotent call:\nfirst:  %s\nsecond: %s", firstContent, secondContent)
	}

	groups := countSessionStartCommandOccurrences(t, claudeSettingsPath(projectDir))
	if groups != 1 {
		t.Fatalf("expected exactly one hook occurrence, got %d", groups)
	}
}

func TestEnsureClaudeSessionStartHookPreservesUnrelatedSettings(t *testing.T) {
	projectDir := t.TempDir()
	path := claudeSettingsPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	existing := `{
  "permissions": {"allow": ["Bash(go test:*)"]},
  "apiKeyHelper": "some-helper.sh",
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo hi"}]}]
  }
}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	changed, _, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		t.Fatalf("ensureClaudeSessionStartHook returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true when adding a new SessionStart hook")
	}

	settings := decodeSettings(t, path)
	if _, ok := settings["permissions"]; !ok {
		t.Fatal("expected unrelated 'permissions' key to be preserved")
	}
	if _, ok := settings["apiKeyHelper"]; !ok {
		t.Fatal("expected unrelated 'apiKeyHelper' key to be preserved")
	}
	hooksField, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("expected 'hooks' object to be preserved")
	}
	if _, ok := hooksField["PreToolUse"]; !ok {
		t.Fatal("expected existing PreToolUse hook to be preserved")
	}
	if _, ok := hooksField["SessionStart"]; !ok {
		t.Fatal("expected SessionStart hook to be added")
	}
}

func TestEnsureClaudeSessionStartHookPreservesExistingSessionStartHooks(t *testing.T) {
	projectDir := t.TempDir()
	path := claudeSettingsPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	existing := `{
  "hooks": {
    "SessionStart": [
      {"matcher": "resume", "hooks": [{"type": "command", "command": "echo resumed"}]}
    ]
  }
}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, _, err := ensureClaudeSessionStartHook(projectDir); err != nil {
		t.Fatalf("ensureClaudeSessionStartHook returned error: %v", err)
	}

	settings := decodeSettings(t, path)
	hooksField := settings["hooks"].(map[string]any)
	groups := hooksField["SessionStart"].([]any)
	if len(groups) != 2 {
		t.Fatalf("expected 2 SessionStart groups (existing + ours), got %d", len(groups))
	}
	if !sessionStartHookPresent(settings) {
		t.Fatal("expected our command to be detected as present")
	}
}

func TestEnsureClaudeSessionStartHookDetectsAlreadyPresentCommand(t *testing.T) {
	projectDir := t.TempDir()
	path := claudeSettingsPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	existing := `{
  "hooks": {
    "SessionStart": [
      {"hooks": [{"type": "command", "command": "allmyagents context --session-start-hook"}]}
    ]
  }
}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	changed, _, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		t.Fatalf("ensureClaudeSessionStartHook returned error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false when the command is already present")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	return string(data)
}

func decodeSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	var settings map[string]any
	if err := json.Unmarshal([]byte(readFile(t, path)), &settings); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	return settings
}

func sessionStartHookPresentInFile(t *testing.T, path string) bool {
	t.Helper()
	return sessionStartHookPresent(decodeSettings(t, path))
}

func countSessionStartCommandOccurrences(t *testing.T, path string) int {
	t.Helper()
	settings := decodeSettings(t, path)
	count := 0
	for _, group := range sessionStartGroups(settings) {
		groupMap, ok := group.(map[string]any)
		if !ok {
			continue
		}
		hooks, ok := groupMap["hooks"].([]any)
		if !ok {
			continue
		}
		for _, hook := range hooks {
			hookMap, ok := hook.(map[string]any)
			if !ok {
				continue
			}
			if command, _ := hookMap["command"].(string); command == sessionStartHookCommand {
				count++
			}
		}
	}
	return count
}
