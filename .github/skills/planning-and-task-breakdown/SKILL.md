---
name: planning-and-task-breakdown
description: "Breaks work into ordered tasks. Use when you have a spec or requirements and need to break work into implementable tasks. Use when a task feels too large to start, when you need to estimate scope, or when parallel work is possible."
---

# Planning and Task Breakdown

## Overview

Decompose work into small, verifiable tasks with explicit acceptance criteria. Every task should be small enough to implement, test, and verify in a single focused session.

## When to Use

- You have a spec and need implementable units
- A task feels too large or vague to start
- Work needs to be parallelized
- The implementation order isn't obvious

**When NOT to use:** Single-file changes with obvious scope.

## The Planning Process

### Step 1: Enter Plan Mode

Read-only — no code writing:
- Read the spec and relevant codebase sections
- Identify existing patterns and conventions
- Map dependencies between components
- Note risks and unknowns

### Step 2: Identify the Dependency Graph

Map what depends on what. Implementation order follows the dependency graph bottom-up.

### Step 3: Slice Vertically

Build one complete path through the stack, not horizontal layers:

**Bad:** Task 1: Build entire DB schema → Task 2: Build all API endpoints → Task 3: Build all UI

**Good:** Task 1: User can create an account (schema + API + UI) → Task 2: User can log in → Task 3: User can create a task

### Step 4: Write Tasks

```markdown
## Task [N]: [Short descriptive title]

**Description:** What this task accomplishes.

**Acceptance criteria:**
- [ ] [Specific, testable condition]

**Verification:**
- [ ] Tests pass
- [ ] Build succeeds

**Dependencies:** [Task numbers or "None"]
**Files likely touched:** [list]
**Estimated scope:** [Small: 1-2 files | Medium: 3-5 files | Large: 5+ files]
```

### Step 5: Order and Checkpoint

- Dependencies satisfied (build foundation first)
- Each task leaves the system working
- Checkpoints every 2-3 tasks
- High-risk tasks early (fail fast)

## Task Sizing

| Size | Files | Example |
|------|-------|---------|
| **XS** | 1 | Add a validation rule |
| **S** | 1-2 | Add a new API endpoint |
| **M** | 3-5 | User registration flow |
| **L** | 5-8 | Search with filtering — break down further |
| **XL** | 8+ | **Too large** |

If a task is L or larger, break it down. Agents perform best on S and M tasks.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'll figure it out as I go" | That's how you get tangled messes and rework |
| "The tasks are obvious" | Write them down — explicit tasks surface hidden dependencies |
| "Planning is overhead" | Planning IS the task. Implementation without a plan is just typing |

## Red Flags

- Starting implementation without a written task list
- Tasks without acceptance criteria
- No verification steps in the plan
- All tasks are XL-sized
- Dependency order not considered

## Verification

- [ ] Every task has acceptance criteria
- [ ] Every task has a verification step
- [ ] Dependencies identified and ordered correctly
- [ ] No task touches more than ~5 files
- [ ] Checkpoints exist between major phases
- [ ] Human has reviewed and approved the plan
