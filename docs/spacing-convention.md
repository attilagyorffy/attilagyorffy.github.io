# Spacing Convention

The site uses the existing `--space-*` scale for layout rhythm and keeps that
scale consistent across page chrome and prose.

## Flow rhythm

- Use bottom margins for default prose flow.
- Paragraphs, lists, code blocks, and metadata blocks own their trailing space.
- Do not stack a top margin on the next flow element unless it is a deliberate
  section break.

## Token usage

- `--space-xs` is for compact control spacing such as nav link padding and
  stacked navigation gaps.
- `--space-s` is for small callout padding and offset positioning.
- `--space-m` is for related content within the same block.
- `--space-l` is the default block rhythm for prose content.
- `--space-xl` is for separators that need extra breathing room.
- `--space-2xl` is for major page-region separation.

## Exceptions

- Decorative inline spacing can stay in `em` units when it needs to scale with
  the surrounding text.
- Border widths, radii, and fixed component sizes are not part of the spacing
  rhythm and can stay as raw values when they are not layout gaps.
