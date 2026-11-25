---
title: field notes from mega's roadmap
state: researching
created: 2026-11-14
tags: [story, spike]
---

A running record of what working mega's roadmap shows about ranger and about roadmap and issue management in general. Michael, 2026-11-14: "we're also going to be using this as a study to make improvements in ranger and our roadmap (and issue) management in general." Kept by the planning agent on mega, appended as things come up. Each note says what happened and which card here it bears on. It records; it does not design.

`batches-slices-and-follow-ups` holds the first week's observations. These are what came after.

## 2026-11-13 and 14: a stocktake, an epic, and daily triage

- **An epic has nowhere to sit.** The install epic was written as a card that lists its stories by file name in prose. In a lane it is one row among its own stories. It was first put at the top of researching as a header, then parked in evaluating, where it reads as built. Bears on `linking` and `batches-slices-and-follow-ups`.
- **Renaming a card breaks its references.** The epic's file was renamed when a word changed. Four other files named it by slug and each had to be found and edited, and the journal's older entries now point at a name that no longer exists. Bears on `linking`.
- **A closed card leaves dead references behind.** A card removed after its work was done was still cited by `AGENTS.md` and a README two days later. A reviewer found it. Bears on `linking`.
- **A lane's order carries groups that exist nowhere.** Horizon was ordered as five groups: small and known, the direction, lifecycle, distribution, more than one host. The groups were in the message that proposed the order and nowhere on the board. Researching likewise has "before the demo" and "after" with no line between them.
- **Evaluating has exit conditions that live in heads.** "Soak until automated builds exist." "Close together with the fix that rewrites the same file." "Its real test is the next harness going through the same path." Each was said in conversation and written into the card's prose by hand.
- **A card waiting on a decision looks like any other.** Twice a card could not move until Michael decided something: an approval to download harness builds, and whether to add an approval step for project MCP servers. Nothing on the board marked either as waiting on him, and both were carried in the agent's replies until answered.
- **Quick captures arrive without a kind.** Six cards captured in a sentence or two came in untagged. Triage added the tag, a background note and a lane each time. Bears on `improved-capture-metadata`.
- **Two voices in one card.** Michael's capture is a line or two in his words. The planning agent then appends a "background" section, dated and signed, so the original stays as written. That convention appeared by itself and is now on a dozen cards.
- **A follow-ups card ages unevenly.** Review leftovers were bundled into one card of four unrelated items. One item then became low priority for a stated reason and the other three did not, and the card can only be ordered as a whole. Bears on `issues-or-punchlist`.
- **Findings became cards and notes at once.** A rehearsal produced three new cards and notes appended to two existing ones. The notes are as much its output as the cards and are invisible on the board.
- **Built but not moved, again.** Four cards sat in researching after their work was committed, until a pass over the board caught them. Already in `batches-slices-and-follow-ups`; it keeps happening.
- **An agent orders by editing the sidecar.** `ranger state` moves a card between lanes. Position within a lane has no command, so the agent rewrites `.ranger/order.yaml` with a script, while Michael drags cards in the board and sometimes has his own changes to the same file staged.
- **Who may move what was settled in words.** The agent may triage and order inbox, horizon and researching alone. Anything touching building or evaluating is discussed first. That rule lives in the agent's memory and nowhere ranger can see.
- **Cards in other repositories.** Work on mega produced cards for push, df, dfw and ranger. Each sits uncommitted in its own repository until Michael visits it. Bears on `cross-repo-aggregation`.
- **A lane's order means something different in each lane.** Researching is ordered by what to work next. Evaluating, Michael said on 2026-11-14, he keeps "in the rough order the items were implemented", newest at the top. Horizon is ordered in groups. Three lanes, three meanings of position, none of them stated on the board.
- **Where this is heading.** Michael, the same day: "i suspect the velocity we're developing with mega will result in a substantial overhaul of ranger, and possibly integrating ranger's new functionality directly into mega. or at least making them work together very well." Mega's card `mega-as-the-development-hub` names ranger inside mega as one of its pieces.
- **A card became an epic within the hour.** A capture about backing up harness configuration was written up as one small card. Two messages later it was the first layer of an epic about rebuilding a workspace elsewhere, and two older cards turned out to be other layers of the same thing. The small card had to be kept small on purpose, with a section saying what it was now part of, and the two older cards each got a note pointing at their new parent. All of that is prose. Bears on `linking` and `batches-slices-and-follow-ups`.

## 2026-11-15: a review of the board

- **An epic's list of its stories went stale in two days.** The install epic listed nine stories in prose. Two days later two were built, five new cards belonged to it, and the order had changed. Nothing flagged it; a review of the whole board found it, and the list was rewritten by hand into built, next and after. Bears on `linking`.
- **A review of the board is a recurring act with no support.** Three times in three days the whole roadmap was read end to end to answer the same questions: which cards are built but not moved, which are untagged, which have gone stale, what is next, what is waiting on Michael. Each time it was done by listing files and reading frontmatter. The questions are the same every time.
- **"Waiting on Michael" is its own list.** Each review ends with things only he can do: a pass on another machine, a question about his browser, a leftover to retire. They sit on no card that says so, and they are restated in conversation each time until done.

## 2026-11-16: checking that the roadmap is accurate

- **Lanes were right and contents were not.** Every card was in the correct lane. Ten cards said things that had stopped being true because other cards had been built: a requirement that no longer exists, a dependency now met, a finding since fixed, a measurement four days old. Each got a dated "where this stands" section by hand. A card has no way to know that something it mentions has changed.
- **A card's claims about another card are the fragile part.** "Depends on X", "written after Y", "the rehearsal found Z" all went stale the moment X, Y or Z moved. The claims that stayed true were the ones about the card's own subject.

## 2026-11-17

- **Closing two cards left references behind again.** The demo's card and the remote-access card were removed once done. The epic still listed both as stories, one as "next", and two other cards named them. The same thing happened on 2026-11-13 with a different card. Each time it was found by a later pass over the board, not at the moment of closing.
- **A new feature quietly depended on something an older card plans to remove.** The marker for the agent last instructed reads its time from the full scan of each log at start-up. A horizon card exists to stop that scan. Nothing connected the two until a review noticed, and the connection is now a paragraph on the older card.
- **The first pull request from outside brought issues the roadmap had no door for.** Its description listed five failing tests it had not fixed, two of them real product gaps. They arrived on the forge, in prose, on someone else's pull request. They reached the roadmap because the planning agent read the description and copied them onto an existing card. Bears on `issues-or-punchlist`: an outside report is neither a card nor yet an issue of ours, and the link from the card back to the pull request is a sentence.

## 2026-11-18: trimming the board to show it

Michael wanted the roadmap in its clearest shape for a presentation. The board went from 48 cards to 29 in one pass.

- **Three different ways a card left the board, and one act for all of them.** Twelve were closed after real use. Four were folded into another card. Three were taken off because the spec already holds the idea. One was done on the spot and closed. On disk each is the same thing: a deleted file. Only the journal says which was which.
- **Closing needs evidence, and the evidence was outside the board.** "Has this had enough use to close" was answered by reading installed versions, counting snapshots and recalling what the operator had confirmed. A card in evaluating does not know what it is waiting for or whether it has happened.
- **Twenty deletions meant rewording references in ten other files,** including the spec, the system description and a research report. A scan for backticked names with no card behind them found the rest. That scan is the tool this board has lacked every time a card was closed.
- **A fold loses the folded card's history unless it is written into the survivor by hand.** Each survivor got a dated section saying what joined it and why.
- **An epic's place was explained in words.** It sits in evaluating because external operators have built the thing several times while its last stories stay open. No lane says "the aim is being met and work remains".

## 2026-11-18: where ranger is headed, in Michael's words

From his presentation of mega, the seventh of seven future directions: "ranger grows an extensive metadata model. today ranger is just a roadmap board. future ranger manages roadmap epics, punchlists, defects, specs+work orders, and other structured project artifacts."

Everything on this card is evidence for that. The notes group under the things he names:

- **Epics.** An epic with nowhere to sit; its list of stories going stale in two days; a card becoming an epic within the hour; an epic's place in a lane explained in words.
- **Punchlists and defects.** A follow-ups card whose items age unevenly; an outside pull request whose five reported failures had no door into the roadmap.
- **Specs and work orders.** A new feature that leaned on something an older card plans to remove, with nothing connecting the two; a deferred idea that lives in the spec and on a card at once, until a trim had to say which.
- **Structure a card lacks.** No way to say what it waits for, who it waits on, or what it has been evidenced by; position meaning something different in each lane; closing, folding and removing all being the deletion of a file.

On the same day the same idea was recorded from mega's side: its hub epic lists this as the seventh direction and points here.

## 2026-11-19: a work order became a sequence of cards

Mega's simplification started as one spec section and one work order of seven stages. After eight review rounds in a day it was split into eleven cards under an epic, to be built one at a time.

- **An epic's stories have an order, and the board has nowhere to say it.** The sequence lives in three places kept in step by hand: a numbered list in the epic's text, each card's own "fourth in the sequence" line, and the cards' positions in the researching lane. Position in a lane was already carrying other meanings.
- **Two of the eleven are built on a branch and the rest on the main line.** That is a fact about where a card's work lands, said in prose on each card.
- **Each card names a failure it makes impossible.** It is the measure the operator took from an independent review, and every card now has a section for it under the same heading. It is structure by convention: nothing checks that a card has one.
- **A card can be a design to decide before anything is built,** sitting second in a sequence of things to build. The lane says researching for all eleven.
- **The split was a migration out of a document.** Stages of a work order became cards, with their gates, and the work order was left behind marked superseded. Specs, work orders and cards are three forms of one thing at different sizes, and moving between them was all hand work.
- **Dependencies between cards are sentences.** "Easier after", "before, because it removes a difference", "its shape depends on what is decided for". There are three kinds there at least, and none is a link.
