# Layout Proposal

## Core principle

The layout sets the width. Elements fill their container. No child element constrains its own width — only the `.page` container does.

Typography and spacing are already fluid (`clamp()` with `vi` units). The missing piece is the container itself: currently fixed at `720px`, `1100px`, or `1400px` regardless of screen size. Making the container fluid completes the chain — as the viewport grows, the container grows, and everything inside it (text, spacing, grids) grows proportionally.

## One container, one variable

Replace the three fixed `max-width` values with a single CSS variable that modifier classes override:

```css
:root {
    --page-width: clamp(320px, 55vi, 920px);
}

.page {
    max-width: var(--page-width);
    margin: 0 auto;
    padding: var(--page-pad-y) var(--page-pad-x) var(--page-pad-b);
}

.page--wide {
    --page-width: clamp(720px, 75vi, 1200px);
}

.page--gallery {
    --page-width: clamp(720px, 90vi, 1500px);
}
```

The default (55vi, caps at 920px) covers all single-column pages: homepage, about, CV, blog listing, projects, thoughts, styleguide. On a 1920px screen this gives ~1056px before the cap — nearly 50% more than the current 720px. On a 4K screen it reaches the 920px cap, which is still a comfortable reading measure because the type scale grows with it.

## Pages that use each width

**Default (`--page-width`):** Homepage, About, Blog listing, Projects, Thoughts, Photos listing, Styleguide, CV

**`--page-width` via `.page--wide`:** Blog articles (need room for the author/content/date grid)

**`--page-width` via `.page--gallery`:** Photo albums (need room for multi-column masonry)

## What changes

### Homepage

Currently uses a custom `.container` with `max-width: 450px` and inline styles. Change to use `.page` with the default width. The centred-in-viewport flexbox moves to the stylesheet. Hardcoded `padding: 2rem` becomes `var(--page-pad-y) var(--page-pad-x)`. The `font-size: 12px` on `ul` becomes `var(--text-size-sm)`.

The homepage content is sparse (name, one-liner, links) so it will feel more spacious on wide screens. This is intentional — the minimal aesthetic benefits from breathing room.

### Remove `max-width` from child elements

These currently constrain their own width, fighting the layout:

| Element | Current | Change |
|---|---|---|
| `.prose` | `max-width: 720px` | Remove — fills its grid column or `.page` container |
| `article .site-footer` | `max-width: 720px` | Remove — same as prose |
| `.lightbox-caption` | `max-width: 600px` | Remove — the lightbox content area constrains it |

### Article grid

Currently `grid-template-columns: 200px 1fr auto` on desktop. The `200px` author column is fixed — it should stay fixed (it's a sidebar, not content). The `1fr` prose column and `auto` date column grow with the container. No change needed here; removing `.prose`'s own `max-width` is enough.

### Lightbox

`.lightbox-content` and `.lightbox img` use `max-width: 90vw`. These are viewport-relative already and independent of the page container. No change needed.

## Token changes

### Remove

- `--text-size-lead` / `--scale-xl` — the lead paragraph concept is being dropped per the information architecture. Anywhere it was used becomes `--text-size-prose`.

### Add

- `--page-width` in `:root` — the default fluid container width

## What stays the same

- The two media query breakpoints (541px, 921px) — these control layout structure (columns, grid visibility), not sizing
- Icon sizes (`--text-size-icon`, `--text-size-icon-lg`) — UI chrome, fixed
- The lightbox overlay — already viewport-relative
- `em`-based relative sizes on inline `code` elements
- `margin: 0`, `padding: 0` resets

## Reading measure check

The concern with wider containers is that lines become too long to read comfortably. The standard target is 45-75 characters per line.

At the default width cap (920px) minus padding (~3.5rem each side = ~56px each side at 16px base), the content area is ~808px. With `--text-size-prose` resolving to ~1.18rem (~19px) on a wide screen, that gives roughly 65 characters per line in Source Serif 4. This is within the comfortable range.

For `.page--wide`, prose sits in the `1fr` grid column with a 200px sidebar, so the effective prose width is narrower than the full container — also within range.

## Implementation order

1. Add `--page-width` to `:root` and update `.page`, `.page--wide`, `.page--gallery`
2. Remove `max-width` from `.prose`, `article .site-footer`, `.lightbox-caption`
3. Remove `--text-size-lead` / `--scale-xl` and update references
4. Migrate homepage from custom `.container` + inline styles to `.page` + stylesheet classes
5. Test all page types at mobile, laptop, and ultrawide widths
