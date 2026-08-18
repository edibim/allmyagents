package cli

import (
	"fmt"
	"io"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

const usage = `Usage:
  allmyagents init
`

func Run(args []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	if len(args) != 1 {
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("expected one command")
	}

	switch args[0] {
	case "init":
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
	default:
		fmt.Fprint(errOut, usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}
