---
name: performance-optimization
description: "Optimizes application performance. Use when performance requirements exist, when you suspect regressions, or when Core Web Vitals or load times need improvement. Use when profiling reveals bottlenecks."
---

# Performance Optimization

## Overview

Measure before optimizing. Performance work without measurement is guessing. Profile first, identify the actual bottleneck, fix it, measure again.

## When to Use

- Performance requirements exist (load time budgets, response time SLAs)
- Users or monitoring report slow behavior
- Core Web Vitals below thresholds
- Building features handling large datasets or high traffic

**When NOT to use:** Don't optimize before evidence of a problem exists.

## Core Web Vitals Targets

| Metric | Good | Poor |
|--------|------|------|
| LCP (Largest Contentful Paint) | ≤ 2.5s | > 4.0s |
| INP (Interaction to Next Paint) | ≤ 200ms | > 500ms |
| CLS (Cumulative Layout Shift) | ≤ 0.1 | > 0.25 |

## The Optimization Workflow

```
MEASURE → IDENTIFY → FIX → VERIFY → GUARD
```

### Step 1: Measure

**Frontend:** Lighthouse, DevTools Performance tab, web-vitals library
**Backend:** Response time logging, APM, database query timing

### Step 2: Identify the Bottleneck

**Frontend:** Slow LCP (large images, render-blocking resources), High CLS (images without dimensions), Poor INP (heavy JS on main thread), Slow initial load (large bundle)

**Backend:** Slow API responses (N+1 queries, missing indexes), Memory growth (leaked references), CPU spikes (synchronous computation)

### Step 3: Fix Common Anti-Patterns

- **N+1 Queries** → Use joins/includes instead of per-item queries
- **Unbounded Data Fetching** → Paginate with limits
- **Missing Image Optimization** → Add dimensions, lazy loading, responsive sizes, modern formats
- **Unnecessary Re-renders** → Stable references, React.memo for expensive components, useMemo for computations
- **Large Bundle** → Dynamic imports, route-level code splitting
- **Missing Caching** → Cache frequently-read data, HTTP cache headers for static assets

## Performance Budget

```
JS bundle: < 200KB gzipped (initial)
API response: < 200ms (p95)
Lighthouse Performance: ≥ 90
```

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "We'll optimize later" | Performance debt compounds. Fix obvious anti-patterns now. |
| "It's fast on my machine" | Profile on representative hardware and networks. |
| "This optimization is obvious" | If you didn't measure, you don't know. Profile first. |

## Red Flags

- Optimization without profiling data
- N+1 query patterns in data fetching
- List endpoints without pagination
- Images without dimensions, lazy loading, or responsive sizes
- Bundle size growing without review

## Verification

- [ ] Before and after measurements exist (specific numbers)
- [ ] Specific bottleneck identified and addressed
- [ ] Core Web Vitals within "Good" thresholds
- [ ] No N+1 queries in new data fetching code
- [ ] Existing tests still pass
