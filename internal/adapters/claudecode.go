package adapters

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

// Claude Code reads project instructions from CLAUDE.md, but its official,
// documented mechanism for *dynamic* local context is a hook: a shell
// command Claude Code itself runs at the start of every session and whose
// stdout it injects into context. AllMyAgents uses that hook to run
// `allmyagents context --session-start-hook` instead of writing a static
// file — so unlike every other adapter here, the Effective Developer
// Context is never frozen into a committed-looking file at all; Claude
// Code always sees it freshly computed for the session.
//
// See https://code.claude.com/docs/en/memory (SessionStart hooks).
const (
	claudeSettingsDirName          = ".claude"
	claudeSettingsFileName         = "settings.local.json"
	claudeSessionStartHookCommand  = "allmyagents context --session-start-hook"
	claudeSettingsGitExcludePrefix = "**/"
)

var claudeSettingsGitPattern = claudeSettingsGitExcludePrefix + claudeSettingsDirName + "/" + claudeSettingsFileName

// ClaudeCode configures Claude Code's SessionStart hook.
type ClaudeCode struct{}

func (ClaudeCode) Name() string { return "claude-code" }

// Configure mostly ignores rendered: the hook command re-renders the
// context itself at session start, so there is nothing static to write
// here. It's only used to make the Detail message honest about whether a
// Developer Profile actually exists yet.
func (ClaudeCode) Configure(projectDir string, rendered string) Result {
	const agent = "claude-code"

	changed, path, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}

	note := noGitRepositoryNote
	if profile.HasGitRepository(projectDir) {
		if err := profile.EnsureGitExcluded(projectDir, claudeSettingsGitPattern); err != nil {
			return Result{Agent: agent, Status: StatusError, Detail: fmt.Sprintf("hook written to %s but could not git-exclude it: %v", path, err)}
		}
		note = gitExcludedNote
	}

	if !changed {
		return Result{Agent: agent, Status: StatusConfigured, Detail: fmt.Sprintf("SessionStart hook already present at %s", path)}
	}

	detail := fmt.Sprintf("added SessionStart hook to %s (%s)", path, note)
	if strings.TrimSpace(rendered) == "" {
		detail += "; no Developer Profile yet, so Claude Code will show a placeholder until you run `allmyagents init`"
	}
	return Result{Agent: agent, Status: StatusConfigured, Detail: detail}
}

// claudeSettingsPath returns the path to the project's personal,
// never-committed Claude Code settings file:
// <projectDir>/.claude/settings.local.json.
func claudeSettingsPath(projectDir string) string {
	return filepath.Join(projectDir, claudeSettingsDirName, claudeSettingsFileName)
}

// ensureClaudeSessionStartHook makes sure the project's
// .claude/settings.local.json has a SessionStart hook running
// `allmyagents context --session-start-hook`, without disturbing any other
// settings already in that file. It reports whether the file was changed.
func ensureClaudeSessionStartHook(projectDir string) (changed bool, path string, err error) {
	path = claudeSettingsPath(projectDir)

	settings, err := readJSONObject(path)
	if err != nil {
		return false, path, err
	}

	if sessionStartHookPresent(settings) {
		return false, path, nil
	}

	addSessionStartHook(settings)

	if err := writeJSONObject(path, settings); err != nil {
		return false, path, err
	}
	return true, path, nil
}

// sessionStartHookPresent reports whether our SessionStart hook command is
// already present anywhere in the existing hooks.SessionStart groups.
func sessionStartHookPresent(settings map[string]any) bool {
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
			if command, _ := hookMap["command"].(string); command == claudeSessionStartHookCommand {
				return true
			}
		}
	}
	return false
}

func sessionStartGroups(settings map[string]any) []any {
	hooksField, ok := settings["hooks"].(map[string]any)
	if !ok {
		return nil
	}
	groups, ok := hooksField["SessionStart"].([]any)
	if !ok {
		return nil
	}
	return groups
}

// addSessionStartHook appends a new SessionStart hook group for our
// command. It creates the "hooks" and "hooks.SessionStart" containers if
// they don't already exist, and otherwise appends alongside whatever
// SessionStart groups are already there.
func addSessionStartHook(settings map[string]any) {
	hooksField, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooksField = map[string]any{}
		settings["hooks"] = hooksField
	}

	groups, ok := hooksField["SessionStart"].([]any)
	if !ok {
		groups = []any{}
	}

	newGroup := map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": claudeSessionStartHookCommand,
			},
		},
	}

	hooksField["SessionStart"] = append(groups, newGroup)
}
