package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

// withDeveloperProfile isolates ALLMYAGENTS_HOME to a fresh temp dir and
// saves a synthetic Developer Profile there, so tests never depend on
// (or risk touching) a real profile that might exist on the machine
// running them.
func withDeveloperProfile(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
}

// withoutDeveloperProfile isolates ALLMYAGENTS_HOME to a fresh, empty temp
// dir so tests can exercise the "no profile yet" path deterministically.
func withoutDeveloperProfile(t *testing.T) {
	t.Helper()
	t.Setenv("ALLMYAGENTS_HOME", t.TempDir())
}

func TestRunInitProjectConfiguresAllAgentsWithDeveloperProfile(t *testing.T) {
	withDeveloperProfile(t)
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

	for _, agent := range []string{"claude-code", "codex", "copilot-vscode", "gemini-cli", "antigravity"} {
		if !strings.Contains(out.String(), agent+": configured") {
			t.Fatalf("expected %q to be reported as configured, got %q", agent, out.String())
		}
	}

	wantFiles := []string{
		filepath.Join(".claude", "settings.local.json"),
		"AGENTS.md",
		filepath.Join(".github", "copilot-instructions.md"),
		"GEMINI.local.md",
		filepath.Join(".gemini", "settings.json"),
		filepath.Join(".agents", "rules", "allmyagents-context.md"),
	}
	for _, rel := range wantFiles {
		if !profile.Exists(filepath.Join(projectDir, rel)) {
			t.Fatalf("expected %s to be created", rel)
		}
	}

	exclude := readGitExclude(t, projectDir)
	for _, pattern := range []string{
		".allmyagents/",
		"**/.claude/settings.local.json",
		"AGENTS.md",
		".github/copilot-instructions.md",
		"GEMINI.local.md",
		".gemini/settings.json",
		".agents/rules/allmyagents-context.md",
	} {
		if !strings.Contains(exclude, pattern) {
			t.Fatalf("expected %q in git exclude, got %q", pattern, exclude)
		}
	}
}

func TestRunInitProjectWithoutDeveloperProfileSkipsContentAdapters(t *testing.T) {
	withoutDeveloperProfile(t)
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

	if !strings.Contains(out.String(), "claude-code: configured") {
		t.Fatalf("expected claude-code to still be configured without a profile, got %q", out.String())
	}
	for _, agent := range []string{"codex", "copilot-vscode", "gemini-cli", "antigravity"} {
		if !strings.Contains(out.String(), agent+": skipped") {
			t.Fatalf("expected %q to be skipped without a Developer Profile, got %q", agent, out.String())
		}
	}
	if profile.Exists(filepath.Join(projectDir, "AGENTS.md")) {
		t.Fatal("expected no AGENTS.md to be written without a Developer Profile")
	}
}

// TestRunInitProjectReturnsErrorOnMalformedDeveloperProfile proves that a
// corrupted Developer Profile is treated as a real failure (Finding A):
// init-project must not silently degrade to "no profile yet" and report
// overall success — it must fail loudly, before running any adapter.
func TestRunInitProjectReturnsErrorOnMalformedDeveloperProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	if err := os.WriteFile(filepath.Join(home, "developer-profile.json"), []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected an error for a malformed Developer Profile")
	}
	if errOut.Len() == 0 {
		t.Fatal("expected the failure to be reported on stderr")
	}
	if out.Len() != 0 {
		t.Fatalf("expected no adapter to run when context building fails for a real reason, got stdout %q", out.String())
	}
	if profile.Exists(filepath.Join(projectDir, "AGENTS.md")) {
		t.Fatal("expected no AGENTS.md to be written when context building failed")
	}
}

// TestRunInitProjectReturnsErrorOnMalformedSessionOverride proves the same
// for a corrupted project-local session-override.json, even though the
// Developer Profile itself is perfectly valid.
func TestRunInitProjectReturnsErrorOnMalformedSessionOverride(t *testing.T) {
	withDeveloperProfile(t)
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	overridePath := profile.OverridePath(projectDir)
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(overridePath, []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected an error for a malformed session override")
	}
	if out.Len() != 0 {
		t.Fatalf("expected no adapter to run when context building fails for a real reason, got stdout %q", out.String())
	}
}

func TestRunInitProjectIsIdempotent(t *testing.T) {
	withDeveloperProfile(t)
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

	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")
	if countSessionStartCommandOccurrencesInFile(t, settingsPath) != 1 {
		t.Fatal("expected exactly one hook occurrence after running twice")
	}

	if readFile(t, filepath.Join(projectDir, "AGENTS.md")) == "" {
		t.Fatal("expected AGENTS.md to have content")
	}

	exclude := readGitExclude(t, projectDir)
	for _, pattern := range []string{".allmyagents/", "AGENTS.md", "GEMINI.local.md"} {
		if strings.Count(exclude, pattern) != 1 {
			t.Fatalf("expected exactly one %q entry, got exclude content %q", pattern, exclude)
		}
	}
}

func TestRunInitProjectPreservesUnrelatedExistingClaudeSettings(t *testing.T) {
	withDeveloperProfile(t)
	projectDir := t.TempDir()
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")
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

	if !strings.Contains(readFile(t, settingsPath), `"permissions"`) {
		t.Fatal("expected unrelated 'permissions' key to survive init-project")
	}
	if !strings.Contains(out.String(), "claude-code: configured") {
		t.Fatalf("expected claude-code to be configured, got %q", out.String())
	}
}

func TestRunInitProjectOutsideGitRepoStillConfiguresAgents(t *testing.T) {
	withDeveloperProfile(t)
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
	if !profile.Exists(filepath.Join(projectDir, ".claude", "settings.local.json")) {
		t.Fatal("expected the Claude hook to be created even outside a git repository")
	}
	if !profile.Exists(filepath.Join(projectDir, "AGENTS.md")) {
		t.Fatal("expected AGENTS.md to be created even outside a git repository")
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

// TestRunInitProjectTrackedAgentsFileIsSkippedWithoutBlockingOthers verifies
// partial-failure isolation: a pre-existing tracked AGENTS.md makes only
// the Codex integration skip (non-fatal), while the other four agents are
// still configured and the command still exits successfully.
func TestRunInitProjectTrackedAgentsFileIsSkippedWithoutBlockingOthers(t *testing.T) {
	withDeveloperProfile(t)
	projectDir := t.TempDir()
	runGitForTest(t, projectDir, "init", "-q")
	runGitForTest(t, projectDir, "config", "user.email", "test@example.com")
	runGitForTest(t, projectDir, "config", "user.name", "Test")
	teamAgentsMD := "# Team instructions\n\nRun `make test` before committing.\n"
	if err := os.WriteFile(filepath.Join(projectDir, "AGENTS.md"), []byte(teamAgentsMD), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	runGitForTest(t, projectDir, "add", "AGENTS.md")
	runGitForTest(t, projectDir, "commit", "-q", "-m", "add AGENTS.md")
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	if !strings.Contains(out.String(), "codex: skipped") {
		t.Fatalf("expected codex to be skipped, got %q", out.String())
	}
	for _, agent := range []string{"claude-code", "copilot-vscode", "gemini-cli", "antigravity"} {
		if !strings.Contains(out.String(), agent+": configured") {
			t.Fatalf("expected %q to still be configured, got %q", agent, out.String())
		}
	}
	if readFile(t, filepath.Join(projectDir, "AGENTS.md")) != teamAgentsMD {
		t.Fatal("tracked AGENTS.md must not be modified")
	}
}

// TestRunInitProjectReturnsErrorWhenAnAdapterFails verifies that a genuine
// adapter failure (as opposed to a non-fatal skip) makes init-project
// report overall failure via a non-zero-equivalent error, while every
// other adapter still ran to completion.
func TestRunInitProjectReturnsErrorWhenAnAdapterFails(t *testing.T) {
	withDeveloperProfile(t)
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	// Occupy the .github path with a plain file so Copilot's adapter
	// cannot create the .github/ directory it needs — a genuine error,
	// not a tracked-file skip.
	if err := os.WriteFile(filepath.Join(projectDir, ".github"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"init-project"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected an error when an adapter fails")
	}
	if !strings.Contains(out.String(), "copilot-vscode: error") {
		t.Fatalf("expected copilot-vscode to report an error, got %q", out.String())
	}
	for _, agent := range []string{"claude-code", "codex", "gemini-cli", "antigravity"} {
		if !strings.Contains(out.String(), agent+": configured") {
			t.Fatalf("expected %q to still be configured despite copilot-vscode's failure, got %q", agent, out.String())
		}
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

func runGitForTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func readGitExclude(t *testing.T, projectDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	return string(data)
}

func countSessionStartCommandOccurrencesInFile(t *testing.T, path string) int {
	t.Helper()
	return strings.Count(readFile(t, path), "allmyagents context --session-start-hook")
}
