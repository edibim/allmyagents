package profile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// SessionOverride is a sparse, project-scoped set of temporary preference
// overrides. It only ever contains the preferences the developer explicitly
// overrode for the current task/session; it never duplicates the full
// Developer Profile and is never merged back into it.
type SessionOverride struct {
	SchemaVersion int         `json:"schema_version"`
	UpdatedAt     string      `json:"updated_at"`
	Preferences   Preferences `json:"preferences"`
}

// LoadOverride reads the session override file at path. A missing file is
// not an error: it means no overrides are active, so an empty override is
// returned.
func LoadOverride(path string) (SessionOverride, error) {
	if !Exists(path) {
		return SessionOverride{SchemaVersion: 1, Preferences: Preferences{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return SessionOverride{}, fmt.Errorf("read session override: %w", err)
	}

	var override SessionOverride
	if err := json.Unmarshal(data, &override); err != nil {
		return SessionOverride{}, fmt.Errorf("decode session override: %w", err)
	}
	if override.Preferences == nil {
		override.Preferences = Preferences{}
	}
	return override, nil
}

func SaveOverride(path string, override SessionOverride) error {
	return write(path, override, "session override")
}

// ClearOverride removes a temporary override. With an empty preferenceID it
// clears all overrides (deleting the file); otherwise it clears just that
// one preference, deleting the file only if no overrides remain. It never
// touches the Developer Profile.
func ClearOverride(path string, preferenceID string) error {
	if preferenceID == "" {
		return removeOverrideFile(path)
	}

	override, err := LoadOverride(path)
	if err != nil {
		return err
	}
	if _, overridden := override.Preferences[preferenceID]; !overridden {
		return nil
	}

	delete(override.Preferences, preferenceID)
	if len(override.Preferences) == 0 {
		return removeOverrideFile(path)
	}

	override.SchemaVersion = 1
	override.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return SaveOverride(path, override)
}

func removeOverrideFile(path string) error {
	if !Exists(path) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove session override: %w", err)
	}
	return nil
}

// Resolve computes the effective preferences by overlaying the session
// override on top of the persistent Developer Profile, key by key. It
// returns a new map; neither input is modified.
func Resolve(developerProfile DeveloperProfile, override SessionOverride) Preferences {
	effective := make(Preferences, len(developerProfile.Preferences))
	for id, selection := range developerProfile.Preferences {
		effective[id] = selection
	}
	for id, selection := range override.Preferences {
		effective[id] = selection
	}
	return effective
}

// RunOverrideEditor lets the developer set temporary, project-scoped
// preference overrides interactively. It reuses the Developer Profile's
// question/option catalog and selection UI, but reads from and writes to
// the session override file only; the Developer Profile is never modified.
func RunOverrideEditor(in io.Reader, out io.Writer, profilePath string, overridePath string) error {
	developerProfile, err := Load(profilePath)
	if err != nil {
		return err
	}
	override, err := LoadOverride(overridePath)
	if err != nil {
		return err
	}

	interactive := isInteractive(in, out)
	restoreTerminal, _ := enableRawMode(in, out)
	if interactive {
		enterAlternateScreen(out)
	}
	cleanup := func() {
		if interactive {
			exitAlternateScreen(out)
		}
		restoreTerminal()
	}
	defer cleanup()

	changed, err := editOverride(in, out, developerProfile, &override, interactive)
	if err != nil {
		return err
	}
	cleanup()
	cleanup = func() {}

	if !changed {
		fmt.Fprintln(out, "No session override changes made.")
		return nil
	}

	override.SchemaVersion = 1
	override.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := SaveOverride(overridePath, override); err != nil {
		return err
	}
	fmt.Fprintf(out, "Session override updated at %s\n", overridePath)
	return nil
}

func editOverride(in io.Reader, out io.Writer, developerProfile DeveloperProfile, override *SessionOverride, interactive bool) (bool, error) {
	reader := bufio.NewReader(in)
	style := newStyle(interactive)
	changed := false
	message := ""
	const title = "Session Override (this project only)"

	for {
		effective := Resolve(developerProfile, *override)
		preferences := overridePreferenceItems(effective, *override)
		selectedPreference, err := selectPreference(reader, out, preferences, style, message, title)
		if err != nil {
			return false, err
		}
		message = ""
		if selectedPreference.ID == donePreferenceID {
			return changed, nil
		}

		question, ok := questionByID(selectedPreference.ID)
		if !ok {
			continue
		}
		current := effective[question.ID]
		option, selected, err := selectPreferenceOption(reader, out, question, current, style, title)
		if err != nil {
			return false, err
		}
		if !selected {
			continue
		}

		newSelection := selectionFromOption(option)
		if sameSelection(current, newSelection) {
			message = fmt.Sprintf("%s is already set to %s.", preferenceMenuLabel(question.ID), option.Label)
			continue
		}

		if override.Preferences == nil {
			override.Preferences = Preferences{}
		}
		override.Preferences[question.ID] = newSelection
		changed = true
		message = fmt.Sprintf("%s overridden to %s for this project.", preferenceMenuLabel(question.ID), option.Label)
	}
}

func overridePreferenceItems(effective Preferences, override SessionOverride) []preferenceItem {
	items := make([]preferenceItem, 0, len(Questions)+1)
	for _, question := range Questions {
		selection := effective[question.ID]
		description := "Not set"
		if selection.Label != "" {
			description = selection.Label
		}
		if _, overridden := override.Preferences[question.ID]; overridden {
			description += " (overridden)"
		}
		items = append(items, preferenceItem{
			ID:          question.ID,
			Label:       preferenceMenuLabel(question.ID),
			Description: description,
		})
	}
	items = append(items, preferenceItem{
		ID:          donePreferenceID,
		Label:       "Done",
		Description: "Exit session override editing.",
	})
	return items
}

// ShowOverride prints the effective preferences for the current project:
// the Developer Profile overlaid with any active session override, and
// which source each preference came from.
func ShowOverride(out io.Writer, profilePath string, overridePath string) error {
	developerProfile, err := Load(profilePath)
	if err != nil {
		return err
	}
	override, err := LoadOverride(overridePath)
	if err != nil {
		return err
	}

	effective := Resolve(developerProfile, override)

	fmt.Fprintf(out, "Session override file: %s\n", overridePath)
	if len(override.Preferences) == 0 {
		fmt.Fprintln(out, "No active overrides. All preferences come from the Developer Profile.")
	} else {
		fmt.Fprintln(out, "Active overrides for this project.")
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Effective preferences:")

	for _, question := range Questions {
		selection := effective[question.ID]
		label := selection.Label
		if label == "" {
			label = "Not set"
		}
		source := "profile"
		if _, overridden := override.Preferences[question.ID]; overridden {
			source = "override"
		}
		fmt.Fprintf(out, "- %s: %s (%s)\n", preferenceMenuLabel(question.ID), label, source)
	}
	return nil
}
