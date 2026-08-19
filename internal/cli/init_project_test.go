package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

func TestRunInitProjectCreatesClaudeHookAndGitExcludes(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("expected no warnings, got %q", errOut.String())
	}
	if !strings.Contains(out.String(), "Added AllMyAgents SessionStart hook") {
		t.Fatalf("expected confirmation message, got %q", out.String())
	}

	settingsPath := claudeSettingsPath(projectDir)
	settings := decodeSettings(t, settingsPath)
	if !sessionStartHookPresent(settings) {
		t.Fatal("expected SessionStart hook to be present")
	}

	excludeData, err := os.ReadFile(filepath.Join(projectDir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	exclude := string(excludeData)
	if !strings.Contains(exclude, ".allmyagents/") {
		t.Fatalf("expected .allmyagents/ in git exclude, got %q", exclude)
	}
	if !strings.Contains(exclude, "**/.claude/settings.local.json") {
		t.Fatalf("expected settings.local.json pattern in git exclude, got %q", exclude)
	}
}

func TestRunInitProjectIsIdempotent(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	t.Chdir(projectDir)

	for i := range 2 {
		var out bytes.Buffer
		var errOut bytes.Buffer
		if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
			t.Fatalf("Run %d returned error: %v\nstderr: %s", i, err, errOut.String())
		}
	}

	settingsPath := claudeSettingsPath(projectDir)
	if countSessionStartCommandOccurrences(t, settingsPath) != 1 {
		t.Fatalf("expected exactly one hook occurrence after running twice")
	}

	excludeData, err := os.ReadFile(filepath.Join(projectDir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if strings.Count(string(excludeData), ".allmyagents/") != 1 {
		t.Fatalf("expected exactly one .allmyagents/ entry, got exclude content %q", string(excludeData))
	}
	if strings.Count(string(excludeData), "**/.claude/settings.local.json") != 1 {
		t.Fatalf("expected exactly one settings.local.json entry, got exclude content %q", string(excludeData))
	}
}

func TestRunInitProjectPreservesUnrelatedExistingSettings(t *testing.T) {
	projectDir := t.TempDir()
	settingsPath := claudeSettingsPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	existing := `{"permissions": {"allow": ["Bash(go build:*)"]}}`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	settings := decodeSettings(t, settingsPath)
	if _, ok := settings["permissions"]; !ok {
		t.Fatal("expected unrelated 'permissions' key to survive init-project")
	}
	if !sessionStartHookPresent(settings) {
		t.Fatal("expected SessionStart hook to be added alongside existing settings")
	}
}

func TestRunInitProjectOutsideGitRepoStillCreatesClaudeHook(t *testing.T) {
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("expected no warnings outside a git repository, got %q", errOut.String())
	}
	if !profile.Exists(claudeSettingsPath(projectDir)) {
		t.Fatal("expected the Claude hook to be created even outside a git repository")
	}
	if profile.Exists(filepath.Join(projectDir, ".git")) {
		t.Fatal("must not create a .git directory outside a repository")
	}
}

func TestRunInitProjectDoesNotTouchDeveloperProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	original := testDeveloperProfile(t)
	if err := profile.Save(profilePath, original); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	updated, err := profile.Load(profilePath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if updated.UpdatedAt != original.UpdatedAt {
		t.Fatalf("developer profile was modified by init-project: got %s, want %s", updated.UpdatedAt, original.UpdatedAt)
	}
}

func TestRunInitProjectUnexpectedArgumentsPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"init-project", "extra"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected error for unexpected arguments")
	}
	if !strings.Contains(errOut.String(), "allmyagents init-project") {
		t.Fatalf("expected usage output, got %q", errOut.String())
	}
}
