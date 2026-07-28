---
title: additional schema structures
state: researching
created: 2026-07-28
tags: [enhancement]
milestone: v0.1.x
log:
  - stamp: 2026-07-28
    note: design session — scope settled, no schema change; docs/journal/2026-07-28.md
---

the design session (2026-07-28) settled this at much smaller scope than the spike assumed: no new frontmatter fields, no parsing rules, no machine semantics anywhere. two changes.

document sections. the prompt is the unmarked default — everything above the first `##` heading. supporting material lives in named sections below it: `## why` for justification, `## background` for a longer description, vocabulary open. sections are conventional, never validated; a card carrying none is not malformed and the schema table is untouched. unlike tags they drive no colors and no grouping, so near-synonyms cost nothing and need no policing. every existing card stays valid unedited, and one-liners need no ceremony.

inbound contribution is triaged in the pr. parties with repo access contribute roadmap items by pull request, and the back-and-forth stays in the pr thread rather than accreting in the file — a file-resident thread would make every conversational turn a pull request. the merge is the triage decision, so a contributed card can land already in its lane rather than in inbox. `order.yaml` never appears in a contributor pr; the card lands unranked and ranking stays local and post-merge. the merged card carries `source: github:org/repo#N` so the shaping conversation stays reachable from the file. a teammate questioning an existing card uses the same gesture — a pr editing it.

both go into the grimoire's `software/conventions/roadmap-convention.md`. the inbound-pr mode is genuinely new there: the hard rules ("never commit roadmap changes") are written for an agent working in the operator's tree, and the note needs to say which mode its reader is in. the stock `AGENTS.md` paragraph, copied verbatim by adopting repos, takes the section split.

ranger's own change is one thing: the item modal gives the prompt visual primacy, so a long `## background` can't push it out of view on open. nothing else — a body-derived board marker would be a parsing rule over prose, which `richer-log` already rules out.

out of scope: spec and work-order pointers are frontmatter-shaped and belong to `specwork-order-management`; this card re-inflates if that one lands here. `richer-log` stays separate — the log spine stays sparse.
