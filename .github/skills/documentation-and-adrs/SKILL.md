---
name: documentation-and-adrs
description: "Records decisions and documentation. Use when making architectural decisions, changing public APIs, shipping features, or when you need to record context that future engineers and agents will need."
---

# Documentation and ADRs

## Overview

Document decisions, not just code. The most valuable documentation captures the *why* — context, constraints, and trade-offs. Code shows *what*; documentation explains *why it was built this way* and *what alternatives were considered*.

## When to Use

- Making a significant architectural decision
- Choosing between competing approaches
- Adding or changing a public API
- Shipping a feature that changes user-facing behavior
- Explaining the same thing repeatedly

**When NOT to use:** Don't document obvious code. Don't write docs for throwaway prototypes.

## Architecture Decision Records (ADRs)

Write an ADR when choosing a framework, designing a data model, selecting auth strategy, or making any expensive-to-reverse decision.

### ADR Template

Store in `docs/decisions/` with sequential numbering:

```markdown
# ADR-001: [Decision Title]

## Status
Accepted | Superseded by ADR-XXX | Deprecated

## Date
[Date]

## Context
[Requirements, constraints, what prompted this decision]

## Decision
[What we decided and why]

## Alternatives Considered
### [Alternative]
- Pros: [...]
- Cons: [...]
- Rejected: [reason]

## Consequences
[What follows from this decision — trade-offs, requirements, impacts]
```

Don't delete old ADRs. When a decision changes, write a new one that supersedes the old.

## Inline Documentation

- **Comment the why, not the what** — `// Sliding window reset prevents burst attacks` not `// increment counter`
- **Don't comment self-explanatory code**
- **Don't leave commented-out code** — git has history
- **Document known gotchas** with `IMPORTANT:` markers

## API Documentation

- Use JSDoc/TSDoc for public function parameters, returns, throws, and examples
- For REST APIs, maintain OpenAPI/Swagger specs

## README Structure

Every project needs: description, quick start (clone, install, env, run), commands table, architecture overview, contributing guide.

## Changelog

Track what was Added, Fixed, Changed per version with issue references.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The code is self-documenting" | Code shows what, not why, not what was rejected |
| "We'll write docs when the API stabilizes" | APIs stabilize faster when documented. The doc tests the design. |
| "Nobody reads docs" | Agents do. Future engineers do. Your 3-months-later self does. |
| "ADRs are overhead" | A 10-minute ADR prevents a 2-hour debate six months later |

## Red Flags

- Architectural decisions with no written rationale
- Public APIs with no documentation
- README that doesn't explain how to run the project
- Commented-out code instead of deletion
- TODO comments that have been there for weeks

## Verification

- [ ] ADRs exist for significant architectural decisions
- [ ] README covers quick start, commands, and architecture
- [ ] API functions have parameter and return type documentation
- [ ] Known gotchas documented inline
- [ ] No commented-out code remains
