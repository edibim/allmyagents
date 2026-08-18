package profile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Question struct {
	ID      string
	Prompt  string
	Options []Option
}

type Option struct {
	ID          string
	Value       string
	Label       string
	Description string
}

var Questions = []Question{
	{
		ID:     "experience_level",
		Prompt: "What best describes you?",
		Options: []Option{
			{ID: "A", Value: "learning_guided", Label: "I'm learning programming and still need guidance."},
			{ID: "B", Value: "builder_needs_help", Label: "I can build things, but I still need help with unfamiliar problems."},
			{ID: "C", Value: "independent_speed_review", Label: "I can work independently and mainly need AI for speed and review."},
			{ID: "D", Value: "independent_engineering_partner", Label: "I work independently and want AI mainly as a high-level engineering partner."},
		},
	},
	{
		ID:     "learning_method",
		Prompt: "How should AI help you learn?",
		Options: []Option{
			{ID: "A", Value: "socratic", Label: "Socratic", Description: "Ask me questions and guide me to discover the answer myself."},
			{ID: "B", Value: "guided", Label: "Guided", Description: "Explain the reasoning, then let me try or implement it."},
			{ID: "C", Value: "explain_execute", Label: "Explain & Execute", Description: "Explain the important parts, then do the implementation."},
			{ID: "D", Value: "execute", Label: "Execute", Description: "Keep explanations minimal and focus on getting the task done."},
		},
	},
	{
		ID:     "autonomy_level",
		Prompt: "How much should AI do on its own?",
		Options: []Option{
			{ID: "A", Value: "ask_first", Label: "Ask first", Description: "I want approval before significant implementation decisions."},
			{ID: "B", Value: "guided_autonomy", Label: "Guided autonomy", Description: "Make normal implementation decisions, but ask when something important is unclear."},
			{ID: "C", Value: "high_autonomy", Label: "High autonomy", Description: "Execute the task and stop only for meaningful blockers or risky decisions."},
			{ID: "D", Value: "maximum_autonomy", Label: "Maximum autonomy", Description: "Complete the task end-to-end unless explicitly blocked."},
		},
	},
	{
		ID:     "explanation_depth",
		Prompt: "How much do you want AI to explain?",
		Options: []Option{
			{ID: "A", Value: "deep", Label: "Deep", Description: "Explain the reasoning, alternatives, and important details."},
			{ID: "B", Value: "moderate", Label: "Moderate", Description: "Explain important decisions and unfamiliar concepts."},
			{ID: "C", Value: "concise", Label: "Concise", Description: "Tell me what changed and why, without unnecessary detail."},
			{ID: "D", Value: "minimal", Label: "Minimal", Description: "Give me only what I need to continue."},
		},
	},
	{
		ID:     "work_priority",
		Prompt: "What matters most when you're working?",
		Options: []Option{
			{ID: "A", Value: "understanding", Label: "Understanding", Description: "I prefer learning even if the task takes longer."},
			{ID: "B", Value: "balance", Label: "Balance", Description: "I want good understanding without unnecessary slowdown."},
			{ID: "C", Value: "speed", Label: "Speed", Description: "I want the fastest reliable path to completion."},
			{ID: "D", Value: "context_dependent", Label: "Context-dependent", Description: "Switch between learning and speed depending on the task."},
		},
	},
	{
		ID:     "review_style",
		Prompt: "How should AI handle your code review?",
		Options: []Option{
			{ID: "A", Value: "teaching_review", Label: "Teaching review", Description: "Explain mistakes and teach me the underlying patterns."},
			{ID: "B", Value: "engineering_review", Label: "Engineering review", Description: "Focus on correctness, architecture, maintainability, and best practices."},
			{ID: "C", Value: "critical_review", Label: "Critical review", Description: "Actively challenge decisions, complexity, duplication, and possible weaknesses."},
			{ID: "D", Value: "fast_review", Label: "Fast review", Description: "Check the important risks and tell me what must be fixed."},
		},
	},
	{
		ID:     "git_autonomy",
		Prompt: "How should AI handle Git and project changes?",
		Options: []Option{
			{ID: "A", Value: "manual_control", Label: "Manual control", Description: "Give me commands and explanations; I execute them myself."},
			{ID: "B", Value: "guided", Label: "Guided", Description: "Prepare the commands/actions and explain what will happen before important operations."},
			{ID: "C", Value: "autonomous_with_safeguards", Label: "Autonomous with safeguards", Description: "Handle routine Git operations, but require approval for risky actions."},
			{ID: "D", Value: "autonomous", Label: "Autonomous", Description: "Handle the normal Git workflow automatically and stop only for risky or destructive operations."},
		},
	},
	{
		ID:     "uncertainty_policy",
		Prompt: "What should AI do when it is unsure?",
		Options: []Option{
			{ID: "A", Value: "stop_and_ask", Label: "Stop and ask", Description: "Never guess when requirements or intent are unclear."},
			{ID: "B", Value: "investigate_first", Label: "Investigate first", Description: "Inspect the available context and ask only if uncertainty remains."},
			{ID: "C", Value: "safest_reasonable_assumption", Label: "Make the safest reasonable assumption", Description: "Continue when the risk is low and explain the assumption."},
			{ID: "D", Value: "keep_moving", Label: "Keep moving", Description: "Choose the most likely interpretation and continue unless the risk is significant."},
		},
	},
}

func RunOnboarding(in io.Reader, out io.Writer, path string) error {
	if Exists(path) {
		fmt.Fprintf(out, "Developer profile already exists at %s\n", path)
		return nil
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

	selections, err := askQuestions(in, out, Questions, interactive)
	if err != nil {
		return err
	}
	cleanup()
	cleanup = func() {}

	now := time.Now().UTC()
	developerProfile := NewDeveloperProfile(selections, now)
	if err := Save(path, developerProfile); err != nil {
		return err
	}

	writeSummary(out, developerProfile, path)
	return nil
}

func AskQuestions(in io.Reader, out io.Writer, questions []Question) (Preferences, error) {
	return askQuestions(in, out, questions, false)
}

func askQuestions(in io.Reader, out io.Writer, questions []Question, interactive bool) (Preferences, error) {
	reader := bufio.NewReader(in)
	selections := make(Preferences, len(questions))
	style := newStyle(interactive)

	for index, question := range questions {
		option, err := readSelection(reader, out, question, index, len(questions), style)
		if err != nil {
			return nil, err
		}
		selections[question.ID] = Selection{
			OptionID:    option.ID,
			Value:       option.Value,
			Label:       option.Label,
			Description: option.Description,
		}
	}

	return selections, nil
}

func readSelection(reader *bufio.Reader, out io.Writer, question Question, questionIndex int, totalQuestions int, style terminalStyle) (Option, error) {
	selected := 0
	renderQuestion(out, question, questionIndex, totalQuestions, selected, style)

	for {
		key, err := readKey(reader)
		if err != nil {
			return Option{}, err
		}
		switch key {
		case keyUp:
			if selected > 0 {
				selected--
			}
			renderQuestion(out, question, questionIndex, totalQuestions, selected, style)
		case keyDown:
			if selected < len(question.Options)-1 {
				selected++
			}
			renderQuestion(out, question, questionIndex, totalQuestions, selected, style)
		case keyEnter:
			fmt.Fprintln(out)
			return question.Options[selected], nil
		}
	}
}

type key int

const (
	keyUnknown key = iota
	keyUp
	keyDown
	keyEnter
)

func readKey(reader *bufio.Reader) (key, error) {
	input, err := reader.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return keyUnknown, errors.New("input ended before onboarding completed")
		}
		return keyUnknown, err
	}

	switch input {
	case '\r', '\n':
		return keyEnter, nil
	case 'k', 'K':
		return keyUp, nil
	case 'j', 'J':
		return keyDown, nil
	case 0x1b:
		next, err := reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		if next != '[' {
			return keyUnknown, nil
		}
		direction, err := reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		switch direction {
		case 'A':
			return keyUp, nil
		case 'B':
			return keyDown, nil
		}
	}

	return keyUnknown, nil
}

func renderQuestion(out io.Writer, question Question, questionIndex int, totalQuestions int, selected int, style terminalStyle) {
	if style.interactive {
		resetScreen(out)
	}

	fmt.Fprintln(out, style.blue("AllMyAgents setup"))
	fmt.Fprintln(out, style.blue("Developer Profile onboarding"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.dim(fmt.Sprintf("Question %d of %d  %s", questionIndex+1, totalQuestions, progressBar(questionIndex+1, totalQuestions))))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.blue(question.Prompt))
	fmt.Fprintln(out)

	for index, option := range question.Options {
		line := "  " + option.Label
		if index == selected {
			line = style.yellow("> " + option.Label)
		}
		fmt.Fprintln(out, line)
		if option.Description != "" {
			fmt.Fprintln(out, style.dim("    "+option.Description))
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, style.dim("Use Up/Down arrows to move. Press Enter to select."))
}

func progressBar(current int, total int) string {
	if total <= 0 {
		return ""
	}

	const width = 8
	filled := current * width / total
	if filled < 1 {
		filled = 1
	}

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func writeSummary(out io.Writer, developerProfile DeveloperProfile, path string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Developer Profile created")
	fmt.Fprintf(out, "Saved to: %s\n", path)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Summary:")

	for _, question := range Questions {
		selection, ok := developerProfile.Preferences[question.ID]
		if !ok {
			continue
		}
		fmt.Fprintf(out, "- %s: %s\n", summaryLabel(question.ID), selection.Label)
	}
}

func summaryLabel(questionID string) string {
	switch questionID {
	case "experience_level":
		return "Experience"
	case "learning_method":
		return "Learning"
	case "autonomy_level":
		return "Autonomy"
	case "explanation_depth":
		return "Explanation"
	case "work_priority":
		return "Priority"
	case "review_style":
		return "Review"
	case "git_autonomy":
		return "Git"
	case "uncertainty_policy":
		return "Uncertainty"
	default:
		return questionID
	}
}

type terminalStyle struct {
	interactive bool
}

func newStyle(interactive bool) terminalStyle {
	return terminalStyle{interactive: interactive}
}

func (style terminalStyle) blue(text string) string {
	return style.apply("\033[34m", text)
}

func (style terminalStyle) yellow(text string) string {
	return style.apply("\033[33m", text)
}

func (style terminalStyle) dim(text string) string {
	return style.apply("\033[2m", text)
}

func (style terminalStyle) apply(code string, text string) string {
	if !style.interactive {
		return text
	}
	return code + text + "\033[0m"
}

func isInteractive(in io.Reader, out io.Writer) bool {
	inputFile, inputOK := in.(*os.File)
	outputFile, outputOK := out.(*os.File)
	return inputOK && outputOK && isTerminal(inputFile) && isTerminal(outputFile)
}

func enterAlternateScreen(out io.Writer) {
	fmt.Fprint(out, "\033[?1049h\033[H")
}

func exitAlternateScreen(out io.Writer) {
	fmt.Fprint(out, "\033[?1049l")
}

func resetScreen(out io.Writer) {
	fmt.Fprint(out, "\033[H\033[J")
}
