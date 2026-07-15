---
name: security-and-hardening
description: "Hardens code against vulnerabilities. Use when handling user input, authentication, data storage, or external integrations. Use when building any feature that accepts untrusted data, manages sessions, or interacts with third-party services."
---

# Security and Hardening

## Overview

Treat every external input as hostile, every secret as sacred, every authorization check as mandatory. Security isn't a phase — it's a constraint on every line of code that touches user data.

## When to Use

- Building anything that accepts user input
- Implementing authentication or authorization
- Storing or transmitting sensitive data
- Integrating with external APIs
- Handling file uploads, webhooks, or callbacks

## The Three-Tier Boundary System

### Always Do
- Validate all external input at the system boundary
- Parameterize all database queries — never concatenate user input into SQL
- Encode output to prevent XSS (use framework auto-escaping)
- Use HTTPS for all external communication
- Hash passwords with bcrypt/scrypt/argon2
- Set security headers (CSP, HSTS, X-Frame-Options)
- Use httpOnly, secure, sameSite cookies
- Run `npm audit` before every release

### Ask First
- Adding new auth flows or changing auth logic
- Storing new categories of sensitive data
- Changing CORS configuration
- Adding file upload handlers
- Modifying rate limiting

### Never Do
- Commit secrets to version control
- Log sensitive data (passwords, tokens)
- Trust client-side validation as a security boundary
- Use `eval()` or `innerHTML` with user data
- Expose stack traces to users

## OWASP Top 10 Quick Reference

1. **Injection** — Parameterized queries, ORM with validated input
2. **Broken Authentication** — bcrypt (12+ rounds), httpOnly/secure/sameSite cookies
3. **XSS** — Framework auto-escaping, DOMPurify for raw HTML
4. **Broken Access Control** — Check authorization on every protected endpoint, not just authentication
5. **Security Misconfiguration** — helmet middleware, CSP, restricted CORS origins
6. **Sensitive Data Exposure** — Strip sensitive fields from API responses, use env vars for secrets

## Input Validation

Validate with schemas (e.g., Zod) at route handlers. Restrict file uploads by type and size.

## Rate Limiting

Apply general rate limits on `/api/` and stricter limits on `/api/auth/`.

## Secrets Management

```
.env.example  → Committed (template)
.env          → NOT committed (real secrets)
.gitignore must include: .env, .env.local, *.pem, *.key
```

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "This is an internal tool" | Internal tools get compromised. Attackers target the weakest link. |
| "We'll add security later" | Retrofitting is 10x harder. Add it now. |
| "The framework handles security" | Frameworks provide tools, not guarantees. |

## Red Flags

- User input passed directly to queries, shell commands, or HTML
- Secrets in source code or commit history
- API endpoints without auth checks
- Wildcard CORS origins
- No rate limiting on auth endpoints

## Verification

- [ ] `npm audit` shows no critical/high vulnerabilities
- [ ] No secrets in source code or git history
- [ ] All user input validated at system boundaries
- [ ] Auth and authorization checked on every protected endpoint
- [ ] Security headers present in response
- [ ] Error responses don't expose internals
- [ ] Rate limiting active on auth endpoints
