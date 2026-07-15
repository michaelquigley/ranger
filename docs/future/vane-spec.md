---
title: vane
tagline: your roadmap lives in your repo
status: draft
created: 2026-07-13
---

# vane

*your roadmap lives in your repo*

## The Thesis

A project's roadmap is intent, and intent belongs in the same substrate as everything else the project owns: plain files, in git, in the working tree. Forge-hosted project boards and issue trackers put that intent behind somebody else's data model and somebody else's authorization scheme. The board form itself works — it remains a great way to get up to speed on where a project sits — but maintaining it in a separate system is friction that compounds: when the work happens *in* the repos, every round-trip to the forge to create an issue, place it on a board, and keep it current is a context switch the practice increasingly declines to pay, and the boards languish. The same separation stops being tolerable at all the moment agents join the practice. An agent working in a repository already has the one interface that matters: the filesystem. It can read a markdown file, edit a markdown file, and leave the result for review, with zero ceremony. Putting the roadmap behind a forge API means every agent needs credentials, an MCP bridge, and an authorization story, to do something it could otherwise do with `cat`.

vane is two things, and the order matters. First, a convention: a strongly-defined structure for roadmap items as frontmatter-markdown files living in `docs/future/roadmap/`. Second, a tool: a locally-runnable Go binary that reads that structure and presents richer views of it — a CLI for capture and a localhost web UI for the board. The convention is the product; the tool is a reader. Any party that can touch files — Michael in an editor, an agent in a harness, Obsidian, `grep` — is a first-class participant in the roadmap, and vane's UI is just the most comfortable chair in the room.

This replaces the forge project board and the roadmap-shaped use of the forge issue tracker. The issue tracker survives in a demoted role: an inbox where outside users start conversations, from which roadmap-shaped material is manually pulled across.

## What Already Exists Around It

vane lands inside an established practice. Project repos carry a `docs/current/` and `docs/future/` split — reality versus intent. The design-build pipeline moves significant work through four phases: a design session produces a spec, a planning agent grounds it into a work order, mercurius reviews the pair, an implementation agent realizes them, with terminus gating the code. Specs and work orders live in `docs/future/` while in flight and are removed when realized, their value living on in code and `docs/current/`, their archaeology in git history.

What the practice does *not* have is a defined shape for everything upstream of the spec. Deferred work orders, vision notes, roadmap items, and someday-ideas pile up across repos in ad-hoc forms. vane defines that upstream layer — the pool the pipeline draws from — without disturbing the pipeline itself. Specs and work orders remain exactly what they are.

## The Item

One type. The pile of genres that accumulates in `docs/future/` folders — deferred work orders, vision notes, roadmap entries, someday-maybes — turns out to be states of a single thing, not different things. vane calls that thing an **item**: an atomic, roadmap-grain statement of intent.

An item is one markdown file in `docs/future/roadmap/`. Frontmatter carries the machine-readable spine; the body carries whatever prose exists — a single line for a raw idea, pages for a matured framing. The format is deliberately the practice's lingua franca: Obsidian reads it, agents read it, the tool parses it, diffs of it are human-reviewable.

```yaml
---
title: short imperative name
state: inbox | horizon | researching | building | evaluating | done | dropped
created: 2026-07-13
tags: [optional, soft, grouping]
source: optional provenance, e.g. github:michaelquigley/zrok#412
log:
  - stamp: 2026-07-13
    note: optional dated event stamps — the history spine
---

body prose, at whatever weight the idea currently has.
```

The schema, normatively:

| field     | type                            | required |
| --------- | ------------------------------- | -------- |
| `title`   | string                          | yes      |
| `state`   | one of the seven states below   | yes      |
| `created` | date                            | yes      |
| `tags`    | list of strings                 | no       |
| `source`  | string                          | no       |
| `log`     | list of `{stamp: date, note: string}` | no       |

Dates — `created` and `stamp` — are always `YYYY-MM-DD`.

*Malformed* means any document that fails this table: frontmatter that doesn't parse, a duplicate claimed key, a missing required field, or any claimed field — required or optional — that violates its declared shape. Malformed is the condition that flags a card (see the seam census), and a flagged card whose `state:` can't be read appears in the **inbox** lane — inbox is already the lane that means "needs the operator's eye," and a broken file is precisely that. Nothing valid enough to be a file is ever silently omitted from the board. Unknown fields are accepted, preserved byte-for-byte by surgical edits, and ignored by readers: the convention claims the fields above and stays out of everyone else's.

The filename is the slug of the title: `retry-semantics.md`. The slug rule is fixed so every writer — tool, agent, human — derives the same name, and it is ASCII-mechanical, no Unicode tables involved: map `A`–`Z` to `a`–`z`; keep `a`–`z`, `0`–`9`, space, and hyphen; discard every other code point; convert spaces to hyphens; collapse runs of hyphens; trim hyphens from the ends. `Retry Semantics (v2)` → `retry-semantics-v2.md`, and — normative vector for the discard rule — `naïve K-scale` → `nave-scale.md`: the `ï` and the Kelvin-sign `K` are *discarded*, never case-mapped into ASCII, because Unicode-aware lowercasing keeps what ASCII discards and that divergence is exactly what this rule exists to prevent. A title that reduces to nothing gets no derived filename — the writer picks one by hand. One collision rule guards every path into `roadmap/` — retitle-rename and new-item capture alike: landing on a filename that already exists fails rather than clobbers, the in-flight content survives (the temp file keeps everything typed), and both paths are reported so the operator resolves the collision, usually by retitling the newcomer. The title is the truth and the filename is derived from it, so a file whose name isn't the slug of its title is a *detectable* condition — any reader can run the rule — and the tool surfaces it as a visibly flagged card, repaired by one rename gesture, never silently "fixed." Capture writes to a temporary file first and renames into place once a title exists, so the slug rule never gets in the way of speed.

The `log:` field is the item's history spine: optional event stamps for the few moments worth keeping in the working tree — chiefly a spec drawing on the item (see below). Each entry is a `stamp:` date and a one-line `note:`, and nothing more; the shape is deliberately too small to grow a changelog in. A stamp records that something happened on a date and is never edited afterward, so it cannot drift. The spine is deliberately sparse and is not a changelog; git history remains the archaeology. An item with no log is a perfectly healthy item.

Items do not nest. If a cluster of items wants to travel together, that is either what `tags` are for (soft grouping) or what spawning a spec is for (hard consolidation). Nesting is where simple models go to die, and vane declines the invitation.

## The Lifecycle

Seven states, five of them lifted directly from years of forge-board practice, two added to name what that practice was missing. The lifecycle is **descriptive, not enforced**: any state may move to any state, in the UI as in the files, and the named paths below are the expected weather, not walls. The substrate can't hold a transition graph anyway — anyone with an editor can set `state:` to anything — so the tool doesn't pretend to one; the judgment gate for a nonsensical transition is the same as for every other edit, the diff and the operator. "Terminal" below means *no expected onward state*, not *locked*.

- **inbox** — untriaged. Ideas, drive-by captures, material pulled from forge issues. Getting something into the inbox should cost nothing; that is the CLI's first job.
- **horizon** — triaged and deliberately at rest. This is the state the old boards lacked, and its absence is why vision notes and deferred work orders piled up as orphaned documents. A horizon item is not backlog that failed — its job is to sit still. Vision-register material lives here, possibly forever, legitimately.
- **researching** — being shaped: thinking underway, design sessions happening, a spec possibly being drawn from it.
- **building** — implementation in flight. The pipeline's artifacts (spec, work order, terminus reviews) carry the fine grain of execution; the item's state carries the one-glance weather.
- **evaluating** — built, and being lived with. Evaluation here means soak: running the thing in the practice to see whether it is actually what was wanted. It exits forward to done when it holds, or backward to researching (or out to dropped) when it doesn't.
- **done** — realized.
- **dropped** — declined.

Done and dropped are terminal states, not deletions. Marking an item done is a state transition; removing the file is a separate, deliberate curation act by the operator, probably batched. The tool never deletes. The board gets a visible shipped lane for as long as that's useful, the working tree never accumulates permanent residue, and git history remains the archaeology for anything removed — the same discipline the pipeline already applies to realized specs.

## Items and Specs: Spawn, Not Grow

A spec can be born from one or more items, so an item cannot *become* a spec — there is no single file to grow. The relation lives in links, not containment, and the link is unidirectional: when the design phase picks up a thread, the resulting spec records which items fed it in a `sources:` frontmatter list — repo-relative item paths, e.g. `docs/future/roadmap/retry-semantics.md` — and those items (typically) take the researching state. Items carry no pointer back — a maintained two-ended link is two chances to drift. And `sources:` itself is **provenance, not a live link**: true when the spec was born, never maintained afterward, exactly like the `source:` line an item carries for a forge issue. If an item retitles after a spec draws on it, the recorded path goes quietly historical, and git carries the archaeology — nobody promises otherwise, so nothing is broken.

What the item does take is a stamp in its log — `stamp: 2026-08-02`, `note: spec drawn — docs/future/frame-composition-spec.md`. The stamp matters because specs are ephemeral — a realized spec is removed from `docs/future/`, and its `sources:` list leaves the working tree with it, while the item lives on through building and evaluating. The stamp is the trace that survives, and it keeps vane's reader scoped to `roadmap/` alone. It is *human-visible* provenance — a line on the card for the operator's eye — and v1 assigns no machine-derived semantics to log notes: if the board ever needs to mechanically know that an item feeds a spec, that is a future field earning its way into the schema, not a parsing rule over prose. The item remains the roadmap-grain record; the spec is the design-grain artifact; neither pretends to be the other.

## The Write Model

vane's tool is read-only by default and read-write by intent, and the write model is the load-bearing safety property of the whole design: **the working tree is the write buffer, and git is the judgment gate.** Every edit the tool makes — a state flip, a reorder, a body edit — is a file write into the working tree, and nothing more. The tool never commits, never pushes, never touches git at all. The uncommitted diff *is* the pending-review queue, and the operator (or sexton, at the operator's direction) decides what enters history.

This is the same safeguard the grimoire relies on — nothing settles into the durable record without a judgment gate — implemented mechanically by infrastructure that already exists. It is also the answer to the agent question: agents can participate in the roadmap freely, because participation produces reviewable diffs, not committed facts.

Edits are surgical. A gesture changes the lines that express it and leaves every other byte alone — no field reordering, no requoting, no comments dropped by a decode-and-encode cycle. A state flip that arrives as a twenty-line reformat defeats the reviewable-diff thesis as surely as a hidden commit would. The work order should carry round-trip tests — state, title, and body edits against hand-formatted item files, and reorder and prune gestures against hand-formatted `order.yaml` fixtures with comments — verifying every gesture alters only the lines that express it.

One mechanical guard rides along: every read that paints a view delivers each file's content hash alongside its content, every mutation carries those hashes back, and the tool compares them against the disk before writing; a mismatch refuses the write and reloads instead. The hash the guard checks is always the one from the read that produced the state the operator acted on — never a fresh re-read at write time, which would only compare the disk to itself and wave everything through. (This is ordinary optimistic concurrency, ETag/If-Match shaped, spelled out because the server is otherwise stateless.) Absence is a version too: a read that finds no `order.yaml` reports an explicit absent sentinel, and a mutation whose expected state is *absent* succeeds only by creating the file no-clobber — if another writer created it first, the creation fails and reloads, exactly like any other hash mismatch. The working tree is a shared buffer — agents and editors write into it while the board is open — and the guard keeps the tool from writing blind over a change it hasn't seen. It is best-effort conflict *detection*, not a lock: a writer can still slip into the microseconds between check and write, and the design accepts that, because the deep safety net was never this check — it's the write model itself. Only uncommitted working-tree state is ever at risk, and git is the judgment gate. For the gestures that touch two files (transition-and-place, retitle-of-ranked), the guard preflights both hash checks before the first write, and any partial failure is reported plainly and followed by a reload — a half-applied gesture is benign under the git gate, but the UI shouldn't leave the operator guessing about it.

## Ordering

Lanes are ordered; priority is real. The order lives in a sidecar file, `docs/future/roadmap/order.yaml`: per-lane lists of item filenames, top to bottom. Reordering a lane is reordering lines in a list — a gesture every hand, human or agent or tool, already knows how to make — and it touches exactly one file no matter how many items move.

```yaml
researching:
  - retry-semantics.md
  - board-capture.md
horizon:
  - frame-composition.md
```

The shape is the whole convention: a mapping of lane (state) names to ordered filename lists. The file is optional — absent means every item is unranked, which is exactly the state of a fresh repo; capture never creates it. A duplicate filename within the file: first occurrence wins, later ones are ignored. A key that isn't a state name: ignored. A duplicate *lane key*, though, makes the file unreadable — YAML loaders disagree about which mapping survives, so guessing isn't self-healing, it's coin-flipping — and an unreadable `order.yaml` is a repository-level fail-fast (see the seam census): loud, at startup, fix the file and go. Entries are single lines, same tier: a filename is one line of text, and an entry written as a multi-line scalar makes the file unreadable — every surgical operation targets entry lines, and bytes that span lines would strand under a reorder or prune, turning the tool's own write into corruption. And the rules apply in a fixed order — invalid entries are discarded first (files that don't exist, unknown lane keys, items listed under a lane they're not in), *then* first-occurrence-wins runs among what survives — so a stale line can never shadow a valid one. Everything discarded is pruned opportunistically, like every other stale reference below. Pruning demands positive evidence, though: a lane-mismatch entry is discarded only when the item exists *and parses with a valid, different state*. An entry naming a file that exists but can't currently be read is retained — a half-saved editor write is not proof of anything, and a rank shouldn't die in the race window of someone else's save. Only entries for files that don't exist at all prune unconditionally.

Malformed cards get deterministic positions too, since the board promises to show them. A card whose state can't be read belongs to inbox *for all ordering purposes*; any entry retained for it in another lane is inert — held on disk, participating in no duplicate resolution, shadowing nothing. And a card without a valid `created` sorts after every dated unranked card, by filename ascending. Every card the board displays has exactly one computable position.

An earlier draft carried ordering as a per-item rank field with sparse fractional keys. It worked on paper and cost too much in practice-shaped ways: the key grammar was the one part of the convention nobody could operate by hand without instruction, and the order itself was illegible — scattered across N files, reconstructible only by collecting and sorting. The sidecar inverts both: nothing to teach, and the order is readable at a glance.

The semantics are self-healing, because the sidecar is a reference and references drift. An entry naming a file that doesn't exist, or an item no longer in that lane, is ignored — and pruned opportunistically the next time the tool writes the file. An item absent from its lane's list is unranked: it sorts below every listed item — arrived, not yet placed — ordered by `created` ascending, then filename ascending. The board never breaks over a stale line; it degrades to new-arrivals-at-the-bottom, which is where they belonged anyway.

Agents don't order. An agent writing an item never touches `order.yaml`; the item lands unranked, and position is assigned at triage, by the operator. Priority is a judgment about the operator's energy and context — nothing an agent should guess at. When rank was a frontmatter field this was a behavioral rule; the sidecar makes it structural, because the write surfaces no longer overlap.

One gesture deserves calling out, because the self-healing rule would otherwise mishandle it: retitling a ranked item. The rename changes the filename the sidecar points at, and left alone, self-healing would quietly unrank the item — a pure wording edit changing priority, which nobody asked for. So when the tool performs a rename, it also replaces the old filename in `order.yaml` in place, position preserved: the retitle of a ranked item is honestly a two-file gesture. A hand rename that forgets the sidecar is not an error; it falls back to self-healing — the item lands unranked and gets re-placed at the next triage.

The same discipline governs transitions of ranked items. A state flip that left the old-lane entry behind would create a ghost: pruned, the item returns to that lane unranked; unpruned, its old priority quietly revives — the outcome decided by whether some unrelated write happened to prune in between. So when the tool transitions a ranked item, it removes the old-lane entry in the same gesture, and a transition-and-place moves the entry rather than adding a second one. Leaving a lane always costs your place in it — deterministically. As with renames, a hand edit that forgets the sidecar isn't an error; self-healing catches it.

This is the general principle showing through a specific file: because files are the truth, every gesture in the UI is an edit someone will review. Gestures should touch the minimum surface that expresses them — a drag within a lane edits the order file and nothing else; a state flip of an unranked item edits the item and nothing else; a gesture that moves a ranked item touches both files, because the gesture said two things.

## The Forge Inbox

Private repos need no ingestion story — the operator is their only real user and simply writes items directly. Public repos collect issues from outside users, and the forge remains the right surface for those conversations: it's where the users are, and replies belong there.

Ingestion is therefore **one-way, pull-based, lossy by design, and in v1, manual**. When an issue turns out to be roadmap-shaped, the operator creates an inbox item carrying a `source:` provenance line pointing back at it. From that moment the item is the truth for intent and the issue remains the truth for conversation. No comment mirroring, no status sync-back, no webhooks, no credentials held by the tool. Closing the loop on the forge — a "this is on the roadmap" comment, closing the issue when the item ships — is the operator's courtesy, not the tool's obligation.

The `source:` field is the entire ingestion contract. Tooling to populate it (an ingest command shelling out to `gh` and gitea's equivalent) can arrive later without the convention changing shape.

## The Tool

A single Go binary, `vane`, operating on the repo it's invoked in. The root is found by a single upward walk: at each ancestor, starting from the working directory, a `docs/future/roadmap/` claims the root; failing that, any entry named `.git` — directory or file, since worktrees and submodules use the file form — claims it, and the walk never continues past the first `.git`, because a repo boundary is a wall, not a waypoint. If the walk exhausts without either marker, the working directory itself is the root. Capture from three directories deep lands in the repo's roadmap, not a nested one; capture in a nested repo lands in *that* repo, never the enclosing one. Checking that `.git` *exists* is a root marker, a filename test, not git integration (see the seam census).

**CLI.** The essential surface is capture, and the contract is one sentence: `vane [title]` opens the operator's editor (via a `VANE_EDITOR` / `EDITOR` cascade) on the whole item — a frontmatter skeleton carrying `title:` from the argument or empty, `state: inbox`, and `created:` today (the process-local calendar date — the date the operator experienced when the idea arrived), plus an empty body — so what you edit is exactly what lands, with no special first-line rules. On save-and-exit, a present title derives the slug and the temp file renames into `roadmap/`; an empty title cancels the capture and prints the temp file's path, so words already typed are never lost. A non-empty title whose slug reduces to nothing gets the same treatment, except the printed instruction says to pick a filename and rename the temp file by hand. The temp file itself lives in `roadmap/`, not the system temp directory — a `.capture-` prefix marks it as a non-item that enumeration ignores, and it stays put on cancellation. Even a canceled capture remains inside the working tree, reviewable and resumable, never outside the judgment gate. Capture creates `docs/future/roadmap/` on demand when it's missing — the first idea in a fresh repo is exactly the moment entry must cost nothing. Listing and state transitions at the CLI are cheap to add and probably earn their place, but capture is the non-negotiable — the inbox only works if entry costs nothing.

**Web UI.** `vane serve` presents a localhost board: lanes rendered from states, items ordered by `order.yaml`, drag to reorder or transition, click to view and edit, and capture — a new item created directly from the board, landing in the inbox lane as a fresh file. Capture in the UI matters for the same reason it matters at the CLI: the inbox only works if entry costs nothing, and when the board is already open, dropping to a terminal is the friction. Everything the UI does resolves to file writes under the write model above. And the mirror-image rule holds on the read side: the server holds no authoritative state — every board render and API read rebuilds from a fresh read of the disk, because files are the truth and agents and editors keep writing them while the board is open. At roadmap scale on localhost, re-reading is free. v1 surfaces changes on manual browser refresh; a reload button or modest polling is a later nicety, not a model change. Plain `localhost:port` in a browser for now; the desktop-native shape is a future concern (see deferrals).

**Implementation dispositions.** Go, with `github.com/michaelquigley/df/dl` for logging and `github.com/michaelquigley/df/dd` for frontmatter/YAML handling, per the standing convention. The domain model — items, states, ordering — stays free of both rendering and transport knowledge; the CLI renderer, the HTTP layer, and the web UI are all walkers over the same model. (See the seam census.)

## Scenarios

**Capture mid-flow.** Michael is deep in zrok work and an idea surfaces about a different project. He switches to that repo's directory, runs `vane`, types two lines into the editor that opens, closes it. An inbox item exists as an uncommitted file. Total cost: seconds. The idea is caught; the flow resumes.

**An agent proposes.** A Claude Code instance working in a repo notices that a deferred concern from a just-realized spec deserves roadmap presence. It writes `docs/future/roadmap/retry-semantics.md` with inbox state and a one-paragraph body — a plain file write, no MCP, no auth. The item shows up in Michael's next `git status` and on the board. He reads the diff, adjusts a line, commits it — or discards it. The agent participated; the judgment gate held.

**Triage.** Saturday morning, coffee, `vane serve`. Four inbox items from the week. One is noise — dropped. One is a real idea with no near-term energy — horizon. Two are alive — researching, ranked against what's already there by dragging. Five file edits sit in the working tree; one commit records the triage.

**A spec is born.** Three researching items in the frame repo turn out to be one piece of work. A design session produces a spec in `docs/future/`; the spec's frontmatter names the three items in `sources:`; each item takes a dated stamp in its log. The board now tells the truth at a glance: three intents, one design, in motion.

**An outside user's idea ships.** A zrok user files a GitHub issue proposing a capability. It's roadmap-shaped, so Michael creates an inbox item with `source: github:openziti/zrok#NNN` and replies on the issue that it's under consideration. Months later the item moves through building and evaluating to done; Michael closes the issue with a pointer to the release. The conversation lived on the forge; the intent lived in the repo; neither system pretended to be the other.

## Seam Census

Boundary calls made in this design, recorded for downstream review (mercurius on this spec; terminus on the eventual code).

- **model / render** — **separate.** The item model feeds at least three renderers from day one: CLI output, the web UI, and raw markdown read by humans and agents. Multiple consumers meet at the model, so the boundary earns its cost. No rendering or formatting behavior on domain types.
- **model / transport** — **separate** (unconditional, per standing disposition). The model knows nothing of HTTP, JSON wire shapes, or the browser. The serve layer translates at the edge.
- **filesystem / board freshness** — **the server holds no authoritative state.** Every render and API read rebuilds from disk; a startup snapshot that goes stale against first-class filesystem writers would betray the shared-working-tree model. v1's browser-side surfacing is manual refresh.
- **model / document persistence** — **separate.** Domain types (items, states, ordering) know nothing of frontmatter parsing, raw bytes, file paths, hashes, or filesystem writes; a document layer owns parse, surgical patch, and the content-hash guard. The model never serializes itself. This is the entry that makes the surgical-write commitment reviewable at code time: a diff where a domain type learns to write is an adjudicable finding.
- **contract circumvention** — **the tool never touches git.** This is the load-bearing facade of the design: all persistence is working-tree file writes; commit, push, and history are the operator's jurisdiction exclusively. Any future feature that wants the tool to commit — however convenient — reaches around the judgment gate and should be treated as a design change, not an enhancement. Review should catch any diff that imports git plumbing. One clarification so root discovery doesn't read as a breach: testing whether an entry named `.git` *exists* — file or directory, never opened, never read — is a filename check used as a root marker, not git integration; the boundary is about persistence, history, and judgment.
- **error by tier** — fail-fast is reserved for repository-level failures: the roadmap directory missing or unreadable at `serve` startup, an unreadable `order.yaml` — except capture, which creates a missing directory on demand rather than failing. A single bad item is never repository-level: malformed frontmatter or an unparseable file degrades that one item into a visibly flagged card rather than downing the board — one bad file shouldn't hide the other forty, and the flagged card is the error report. The serve loop wraps, logs via `dl`, and continues on per-request failures. Revisit if item-file corruption in practice turns out to deserve harder failure.
- **convention / tool ownership** — **the convention is primary; the tool is a reader.** Nothing about the file format may exist only in the tool's understanding. If vane's binary vanished, the roadmap remains fully legible and operable by hand. Revisit-condition: none foreseen; this is the thesis.

## Deferred (and Why)

- **Cross-repo aggregation.** A single board over every repo on disk (github clones, HQ gitea clones) is a genuinely useful future shape — but it drags in repo discovery, and the per-repo convention has to prove itself first. The convention is designed so aggregation is purely additive: a future reader walks N repos instead of one.
- **Forge ingestion tooling.** The `source:` field defines the contract now; the `gh`/gitea plumbing arrives only if manual item creation proves to be real friction. Issue volume today doesn't justify the auth surface.
- **dfw integration.** vane looks like the shape dfw was designed for, but coupling two unshipped projects is a known trap. v1 is plain localhost-in-a-browser; the desktop-native wrap is a later, separate move once both projects stand on their own.
- **Done-item reporting.** "What shipped last quarter" becomes a git-log question once done items are curated away. If that reporting need gets real, the answer is the tool reading history — and that would mean revisiting the no-git seam through this same review gate, a recorded design change rather than an enhancement. A future shape either way, not a reason to keep corpses in the working tree.
- **Automated state sync with the pipeline.** Items in building could theoretically track pipeline events automatically. Declined for now: the operator flipping coarse states by hand is cheap, and automation here would couple vane to pipeline internals it has no business knowing.
