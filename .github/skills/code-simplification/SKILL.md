---
name: code-simplification
description: "Simplifies code for clarity. Use when refactoring code for clarity without changing behavior. Use when code works but is harder to read, maintain, or extend than it should be."
---

# Code Simplification

## Overview

Simplify code by reducing complexity while preserving exact behavior. The goal isn't fewer lines — it's code easier to read, understand, modify, and debug. Test: "Would a new team member understand this faster than the original?"

## When to Use

- After a feature works but feels heavier than needed
- During code review when complexity is flagged
- When encountering deeply nested logic or unclear names
- When refactoring code written under pressure

**When NOT to use:** Code is already clean, you don't understand it yet, or performance-critical code where "simpler" would be measurably slower.

## The Five Principles

1. **Preserve behavior exactly** — All inputs, outputs, side effects, error behavior unchanged
2. **Follow project conventions** — Match the codebase, don't impose external preferences
3. **Prefer clarity over cleverness** — Explicit beats compact when compact requires a mental pause
4. **Maintain balance** — Don't inline too aggressively, combine unrelated logic, or optimize for line count
5. **Scope to what changed** — Default to recently modified code. Avoid drive-by refactors.

## The Simplification Process

### Step 1: Understand Before Touching (Chesterton's Fence)

Before removing anything, understand why it exists. Check git blame for context.

### Step 2: Identify Opportunities

**Structural:** Deep nesting → guard clauses. Long functions → split. Nested ternaries → if/else. Boolean flags → options objects.

**Naming:** Generic names (`data`, `temp`) → descriptive names. Abbreviated names → full words.

**Redundancy:** Duplicated logic → shared function. Dead code → remove. Unnecessary wrappers → inline.

### Step 3: Apply Incrementally

One simplification at a time. Run tests after each. Submit refactoring separately from features.

**Rule of 500:** If touching 500+ lines, use automation (codemods, AST transforms) instead of manual edits.

### Step 4: Verify the Result

Compare before and after. If the "simplified" version is harder to understand, revert.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "It's working, no need to touch it" | Hard-to-read code will be hard to fix when it breaks |
| "Fewer lines is always simpler" | A 1-line nested ternary is not simpler than a 5-line if/else |
| "I'll refactor while adding this feature" | Separate them. Mixed changes are harder to review and revert. |
| "The original author must have had a reason" | Check git blame. But accumulated complexity often has no reason. |

## Red Flags

- Simplification that requires modifying tests (you changed behavior)
- "Simplified" code that's longer and harder to follow
- Renaming to match your preferences rather than project conventions
- Removing error handling for "cleanliness"
- Simplifying code you don't fully understand

## Verification

- [ ] All existing tests pass without modification
- [ ] Build succeeds with no new warnings
- [ ] Each simplification is a reviewable, incremental change
- [ ] Simplified code follows project conventions
- [ ] No error handling removed or weakened
- [ ] No dead code left behind
