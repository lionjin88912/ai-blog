# Project Guidelines

## Development Lifecycle

This project uses structured skills for every phase of development. Skills activate automatically based on task context, or invoke them directly via `/` slash commands.

```
DEFINE → PLAN → BUILD → VERIFY → REVIEW → SHIP
/spec    /plan   /build  /test   /review  /ship
```

## Skills Reference

| Phase | Skill | Triggers |
|-------|-------|----------|
| Define | `idea-refine` | Brainstorming, exploring concepts |
| Define | `spec-driven-development` | New project/feature, unclear requirements |
| Plan | `planning-and-task-breakdown` | Breaking work into tasks |
| Build | `incremental-implementation` | Multi-file changes |
| Build | `test-driven-development` | Any logic, bug fix, or behavior change |
| Build | `context-engineering` | Session setup, quality degradation |
| Build | `source-driven-development` | Framework-specific code |
| Build | `frontend-ui-engineering` | UI components, layouts, state |
| Build | `api-and-interface-design` | API endpoints, type contracts |
| Verify | `browser-testing-with-devtools` | Browser-rendered output |
| Verify | `debugging-and-error-recovery` | Test failures, build breaks, unexpected errors |
| Review | `code-review-and-quality` | Before merging any change |
| Review | `code-simplification` | Reducing complexity without changing behavior |
| Review | `security-and-hardening` | User input, auth, data storage |
| Review | `performance-optimization` | Load times, response times, Core Web Vitals |
| Ship | `git-workflow-and-versioning` | Every code change |
| Ship | `ci-cd-and-automation` | CI pipelines, deployment |
| Ship | `deprecation-and-migration` | Removing old systems, migrating users |
| Ship | `documentation-and-adrs` | Architectural decisions, API docs |
| Ship | `shipping-and-launch` | Production deployments |

## Conventions

- Write tests before code (Red-Green-Refactor)
- Build in vertical slices — implement, test, verify, commit
- Commit early, commit often — atomic commits with descriptive messages
- Validate at boundaries — trust internal code, validate external input
- Measure before optimizing — profile first, fix what data shows

## Boundaries

### Always
- Run tests before committing
- Validate all user input at system boundaries
- Follow existing patterns in the codebase
- Keep commits atomic and focused (~100 lines)

### Ask First
- Adding new dependencies
- Changing database schema
- Modifying CI/CD configuration
- Changing authentication or authorization logic

### Never
- Commit secrets or credentials
- Disable tests to make CI pass
- Skip code review
- Use `eval()` or `innerHTML` with user data
- Force-push to shared branches
