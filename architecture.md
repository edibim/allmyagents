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
AI Executor  
↓  
Repository

Codex is the first AI executor used for V0.

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

V0 uses Codex.

The architecture must keep the executor replaceable so that additional AI agents can be supported later.

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

The initial interaction model may eventually expose commands such as:

- allmyagents init
- allmyagents profile
- allmyagents project
- allmyagents run
- allmyagents verify

Exact commands are not finalized yet.

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

## 7. Future Multi-Agent Support

The architecture should eventually allow:

AllMyAgents  
├── Codex  
├── Claude Code  
└── Other Agents

However, multi-agent orchestration is explicitly outside V0.

The first objective is to prove that the intelligence layer provides value independently of the executor.

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