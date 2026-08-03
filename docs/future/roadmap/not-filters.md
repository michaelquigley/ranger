---
title: "\"not\" filters"
state: evaluating
created: 2026-08-03
tags: [feature]
milestone: v0.1.x
log:
  - stamp: 2026-08-03
    note: alt-click exclusion built — docs/journal/2026-08-03.md
---

alt-click a tag chip, subsystem chip, or milestone badge to filter it _out_. plain click still includes; clicking into the opposite polarity flips rather than stacking. excluded pills read `not x` in the header filter bar. soak question: does alt-click cover the filter-out need, or does exclusion eventually want a keyboard-free affordance too?

## background

the original spike — whether a shift-click or ctrl-click is even capturable — dissolved on inspection: react mouse events carry modifier booleans, and the chip handlers already receive the event. ctrl-click was rejected (macos context-menu gesture; browsers swallow the click), and alt-click won over shift-click, which conventionally means extend-selection. search negation was left out deliberately — that's query-syntax territory, not a chip gesture.
