package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	claudeSettingsDirName   = ".claude"
	claudeSettingsFileName  = "settings.local.json"
	sessionStartHookCommand = "allmyagents context --session-start-hook"
)

// claudeSettingsPath returns the path to the project's personal, never-committed
// Claude Code settings file: <projectDir>/.claude/settings.local.json.
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

// readJSONObject reads a JSON object file as a generic map so that any
// keys AllMyAgents doesn't know about are preserved untouched when the
// file is written back. A missing file yields an empty object.
func readJSONObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return settings, nil
}

func writeJSONObject(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s directory: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
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
			if command, _ := hookMap["command"].(string); command == sessionStartHookCommand {
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
				"command": sessionStartHookCommand,
			},
		},
	}

	hooksField["SessionStart"] = append(groups, newGroup)
}
