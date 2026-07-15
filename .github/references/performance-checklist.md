# Performance Checklist

Quick reference for web application performance. Use alongside the `performance-optimization` skill.

## Core Web Vitals Targets

| Metric | Good | Needs Work | Poor |
|--------|------|------------|------|
| LCP (Largest Contentful Paint) | ≤ 2.5s | ≤ 4.0s | > 4.0s |
| INP (Interaction to Next Paint) | ≤ 200ms | ≤ 500ms | > 500ms |
| CLS (Cumulative Layout Shift) | ≤ 0.1 | ≤ 0.25 | > 0.25 |

## TTFB Diagnosis

When TTFB > 800ms, check DevTools Network waterfall:

- [ ] **DNS** slow → `<link rel="dns-prefetch">` or `<link rel="preconnect">`
- [ ] **TCP/TLS** slow → enable HTTP/2, edge deployment, keep-alive
- [ ] **Server** slow → profile backend, check slow queries, add caching

## Frontend Checklist

### Images
- [ ] Modern formats (WebP, AVIF)
- [ ] Responsive sizes (`srcset` and `sizes`)
- [ ] Explicit `width` and `height` (prevents CLS)
- [ ] Below-fold: `loading="lazy"` and `decoding="async"`
- [ ] LCP image: `fetchpriority="high"`, no lazy loading

### JavaScript
- [ ] Bundle < 200KB gzipped (initial load)
- [ ] Code splitting with dynamic `import()` for routes
- [ ] Tree shaking enabled (ESM, `sideEffects: false`)
- [ ] No blocking JS in `<head>` (use `defer` or `async`)
- [ ] Long tasks (> 50ms) broken up for main thread availability
- [ ] `scheduler.yield()` or `yieldToMain` in long-running loops
- [ ] Non-critical work deferred out of event handlers
- [ ] Third-party scripts: `async`/`defer`, audited for size

### CSS
- [ ] Critical CSS inlined or preloaded
- [ ] No render-blocking CSS for non-critical styles
- [ ] No CSS-in-JS runtime cost in production (use extraction)

### Fonts
- [ ] 2–3 families, 2–3 weights each
- [ ] WOFF2 only, self-hosted when possible
- [ ] LCP-critical fonts preloaded
- [ ] `font-display: swap` (or `optional` for non-critical)
- [ ] Subsetted via `unicode-range`
- [ ] Fallback metrics adjusted (`size-adjust`, `ascent-override`)

### Network
- [ ] Static assets: long `max-age` + content hashing
- [ ] API responses cached (`Cache-Control`)
- [ ] HTTP/2 or HTTP/3 enabled
- [ ] `<link rel="preconnect">` for known origins
- [ ] No unnecessary redirects

### Rendering
- [ ] No layout thrashing (forced synchronous layouts)
- [ ] Animations use `transform` and `opacity` (GPU-accelerated)
- [ ] Long lists virtualized (e.g., `react-window`)
- [ ] `content-visibility: auto` for off-screen sections
- [ ] No `unload` handlers or `Cache-Control: no-store` on HTML (preserves bfcache)

## Backend Checklist

### Database
- [ ] No N+1 query patterns (use eager loading / joins)
- [ ] Queries have appropriate indexes
- [ ] List endpoints paginated (never unbounded `SELECT *`)
- [ ] Connection pooling configured
- [ ] Slow query logging enabled

### API
- [ ] Response times < 200ms (p95)
- [ ] No synchronous heavy computation in request handlers
- [ ] Bulk operations instead of loops
- [ ] Response compression (gzip/brotli)
- [ ] Appropriate caching (in-memory, Redis, CDN)

### Infrastructure
- [ ] CDN for static assets
- [ ] Server close to users (or edge deployment)
- [ ] Horizontal scaling configured
- [ ] Health check endpoint for load balancer

## Measurement

```bash
npx lighthouse https://localhost:3000 --output json     # Lighthouse CLI
npx webpack-bundle-analyzer stats.json                  # Bundle analysis
npx vite-bundle-visualizer                              # Vite bundle analysis
```

```typescript
import { onLCP, onINP, onCLS } from 'web-vitals';
onLCP(console.log); onINP(console.log); onCLS(console.log);
```

## Anti-Patterns

| Anti-Pattern | Impact | Fix |
|---|---|---|
| N+1 queries | Linear DB load growth | Use joins, includes, batch loading |
| Unbounded queries | Memory exhaustion | Always paginate, add LIMIT |
| Missing indexes | Slow reads at scale | Add indexes for filtered/sorted columns |
| Layout thrashing | Jank, dropped frames | Batch DOM reads, then batch writes |
| Unoptimized images | Slow LCP | WebP, responsive sizes, lazy load |
| Large bundles | Slow TTI | Code split, tree shake, audit deps |
| Blocking main thread | Poor INP | Chunk long tasks with `scheduler.yield()` |
| Memory leaks | Growing memory | Clean up listeners, intervals, refs |
