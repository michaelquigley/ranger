---
title: order.yaml -> .ranger/order.yaml
state: evaluating
created: 2026-08-05
tags: [enhancement]
milestone: v0.1.x
log:
  - stamp: 2026-08-14
    note: built — docs/journal/2026-08-14.md
---

now that we're storing filters in `.ranger/filters.yaml`, we might as well consolidate the `order.yaml` into `.ranger/order.yaml` also.

## realized

ranking moved to `docs/future/roadmap/.ranger/order.yaml`, beside the saved filters. `.ranger/` is ranger's own two files; the roadmap directory is items and assets again.

Existing roadmaps migrate themselves: `Load` notices a legacy `order.yaml` at the roadmap root and moves it into place once, before it reads. The move goes through `FinalizeLink` — the no-clobber funnel capture and rename already use — so the operator's bytes cross unchanged and a `.ranger/order.yaml` already in place wins atomically rather than by a check the next syscall could invalidate. This is the one write ranger performs without being asked for it; the alternative was a board that silently dropped the ranking of every roadmap written before the move.

## soak

- Is an unasked-for write at load time the right call in practice, or does it want to be visible somewhere — a board notice, a log line — rather than purely silent?
- `.ranger/` was defined as tool state and order.yaml was its explicit contrast case; the definition widened to "ranger's own files." Does that hold up as the directory gains anything else, or does the operator-authored/tool-rendered split want naming inside it?
