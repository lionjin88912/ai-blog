---
name: code-review-and-quality
description: "Conducts multi-axis code review. Use before merging any change. Use when reviewing code written by yourself, another agent, or a human. Use when assessing code quality across multiple dimensions."
---

# Code Review and Quality

## Overview

Multi-dimensional code review with quality gates. Every change gets reviewed. Review covers five axes: correctness, readability, architecture, security, and performance.

**Approval standard:** Approve when it definitely improves overall code health, even if imperfect. Don't block because it isn't how you would have written it.

## When to Use

- Before merging any PR or change
- After completing a feature implementation
- When evaluating code from another agent or model
- After any bug fix

## The Five-Axis Review

1. **Correctness** — Matches spec? Edge cases handled? Error paths? Tests correct?
2. **Readability** — Names descriptive? Control flow straightforward? No "clever" tricks? Could it be simpler?
3. **Architecture** — Follows existing patterns? Clean module boundaries? Appropriate abstraction level?
4. **Security** — Input validated? Secrets out of code? Auth checked? SQL parameterized? External data untrusted?
5. **Performance** — N+1 patterns? Unbounded loops? Unnecessary re-renders? Missing pagination?

## Review Process

1. **Understand context** — What is this change trying to accomplish?
2. **Review tests first** — Tests reveal intent and coverage
3. **Review implementation** — Walk through with five axes
4. **Categorize findings:**
   - *(no prefix)* — Required change, must address
   - **Critical:** — Blocks merge (security, data loss)
   - **Nit:** — Minor, optional
   - **Optional/Consider:** — Worth considering but not required
   - **FYI** — Informational, no action needed
5. **Verify the verification** — Tests run? Build pass? Manual testing done?

## Change Sizing

```
~100 lines  → Good
~300 lines  → Acceptable for single logical change
~1000 lines → Too large — split it
```

Separate refactoring from feature work. They are two separate changes.

## Key Practices

- **Dead code hygiene:** List orphaned code after refactoring, ask before deleting
- **Dependency discipline:** Before adding any dependency, check if existing stack solves it, check size, maintenance, vulnerabilities, license
- **Honesty:** Don't rubber-stamp. Don't soften real issues. Push back on problems.
- **Speed:** Respond within one business day. Quick feedback reduces frustration.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "It works, that's good enough" | Working but unreadable/insecure code creates compounding debt |
| "AI-generated code is probably fine" | AI code needs more scrutiny, not less |
| "We'll clean it up later" | Later never comes. The review is the quality gate. |

## Red Flags

- PRs merged without any review
- "LGTM" without evidence of review
- Large PRs "too big to review properly"
- No regression tests with bug fix PRs
- Accepting "I'll fix it later"

## Verification

- [ ] All Critical issues resolved
- [ ] All Important issues resolved or explicitly deferred
- [ ] Tests pass
- [ ] Build succeeds
- [ ] Verification story documented
