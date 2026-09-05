package adapters

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

// Gemini CLI reads a hierarchical GEMINI.md by default, but the filename(s)
// it searches for are configurable via the documented context.fileName
// setting in .gemini/settings.json (see
// https://geminicli.com/docs/cli/gemini-md/), which accepts an array of
// names. Rather than writing into GEMINI.md itself — the conventional home
// for a team's real, shared project context — AllMyAgents adds its own
// distinct filename, GEMINI.local.md, to that search list and writes the
// Effective Developer Context there instead. Both the settings file and
// the content file are only ever touched when they are not already
// tracked by git.
const (
	geminiSettingsFile = ".gemini/settings.json"
	geminiContentFile  = "GEMINI.local.md"
)

// Gemini configures Gemini CLI to also read GEMINI.local.md.
type Gemini struct{}

func (Gemini) Name() string { return "gemini-cli" }

func (Gemini) Configure(projectDir string, rendered string) Result {
	const agent = "gemini-cli"

	if strings.TrimSpace(rendered) == "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: noProfileYetDetail}
	}

	if profile.IsTracked(projectDir, geminiSettingsFile) {
		return Result{Agent: agent, Status: StatusSkipped, Detail: fmt.Sprintf(
			"%s is tracked by git — AllMyAgents does not modify a project's shared config", geminiSettingsFile,
		)}
	}
	if profile.IsTracked(projectDir, geminiContentFile) {
		return Result{Agent: agent, Status: StatusSkipped, Detail: fmt.Sprintf(
			"%s already exists and is tracked by git — AllMyAgents does not modify a project's shared instructions", geminiContentFile,
		)}
	}

	settingsPath := filepath.Join(projectDir, geminiSettingsFile)
	settings, err := readJSONObject(settingsPath)
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	ensureGeminiContextFileName(settings, geminiContentFile)
	if err := writeJSONObject(settingsPath, settings); err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	if err := profile.EnsureGitExcluded(projectDir, geminiSettingsFile); err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: fmt.Sprintf("wrote %s but could not git-exclude it: %v", geminiSettingsFile, err)}
	}

	skipReason, note, err := writeManagedFile(projectDir, geminiContentFile, buildContent(rendered))
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	if skipReason != "" {
		// Already checked above; kept as defense in depth against a
		// concurrent change between the two checks.
		return Result{Agent: agent, Status: StatusSkipped, Detail: skipReason}
	}

	return Result{Agent: agent, Status: StatusConfigured, Detail: fmt.Sprintf(
		"wrote %s and added it to context.fileName in %s (%s)", geminiContentFile, geminiSettingsFile, note,
	)}
}

// ensureGeminiContextFileName makes sure settings["context"]["fileName"]
// includes want, preserving whatever names (a single string or an array)
// were already configured, and defaulting to Gemini's own "GEMINI.md" when
// nothing was configured yet.
func ensureGeminiContextFileName(settings map[string]any, want string) {
	contextField, ok := settings["context"].(map[string]any)
	if !ok {
		contextField = map[string]any{}
		settings["context"] = contextField
	}

	var names []string
	switch v := contextField["fileName"].(type) {
	case string:
		names = []string{v}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				names = append(names, s)
			}
		}
	}
	if len(names) == 0 {
		names = []string{"GEMINI.md"}
	}

	for _, n := range names {
		if n == want {
			contextField["fileName"] = toAnySlice(names)
			return
		}
	}
	names = append(names, want)
	contextField["fileName"] = toAnySlice(names)
}

func toAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
