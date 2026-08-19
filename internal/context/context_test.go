package context

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

func TestBuildResolvesProfileAndOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	overridePath := profile.OverridePath(projectDir)
	override := profile.SessionOverride{
		SchemaVersion: 1,
		Preferences: profile.Preferences{
			"learning_method": {OptionID: "D", Value: "execute", Label: "Execute"},
		},
	}
	if err := profile.SaveOverride(overridePath, override); err != nil {
		t.Fatalf("SaveOverride returned error: %v", err)
	}

	ctx, err := Build(projectDir)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if ctx.Preferences["learning_method"].Value != "execute" {
		t.Fatalf("learning_method = %#v, want execute", ctx.Preferences["learning_method"])
	}
	if !ctx.Overridden["learning_method"] {
		t.Fatal("expected learning_method to be marked overridden")
	}
	if ctx.Overridden["experience_level"] {
		t.Fatal("did not expect experience_level to be marked overridden")
	}
}

func TestBuildWithNoOverrideMarksNothingOverridden(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	ctx, err := Build(projectDir)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(ctx.Overridden) != 0 {
		t.Fatalf("expected no overrides, got %#v", ctx.Overridden)
	}
	if ctx.Preferences["learning_method"].Value != "socratic" {
		t.Fatalf("expected profile value to flow through unchanged, got %#v", ctx.Preferences["learning_method"])
	}
}

func TestBuildMissingProfileReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()

	if _, err := Build(projectDir); err == nil {
		t.Fatal("expected error when the developer profile does not exist")
	}
}

func TestRenderIncludesLabelsAndOverrideMarker(t *testing.T) {
	ctx := EffectiveContext{
		Preferences: profile.Preferences{
			"learning_method": {Value: "execute", Label: "Execute"},
		},
		Overridden: map[string]bool{"learning_method": true},
	}

	rendered := Render(ctx)
	if !strings.Contains(rendered, "Learning method: Execute (session override for this project)") {
		t.Fatalf("expected rendered override marker, got %q", rendered)
	}
}

func TestRenderOmitsPreferencesNotPresent(t *testing.T) {
	ctx := EffectiveContext{Preferences: profile.Preferences{}}

	rendered := Render(ctx)
	if strings.Contains(rendered, "Experience:") {
		t.Fatalf("did not expect unset preferences in rendered output, got %q", rendered)
	}
}

func TestRenderDoesNotLeakFilesystemPaths(t *testing.T) {
	ctx := EffectiveContext{
		Preferences: profile.Preferences{
			"learning_method": {Value: "socratic", Label: "Socratic"},
		},
	}

	rendered := Render(ctx)
	if strings.Contains(rendered, "/") || strings.Contains(rendered, ".json") {
		t.Fatalf("rendered context must not contain filesystem paths, got %q", rendered)
	}
}

func testProfile(t *testing.T) profile.DeveloperProfile {
	t.Helper()

	selections := make(profile.Preferences, len(profile.Questions))
	for _, question := range profile.Questions {
		option := question.Options[0]
		selections[question.ID] = profile.Selection{OptionID: option.ID, Value: option.Value, Label: option.Label, Description: option.Description}
	}
	return profile.NewDeveloperProfile(selections, time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC))
}
