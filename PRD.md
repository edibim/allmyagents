# AllMyAgents — Product Requirements Document

**Status:** Draft V0

**Repository:** `allmyagents`

**Target:** V0 MVP

---

## 1. Product Summary

AllMyAgents is a local-first personal engineering intelligence layer for AI-assisted software development.

It is designed to preserve useful developer and project context across AI coding sessions and different AI agents, reducing the need for developers to repeatedly reconstruct who they are, how they work, what they are learning, what they are building, and what has already been decided.

The product is not itself the coding agent. It provides the intelligence, context, policies, and workflow around the agent.

---

## 2. Problem

AI-assisted development currently creates several forms of friction:

- Developers repeatedly provide personal preferences, learning style, and working conventions to AI.
- Project context must be rediscovered or manually transferred between sessions and tools.
- Decisions and their reasoning can be lost over time.
- Developers using multiple AI agents may need to move context between them manually.
- AI can produce working code without ensuring that the developer understands what was implemented.
- Developers must spend time reviewing whether AI-generated work follows project requirements, architecture, coding rules, and simplicity constraints.

This creates unnecessary token usage, repeated prompting, context loss, and reduced developer control.

AllMyAgents should reduce this friction without requiring the developer to become skilled at prompt engineering.

---

## 3. Product Hypothesis

If useful developer and project knowledge can be maintained locally and supplied automatically to the appropriate AI workflow, then developers can use AI agents more efficiently without repeatedly reconstructing their context.

The system should reduce friction while preserving:

1. developer control,
2. developer understanding,
3. project consistency,
4. useful long-term engineering knowledge.

The system should provide relevant context rather than blindly sending all available context to the AI.

This is a hypothesis to be validated through real usage.

---

## 4. Target Users

### Primary V0 users

Learner and junior developers who use AI coding agents and want to improve productivity while still understanding and learning from the work.

### Broader target

The architecture must support developers at different experience levels.

The behavior must be configurable rather than hardcoded around one user type.

A developer should be able to use the product without needing advanced prompt-engineering skills.

---

## 5. Core Product Promise

A developer should be able to use different AI agents without losing their personal, learning, or project context.

The system should know:

- who the developer is,
- how the developer prefers to work and learn,
- what the developer is building,
- what has already been decided,
- what the current task requires,
- what knowledge is relevant to the current task.

The developer should not need to repeatedly reconstruct this context manually.

Developer context should be reusable across projects.

Project context must remain specific to the relevant project.

---

## 6. Developer Profile

On first run, AllMyAgents will guide the developer through a short onboarding flow.

The onboarding uses predefined questions with selectable answers rather than requiring the developer to write prompts.

The initial V0 onboarding contains eight questions.

### Question 1 — What best describes you?

**A.** I'm learning programming and still need guidance.

**B.** I can build things, but I still need help with unfamiliar problems.

**C.** I can work independently and mainly need AI for speed and review.

**D.** I work independently and want AI mainly as a high-level engineering partner.

This establishes the developer's experience and expected level of guidance.

### Question 2 — How should AI help you learn?

**A. Socratic** — Ask me questions and guide me to discover the answer myself.

**B. Guided** — Explain the reasoning, then let me try or implement it.

**C. Explain & Execute** — Explain the important parts, then do the implementation.

**D. Execute** — Keep explanations minimal and focus on getting the task done.

This establishes the preferred learning method.

### Question 3 — How much should AI do on its own?

**A. Ask first** — I want approval before significant implementation decisions.

**B. Guided autonomy** — Make normal implementation decisions, but ask when something important is unclear.

**C. High autonomy** — Execute the task and stop only for meaningful blockers or risky decisions.

**D. Maximum autonomy** — Complete the task end-to-end unless explicitly blocked.

This establishes the developer's preferred autonomy and approval threshold.

### Question 4 — How much do you want AI to explain?

**A. Deep** — Explain the reasoning, alternatives, and important details.

**B. Moderate** — Explain important decisions and unfamiliar concepts.

**C. Concise** — Tell me what changed and why, without unnecessary detail.

**D. Minimal** — Give me only what I need to continue.

This establishes explanation depth independently from the learning method.

### Question 5 — What matters most when you're working?

**A. Understanding** — I prefer learning even if the task takes longer.

**B. Balance** — I want good understanding without unnecessary slowdown.

**C. Speed** — I want the fastest reliable path to completion.

**D. Context-dependent** — Switch between learning and speed depending on the task.

This establishes the developer's priority when learning and execution speed compete.

### Question 6 — How should AI handle your code review?

**A. Teaching review** — Explain mistakes and teach me the underlying patterns.

**B. Engineering review** — Focus on correctness, architecture, maintainability, and best practices.

**C. Critical review** — Actively challenge decisions, complexity, duplication, and possible weaknesses.

**D. Fast review** — Check the important risks and tell me what must be fixed.

This establishes review style and strictness.

### Question 7 — How should AI handle Git and project changes?

**A. Manual control** — Give me commands and explanations; I execute them myself.

**B. Guided** — Prepare the commands/actions and explain what will happen before important operations.

**C. Autonomous with safeguards** — Handle routine Git operations, but require approval for risky actions.

**D. Autonomous** — Handle the normal Git workflow automatically and stop only for risky or destructive operations.

This establishes operational and Git autonomy.

### Question 8 — What should AI do when it is unsure?

**A. Stop and ask** — Never guess when requirements or intent are unclear.

**B. Investigate first** — Inspect the available context and ask only if uncertainty remains.

**C. Make the safest reasonable assumption** — Continue when the risk is low and explain the assumption.

**D. Keep moving** — Choose the most likely interpretation and continue unless the risk is significant.

This establishes uncertainty tolerance and risk behavior.

### Developer Profile Structure

The profile should represent these preferences as structured data, rather than storing them as one large static prompt.

The profile may contain:

- experience level,
- learning method,
- autonomy level,
- explanation depth,
- work priority,
- review style,
- Git autonomy,
- uncertainty policy.

The profile should remain reusable across different projects and AI agents.

### Profile Editing

The developer must be able to review and change individual profile preferences after onboarding.

Changing one preference must not require repeating the entire onboarding process.

The product should provide a simple way to modify individual preferences.

The exact CLI/interface for editing the profile is an implementation detail and should be determined by the architecture.

### Adaptive Interaction

A developer may have a persistent profile while temporarily changing the interaction behavior for a specific task or session.

For example:

- persistent profile: Socratic,
- current task: Fast.

A temporary task/session preference must not permanently overwrite the developer's default profile unless the developer explicitly chooses to update the profile.

The system should support switching between learning-focused and execution-focused behavior without requiring the developer to recreate their profile.

The system may adapt to demonstrated progress, but the developer remains in control.

---

## 7. Persistent Engineering Knowledge

AllMyAgents must not simply archive conversations.

It should preserve useful structured engineering knowledge, including where appropriate:

### Developer knowledge

- preferences,
- learning style,
- demonstrated skills,
- development goals,
- useful long-term patterns.

### Project knowledge

- project purpose,
- PRD requirements,
- architecture,
- constraints,
- current state,
- relevant repository structure.

### Decision history

- important decisions,
- alternatives considered,
- rejected approaches,
- reasons for decisions,
- outcomes and lessons.

Only valuable knowledge should persist.

The system should not automatically persist every AI response or conversation.

Only confirmed and useful information should become persistent state.

---

## 8. Context Strategy

AllMyAgents should provide relevant context to the AI executor rather than blindly providing all stored information.

Context may be composed from:

- Developer Profile,
- Project State,
- current task requirements,
- repository rules,
- relevant project files,
- relevant previous decisions.

The context system should minimize unnecessary repetition and token usage.

The system should determine which stored information is relevant to the current task.

The objective is:

> Relevant context, not maximum context.

---

## 9. Storage Strategy

AllMyAgents will be **local-first**.

V0 will store the core intelligence locally on the developer's machine.

Developer intelligence and project intelligence must remain conceptually and physically separated.

The initial model is:

- Developer intelligence: `~/.allmyagents/`
- Project intelligence: `<project>/.allmyagents/`

Developer intelligence is reusable across projects.

Project intelligence belongs to the relevant project.

Local developer profiles and project state must not be committed to public repositories by default.

A future version may provide optional encrypted cloud synchronization.

Cloud infrastructure is explicitly outside the V0 implementation scope.

---

## 10. Agent Strategy

AllMyAgents is agent-agnostic.

It is not intended to become another coding agent.

### V0

Codex is the first execution integration.

AllMyAgents provides context and workflow intelligence to the executor.

The executor must remain replaceable.

### Future

The architecture may support additional agents such as:

- Claude Code,
- Gemini-based agents,
- other compatible coding agents.

Multi-agent orchestration is future scope, not a V0 requirement.

---

## 11. Core Engineering Workflow

The intended workflow is:

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

### Understand

Identify the user's intent and relevant project requirements.

### Plan

Determine the appropriate approach using project rules, architecture, decisions, and developer preferences.

### Implement

Use the configured AI executor.

The executor may implement a clearly defined task autonomously within the requirements and architectural boundaries.

The developer should not need to approve every implementation detail.

### Test

Run the project's relevant tests and build checks.

### Verify

Check the implementation against:

- task requirements,
- project architecture,
- coding rules,
- simplicity/KISS principles,
- changed files,
- unnecessary complexity,
- relevant risks.

### Explain

Produce a human-readable explanation of:

- what was implemented,
- why it was implemented,
- how important functions/components work,
- important decisions,
- relevant audit/hand-off information.

The explanation should be appropriate to the developer's configured explanation depth and learning preferences.

### Human Approval

The developer remains the final authority over important changes.

Important or risky operations must remain subject to developer control.

### Update Project State

After approval, only confirmed and valuable project knowledge is persisted.

---

## 12. Autonomy

Autonomy is adaptive rather than a single global switch.

The system should consider:

- developer preferences,
- project rules,
- task context,
- action risk.

V0 must support human-controlled execution.

For example, a developer may choose to receive Git commands and execute them manually rather than allowing the tool to push automatically.

The system must not automatically:

- push to remote repositories,
- force-push,
- merge branches,
- create releases,
- perform destructive operations,

unless explicitly permitted by the developer and supported by the configured workflow.

The developer must be able to maintain manual control even when the AI executor operates with high autonomy.

Future versions may support higher levels of autonomy.

---

## 13. V0 Scope

The V0 must demonstrate the core workflow on a real project.

### Included

- first-run onboarding,
- local Developer Profile,
- editable Developer Profile,
- individual profile preference updates,
- temporary task/session interaction overrides,
- project detection,
- project context,
- task context,
- relevant context selection,
- Codex execution integration,
- testing,
- technical verification,
- human-readable learning/handoff report,
- human approval,
- project-state update.

### V0 onboarding

The first-run onboarding must:

1. ask the eight predefined profile questions,
2. provide four selectable answers for each question,
3. create structured Developer Profile data,
4. store the profile locally,
5. allow the developer to later modify individual preferences.

The onboarding must not require the developer to write prompts.

### Validation target

The first real validation will be performed by using AllMyAgents on a real Milestone 3 coding task.

The objective is to determine whether the tool actually reduces repeated context transfer while preserving developer understanding and control.

---

## 14. Non-Goals for V0

The following are explicitly out of scope:

- multi-agent orchestration,
- cloud synchronization,
- SaaS backend,
- subscriptions/payments,
- model marketplace,
- large graphical UI,
- automatic Git push by default,
- fully autonomous development,
- conversation-history archiving,
- advanced knowledge graph infrastructure unless required by the V0 implementation.

These may be reconsidered after validation.

---

## 15. Success Criteria

V0 is successful only if real usage demonstrates measurable improvement.

We will evaluate:

### Context reduction

Can the developer start a task without repeatedly explaining their identity, learning preferences, project context, and working conventions?

### Workflow reduction

Does the developer perform fewer manual context-transfer steps between ChatGPT, Codex, project files, and other tools?

### Token/time efficiency

Does the workflow reduce unnecessary prompting, repeated explanations, and context reconstruction?

### Developer control

Can the developer understand and approve important AI actions without losing control of the process?

### Developer understanding

Can the developer explain the resulting implementation after using the system?

### Project consistency

Does the implementation remain aligned with project requirements, architecture, coding rules, and simplicity constraints?

### Profile usefulness

Does the Developer Profile meaningfully change AI behavior in accordance with the developer's selected preferences?

### Profile flexibility

Can the developer change an individual preference without repeating the entire onboarding process?

Can the developer temporarily change interaction behavior for a specific task without permanently changing their default profile?

### State continuity

After completing a task, can the next task benefit from the confirmed knowledge without rebuilding the context manually?

---

## 16. Product Principle

The core product principle is:

> **Remove unnecessary friction from the developer's interaction with AI without removing the developer from the engineering process.**

The goal is not to replace the developer.

The goal is to make the developer more capable, informed, and efficient when working with AI.

The product should save time without turning development into uncontrolled vibe coding.

The developer remains the owner of the engineering decisions and should be able to understand the work produced with AI assistance.

---

## 17. Validation Strategy

The first user is the project creator.

The product will be dogfooded on real development work before being presented to other developers.

Initial validation sequence:

1. Build V0.
2. Use V0 on a real M3 task.
3. Compare the workflow against the current manual workflow.
4. Record friction, failures, time savings, context reduction, and learning quality.
5. Improve V0 based on evidence.
6. Use it on additional real tasks.
7. Test with a small group of other developers.
8. Re-evaluate the product hypothesis using real data.

If the product does not produce measurable improvement, the scope and product thesis must be reconsidered rather than forcing the original idea.

---

## 18. Current Product Direction

The long-term product direction is:

**Personal Engineering Intelligence**

A local-first intelligence layer that becomes more useful over time by maintaining structured knowledge about the developer, their projects, their decisions, their progress, and their working preferences, while remaining independent of any single AI model or coding agent.