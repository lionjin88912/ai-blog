# Accessibility Checklist

Quick reference for WCAG 2.1 AA compliance. Use alongside the `frontend-ui-engineering` skill.

## Essential Checks

### Keyboard Navigation
- [ ] All interactive elements focusable via Tab
- [ ] Focus order follows visual/logical order
- [ ] Focus is visible (outline/ring on focused elements)
- [ ] Custom widgets have keyboard support (Enter, Escape)
- [ ] No keyboard traps
- [ ] Skip-to-content link at top of page
- [ ] Modals trap focus while open, return focus on close

### Screen Readers
- [ ] All images have `alt` text (or `alt=""` for decorative)
- [ ] All form inputs have associated labels (`<label>` or `aria-label`)
- [ ] Buttons and links have descriptive text (not "Click here")
- [ ] Icon-only buttons have `aria-label`
- [ ] Page has one `<h1>`, headings don't skip levels
- [ ] Dynamic content announced (`aria-live` regions)
- [ ] Tables have `<th>` headers with scope

### Visual
- [ ] Text contrast ≥ 4.5:1 (normal) or ≥ 3:1 (large, 18px+)
- [ ] UI components contrast ≥ 3:1 against background
- [ ] Color is not the only way to convey information
- [ ] Text resizable to 200% without breaking layout
- [ ] No content flashing > 3 times per second

### Forms
- [ ] Every input has a visible label
- [ ] Required fields indicated (not by color alone)
- [ ] Error messages specific and associated with the field
- [ ] Error state visible by more than color (icon, text, border)
- [ ] Known fields use autocomplete (`type="email" autocomplete="email"`)

### Content
- [ ] Language declared (`<html lang="en">`)
- [ ] Page has a descriptive `<title>`
- [ ] Links distinguish from surrounding text (not by color alone)
- [ ] Touch targets ≥ 44x44px on mobile

## Common HTML Patterns

### Buttons vs. Links
```html
<button onClick={handleDelete}>Delete Task</button>   <!-- Actions -->
<a href="/tasks/123">View Task</a>                     <!-- Navigation -->
<div onClick={handleDelete}>Delete</div>               <!-- BAD -->
```

### Form Labels
```html
<label htmlFor="email">Email address</label>
<input id="email" type="email" required />

<input type="search" aria-label="Search tasks" />      <!-- Hidden label -->
```

### ARIA Roles
```html
<nav aria-label="Main navigation">...</nav>
<div role="status" aria-live="polite">Task saved</div>
<div role="alert">Error: Title is required</div>
<dialog aria-modal="true" aria-labelledby="dialog-title">
  <h2 id="dialog-title">Confirm Delete</h2>
</dialog>
<div aria-busy="true" aria-label="Loading tasks"><Spinner /></div>
```

## ARIA Live Regions

| Value | Behavior | Use For |
|-------|----------|---------|
| `aria-live="polite"` | Announced at next pause | Status updates, confirmations |
| `aria-live="assertive"` | Announced immediately | Errors, time-sensitive alerts |
| `role="status"` | Same as `polite` | Status messages |
| `role="alert"` | Same as `assertive` | Error messages |

## Testing Tools

```bash
npx axe-core           # Programmatic accessibility testing
npx pa11y              # CLI accessibility checker
# Chrome DevTools → Lighthouse → Accessibility
# Chrome DevTools → Elements → Accessibility tree
# macOS: VoiceOver (Cmd+F5) | Windows: NVDA | Linux: Orca
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|---|---|---|
| `div` as button | Not focusable, no keyboard | Use `<button>` |
| Missing `alt` text | Images invisible to readers | Add descriptive `alt` |
| Color-only states | Invisible to color-blind users | Add icons, text, patterns |
| Autoplaying media | Disorienting | Add controls, don't autoplay |
| Custom dropdown, no ARIA | Unusable by keyboard/reader | Use `<select>` or ARIA listbox |
| Removing focus outlines | Users can't see focus | Style outlines, don't remove |
| Empty links/buttons | No description announced | Add text or `aria-label` |
| `tabindex > 0` | Breaks natural tab order | Use `0` or `-1` only |
