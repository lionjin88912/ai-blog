---
name: shipping-and-launch
description: "Prepares production launches. Use when deploying to production. Use when you need a pre-launch checklist, monitoring setup, staged rollout plan, or rollback strategy."
---

# Shipping and Launch

## Overview

Ship with confidence. The goal is to deploy safely with monitoring in place, a rollback plan ready, and a clear definition of success. Every launch should be reversible, observable, and incremental.

## When to Use

- Deploying a feature to production
- Releasing a significant change
- Migrating data or infrastructure
- Any deployment that carries risk (all of them)

## Pre-Launch Checklist

### Code Quality
- [ ] All tests pass (unit, integration, e2e)
- [ ] Build succeeds with no warnings
- [ ] Lint and type checking pass
- [ ] Code reviewed and approved
- [ ] No TODO comments or `console.log` debugging in production code

### Security
- [ ] No secrets in code or version control
- [ ] `npm audit` clean
- [ ] Input validation on all user-facing endpoints
- [ ] Auth checks in place
- [ ] Security headers configured, CORS restricted

### Performance
- [ ] Core Web Vitals within "Good" thresholds
- [ ] No N+1 queries in critical paths
- [ ] Images optimized, bundle within budget
- [ ] Caching configured

### Accessibility
- [ ] Keyboard navigation works
- [ ] Screen reader support
- [ ] Color contrast meets WCAG 2.1 AA
- [ ] axe-core or Lighthouse clean

### Infrastructure
- [ ] Environment variables set in production
- [ ] Database migrations ready
- [ ] Logging and error reporting configured
- [ ] Health check endpoint exists

## Feature Flag Strategy

```
DEPLOY (flag OFF) → ENABLE for team → GRADUAL ROLLOUT (5% → 25% → 50% → 100%) → CLEAN UP flag
```

Every flag has an owner and expiration. Clean up within 2 weeks of full rollout.

## Staged Rollout

### Rollout Decision Thresholds

| Metric | Advance | Hold | Roll Back |
|--------|---------|------|-----------|
| Error rate | Within 10% baseline | 10-100% above | >2x baseline |
| P95 latency | Within 20% baseline | 20-50% above | >50% above |
| Business metrics | Neutral/positive | <5% decline | >5% decline |

### When to Roll Back Immediately

- Error rate >2x baseline
- P95 latency >50% increase
- User-reported issues spike
- Data integrity or security issues

## Monitoring

**Application:** Error rate, response time (p50/p95/p99), request volume, business metrics

**Client:** Core Web Vitals, JS errors, API errors from client perspective

**Post-launch (first hour):** Health check 200, error dashboard clean, latency normal, critical flow works, logs flowing.

## Rollback Strategy

Document before deploying:
- Trigger conditions
- Rollback steps (disable flag or deploy previous version)
- Database considerations
- Time to rollback (flag: <1 min, redeploy: <5 min, DB: <15 min)

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "It works in staging" | Production has different data, traffic, and edge cases |
| "We don't need feature flags" | Every feature benefits from a kill switch |
| "We'll add monitoring later" | You can't debug what you can't see. Add before launch. |
| "Rolling back is admitting failure" | Rolling back is responsible engineering |

## Red Flags

- Deploying without a rollback plan
- No monitoring in production
- Big-bang releases with no staging
- Feature flags with no owner or expiration
- "It's Friday afternoon, let's ship it"

## Verification

Before deploying:
- [ ] Pre-launch checklist completed
- [ ] Feature flag configured
- [ ] Rollback plan documented
- [ ] Monitoring dashboards set up

After deploying:
- [ ] Health check returns 200
- [ ] Error rate normal
- [ ] Latency normal
- [ ] Critical user flow works
- [ ] Rollback verified ready
