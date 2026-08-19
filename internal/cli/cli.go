package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

const usage = `Usage:
  allmyagents init
  allmyagents profile
  allmyagents override
  allmyagents override show
  allmyagents override clear [preference-id]
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
