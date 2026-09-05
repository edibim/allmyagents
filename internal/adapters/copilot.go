package adapters

import (
	"fmt"
	"strings"
)

// GitHub Copilot in VS Code auto-discovers .github/copilot-instructions.md
// at the workspace root with no settings required — a stable, documented
// mechanism (see
// https://code.visualstudio.com/docs/agent-customization/custom-instructions).
// That path is conventionally a team's real, shared repository
// instructions, committed to version control, so this adapter only ever
// manages it when it isn't already tracked by git (see writeManagedFile) —
// it never appends to or overwrites a team's real instructions file.
const copilotInstructionsFile = ".github/copilot-instructions.md"

// Copilot configures GitHub Copilot's repository custom instructions file.
type Copilot struct{}

func (Copilot) Name() string { return "copilot-vscode" }

func (Copilot) Configure(projectDir string, rendered string) Result {
	const agent = "copilot-vscode"

	if strings.TrimSpace(rendered) == "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: noProfileYetDetail}
	}

	skipReason, note, err := writeManagedFile(projectDir, copilotInstructionsFile, buildContent(rendered))
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	if skipReason != "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: skipReason}
	}
	return Result{Agent: agent, Status: StatusConfigured, Detail: fmt.Sprintf("wrote %s with your effective developer context (%s)", copilotInstructionsFile, note)}
}
