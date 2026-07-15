---
name: source-driven-development
description: "Grounds every implementation decision in official documentation. Use when you want authoritative, source-cited code free from outdated patterns. Use when building with any framework or library where correctness matters."
---

# Source-Driven Development

## Overview

Every framework-specific code decision must be backed by official documentation. Don't implement from memory — verify, cite, and let the user see your sources. Training data goes stale, APIs get deprecated, best practices evolve.

## When to Use

- Building code that follows current best practices for a framework
- Implementing features where the framework's recommended approach matters
- Reviewing code that uses framework-specific patterns

**When NOT to use:** Pure logic (loops, conditionals), renaming/typo fixes, or when the user explicitly wants speed over verification.

## The Process

```
DETECT ──→ FETCH ──→ IMPLEMENT ──→ CITE
```

### Step 1: Detect Stack and Versions

Read the project's dependency file (`package.json`, `Cargo.toml`, `pyproject.toml`, etc.) to identify exact versions. State what you found explicitly. If versions are ambiguous, ask.

### Step 2: Fetch Official Documentation

Fetch the specific documentation page for the feature you're implementing.

**Source hierarchy:**
1. Official documentation (react.dev, docs.djangoproject.com)
2. Official blog / changelog
3. Web standards references (MDN, web.dev)
4. Browser/runtime compatibility (caniuse.com)

**Not authoritative:** Stack Overflow, blog posts, tutorials, your own training data.

### Step 3: Implement Following Documented Patterns

- Use API signatures from the docs, not from memory
- If docs show a new way, use the new way
- If docs deprecate a pattern, don't use the deprecated version
- If docs don't cover something, flag it as unverified

When docs conflict with existing project code, surface the conflict — don't silently pick one.

### Step 4: Cite Your Sources

```typescript
// React 19 form handling with useActionState
// Source: https://react.dev/reference/react/useActionState#usage
const [state, formAction, isPending] = useActionState(submitOrder, initialState);
```

If you cannot find documentation: `UNVERIFIED: Could not find official documentation for this pattern.`

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'm confident about this API" | Confidence is not evidence. Training data contains outdated patterns. |
| "Fetching docs wastes tokens" | Hallucinating an API wastes more. One fetch prevents hours of rework. |
| "This is a simple task" | Simple tasks with wrong patterns become templates copied into ten components. |

## Red Flags

- Writing framework code without checking docs for that version
- Using "I believe" or "I think" about an API instead of citing
- Citing Stack Overflow instead of official documentation
- Using deprecated APIs from training data

## Verification

- [ ] Framework versions identified from dependency file
- [ ] Official documentation fetched for framework-specific patterns
- [ ] Code follows current version's documented patterns
- [ ] Non-trivial decisions include source citations
- [ ] No deprecated APIs used
- [ ] Anything unverified is explicitly flagged
