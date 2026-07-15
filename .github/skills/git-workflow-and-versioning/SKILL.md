---
name: git-workflow-and-versioning
description: "Structures git workflow practices. Use when making any code change. Use when committing, branching, resolving conflicts, or when organizing work across multiple parallel streams."
---

# Git Workflow and Versioning

## Overview

Treat commits as save points, branches as sandboxes, and history as documentation. With AI agents generating code at speed, disciplined version control keeps changes manageable, reviewable, and reversible.

## When to Use

Always. Every code change flows through git.

## Core Principles

### Trunk-Based Development

Keep `main` always deployable. Work in short-lived feature branches (1-3 days). Feature flags > long branches.

### 1. Commit Early, Commit Often

Each successful increment gets its own commit. Don't accumulate large uncommitted changes.

### 2. Atomic Commits

Each commit does one logical thing. Don't mix unrelated changes.

### 3. Descriptive Messages

```
<type>: <short description>

<optional body explaining why>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

First line: short, imperative, standalone. Body: context and reasoning not visible in code.

### 4. Keep Concerns Separate

Don't combine formatting with behavior changes. Don't combine refactors with features.

### 5. Size Your Changes

~100 lines → Easy to review. ~300 lines → Acceptable. ~1000+ lines → Split.

## The Save Point Pattern

```
Change → Test passes? → Commit → Continue
                    └→ Test fails? → Revert to last commit → Investigate
```

Never lose more than one increment of work.

## Branching Strategy

```
feature/<description>   fix/<description>   chore/<description>   refactor/<description>
```

Branch from main, keep short-lived, delete after merge.

## Pre-Commit Hygiene

```bash
git diff --staged                                    # Check what you're committing
git diff --staged | grep -i "password\|secret\|token" # No secrets
npm test && npm run lint && npx tsc --noEmit         # Tests, lint, types
```

## Change Summaries

After modifications, provide:
- **CHANGES MADE:** What was changed
- **THINGS I DIDN'T TOUCH:** Scope discipline
- **POTENTIAL CONCERNS:** Assumptions to validate

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'll commit when the feature is done" | One giant commit is impossible to review, debug, or revert |
| "The message doesn't matter" | Messages are documentation for future you and future agents |
| "I'll squash it all later" | Squashing destroys the development narrative |

## Red Flags

- Large uncommitted changes accumulating
- Messages like "fix", "update", "misc"
- Formatting mixed with behavior changes
- Committing `node_modules/`, `.env`, or build artifacts
- Force-pushing to shared branches

## Verification

- [ ] Commit does one logical thing
- [ ] Message explains the why
- [ ] Tests pass before committing
- [ ] No secrets in the diff
- [ ] `.gitignore` covers standard exclusions
