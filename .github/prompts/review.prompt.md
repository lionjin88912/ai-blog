---
description: "Review before merge. Five-axis code review covering correctness, readability, architecture, security, and performance."
agent: "agent"
argument-hint: "Code or changes to review"
---

Conduct a thorough code review of the described changes.

Follow the `code-review-and-quality` skill. For simplification opportunities, also reference `code-simplification`. For security concerns, reference `security-and-hardening`. For performance concerns, reference `performance-optimization`.

## Process

1. Understand the context — what is this change trying to accomplish?
2. Review tests first — do they exist, test behavior, cover edge cases?
3. Review implementation across five axes:
   - **Correctness:** Matches spec? Edge cases? Error paths?
   - **Readability:** Clear names? Straightforward flow? No clever tricks?
   - **Architecture:** Follows patterns? Clean boundaries? Right abstraction level?
   - **Security:** Input validated? Secrets out? Auth checked? External data untrusted?
   - **Performance:** N+1 patterns? Unbounded ops? Missing pagination?
4. Categorize every finding: Critical / Required / Nit / Optional / FYI
5. Check change size (~100 lines ideal, ~1000+ must be split)
6. Verify the verification — tests run? Build pass?

## Output

A structured review with categorized findings and a verdict (Approve / Request Changes).
