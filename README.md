# AllMyAgents

A local-first personal engineering intelligence layer for AI-assisted software development.

### Your developer identity shouldn't belong to one AI

AI assistants are getting better at remembering you. But that memory usually stays inside the assistant that created it.

**AllMyAgents takes a different approach: your developer identity belongs to you.**

Define how you work once, keep it locally, and carry that context across the AI coding agents you use — Claude Code, Codex, Copilot, Gemini, and Antigravity.

**AI memory makes the AI remember you. AllMyAgents makes your engineering identity portable.**

AllMyAgents is not a coding agent. It is the layer that sits around one — preserving who you are, how you work, and what your project already knows, so you stop re-explaining it every session.

## What is AllMyAgents?

AI coding agents are good at writing code. They are bad at remembering you.

...

## What is AllMyAgents?

AI coding agents are good at writing code. They are bad at remembering you.

Every new session, developers re-explain their experience level, how they like to learn, how much autonomy to give the AI, what the project is, and what has already been decided. AllMyAgents exists to hold that context locally and make it available to the AI-assisted workflow automatically — without becoming another coding agent itself.

## The Problem

AI-assisted development creates friction that has nothing to do with the AI's coding ability:

- Personal preferences and learning style get re-typed every session.
- Project context has to be rediscovered or manually pasted in.
- Decisions and the reasoning behind them get lost over time.
- Switching between AI tools means moving context by hand.
- AI can produce working code without the developer actually understanding it.
- Developers still have to manually check whether generated code follows project rules and architecture.

This costs time, tokens, and developer control.

## The Idea

If useful developer and project knowledge is maintained locally and supplied automatically to the right workflow, developers can use AI agents without reconstructing their context every time — while staying in control of the outcome.

AllMyAgents is built around a specific tradeoff:

> Relevant context, not maximum context.

The system should figure out what's actually relevant to the current task rather than dumping everything it knows at the AI. And whatever the AI does with that context, the developer stays the one who approves, understands, and owns the result.

## How It's Meant to Work

AllMyAgents is designed around one core engineering loop:

```
Understand → Plan → Implement → Test → Verify → Explain → Human Approval → Update Project State
```

The AI executor implements within boundaries set by the developer's profile, the project's architecture, and its own requirements — then the work is tested, verified against those requirements, explained in human-readable terms, and only persisted once approved.

**This is the target design, not what's shipped today.** The currently shipped V0 implements the *context* half of this loop — Developer Profile, Session Overrides, and Effective Developer Context delivered automatically to five supported agents. Test, Verify, Explain, Human Approval, and Update Project State describe where this is headed, not something any agent integration runs today. See "Current Status: V0" below for exactly what's implemented.

AllMyAgents is agent-agnostic: it computes one **Effective Developer Context** (your Developer Profile, optionally overridden for the current project) and hands it to whichever AI coding agent you actually use, via that agent's own documented mechanism. V0 supports **Claude Code, Codex CLI, GitHub Copilot in VS Code, Gemini CLI, and Google Antigravity** — see [`docs/integrations.md`](docs/integrations.md) for exactly how each one works, what's verified versus experimental, and how to check it's working. The executor is intentionally kept replaceable — AllMyAgents provides the context and workflow intelligence around it, not the execution engine itself.

## Local-First by Design

Developer intelligence and project intelligence are deliberately kept separate:

```
~/.allmyagents/                    <project>/.allmyagents/
  developer profile                  project-specific state
  reusable across every project      belongs to this project only
```

Everything lives on the developer's machine by default and is not committed to public repositories. A future version may add optional encrypted cloud sync, but that is explicitly out of scope for V0.

This also governs how `init-project` behaves: some agents (Codex, Copilot, Gemini CLI, Antigravity) only have a working-tree file as their documented way to read local instructions, so AllMyAgents writes one — clearly marked as AllMyAgents-generated, and **never** written to a path your project already tracks in git or a pre-existing file it didn't create itself. Your personal context is never mixed into a repository's real, shared instructions. When a real Git repository is present, the file is also added to `.git/info/exclude` (never `.gitignore`); if there isn't one yet, `init-project` says so plainly rather than claiming protection it can't provide. See [`docs/integrations.md`](docs/integrations.md) for the exact behavior per agent.

## Getting Started

### Requirements

- Go 1.25 or newer

### Build

```bash
git clone https://github.com/edibim/allmyagents.git
cd allmyagents
go build -o allmyagents ./cmd/allmyagents
```

### Usage

```bash
./allmyagents init                            # Starts the interactive Developer Profile onboarding
./allmyagents profile                         # review and edit individual profile preferences
./allmyagents override                        # Select a preference and temporarily change it for the current project.
./allmyagents override show                   # see effective preferences (profile + active overrides)
./allmyagents override clear [preference-id]  # clear one override, or all if no id is given
./allmyagents context                         # print the Effective Developer Context for the current project
./allmyagents init-project                    # wire up every supported agent for the current project
```

Your Developer Profile lives at `~/.allmyagents/developer-profile.json` and follows you across every project. Session overrides live inside the project you run the command from and are never written back into your profile.

### First setup, end to end

```bash
./allmyagents init            # once per machine: answer 8 questions to create your Developer Profile
cd path/to/your/project       # any repository — yours doesn't need to know about AllMyAgents
allmyagents init-project      # once per project: wires up every supported agent it can
```

`init-project` prints one line per agent — `configured`, `skipped: <reason>`, or `error: <reason>` — and never touches a file your project already tracks in git. A `skipped` result is normal and expected (for example, when your repo already has its own real, shared `AGENTS.md`); an `error` means the command exits with a non-zero status so you notice. Re-running `init-project` at any time is safe: it's idempotent and doubles as a status check.

Then just start your agent of choice in that project directory — Claude Code, Codex CLI, GitHub Copilot in VS Code, Gemini CLI, or Google Antigravity — and it should already know how you like to work. See [`docs/integrations.md`](docs/integrations.md) for the exact mechanism and verification steps per agent, and what to do if `init-project` skipped one.

## Current Status: V0

AllMyAgents is early. V0 is not a finished product — it's a working slice built to validate a hypothesis: that locally maintained context can meaningfully reduce AI-assisted development friction. It is being validated through real, dogfooded engineering work before going any further.

**What "V0" means here, unambiguously:** V0 is the concrete, shipped slice listed under "Implemented" directly below — Developer Profile, Session Overrides, Effective Developer Context, and agent integrations. It is *not* the full loop described above under "How It's Meant to Work"; that remains the target design this is built toward, tracked under "V0 Target."

### Implemented

The developer profile foundation is built and working today:

- **First-run onboarding** — eight structured questions (experience, learning style, autonomy, explanation depth, work priority, review style, Git autonomy, uncertainty handling), answered by selecting an option, never by writing a prompt.
- **Structured Developer Profile** — preferences are stored as data, not a giant static prompt, and persisted locally.
- **Interactive profile editing** — change one preference at a time without redoing onboarding.
- **Temporary session/task overrides** — override individual preferences for the current project without touching the persistent profile. Overrides are sparse (only what you explicitly change), scoped to the project, and stay in effect until you explicitly clear them.
- **Effective Developer Context** — `allmyagents context` deterministically resolves your Developer Profile and any active Session Override into one Markdown document, with no filesystem paths or raw internals in the output.
- **Agent integrations** — `allmyagents init-project` wires up Claude Code, Codex CLI, GitHub Copilot (VS Code), Gemini CLI, and Google Antigravity for the current project, each via its own adapter and its agent's real documented mechanism. See [`docs/integrations.md`](docs/integrations.md) for per-agent status (verified vs. configuration-verified vs. experimental).

### V0 Target

Still to be built, per the product requirements:

- Task context assembly beyond the Developer Profile (e.g. pulling in relevant project/task-specific information)
- Automated testing and technical verification of the AI executor's output
- Human-readable handoff / learning report
- Human approval gate
- Confirmed project-state updates

## What V0 Is Not

Explicitly out of scope for this version:

- Cloud synchronization or a SaaS backend
- Subscriptions or payments
- Multi-agent orchestration (agents coordinating with each other — each supported agent independently receiving your context is not the same thing)
- Fully autonomous development
- Automatic Git push by default
- A large graphical UI

These may be reconsidered later, but only once V0 proves the core idea.

## Product Principle

> **Remove unnecessary friction from the developer's interaction with AI without removing the developer from the engineering process.**

The goal isn't to replace the developer. It's to make the developer more capable, informed, and efficient while staying the owner of every engineering decision.

## Validation

The first user is the project's own creator. AllMyAgents is being used on real development work, and the plan is straightforward: build, use it for real, measure whether it actually reduces friction, and adjust — including reconsidering the approach entirely if the evidence doesn't support it.

## Documentation

- [`docs/PRD.md`](docs/PRD.md) — product requirements and V0 scope
- [`docs/architecture.md`](docs/architecture.md) — architecture decisions and system design
- [`docs/integrations.md`](docs/integrations.md) — per-agent integration mechanism, setup, verification, and limitations
- [`docs/AGENTS.md`](docs/AGENTS.md) — engineering rules for AI-assisted development in this repository
