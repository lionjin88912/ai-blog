---
name: context-engineering
description: "Optimizes agent context setup. Use when starting a new session, when agent output quality degrades, when switching between tasks, or when you need to configure rules files and context for a project."
---

# Context Engineering

## Overview

Feed agents the right information at the right time. Context is the single biggest lever for agent output quality — too little and the agent hallucinates, too much and it loses focus.

## When to Use

- Starting a new coding session
- Agent output quality is declining
- Switching between different parts of a codebase
- Setting up a new project for AI-assisted development

## The Context Hierarchy

```
1. Rules Files (CLAUDE.md, copilot-instructions.md)  ← Always loaded
2. Spec / Architecture Docs                          ← Per feature/session
3. Relevant Source Files                              ← Per task
4. Error Output / Test Results                        ← Per iteration
5. Conversation History                               ← Accumulates
```

### Level 1: Rules Files

Create a rules file covering: tech stack, commands, code conventions, boundaries, and one example of a well-written component.

### Level 2: Specs and Architecture

Load only the relevant spec section, not the entire spec.

### Level 3: Relevant Source Files

Before editing, read the file. Before implementing a pattern, find an existing example. Load: files to modify, related tests, similar patterns, type definitions.

### Level 4: Error Output

Feed specific errors back, not entire 500-line outputs.

### Level 5: Conversation Management

Start fresh sessions when switching features. Summarize progress when context gets long.

## Context Packing Strategies

- **Brain Dump:** At session start, provide everything needed in a structured block
- **Selective Include:** Only include what's relevant to the current task
- **Hierarchical Summary:** Maintain a project map, load relevant sections

## Confusion Management

**When context conflicts:** Surface the discrepancy explicitly. Don't silently pick one interpretation.

**When requirements are incomplete:** Check existing code for precedent. If none, stop and ask.

**Inline planning:** Emit a lightweight plan before executing multi-step tasks.

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Context starvation | Load rules file + relevant source files before each task |
| Context flooding | Include only task-relevant context. Aim for <2,000 lines |
| Stale context | Start fresh sessions when context drifts |
| Missing examples | Include one example of the pattern to follow |
| Silent confusion | Surface ambiguity explicitly |

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The agent should figure out the conventions" | Write a rules file — 10 minutes saves hours |
| "More context is always better" | Performance degrades with too many instructions. Be selective. |

## Verification

- [ ] Rules file exists covering tech stack, commands, conventions, boundaries
- [ ] Agent output follows project patterns
- [ ] Agent references actual project files (not hallucinated ones)
- [ ] Context refreshed when switching major tasks
