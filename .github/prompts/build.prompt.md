---
description: "Build incrementally. Implements one task or slice at a time — test, verify, commit."
agent: "agent"
argument-hint: "Task or feature slice to implement"
---

Implement the described task following a disciplined build process.

Follow the `incremental-implementation` and `test-driven-development` skills. For framework-specific code, also follow the `source-driven-development` skill. For UI work, follow `frontend-ui-engineering`. For API work, follow `api-and-interface-design`.

## Process

1. Read relevant source files and find existing patterns to follow
2. Write a failing test first (RED)
3. Write the minimum code to make it pass (GREEN)
4. Refactor while tests stay green (REFACTOR)
5. Verify: tests pass, build succeeds, lint clean
6. Commit with a descriptive message
7. Provide a change summary (CHANGES MADE / DIDN'T TOUCH / CONCERNS)

## Rules

- One logical thing per increment
- Keep it compilable after every change
- Touch only what the task requires — no scope expansion
- Simplest thing that could work first
