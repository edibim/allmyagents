// Package context resolves and renders the Effective Context AllMyAgents
// hands to AI executors: the persistent Developer Profile overlaid with any
// active, project-scoped Session Override. It is the single authoritative
// resolution path — executor integrations must build on it rather than
// reading the Developer Profile or Session Override files directly, so
// different executors can never disagree on what the "effective" context
// is.
package context

import (
	"fmt"
	"strings"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

// EffectiveContext is the resolved view of how the current developer wants
// AI assistance to work in the current project.
type EffectiveContext struct {
	// Preferences is the Developer Profile overlaid with the Session
	// Override, keyed by question ID.
	Preferences profile.Preferences
	// Overridden marks which preference IDs came from the Session
	// Override rather than the persistent Developer Profile.
	Overridden map[string]bool
}

// Build loads the Developer Profile and the Session Override for
// projectDir and resolves them into an EffectiveContext.
func Build(projectDir string) (EffectiveContext, error) {
	profilePath, err := profile.DefaultPath()
	if err != nil {
		return EffectiveContext{}, err
	}
	developerProfile, err := profile.Load(profilePath)
	if err != nil {
		return EffectiveContext{}, err
	}

	overridePath := profile.OverridePath(projectDir)
	override, err := profile.LoadOverride(overridePath)
	if err != nil {
		return EffectiveContext{}, err
	}

	overridden := make(map[string]bool, len(override.Preferences))
	for id := range override.Preferences {
		overridden[id] = true
	}

	return EffectiveContext{
		Preferences: profile.Resolve(developerProfile, override),
		Overridden:  overridden,
	}, nil
}

// Render turns an EffectiveContext into the Markdown document AllMyAgents
// hands to an AI executor. It contains only resolved preference labels —
// no filesystem paths, no raw profile internals, no project source.
func Render(ctx EffectiveContext) string {
	var b strings.Builder
	b.WriteString("# AllMyAgents — Developer Context\n\n")
	b.WriteString("This describes how the current developer wants AI assistance to work. ")
	b.WriteString("It comes from their local AllMyAgents Developer Profile, optionally overridden for this project. ")
	b.WriteString("Treat it as working preferences, not as project requirements.\n\n")

	for _, question := range profile.Questions {
		selection, ok := ctx.Preferences[question.ID]
		if !ok || selection.Label == "" {
			continue
		}
		line := fmt.Sprintf("- %s: %s", profile.PreferenceLabel(question.ID), selection.Label)
		if ctx.Overridden[question.ID] {
			line += " (session override for this project)"
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
