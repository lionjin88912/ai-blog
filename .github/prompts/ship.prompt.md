---
description: "Ship to production. Pre-launch checklist, feature flags, staged rollout, and rollback plan."
agent: "agent"
argument-hint: "Feature or release to ship"
---

Prepare the described feature or release for production deployment.

Follow the `shipping-and-launch` skill. Also reference `git-workflow-and-versioning` for commit discipline and `ci-cd-and-automation` for pipeline verification.

## Process

1. Run the pre-launch checklist:
   - Code quality: tests pass, build clean, lint clean, reviewed
   - Security: no secrets, audit clean, input validated, auth checked, headers set
   - Performance: Web Vitals good, no N+1, images optimized, bundle within budget
   - Accessibility: keyboard nav, screen reader, contrast, axe-core clean
   - Infrastructure: env vars set, migrations ready, logging configured, health check exists
2. Configure feature flag (if applicable)
3. Document the rollback plan (trigger conditions, steps, time-to-rollback)
4. Plan the staged rollout (staging → production flag OFF → team → canary 5% → gradual)
5. Set up monitoring dashboards (error rate, latency, business metrics)

## Output

A launch readiness report with checklist status, rollback plan, and rollout sequence.
