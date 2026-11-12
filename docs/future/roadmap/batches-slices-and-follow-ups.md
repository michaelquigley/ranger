---
title: batches, slices and follow-ups
state: researching
created: 2026-11-12
tags: [feature, spike]
milestone: v0.1.x
---

Give the roadmap a way to say how cards relate when work is cut and grouped, so those relations stop living in prose and in the operator's head. Observed from an agent's side over a week of building metawoo mega through its roadmap; a note for when the operator decides what ranger should learn, not a design.

## what kept being improvised

- **An epic cut into slices.** Three "improvements" cards became six slice cards. The only record of which bullet went where was strikethrough text on the parent with "(moved to the X card)" after it, and the parents had to be walked by hand to see when they were fully carved.
- **Small cards run as one batch.** Four one-line cards were always going to be built together. Nothing on the board says so; their order in the lane is the only hint, and one of the four sat ten places below the others.
- **Ride-alongs.** A one-line fix left open by one batch was attached to the next card that touched the same surface, as a section added to that card. If that card's prompt had already gone out, the item was silently dropped and had to be noticed and moved again.
- **Follow-up cards.** Each larger batch left a small "X follow-ups" card. Its link back to the card it follows is its title and a sentence.
- **Cards that gate others.** A probe or a tool change in another repository had to land first; a rename had to precede two cards that would bake in a name. Ordering carried this, with the reason in a journal.
- **Built but not yet moved.** Cards stayed in building or researching after their work was committed, because moving them is a separate gesture from committing.

## questions

- Is the missing primitive a relation between cards (part of, follows, rides with, waits for), a grouping above the card (a batch), or only links plus a view that reads them? The existing `linking` card may already be most of it.
- Should a lane show a batch as one row that opens?
- Can a card know that the work it describes has been committed, without ranger learning about git?
