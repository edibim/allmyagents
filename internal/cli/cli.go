package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	allmyagentscontext "github.com/n7ptd2xr8c-cell/allmyagents/internal/context"
	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

const usage = `Usage:
  allmyagents init
  allmyagents profile
  allmyagents override
  allmyagents override show
  allmyagents override clear [preference-id]
  allmyagents context
  allmyagents init-project
`

func Run(args []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	if len(args) < 1 {
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("expected one command")
	}

	switch args[0] {
	case "init":
		if len(args) != 1 {
			fmt.Fprint(errOut, usage)
			return fmt.Errorf("expected one command")
		}
		path, err := profile.DefaultPath()
		if err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		if err := profile.RunOnboarding(in, out, path); err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		return nil
	case "profile":
		if len(args) != 1 {
			fmt.Fprint(errOut, usage)
			return fmt.Errorf("expected one command")
		}
		path, err := profile.DefaultPath()
		if err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		if err := profile.RunEditor(in, out, path); err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		return nil
	case "override":
		return runOverride(args[1:], in, out, errOut)
	case "context":
		return runContext(args[1:], out, errOut)
	case "init-project":
		return runInitProject(args[1:], out, errOut)
	default:
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runOverride(rest []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	profilePath, err := profile.DefaultPath()
	if err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}
	overridePath := profile.OverridePath(projectDir)

	if err := profile.EnsureProjectStateExcluded(projectDir); err != nil {
		fmt.Fprintf(errOut, "Warning: could not add .allmyagents/ to git exclude: %v\n", err)
	}

	if len(rest) == 0 {
		if err := profile.RunOverrideEditor(in, out, profilePath, overridePath); err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		return nil
	}

	switch rest[0] {
	case "show":
		if len(rest) != 1 {
			fmt.Fprint(errOut, usage)
			return fmt.Errorf("unexpected arguments after show")
		}
		if err := profile.ShowOverride(out, profilePath, overridePath); err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		return nil
	case "clear":
		if len(rest) > 2 {
			fmt.Fprint(errOut, usage)
			return fmt.Errorf("unexpected arguments after clear")
		}
		preferenceID := ""
		if len(rest) == 2 {
			preferenceID = rest[1]
		}
		if err := profile.ClearOverride(overridePath, preferenceID); err != nil {
			fmt.Fprintf(errOut, "Error: %v\n", err)
			return err
		}
		if preferenceID == "" {
			fmt.Fprintf(out, "Session override cleared at %s\n", overridePath)
		} else {
			fmt.Fprintf(out, "Cleared override for %s at %s\n", preferenceID, overridePath)
		}
		return nil
	default:
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("unknown override command %q", rest[0])
	}
}

const sessionStartHookFlag = "--session-start-hook"

// runContext prints the Effective Context (Developer Profile overlaid with
// any active Session Override) for the current project. With
// --session-start-hook it wraps the same content in the
// hookSpecificOutput.additionalContext JSON form Claude Code's SessionStart
// hook expects, and degrades gracefully instead of failing when no
// Developer Profile exists yet, since that path runs unattended on every
// session start.
func runContext(rest []string, out io.Writer, errOut io.Writer) error {
	sessionStartHook := false
	for _, arg := range rest {
		if arg != sessionStartHookFlag {
			fmt.Fprint(errOut, usage)
			return fmt.Errorf("unknown context flag %q", arg)
		}
		sessionStartHook = true
	}

	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}

	effectiveContext, buildErr := allmyagentscontext.Build(projectDir)
	if buildErr != nil {
		if sessionStartHook {
			return writeSessionStartHookOutput(out, "AllMyAgents: no Developer Profile found yet. Run `allmyagents init` to set one up.")
		}
		fmt.Fprintf(errOut, "Error: %v\n", buildErr)
		return buildErr
	}

	rendered := allmyagentscontext.Render(effectiveContext)
	if sessionStartHook {
		return writeSessionStartHookOutput(out, rendered)
	}
	fmt.Fprint(out, rendered)
	return nil
}

type sessionStartHookOutput struct {
	HookSpecificOutput sessionStartHookSpecificOutput `json:"hookSpecificOutput"`
}

type sessionStartHookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

func writeSessionStartHookOutput(out io.Writer, additionalContext string) error {
	payload := sessionStartHookOutput{
		HookSpecificOutput: sessionStartHookSpecificOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: additionalContext,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}

const claudeSettingsGitPattern = "**/" + claudeSettingsDirName + "/" + claudeSettingsFileName

// runInitProject activates AllMyAgents' Claude Code integration for the
// current project: it wires up a SessionStart hook in
// .claude/settings.local.json so Claude Code receives the Developer
// Profile automatically, and keeps both .allmyagents/ and that settings
// file out of the project's Git status. It never touches a tracked file,
// and is safe to run repeatedly or outside a Git repository.
func runInitProject(rest []string, out io.Writer, errOut io.Writer) error {
	if len(rest) != 0 {
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("unexpected arguments after init-project")
	}

	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}

	if err := profile.EnsureProjectStateExcluded(projectDir); err != nil {
		fmt.Fprintf(errOut, "Warning: could not add .allmyagents/ to git exclude: %v\n", err)
	}

	changed, settingsPath, err := ensureClaudeSessionStartHook(projectDir)
	if err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}

	if err := profile.EnsureGitExcluded(projectDir, claudeSettingsGitPattern); err != nil {
		fmt.Fprintf(errOut, "Warning: could not add %s to git exclude: %v\n", claudeSettingsFileName, err)
	}

	if changed {
		fmt.Fprintf(out, "Added AllMyAgents SessionStart hook to %s\n", settingsPath)
	} else {
		fmt.Fprintf(out, "AllMyAgents SessionStart hook already present at %s\n", settingsPath)
	}
	fmt.Fprintln(out, "Claude Code will now receive your Developer Profile automatically when you start a session in this project.")
	return nil
}
