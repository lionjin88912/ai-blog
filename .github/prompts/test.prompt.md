---
description: "Prove it works. Writes tests, runs the suite, and verifies behavior."
agent: "agent"
argument-hint: "Code or feature to test"
---

Write tests for the described code or feature, then run them to verify correctness.

Follow the `test-driven-development` skill. For browser-rendered output, also follow `browser-testing-with-devtools`. For failures, follow `debugging-and-error-recovery`.

## Process

1. Identify what needs testing (new behavior, bug fix, changed functionality)
2. For bug fixes: write a reproduction test first (Prove-It Pattern)
3. Write tests following the test pyramid (80% unit, 15% integration, 5% E2E)
4. Use Arrange-Act-Assert structure, descriptive names, one assertion per concept
5. Run the full test suite and verify all pass
6. Report coverage if tracked

## Test Quality Checks

- Tests verify behavior, not implementation details
- Each test is self-contained (DAMP over DRY)
- Real implementations preferred over mocks (mock only at boundaries)
- No flaky tests (deterministic assertions, isolated state)
