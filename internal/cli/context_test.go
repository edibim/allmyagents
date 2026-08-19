package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

func TestRunContextPrintsRenderedPreferences(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"context"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "Learning method: Socratic") {
		t.Fatalf("expected rendered preference, got %q", out.String())
	}
	if strings.Contains(out.String(), home) {
		t.Fatalf("rendered context must not leak filesystem paths, got %q", out.String())
	}
}

func TestRunContextSessionStartHookFlagEmitsStructuredJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"context", "--session-start-hook"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	var payload struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if payload.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("hookEventName = %q, want SessionStart", payload.HookSpecificOutput.HookEventName)
	}
	if !strings.Contains(payload.HookSpecificOutput.AdditionalContext, "Learning method: Socratic") {
		t.Fatalf("additionalContext missing rendered preference, got %q", payload.HookSpecificOutput.AdditionalContext)
	}
}

func TestRunContextSessionStartHookGracefulWhenNoProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"context", "--session-start-hook"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	var payload struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(payload.HookSpecificOutput.AdditionalContext, "allmyagents init") {
		t.Fatalf("expected a helpful message pointing at 'allmyagents init', got %q", payload.HookSpecificOutput.AdditionalContext)
	}
}

func TestRunContextMissingProfileReturnsErrorWithoutHookFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"context"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected error when no developer profile exists")
	}
	if !strings.Contains(errOut.String(), "read developer profile") {
		t.Fatalf("expected missing profile error, got %q", errOut.String())
	}
}

func TestRunContextUnknownFlagPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Run([]string{"context", "--bogus"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(errOut.String(), "allmyagents context") {
		t.Fatalf("expected usage output, got %q", errOut.String())
	}
}
