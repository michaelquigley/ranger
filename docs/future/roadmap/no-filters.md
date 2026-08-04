---
title: "\"no\" filters"
state: evaluating
created: 2026-08-04
tags: [feature]
milestone: v0.1.x
log:
  - stamp: 2026-08-04
    note: built as `no:` tokens in the search box, compiled to filter state — docs/journal/2026-08-04.md
---

filter cards where a dimension (tags, subsystems, milestone) is not specified at all. typed as `no:tags`, `no:subsystems`, or `no:milestone` in the search box — the token leaves the input and becomes a dashed filter pill; remaining text stays search. soak questions: does the typed gesture get discovered and remembered, and does the exact-token spelling (`no:` plus the frontmatter key) feel right in the fingers?

## background

the card's real question was where the gesture lives, since absence has no chip to click. three alternatives were weighed against the search-box idea: ghost chips on cards (consistent but permanent ink on every card), a standing header control or facet panel (discoverable but standing chrome), and a convention-first cut with no gesture at all (hand-authored filters.yaml only). michael called it — `no:` in the search box, but as *entry gesture, not home*: the token compiles into ordinary filter state rather than living in the search query, because absence lenses ("unmilestoned backlog") are exactly what saved filters exist for and the search query is deliberately unsaveable. absence is dimension-level, not value-level, occupies its dimension exclusively (displacement both ways), and rides filters.yaml as `no_*` booleans. a wider search-box query language (`tag:epic`) was deliberately left on the table as its own future card — the recognizer seam admits it without rework.
