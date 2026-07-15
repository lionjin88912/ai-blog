---
name: deprecation-and-migration
description: "Manages deprecation and migration. Use when removing old systems, APIs, or features. Use when migrating users from one implementation to another. Use when deciding whether to maintain or sunset existing code."
---

# Deprecation and Migration

## Overview

Code is a liability, not an asset. Every line has ongoing cost — bugs, dependencies, security patches, onboarding. Deprecation removes code that no longer earns its keep. Migration moves users safely from old to new.

## When to Use

- Replacing an old system, API, or library
- Sunsetting a feature no longer needed
- Consolidating duplicate implementations
- Removing dead code nobody owns
- Planning lifecycle of a new system (deprecation planning starts at design time)

## Core Principles

- **Code is a liability** — value is the functionality, not the code itself
- **Hyrum's Law makes removal hard** — every observable behavior is depended on
- **Deprecation planning starts at design time** — ask "How would we remove this in 3 years?"

## The Deprecation Decision

1. Does this system still provide unique value? → If yes, maintain it.
2. How many consumers depend on it? → Quantify migration scope.
3. Does a replacement exist? → If no, build it first.
4. What's the migration cost? → If trivially automated, do it.
5. What's the cost of NOT deprecating? → Security risk, engineer time, complexity.

## Compulsory vs Advisory

| Type | When | Mechanism |
|------|------|-----------|
| Advisory | Old system is stable | Warnings, documentation, nudges |
| Compulsory | Security issues or unsustainable cost | Hard deadline + migration tooling |

Default to advisory. Compulsory requires providing migration tooling and support.

## The Migration Process

1. **Build the replacement** — Don't deprecate without a working alternative proven in production
2. **Announce and document** — Deprecation notice, replacement, reason, migration guide
3. **Migrate incrementally** — One consumer at a time, verify behavior matches
4. **Remove the old system** — Only after all consumers migrated and zero active usage verified

**The Churn Rule:** If you own the deprecated infrastructure, you are responsible for migrating users.

## Migration Patterns

- **Strangler:** Run old and new in parallel, route traffic incrementally
- **Adapter:** Old interface wrapping new implementation
- **Feature Flag:** Switch consumers one at a time

## Zombie Code

Code nobody owns but everybody depends on. Signs: no commits in 6+ months, no maintainer, failing tests nobody fixes. Response: assign an owner or deprecate with a plan.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "It still works, why remove it?" | Unmaintained code accumulates security debt silently |
| "Someone might need it later" | Rebuilding is cheaper than maintaining unused code |
| "Users will migrate on their own" | They won't. Provide tooling or do it yourself. |

## Red Flags

- Deprecated systems with no replacement
- Deprecation announcements without migration tooling
- Zombie code with no owner and active consumers
- New features added to a deprecated system

## Verification

- [ ] Replacement is production-proven
- [ ] Migration guide exists with concrete steps
- [ ] All active consumers migrated (verified by metrics)
- [ ] Old code, tests, docs fully removed
- [ ] No references to deprecated system remain
