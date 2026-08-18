# AGENTS.md

## Purpose

This repository contains AllMyAgents, a local-first intelligence layer for AI-assisted software development.

Agents working in this repository must optimize for correctness, simplicity, maintainability, and developer understanding.

---

## Before Making Changes

1. Read the relevant project documentation before changing code.
2. Inspect the repository structure before making assumptions.
3. Read the relevant sections of `PRD.md`.
4. Check existing architecture and implementation before introducing new patterns.
5. Identify the exact scope of the task.

Do not modify unrelated files.

---

## Requirements and Architecture

- `PRD.md` defines product requirements and V0 scope.
- Architecture decisions must be respected.
- Do not introduce functionality outside the task scope without explicit approval.
- If requirements conflict or are ambiguous, stop and ask before implementing.
- Prefer the simplest solution that satisfies the requirements.

---

## Coding Principles

Follow these principles:

- Keep It Simple.
- Prefer readable code over clever code.
- Keep functions focused and reasonably small.
- Avoid unnecessary abstractions.
- Avoid duplicated logic where practical.
- Reuse existing project patterns when they are appropriate.
- Do not introduce dependencies without a clear reason.
- Do not over-engineer for hypothetical future requirements.

Implementation details are flexible as long as they remain within the project's requirements and architectural boundaries.

---

## Testing

Before considering a task complete:

1. Run relevant tests.
2. Check for compilation/build errors.
3. Verify that the implementation satisfies the task requirements.
4. Check that unrelated behavior was not unnecessarily changed.

If tests fail, investigate the cause and fix the implementation before reporting completion.

Do not claim a task is complete only because the code compiles.

---

## Verification

Before completion, review:

- changed files,
- task requirements,
- project architecture,
- coding principles,
- unnecessary complexity,
- duplicated logic,
- unintended side effects.

The question is not only:

> "Does the code work?"

Also ask:

> "Does the code belong here, and is this the simplest appropriate solution?"

---

## Git Rules

Agents must not automatically:

- push to a remote repository,
- force push,
- merge branches,
- create releases,
- modify the default branch.

Do not create commits unless explicitly requested.

When Git commands are needed, provide the commands clearly so the developer can review and execute them.

Keep commits focused and meaningful.

---

## Security and Privacy

Never commit:

- API keys,
- passwords,
- access tokens,
- private credentials,
- `.env` files containing secrets,
- personal developer profiles,
- local AllMyAgents state.

Check the repository before committing sensitive information.

---

## Developer Control

The developer remains responsible for approving important changes.

Agents should explain significant implementation decisions when requested.

Do not hide changes behind unnecessary automation.

If an operation has potentially destructive or irreversible consequences, ask for confirmation before performing it.

---

## Completion Report

When a task is complete, provide or generate a concise report containing:

- what was changed,
- why it was changed,
- important implementation details,
- tests performed,
- verification results,
- relevant decisions or trade-offs,
- anything the developer should understand before review.

The developer should be able to understand the completed work without reconstructing the entire coding session.

---

## General Rule

> Follow the project's requirements strictly, but keep implementation simple and allow reasonable engineering judgment within those boundaries.