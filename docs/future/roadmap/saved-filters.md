---
title: saved filters
state: evaluating
created: 2026-08-03
tags: [feature]
milestone: v0.1.x
log:
  - stamp: 2026-08-04
    note: built as repo-published .ranger/filters.yaml — docs/journal/2026-08-04.md
  - stamp: 2026-08-04
    note: active pill now carries an accent highlight, derived by filter equality — docs/journal/2026-08-04.md
---

save the active board filter under a name; apply it back with one click from a header pill. the set lives in `docs/future/roadmap/.ranger/filters.yaml`, published with the roadmap so saved lenses travel with the cards. soak questions: does whole-set-replace feel right as the mutation shape, and do the header pills stay quiet enough as the set grows?

## background

the storage seam was the card's real spike: browser state (localStorage), url serialization, or a repo file. michael called it — a repo file, because saved filters want to be published into the actual work alongside the cards. that made this a server-side feature: a third document kind (the first tool-rendered one, re-emitted whole on save rather than surgically patched), a `filtersVersion` hash guard, and board-payload delivery so the pills ride the live-reload clock. an unreadable filters.yaml degrades rather than failing the board. the search query is deliberately not part of a saved filter.
