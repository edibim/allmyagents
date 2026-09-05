package adapters

import (
	"fmt"
	"strings"
)

// Codex CLI auto-discovers an AGENTS.md file at the repository root (and in
// nested directories) with no configuration required — an official,
// documented mechanism (see
// https://developers.openai.com/codex/guides/agents-md). Codex has no
// dynamic hook system, so AllMyAgents writes a static, git-excluded
// AGENTS.md carrying the current Effective Developer Context. Because
// AGENTS.md is also the conventional home for a team's real, shared
// project instructions, this adapter only ever manages the file when it
// isn't already tracked by git (see writeManagedFile) — it never appends
// to or overwrites a real team AGENTS.md.
const codexAgentsFile = "AGENTS.md"

// Codex configures Codex CLI's AGENTS.md discovery.
type Codex struct{}

func (Codex) Name() string { return "codex" }

func (Codex) Configure(projectDir string, rendered string) Result {
	const agent = "codex"

	if strings.TrimSpace(rendered) == "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: noProfileYetDetail}
	}

	skipReason, note, err := writeManagedFile(projectDir, codexAgentsFile, buildContent(rendered))
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	if skipReason != "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: skipReason}
	}
	return Result{Agent: agent, Status: StatusConfigured, Detail: fmt.Sprintf("wrote %s with your effective developer context (%s)", codexAgentsFile, note)}
}
