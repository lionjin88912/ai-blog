---
name: idea-refine
description: "Refines ideas iteratively through structured divergent and convergent thinking. Use when you have a rough concept that needs exploration, when brainstorming, or when stress-testing a plan."
---

# Idea Refine

## Overview

Turn vague ideas into sharp, actionable concepts through three phases: expand, evaluate, converge. Push toward the simplest version that solves the real problem.

## When to Use

- You have a rough concept that needs exploration
- Stress-testing an existing plan
- Starting a new project and need clarity on direction

**When NOT to use:** Requirements are already clear and specific.

## Process

### Phase 1: Understand & Expand (Divergent)

1. **Restate the idea** as a "How Might We" problem statement
2. **Ask 3-5 sharpening questions:**
   - Who is this for, specifically?
   - What does success look like?
   - What are the real constraints?
   - What's been tried before?
   - Why now?
3. **Generate 5-8 idea variations** using these lenses:
   - Inversion: "What if we did the opposite?"
   - Constraint removal: "What if budget/time weren't factors?"
   - Simplification: "What's the version that's 10x simpler?"
   - Combination: "What if we merged this with [adjacent idea]?"

### Phase 2: Evaluate & Converge

1. **Cluster** resonant ideas into 2-3 distinct directions
2. **Stress-test** each against:
   - **User value:** Painkiller or vitamin?
   - **Feasibility:** What's the hardest part?
   - **Differentiation:** What makes this genuinely different?
3. **Surface hidden assumptions** — what you're betting is true but haven't validated

Be honest, not supportive. Push back on weak ideas with specificity.

### Phase 3: Sharpen & Ship

Produce a markdown one-pager:

```markdown
# [Idea Name]

## Problem Statement
[One-sentence "How Might We" framing]

## Recommended Direction
[The chosen direction and why]

## Key Assumptions to Validate
- [ ] [Assumption — how to test it]

## MVP Scope
[Minimum version that tests the core assumption]

## Not Doing (and Why)
- [Thing] — [reason]
```

The "Not Doing" list is the most valuable part. Focus is about saying no.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "Let's just start building" | Clarity before code prevents rework |
| "This idea is obviously good" | Every good idea has hidden assumptions worth surfacing |
| "We need more variations" | 5-8 considered variations beat 20 shallow ones |

## Red Flags

- Skipping the "who is this for" question
- No assumptions surfaced before committing
- Yes-machining weak ideas
- Jumping to output without exploring alternatives

## Verification

- [ ] Clear "How Might We" problem statement exists
- [ ] Target user and success criteria defined
- [ ] Multiple directions explored
- [ ] Hidden assumptions listed with validation strategies
- [ ] "Not Doing" list makes trade-offs explicit
- [ ] User confirmed direction before implementation
