---
name: api-and-interface-design
description: "Guides stable API and interface design. Use when designing APIs, module boundaries, or any public interface. Use when creating REST or GraphQL endpoints, defining type contracts, or establishing boundaries between frontend and backend."
---

# API and Interface Design

## Overview

Design stable, well-documented interfaces that are hard to misuse. Good interfaces make the right thing easy and the wrong thing hard. Applies to REST APIs, GraphQL schemas, module boundaries, component props, and any code-to-code surface.

## When to Use

- Designing new API endpoints
- Defining module boundaries or contracts
- Creating component prop interfaces
- Changing existing public interfaces

## Core Principles

### Hyrum's Law

Every observable behavior becomes depended on. Be intentional about what you expose. Don't leak implementation details.

### Contract First

Define the interface before implementing it. The contract is the spec.

### Consistent Error Semantics

Pick one error strategy and use it everywhere:

```typescript
interface APIError {
  error: {
    code: string;        // Machine-readable: "VALIDATION_ERROR"
    message: string;     // Human-readable
    details?: unknown;
  };
}
// 400 → Invalid data | 401 → Not authenticated | 403 → Not authorized
// 404 → Not found | 409 → Conflict | 422 → Validation failed | 500 → Server error
```

### Validate at Boundaries

Trust internal code. Validate at system edges where external input enters. Third-party API responses are untrusted data — validate their shape before use.

### Prefer Addition Over Modification

Extend interfaces with optional fields. Never change existing field types or remove fields.

### Predictable Naming

| Pattern | Convention |
|---------|-----------|
| REST endpoints | Plural nouns, no verbs: `GET /api/tasks` |
| Query params | camelCase: `?sortBy=createdAt` |
| Response fields | camelCase: `{ createdAt }` |
| Boolean fields | is/has/can prefix: `isComplete` |

## REST API Patterns

- Paginate list endpoints from the start
- Use PATCH for partial updates
- Use query parameters for filters
- Separate input types from output types

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "We'll document the API later" | The types ARE the documentation. Define them first. |
| "We don't need pagination for now" | You will at 100+ items. Add it from the start. |
| "Nobody uses that undocumented behavior" | Hyrum's Law: if observable, somebody depends on it. |

## Red Flags

- Endpoints returning different shapes depending on conditions
- Inconsistent error formats across endpoints
- Validation scattered throughout internal code
- List endpoints without pagination
- Verbs in REST URLs (`/api/createTask`)

## Verification

- [ ] Every endpoint has typed input and output schemas
- [ ] Error responses follow a single consistent format
- [ ] Validation happens at system boundaries only
- [ ] List endpoints support pagination
- [ ] New fields are additive and optional
- [ ] Naming follows consistent conventions
