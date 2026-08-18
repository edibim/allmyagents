# AllMyAgents --- Product Requirements Document

**Status:** Draft V0\
**Repository:** `allmyagents`\
**Target:** V0 MVP

------------------------------------------------------------------------

## 1. Product Summary

AllMyAgents is a local-first personal engineering intelligence layer for
AI-assisted software development.

It is designed to preserve useful developer and project context across
AI coding sessions and different AI agents, reducing the need for
developers to repeatedly reconstruct who they are, how they work, what
they are learning, what they are building, and what has already been
decided.

The product is not itself the coding agent. It provides the
intelligence, context, policies, and workflow around the agent.

------------------------------------------------------------------------

## 2. Problem

AI-assisted development currently creates several forms of friction:

-   Developers repeatedly provide personal preferences, learning style,
    and working conventions to AI.
-   Project context must be rediscovered or manually transferred between
    sessions and tools.
-   Decisions and their reasoning can be lost over time.
-   Developers using multiple AI agents may need to move context between
    them manually.
-   AI can produce working code without ensuring that the developer
    understands what was implemented.
-   Developers must spend time reviewing whether AI-generated work
    follows project requirements, architecture, coding rules, and
    simplicity constraints.

This creates unnecessary token usage, repeated prompting, context loss,
and reduced developer control.

------------------------------------------------------------------------

## 3. Product Hypothesis

If useful developer and project knowledge can be maintained locally and
supplied automatically to the appropriate AI workflow, then developers
can use AI agents more efficiently without repeatedly reconstructing
their context.

The system should reduce friction while preserving:

1.  developer control,
2.  developer understanding,
3.  project consistency,
4.  useful long-term engineering knowledge.

This is a hypothesis to be validated through real usage.

------------------------------------------------------------------------

## 4. Target Users

### Primary V0 users

Learner and junior developers who use AI coding agents and want to
improve productivity while still understanding and learning from the
work.

### Broader target

The architecture must support developers at different experience levels.

The behavior should be configurable rather than hardcoded around one
user type.

------------------------------------------------------------------------

## 5. Core Product Promise

A developer should be able to use different AI agents without losing
their personal, learning, or project context.

The system should know:

-   who the developer is,
-   how the developer prefers to work and learn,
-   what the developer is building,
-   what has already been decided,
-   what the current task requires,
-   and what knowledge is relevant to the current task.

The developer should not need to repeatedly reconstruct this context
manually.

------------------------------------------------------------------------

## 6. Developer Profile

On first run, AllMyAgents will guide the developer through a short
onboarding flow.

The profile may include:

-   developer role/experience level,
-   goals,
-   learning preferences,
-   explanation depth,
-   autonomy preferences,
-   coding philosophy,
-   Git workflow preferences,
-   adaptive learning preference.

The profile is structured data, not a large static prompt.

### Learning behavior

A developer may have a persistent learning profile while temporarily
changing the interaction mode.

Example modes:

-   Deep
-   Balanced
-   Fast
-   Execute

A learner may normally use Socratic behavior but switch to Fast mode for
a time-sensitive task.

The system may adapt to demonstrated progress, but the developer remains
in control.

------------------------------------------------------------------------

## 7. Persistent Engineering Knowledge

AllMyAgents must not simply archive conversations.

It should preserve useful structured engineering knowledge, including
where appropriate:

### Developer knowledge

-   preferences,
-   learning style,
-   demonstrated skills,
-   development goals,
-   useful long-term patterns.

### Project knowledge

-   project purpose,
-   PRD requirements,
-   architecture,
-   constraints,
-   current state,
-   relevant repository structure.

### Decision history

-   important decisions,
-   alternatives considered,
-   rejected approaches,
-   reasons for decisions,
-   outcomes and lessons.

Only valuable knowledge should persist.

------------------------------------------------------------------------

## 8. Storage Strategy

AllMyAgents will be **local-first**.

V0 will store the core intelligence locally on the developer's machine.

A future version may provide optional encrypted cloud synchronization.

Cloud infrastructure is explicitly outside the V0 implementation scope.

------------------------------------------------------------------------

## 9. Agent Strategy

AllMyAgents is agent-agnostic.

It is not intended to become another coding agent.

### V0

Codex is the first execution integration.

AllMyAgents provides context and workflow intelligence to the executor.

### Future

The architecture may support additional agents such as:

-   Claude Code,
-   Gemini-based agents,
-   other compatible coding agents.

Multi-agent orchestration is future scope, not a V0 requirement.

------------------------------------------------------------------------

## 10. Core Engineering Workflow

The intended workflow is:

``` text
Understand
    ↓
Plan
    ↓
Implement
    ↓
Test
    ↓
Verify
    ↓
Explain
    ↓
Human Approval
    ↓
Update Project State
```

### Understand

Identify the user's intent and relevant project requirements.

### Plan

Determine the appropriate approach using project rules, architecture,
decisions, and developer preferences.

### Implement

Use the configured AI executor.

### Test

Run the project's relevant tests.

### Verify

Check the implementation against:

-   task requirements,
-   project architecture,
-   coding rules,
-   simplicity/KISS principles,
-   changed files,
-   unnecessary complexity,
-   relevant risks.

### Explain

Produce a human-readable explanation of:

-   what was implemented,
-   why it was implemented,
-   how important functions/components work,
-   important decisions,
-   relevant audit/hand-off information.

### Human Approval

The developer remains the final authority over important changes.

### Update Project State

After approval, only confirmed and valuable project knowledge is
persisted.

------------------------------------------------------------------------

## 11. Autonomy

Autonomy is adaptive rather than a single global switch.

The system should consider:

-   developer preferences,
-   project rules,
-   task context,
-   action risk.

The V0 should support human-controlled execution.

For example, a developer may choose to receive Git commands and execute
them manually rather than allowing the tool to push automatically.

Future versions may support higher levels of autonomy.

------------------------------------------------------------------------

## 12. V0 Scope

The V0 must demonstrate the complete core workflow on a real project.

### Included

-   first-run onboarding,
-   local Developer Profile,
-   project detection,
-   project context,
-   task context,
-   Codex execution integration,
-   testing,
-   technical verification,
-   human-readable learning/handoff report,
-   human approval,
-   project-state update.

### Validation target

The first real validation will be performed by using AllMyAgents on a
real Milestone 3 coding task.

The objective is to determine whether the tool actually reduces repeated
context transfer while preserving developer understanding and control.

------------------------------------------------------------------------

## 13. Non-Goals for V0

The following are explicitly out of scope:

-   multi-agent orchestration,
-   cloud synchronization,
-   SaaS backend,
-   subscriptions/payments,
-   model marketplace,
-   large graphical UI,
-   automatic Git push by default,
-   fully autonomous development,
-   conversation-history archiving,
-   advanced knowledge graph infrastructure unless required by the V0
    implementation.

These may be reconsidered after validation.

------------------------------------------------------------------------

## 14. Success Criteria

V0 is successful only if real usage demonstrates measurable improvement.

We will evaluate:

### Context reduction

Can the developer start a task without repeatedly explaining their
identity, learning preferences, project context, and working
conventions?

### Workflow reduction

Does the developer perform fewer manual context-transfer steps between
ChatGPT, Codex, project files, and other tools?

### Token/time efficiency

Does the workflow reduce unnecessary prompting, repeated explanations,
and context reconstruction?

### Developer control

Can the developer understand and approve important AI actions without
losing control of the process?

### Developer understanding

Can the developer explain the resulting implementation after using the
system?

### Project consistency

Does the implementation remain aligned with project requirements,
architecture, coding rules, and simplicity constraints?

### State continuity

After completing a task, can the next task benefit from the confirmed
knowledge without rebuilding the context manually?

------------------------------------------------------------------------

## 15. Product Principle

The core product principle is:

> **Remove unnecessary friction from the developer's interaction with AI
> without removing the developer from the engineering process.**

The goal is not to replace the developer.

The goal is to make the developer more capable, informed, and efficient
when working with AI.

------------------------------------------------------------------------

## 16. Validation Strategy

The first user is the project creator.

The product will be dogfooded on real development work before being
presented to other developers.

Initial validation sequence:

1.  Build V0.
2.  Use V0 on a real M3 task.
3.  Compare the workflow against the current manual workflow.
4.  Record friction, failures, time savings, context reduction, and
    learning quality.
5.  Improve V0 based on evidence.
6.  Use it on additional real tasks.
7.  Test with a small group of other developers.
8.  Re-evaluate the product hypothesis using real data.

If the product does not produce measurable improvement, the scope and
product thesis must be reconsidered rather than forcing the original
idea.

------------------------------------------------------------------------

## 17. Current Product Direction

The long-term product direction is:

**Personal Engineering Intelligence**

A local-first intelligence layer that becomes more useful over time by
maintaining structured knowledge about the developer, their projects,
their decisions, their progress, and their working preferences, while
remaining independent of any single AI model or coding agent.
