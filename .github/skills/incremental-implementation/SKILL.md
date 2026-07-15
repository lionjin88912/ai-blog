---
name: incremental-implementation
description: "Delivers changes incrementally. Use when implementing any feature or change that touches more than one file. Use when a task feels too big to land in one step."
---

# Incremental Implementation

## Overview

Build in thin vertical slices — implement one piece, test it, verify it, then expand. Each increment leaves the system in a working, testable state.

## When to Use

- Implementing any multi-file change
- Building a new feature from a task breakdown
- Any time you're tempted to write more than ~100 lines before testing

**When NOT to use:** Single-file, single-function changes with minimal scope.

## The Increment Cycle

```
Implement ──→ Test ──→ Verify ──→ Commit ──→ Next slice
```

For each slice:
1. Implement the smallest complete piece of functionality
2. Run the test suite (or write a test if none exists)
3. Verify the slice works (tests pass, build succeeds)
4. Commit with a descriptive message
5. Move to the next slice

## Slicing Strategies

**Vertical Slices (Preferred):** Build one complete path through the stack per slice.

**Contract-First:** Define the API contract first, then implement backend and frontend in parallel.

**Risk-First:** Tackle the riskiest piece first. If it fails, you discover it early.

## Implementation Rules

**Rule 0: Simplicity First** — Ask "What is the simplest thing that could work?" Three similar lines is better than a premature abstraction.

**Rule 0.5: Scope Discipline** — Touch only what the task requires. Don't "clean up" adjacent code, refactor unrelated imports, or add features not in the spec.

**Rule 1: One Thing at a Time** — Each increment changes one logical thing.

**Rule 2: Keep It Compilable** — After each increment, the project must build and tests must pass.

**Rule 3: Feature Flags for Incomplete Features** — If a feature isn't ready for users, put it behind a flag.

**Rule 4: Safe Defaults** — New code defaults to conservative behavior.

**Rule 5: Rollback-Friendly** — Each increment should be independently revertable.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'll test it all at the end" | Bugs compound. A bug in Slice 1 makes Slices 2-5 wrong. |
| "It's faster to do it all at once" | Until something breaks and you can't find which of 500 lines caused it. |
| "These changes are too small to commit" | Small commits are free. Large commits hide bugs. |
| "This refactor is small enough to include" | Refactors mixed with features make both harder to debug. |

## Red Flags

- More than 100 lines written without running tests
- Multiple unrelated changes in a single increment
- "Let me just quickly add this too" scope expansion
- Build or tests broken between increments
- Building abstractions before the third use case

## Verification

- [ ] Each increment was individually tested and committed
- [ ] Full test suite passes
- [ ] Build is clean
- [ ] Feature works end-to-end as specified
- [ ] No uncommitted changes remain
