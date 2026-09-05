package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/adapters"
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

// runInitProject activates AllMyAgents' agent integrations for the current
// project: Claude Code, Codex, GitHub Copilot (VS Code), Gemini CLI, and
// Google Antigravity, each via its own adapter (see internal/adapters).
// Every adapter runs independently — one failing or being skipped never
// stops the others — and none of them ever modify a file already tracked
// by Git, so a project's real, shared instructions are never overwritten
// or mixed with AllMyAgents' personal, local-only context. It keeps
// .allmyagents/ out of the project's Git status and is safe to run
// repeatedly or outside a Git repository.
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

	effectiveContext, buildErr := allmyagentscontext.Build(projectDir)
	rendered := ""
	switch {
	case buildErr == nil:
		rendered = allmyagentscontext.Render(effectiveContext)
	case errors.Is(buildErr, fs.ErrNotExist):
		// No Developer Profile yet: a legitimate, non-fatal state. Adapters
		// that need static content report their own "skipped" reason for
		// this; leaving rendered empty is how they detect it.
	default:
		// A real failure — malformed Developer Profile or Session Override
		// JSON, an unresolvable home directory, a permissions error, etc.
		// This must not be silently treated as "no profile yet": report it
		// and stop before running any adapter, rather than letting five
		// adapters report a misleading "no Developer Profile" skip.
		err := fmt.Errorf("could not build Effective Developer Context: %w", buildErr)
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}

	results := adapters.RunAll(projectDir, rendered)

	failed := false
	for _, result := range results {
		fmt.Fprintf(out, "%s: %s — %s\n", result.Agent, result.Status, result.Detail)
		if result.Status == adapters.StatusError {
			failed = true
		}
	}

	if failed {
		err := fmt.Errorf("one or more agent integrations failed; see the errors above")
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return err
	}
	return nil
}
