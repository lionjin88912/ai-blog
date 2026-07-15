---
name: spec-driven-development
description: "Creates specs before coding. Use when starting a new project, feature, or significant change and no specification exists yet. Use when requirements are unclear, ambiguous, or only exist as a vague idea."
---

# Spec-Driven Development

## Overview

Write a structured specification before writing any code. The spec defines what we're building, why, and how we'll know it's done. Code without a spec is guessing.

## When to Use

- Starting a new project or feature
- Requirements are ambiguous or incomplete
- The change touches multiple files or modules
- The task would take more than 30 minutes to implement

**When NOT to use:** Single-line fixes, typo corrections, or unambiguous self-contained changes.

## The Gated Workflow

```
SPECIFY ──→ PLAN ──→ TASKS ──→ IMPLEMENT
   │          │        │          │
   ▼          ▼        ▼          ▼
 Human      Human    Human      Human
 reviews    reviews  reviews    reviews
```

### Phase 1: Specify

Surface assumptions immediately before writing any spec content:

```
ASSUMPTIONS I'M MAKING:
1. [Assumption about tech stack]
2. [Assumption about architecture]
→ Correct me now or I'll proceed with these.
```

Write a spec covering six core areas:

1. **Objective** — What we're building, who it's for, what success looks like
2. **Commands** — Full executable commands (build, test, lint, dev)
3. **Project Structure** — Where source code, tests, and docs live
4. **Code Style** — One real code snippet showing the style
5. **Testing Strategy** — Framework, locations, coverage expectations
6. **Boundaries:**
   - **Always do:** Run tests before commits, validate inputs
   - **Ask first:** Database schema changes, adding dependencies
   - **Never do:** Commit secrets, remove failing tests

Reframe vague requirements as concrete success criteria:

```
REQUIREMENT: "Make the dashboard faster"
REFRAMED: Dashboard LCP < 2.5s, initial data load < 500ms, CLS < 0.1
```

### Phase 2: Plan

Generate a technical implementation plan: components, dependencies, order, risks.

### Phase 3: Tasks

Break the plan into discrete tasks with acceptance criteria and verification steps.

### Phase 4: Implement

Execute tasks one at a time following `incremental-implementation` and `test-driven-development`.

## Keeping the Spec Alive

- Update when decisions or scope change
- Commit the spec alongside code
- Reference spec sections in PRs

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "This is simple, I don't need a spec" | Simple tasks still need acceptance criteria. A two-line spec is fine. |
| "I'll write the spec after I code it" | That's documentation, not specification. The value is clarity *before* code. |
| "The spec will slow us down" | A 15-minute spec prevents hours of rework. |

## Red Flags

- Starting code without written requirements
- Implementing features not in any spec
- Making architectural decisions without documenting them

## Verification

- [ ] Spec covers all six core areas
- [ ] Human has reviewed and approved the spec
- [ ] Success criteria are specific and testable
- [ ] Boundaries (Always/Ask First/Never) are defined
- [ ] Spec is saved to a file in the repository
