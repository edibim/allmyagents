# AllMyAgents

> Personal engineering intelligence for AI-assisted software development.

AllMyAgents is a local-first intelligence layer designed to help developers work with AI coding agents without repeatedly rebuilding their personal and project context.

The goal is simple:

> Use different AI agents without losing who you are, how you work, what you are building, or what has already been decided.

## Why AllMyAgents?

AI coding tools can write code quickly, but developers still spend time:

- repeating personal preferences and learning instructions,
- reconstructing project context,
- moving context between AI tools,
- reviewing whether generated code follows project rules,
- understanding what the AI actually changed,
- documenting decisions and updating project state.

AllMyAgents aims to remove that unnecessary friction while keeping the developer in control.

## Current Status

🚧 **Early-stage / V0 development**

This repository is currently being built and validated through real development work.

The first validation target is a real Milestone 3 coding task.

The product direction may change based on evidence from real usage.

## V0 Goal

The first MVP aims to demonstrate the complete engineering workflow:

Onboarding
→ Developer Profile
→ Project Context
→ Task Context
→ AI Execution
→ Tests
→ Verification
→ Learning / Handoff Report
→ Human Approval
→ Project State Update

The first execution integration will use Codex.

AllMyAgents is intended to remain agent-agnostic, so other AI coding agents can be supported later.

## Core Principles

### Local-first

Personal developer intelligence is stored locally by default.

Future versions may provide optional encrypted cloud synchronization.

### Developer-controlled

The system should reduce repetitive work without removing the developer from the engineering process.

### Learning-aware

Developers can configure how they want AI to teach, explain, and assist them.

Possible interaction modes may include:

- learning-focused,
- balanced,
- fast,
- execution-focused.

The exact interaction model is still being validated.

### Persistent engineering knowledge

AllMyAgents should preserve useful structured knowledge rather than simply storing entire conversations.

This may include:

- developer preferences,
- learning state,
- project requirements,
- architecture,
- decisions,
- current project state,
- useful lessons.

## Architecture Direction

The intended high-level architecture is:

Developer
    ↓
AllMyAgents
    ├── Developer Profile
    ├── Project Intelligence
    ├── Task Context
    ├── Workflow / Autonomy
    ├── Verification
    └── State Management
    ↓
AI Executor
    ↓
Software Project

The intelligence layer belongs to AllMyAgents.

The AI executor can change.

## Repository

The project is currently structured as a Go application.

Important documents:

- `PRD.md` — product requirements and V0 scope
- `architecture.md` — architecture decisions and system design
- `AGENTS.md` — engineering rules for AI-assisted development

Additional project documentation will be added as the architecture evolves.

## Development Workflow

Development follows:

feature branch
→ review
→ dev
→ validation
→ main
→ release

The `main` branch represents stable software.

The `dev` branch is used for integration.

Feature work is developed in branches such as:

- `feat/developer-profile`
- `feat/project-context`
- `feat/codex-integration`

## Project Philosophy

AllMyAgents is being built with a simple principle:

> **AI should remove engineering friction, not remove the engineer.**

The product is intentionally being developed through real-world dogfooding before broader validation.