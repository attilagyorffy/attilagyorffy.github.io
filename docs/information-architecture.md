# Information Architecture

Element priority per page type, ordered from highest visual weight to lowest. The layout gives each element its space; elements fill what they're given.

## Homepage

1. Name (h1)
2. Bio text
3. Navigation links

Centred single-column. Minimal, no chrome. Everything scales with the container.

## Blog listing

1. Page title (h1)
2. Post titles (h3) — the primary scanning target
3. Post summaries — secondary, supports scan decision
4. Year headings (h2) — temporal grouping
5. Topic filter tags — navigation aid
6. Intro paragraph

Single-column card list grouped by year.

## Blog article

1. Article title (h1)
2. Section headings (h2) — scannability
3. Prose body — the content itself
4. Code blocks — technical evidence
5. Takeaway callout — key points summary
6. Highlighted text (mark) — emphasis within prose
7. Metadata: reading time, date, author, topic tags
8. Sidebar label — decorative

Uses a grid on desktop (author column | content | date). Prose fills the content column.

## CV

1. Name (h1)
2. Job titles (h2) — scanning target for recruiters
3. Role descriptions — what you did
4. Accomplishment lists (takeaway) — proof
5. Employment period and location — temporal context
6. Skills list — keywords

Single-column prose. Same structure as a blog article without the grid.

## About

1. Page title (h1)
2. Section headings (h2)
3. Prose body
4. Takeaway callout — principles list

Single-column prose. Identical layout to CV.

## Photos listing

1. Album cover thumbnails — visual-first scanning
2. Album titles (h2)
3. Album summary — context
4. Photo count (meta)
5. Intro paragraph

Card list with image previews. Each card is a visual unit.

## Photo album

1. Photos — the content itself, full visual weight
2. Album title (h1) — identification
3. Photo captions — per-image context
4. Intro paragraph — album context
5. Lightbox controls — navigation chrome

Multi-column photo grid. Images fill columns; captions are secondary.

## Projects listing

1. Project titles (h3) — scanning target
2. Project summaries — what it is
3. Category headings (h2) — grouping
4. Intro paragraph

Single-column card list grouped by category. Same structure as blog listing.

## Thoughts

1. Quote text (blockquote) — the content
2. Category headings (h2) — thematic grouping
3. Dates — temporal context
4. Intro paragraph

Single-column quote list grouped by theme.

## Styleguide

1. Section headings (h2) — navigation
2. Visual samples (swatches, type specimens, spacing bars)
3. Token names and values — reference data

Single-column reference page.

---

## Layout patterns

Looking across all pages, there are three structural patterns:

**Single-column prose** — Homepage, About, CV, Styleguide. Content flows in one column. The container width is the only constraint. Elements fill it.

**Single-column card list** — Blog listing, Projects, Thoughts, Photos listing. A sequence of cards, optionally grouped by headings. Cards fill the column width.

**Multi-column grid** — Blog article (author | content | date), Photo album (masonry columns). The grid defines the column widths; content fills each column.

The homepage is structurally single-column prose, just styled differently (centred vertically, monospace font, typing animation).

## Principle

The layout sets the width. Elements inside fill their container. No element sets its own `max-width` — only the outermost `.page` container does. This means changing the container width automatically adjusts everything inside it proportionally, because typography and spacing are already fluid.
