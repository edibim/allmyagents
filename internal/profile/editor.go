package profile

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

const donePreferenceID = "done"

type preferenceItem struct {
	ID          string
	Label       string
	Description string
}

func RunEditor(in io.Reader, out io.Writer, path string) error {
	developerProfile, err := Load(path)
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

	changed, err := editProfile(in, out, &developerProfile, interactive)
	if err != nil {
		return err
	}
	cleanup()
	cleanup = func() {}

	if !changed {
		fmt.Fprintln(out, "No profile changes made.")
		return nil
	}

	developerProfile.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := Update(path, developerProfile); err != nil {
		return err
	}
	fmt.Fprintf(out, "Developer Profile updated at %s\n", path)
	return nil
}

func editProfile(in io.Reader, out io.Writer, developerProfile *DeveloperProfile, interactive bool) (bool, error) {
	reader := bufio.NewReader(in)
	style := newStyle(interactive)
	changed := false
	message := ""

	for {
		preferences := preferenceItems(*developerProfile)
		selectedPreference, err := selectPreference(reader, out, preferences, style, message)
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
		current := developerProfile.Preferences[question.ID]
		option, selected, err := selectPreferenceOption(reader, out, question, current, style)
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

		developerProfile.Preferences[question.ID] = newSelection
		changed = true
		message = fmt.Sprintf("%s changed to %s.", preferenceMenuLabel(question.ID), option.Label)
	}
}

func selectPreference(reader *bufio.Reader, out io.Writer, preferences []preferenceItem, style terminalStyle, message string) (preferenceItem, error) {
	selected := 0
	renderPreferenceList(out, preferences, selected, style, message)

	for {
		key, err := readKey(reader)
		if err != nil {
			return preferenceItem{}, err
		}
		switch key {
		case keyUp:
			if selected > 0 {
				selected--
			}
			renderPreferenceList(out, preferences, selected, style, message)
		case keyDown:
			if selected < len(preferences)-1 {
				selected++
			}
			renderPreferenceList(out, preferences, selected, style, message)
		case keyEnter:
			return preferences[selected], nil
		}
	}
}

func selectPreferenceOption(reader *bufio.Reader, out io.Writer, question Question, current Selection, style terminalStyle) (Option, bool, error) {
	options := append([]Option(nil), question.Options...)
	options = append(options, Option{ID: "BACK", Value: "back", Label: "Back", Description: "Return to the preference list without changing this preference."})
	selected := currentOptionIndex(question, current)
	renderPreferenceEditor(out, question, current, options, selected, style)

	for {
		key, err := readKey(reader)
		if err != nil {
			return Option{}, false, err
		}
		switch key {
		case keyUp:
			if selected > 0 {
				selected--
			}
			renderPreferenceEditor(out, question, current, options, selected, style)
		case keyDown:
			if selected < len(options)-1 {
				selected++
			}
			renderPreferenceEditor(out, question, current, options, selected, style)
		case keyEnter:
			option := options[selected]
			if option.ID == "BACK" {
				return Option{}, false, nil
			}
			return option, true, nil
		}
	}
}

func renderPreferenceList(out io.Writer, preferences []preferenceItem, selected int, style terminalStyle, message string) {
	resetInteractiveScreen(out, style)
	fmt.Fprintln(out, style.blue("AllMyAgents setup"))
	fmt.Fprintln(out, style.blue("Developer Profile"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.dim("Edit one preference at a time."))
	if message != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, style.yellow(message))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.blue("Choose a preference to edit"))
	fmt.Fprintln(out)

	for index, item := range preferences {
		line := "  " + item.Label
		if index == selected {
			line = style.yellow("> " + item.Label)
		}
		fmt.Fprintln(out, line)
		if item.Description != "" {
			fmt.Fprintln(out, style.dim("    "+item.Description))
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, style.dim("Use Up/Down arrows to move. Press Enter to select."))
}

func renderPreferenceEditor(out io.Writer, question Question, current Selection, options []Option, selected int, style terminalStyle) {
	resetInteractiveScreen(out, style)
	fmt.Fprintln(out, style.blue("AllMyAgents setup"))
	fmt.Fprintln(out, style.blue("Developer Profile"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.dim("Editing "+preferenceMenuLabel(question.ID)))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.blue(question.Prompt))
	if current.Label != "" {
		fmt.Fprintln(out, style.dim("Current: "+current.Label))
	}
	fmt.Fprintln(out)

	for index, option := range options {
		label := option.Label
		if option.Value == current.Value && option.ID != "BACK" {
			label += " (current)"
		}
		line := "  " + label
		if index == selected {
			line = style.yellow("> " + label)
		}
		fmt.Fprintln(out, line)
		if option.Description != "" {
			fmt.Fprintln(out, style.dim("    "+option.Description))
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, style.dim("Use Up/Down arrows to move. Press Enter to select."))
}

func preferenceItems(developerProfile DeveloperProfile) []preferenceItem {
	items := make([]preferenceItem, 0, len(Questions)+1)
	for _, question := range Questions {
		selection := developerProfile.Preferences[question.ID]
		description := "Not set"
		if selection.Label != "" {
			description = selection.Label
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
		Description: "Exit profile editing.",
	})
	return items
}

func questionByID(id string) (Question, bool) {
	for _, question := range Questions {
		if question.ID == id {
			return question, true
		}
	}
	return Question{}, false
}

func currentOptionIndex(question Question, current Selection) int {
	for index, option := range question.Options {
		if option.Value == current.Value {
			return index
		}
	}
	return 0
}

func selectionFromOption(option Option) Selection {
	return Selection{
		OptionID:    option.ID,
		Value:       option.Value,
		Label:       option.Label,
		Description: option.Description,
	}
}

func sameSelection(left Selection, right Selection) bool {
	return left.OptionID == right.OptionID &&
		left.Value == right.Value &&
		left.Label == right.Label &&
		left.Description == right.Description
}

func preferenceMenuLabel(questionID string) string {
	switch questionID {
	case "learning_method":
		return "Learning method"
	case "explanation_depth":
		return "Explanation depth"
	case "work_priority":
		return "Work priority"
	case "review_style":
		return "Code review"
	case "uncertainty_policy":
		return "Uncertainty policy"
	default:
		return summaryLabel(questionID)
	}
}

func resetInteractiveScreen(out io.Writer, style terminalStyle) {
	if style.interactive {
		resetScreen(out)
	}
}
