package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

// TestClaudeCodeSessionStartHookRealSubprocess builds the actual
// allmyagents binary and runs it exactly the way Claude Code's official
// SessionStart hook mechanism does: as a subprocess, invoking the literal
// command AllMyAgents' Claude Code adapter installs
// (`allmyagents context --session-start-hook`), with no code path shared
// with the test process other than the compiled binary itself. This is
// the closest thing to a real, fresh-session acceptance test achievable
// without a live Claude Code installation: it verifies the exact stdout
// contract (`hookSpecificOutput.additionalContext`) that Claude Code's
// hook runner parses.
func TestClaudeCodeSessionStartHookRealSubprocess(t *testing.T) {
	binPath := buildAllMyAgentsBinary(t)

	home := t.TempDir()
	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	projectDir := t.TempDir()

	cmd := exec.Command(binPath, "context", "--session-start-hook")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "ALLMYAGENTS_HOME="+home)
	stdout, err := cmd.Output()
	if err != nil {
		t.Fatalf("running compiled binary returned error: %v", err)
	}

	var payload struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(stdout, &payload); err != nil {
		t.Fatalf("subprocess output is not valid JSON: %v\noutput: %s", err, stdout)
	}
	if payload.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("hookEventName = %q, want %q — does not match Claude Code's documented SessionStart hook schema", payload.HookSpecificOutput.HookEventName, "SessionStart")
	}
	if !strings.Contains(payload.HookSpecificOutput.AdditionalContext, "Learning method: Socratic") {
		t.Fatalf("additionalContext missing rendered Developer Profile content, got %q", payload.HookSpecificOutput.AdditionalContext)
	}
}

func buildAllMyAgentsBinary(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "allmyagents")
	cmd := exec.Command("go", "build", "-o", binPath, "github.com/n7ptd2xr8c-cell/allmyagents/cmd/allmyagents")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return binPath
}
