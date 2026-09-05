# AllMyAgents — Architecture

## 1. Architecture Goal

The V0 architecture must prove the core product hypothesis with the smallest reasonable system.

The architecture should be:

- simple,
- local-first,
- maintainable,
- easy to understand,
- easy to change,
- independent from a single AI provider.

Avoid introducing infrastructure that is not required to validate the product.

## 2. V0 Architecture

The V0 architecture follows this flow:

Developer  
↓  
AllMyAgents CLI  
↓  
Developer Profile + Project State + Task Context + Workflow + Verification  
↓  
Context Builder  
↓  
Agent Adapters (one per supported AI executor)  
↓  
AI Executor  
↓  
Repository

V0 delivers the Context Builder's output to five AI executors — Claude Code,
Codex CLI, GitHub Copilot (VS Code), Gemini CLI, and Google Antigravity —
each through its own adapter (§7). This is context delivery, not the full
execution/verification loop described in §8: no adapter today runs an
executor, tests its output, or updates project state on its behalf.

## 3. Core Components

### Developer Profile

Stores persistent developer-specific information such as:

- experience level,
- learning preferences,
- interaction preferences,
- autonomy preferences,
- coding preferences.

This information belongs to the developer and must remain separate from project-specific knowledge.

### Project State

Stores useful project-specific knowledge such as:

- project requirements,
- architecture,
- current state,
- important decisions,
- constraints,
- confirmed lessons.

Project state must be separated from the developer profile.

### Task Context

Represents the context required for the current task.

It may combine:

- developer profile,
- project state,
- task requirements,
- relevant project files,
- applicable repository rules.

The task context should contain only information relevant to the current operation.

### Context Builder

The Context Builder prepares the information required by the executor.

Its responsibility is to transform persistent knowledge and current task information into an appropriate execution context.

It should avoid blindly sending all stored knowledge to the AI.

The objective is:

> Relevant context, not maximum context.

### AI Executor

The executor performs the actual AI-assisted coding work.

V0 delivers context to five executors — Claude Code, Codex CLI, GitHub
Copilot (VS Code), Gemini CLI, and Google Antigravity — through the adapter
architecture in §7. Each is independently replaceable: removing or
disabling any one adapter does not affect the others or the core tool.

AllMyAgents should not depend on the internal behavior of one specific model.

### Verification

Verification checks the result against:

- task requirements,
- project requirements,
- architecture,
- repository rules,
- tests,
- simplicity and maintainability.

Verification should be separate from implementation so that the system can evaluate work after execution.

## 4. Storage Model

V0 is local-first.

The initial model separates:

User Machine

- Developer Intelligence
  - ~/.allmyagents/

- Project Intelligence
  - <project>/.allmyagents/

Developer intelligence is portable across projects.

Project intelligence belongs to the relevant project.

Neither should be committed to a public repository by default.

## 5. CLI Direction

V0 is CLI-first.

The implemented V0 commands are:

- `allmyagents init` — first-run Developer Profile onboarding
- `allmyagents profile` — review/edit individual profile preferences
- `allmyagents override` (and `show` / `clear`) — temporary project-scoped preference overrides
- `allmyagents context` — print the deterministic Effective Developer Context
- `allmyagents init-project` — configure every supported agent adapter for the current project

The CLI is an interface to the underlying intelligence layer, not the intelligence layer itself.

### V0 Project Structure

V0 starts with a minimal Go project structure.

New packages and directories should be introduced only when they are justified by real implementation needs.

Do not create abstractions, packages, or directories only to follow a predefined "production" structure.

The project structure should evolve from actual requirements while preserving simplicity and maintainability.

## 6. Future Local Service

A future version may introduce a local service:

CLI / IDE / Codex / Claude Code / Other Agents  
↓  
Local AllMyAgents Service  
↓  
Core Intelligence

The service should only be introduced when the V0 workflow demonstrates a real need for persistent cross-tool communication.

## 7. V0 Agent Adapter Architecture

V0 implements exactly this, in `internal/adapters`:

AllMyAgents Core (`internal/context`: Developer Profile + Session Override → Effective Developer Context)  
↓  
`internal/adapters` (one adapter per agent, common `Configure(projectDir, rendered) Result` interface)  
├── Claude Code — SessionStart hook (`.claude/settings.local.json`), dynamic, no static file  
├── Codex — `AGENTS.md`  
├── GitHub Copilot (VS Code) — `.github/copilot-instructions.md`  
├── Gemini CLI — `GEMINI.local.md` + `.gemini/settings.json` (`context.fileName`)  
└── Google Antigravity — `.agents/rules/allmyagents-context.md` (experimental — see `docs/integrations.md`)

Two invariants hold for every adapter:

1. **Isolation.** `init-project` runs every adapter independently (a panic is recovered into an error result); one adapter failing or being skipped never prevents, blocks, or corrupts the others. A `skipped` result is non-fatal; an `error` result makes `init-project` report overall failure, but only after every adapter has still run.
2. **Never touch a tracked file.** Codex/Copilot/Gemini/Antigravity's mechanisms are working-tree files that a team conventionally commits for its own real, shared instructions. Before writing any of them, the adapter checks `profile.IsTracked` (via `git ls-files`); if the path is already tracked, it skips rather than overwriting, appending to, or otherwise mixing AllMyAgents' personal context into a project's shared instructions. Every file it does create is marked as AllMyAgents-generated and immediately added to `.git/info/exclude`.

This is context delivery, not multi-agent orchestration: each adapter is independent and none of them coordinate with each other or with a live execution loop. That remains explicitly outside V0 — see `docs/PRD.md` §10.

## 8. Context Flow

The intended V0 flow is:

Developer Profile  
↓  
Project State  
↓  
Task Requirements  
↓  
Repository Rules  
↓  
Context Builder  
↓  
AI Executor  
↓  
Implementation  
↓  
Verification  
↓  
Human Approval  
↓  
Project State Update

The system should not automatically persist every output.

Only confirmed and useful knowledge should become persistent project state.

## 9. Design Principles

### Local-first

Sensitive developer intelligence remains local by default.

### Agent-agnostic

The intelligence layer must not be tightly coupled to one AI provider.

### Context efficiency

The system should provide relevant context instead of repeatedly sending everything.

### Separation of concerns

Developer identity, project knowledge, task context, execution, and verification should remain conceptually separate.

### Human control

Important actions remain subject to developer approval.

### KISS

Prefer the simplest architecture that proves the product hypothesis.

Do not introduce databases, services, graphs, vector stores, or cloud infrastructure without a demonstrated requirement.

## 10. Architecture Decision Rule

When choosing between two valid implementations:

> Prefer the simpler implementation that keeps future replacement possible.

Architecture should evolve from evidence gathered during real usage rather than assumptions about the final product.