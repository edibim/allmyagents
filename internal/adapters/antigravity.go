package adapters

import (
	"fmt"
	"strings"
)

// Google Antigravity's own documentation
// (https://antigravity.google/docs/rules-workflows/) describes workspace
// "Rules" living under .agents/rules/ (with .agent/rules/ kept as a legacy
// fallback) as its repository-level context mechanism — not AGENTS.md,
// despite several third-party guides claiming otherwise. A Rule is a
// Markdown file with frontmatter selecting an activation mode; "Always On"
// is the closest match to "load this every session" available today.
//
// Antigravity's docs are thin (a single page) and its frontmatter schema
// isn't spelled out in as much detail as the other agents here, so this
// integration is treated as experimental: the generated file follows the
// documented directory and the most plausible frontmatter, but should be
// confirmed against the Antigravity UI's Rules view (see
// docs/integrations.md).
const (
	antigravityRulesFile       = ".agents/rules/allmyagents-context.md"
	antigravityRuleFrontMatter = "---\ntrigger: always_on\n---\n\n"
)

// Antigravity configures a Google Antigravity workspace Rule.
type Antigravity struct{}

func (Antigravity) Name() string { return "antigravity" }

func (Antigravity) Configure(projectDir string, rendered string) Result {
	const agent = "antigravity"

	if strings.TrimSpace(rendered) == "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: noProfileYetDetail}
	}

	content := antigravityRuleFrontMatter + buildContent(rendered)
	skipReason, note, err := writeManagedFile(projectDir, antigravityRulesFile, content)
	if err != nil {
		return Result{Agent: agent, Status: StatusError, Detail: err.Error()}
	}
	if skipReason != "" {
		return Result{Agent: agent, Status: StatusSkipped, Detail: skipReason}
	}
	return Result{Agent: agent, Status: StatusConfigured, Detail: fmt.Sprintf(
		"wrote %s as an Always On workspace rule (%s; experimental — verify in Antigravity's Rules view)", antigravityRulesFile, note,
	)}
}
