# Testing Patterns Reference

Quick reference for common testing patterns. Use alongside the `test-driven-development` skill.

## Test Structure (Arrange-Act-Assert)

```typescript
it('describes expected behavior', () => {
  // Arrange: Set up test data and preconditions
  const input = { title: 'Test Task', priority: 'high' };

  // Act: Perform the action being tested
  const result = createTask(input);

  // Assert: Verify the outcome
  expect(result.title).toBe('Test Task');
  expect(result.status).toBe('pending');
});
```

## Naming Conventions

```typescript
// Pattern: [unit] [expected behavior] [condition]
describe('TaskService.createTask', () => {
  it('creates a task with default pending status', () => {});
  it('throws ValidationError when title is empty', () => {});
  it('trims whitespace from title', () => {});
});
```

## Common Assertions

```typescript
// Equality
expect(result).toBe(expected);            // Strict (===)
expect(result).toEqual(expected);         // Deep equality

// Truthiness
expect(result).toBeTruthy();
expect(result).toBeNull();
expect(result).toBeDefined();

// Numbers
expect(result).toBeGreaterThan(5);
expect(result).toBeCloseTo(0.3, 5);

// Strings / Arrays / Objects
expect(result).toMatch(/pattern/);
expect(array).toContain(item);
expect(array).toHaveLength(3);
expect(object).toHaveProperty('key', 'value');

// Errors
expect(() => fn()).toThrow(ValidationError);

// Async
await expect(asyncFn()).resolves.toBe(value);
await expect(asyncFn()).rejects.toThrow(Error);
```

## Mocking Patterns

```typescript
// Mock functions
const mockFn = jest.fn();
mockFn.mockReturnValue(42);
mockFn.mockResolvedValue({ data: 'test' });

// Mock modules
jest.mock('./database', () => ({
  query: jest.fn().mockResolvedValue([{ id: 1 }]),
}));
```

### Mock at Boundaries Only

```
Mock these:                    Don't mock these:
├── Database calls             ├── Internal utility functions
├── HTTP requests              ├── Business logic
├── File system operations     ├── Data transformations
├── External API calls         ├── Validation functions
└── Time/Date (when needed)    └── Pure functions
```

## React/Component Testing

```tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

it('submits the form with entered data', async () => {
  const onSubmit = jest.fn();
  render(<TaskForm onSubmit={onSubmit} />);

  // Find elements by accessible role/label (not test IDs)
  fireEvent.change(screen.getByRole('textbox', { name: /title/i }), {
    target: { value: 'New Task' },
  });
  fireEvent.click(screen.getByRole('button', { name: /create/i }));

  await waitFor(() => {
    expect(onSubmit).toHaveBeenCalledWith({ title: 'New Task' });
  });
});
```

## API / Integration Testing

```typescript
import request from 'supertest';
import { app } from '../src/app';

it('creates a task and returns 201', async () => {
  const response = await request(app)
    .post('/api/tasks')
    .send({ title: 'Test Task' })
    .set('Authorization', `Bearer ${testToken}`)
    .expect(201);

  expect(response.body).toMatchObject({
    id: expect.any(String),
    title: 'Test Task',
    status: 'pending',
  });
});
```

## E2E Testing (Playwright)

```typescript
import { test, expect } from '@playwright/test';

test('user can create and complete a task', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("New Task")');
  await page.fill('[name="title"]', 'Buy groceries');
  await page.click('button:has-text("Create")');
  await expect(page.locator('text=Buy groceries')).toBeVisible();
});
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|---|---|---|
| Testing implementation details | Breaks on refactor | Test inputs/outputs |
| Snapshot everything | No one reviews diffs | Assert specific values |
| Shared mutable state | Tests pollute each other | Setup/teardown per test |
| Testing third-party code | Wastes time | Mock the boundary |
| Skipping tests to pass CI | Hides real bugs | Fix or delete the test |
| Overly broad assertions | Misses regressions | Be specific |
| No async error handling | Swallowed errors | Always `await` async tests |
