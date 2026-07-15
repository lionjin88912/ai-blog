---
name: test-driven-development
description: "Drives development with tests. Use when implementing any logic, fixing any bug, or changing any behavior. Use when you need to prove that code works."
---

# Test-Driven Development

## Overview

Write a failing test before writing the code that makes it pass. For bug fixes, reproduce the bug with a test first. Tests are proof — "seems right" is not done.

## When to Use

- Implementing any new logic or behavior
- Fixing any bug (the Prove-It Pattern)
- Modifying existing functionality
- Any change that could break existing behavior

**When NOT to use:** Pure configuration changes, documentation updates, or static content changes.

## The TDD Cycle

```
RED ──→ GREEN ──→ REFACTOR ──→ (repeat)
```

1. **RED:** Write a test that fails. A test that passes immediately proves nothing.
2. **GREEN:** Write the minimum code to make it pass. Don't over-engineer.
3. **REFACTOR:** With tests green, improve the code without changing behavior. Run tests after every refactor.

## The Prove-It Pattern (Bug Fixes)

```
Bug report → Write reproduction test → Test FAILS (bug confirmed)
→ Implement fix → Test PASSES (fix proven) → Run full suite (no regressions)
```

## The Test Pyramid

```
         /\          E2E (~5%)
        /  \         Integration (~15%)
       /    \        Unit (~80%)
      /______\
```

**The Beyonce Rule:** If you liked it, you should have put a test on it.

## Writing Good Tests

- **Arrange-Act-Assert** pattern for structure
- **One assertion per concept** — each test verifies one behavior
- **DAMP over DRY** — tests should read like specifications, duplication is OK
- **Test state, not interactions** — assert on outcomes, not method calls
- **Prefer real implementations** over mocks. Mock only at boundaries.
- **Name tests descriptively** — `it('sets completedAt when task is completed')`

## Test Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Testing implementation details | Test inputs and outputs, not internal structure |
| Flaky tests | Use deterministic assertions, isolate test state |
| Testing framework code | Only test YOUR code |
| Mocking everything | Prefer real implementations > fakes > stubs > mocks |
| No test isolation | Each test sets up and tears down its own state |

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'll write tests after the code works" | You won't. And post-hoc tests test implementation, not behavior. |
| "This is too simple to test" | Simple code gets complicated. The test documents expected behavior. |
| "Tests slow me down" | Tests slow you now, speed you up on every future change. |
| "I tested it manually" | Manual testing doesn't persist. |

## Red Flags

- Writing code without corresponding tests
- Tests that pass on first run (may not test what you think)
- Bug fixes without reproduction tests
- Skipping tests to make the suite pass

## Verification

- [ ] Every new behavior has a corresponding test
- [ ] All tests pass
- [ ] Bug fixes include a reproduction test that failed before the fix
- [ ] Test names describe the behavior being verified
- [ ] No tests were skipped or disabled
