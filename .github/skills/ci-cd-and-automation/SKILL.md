---
name: ci-cd-and-automation
description: "Automates CI/CD pipeline setup. Use when setting up or modifying build and deployment pipelines. Use when automating quality gates, configuring test runners in CI, or establishing deployment strategies."
---

# CI/CD and Automation

## Overview

Automate quality gates so no change reaches production without passing tests, lint, type checking, and build. CI catches what humans and agents miss, consistently on every change.

**Shift Left:** Catch problems early. A bug caught in linting costs minutes; in production, hours.

**Faster is Safer:** Smaller batches and more frequent releases reduce risk.

## When to Use

- Setting up a new project's CI pipeline
- Adding or modifying automated checks
- Configuring deployment pipelines
- Debugging CI failures

## The Quality Gate Pipeline

```
PR Opened → LINT → TYPE CHECK → UNIT TESTS → BUILD → INTEGRATION → E2E → SECURITY AUDIT → BUNDLE SIZE → Ready
```

**No gate can be skipped.** If lint fails, fix lint — don't disable the rule.

## GitHub Actions Basic CI

```yaml
name: CI
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]
jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '22', cache: 'npm' }
      - run: npm ci
      - run: npm run lint
      - run: npx tsc --noEmit
      - run: npm test -- --coverage
      - run: npm run build
      - run: npm audit --audit-level=high
```

## Feeding CI Failures to Agents

```
CI fails → Copy failure output → Feed to agent → Agent fixes → Push → CI runs again
```

## Deployment Strategies

- **Preview deployments:** Every PR gets a preview environment
- **Feature flags:** Decouple deployment from release. Ship code disabled, enable when ready.
- **Staged rollouts:** Staging → Production (flag OFF) → Canary (5%) → Gradual (25% → 50% → 100%)
- **Rollback plan:** Every deployment must be reversible

## Environment Management

```
.env.example  → Committed (template)
.env          → NOT committed
CI secrets    → GitHub Secrets / vault
```

CI should never have production secrets.

## CI Optimization (when >10 minutes)

1. Cache dependencies
2. Run jobs in parallel (lint, typecheck, test, build)
3. Only run what changed (path filters)
4. Shard test suites across runners

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "CI is too slow" | Optimize the pipeline, don't skip it |
| "This change is trivial" | Trivial changes break builds. CI is fast for trivial changes. |
| "We'll add CI later" | Projects without CI accumulate broken states. Set up on day one. |

## Red Flags

- No CI pipeline in the project
- CI failures ignored or silenced
- Tests disabled in CI to make it pass
- No rollback mechanism
- Secrets in code, not secrets manager

## Verification

- [ ] All quality gates present (lint, types, tests, build, audit)
- [ ] Pipeline runs on every PR and push to main
- [ ] Failures block merge (branch protection)
- [ ] Secrets stored in secrets manager
- [ ] Deployment has rollback mechanism
- [ ] Pipeline runs under 10 minutes
