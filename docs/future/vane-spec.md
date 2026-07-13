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
rank: sparse ordering key
tags: [optional, soft, grouping]
source: optional provenance, e.g. github:michaelquigley/zrok#412
spec: optional path to the spec this item fed
---

body prose, at whatever weight the idea currently has.
```

Items do not nest. If a cluster of items wants to travel together, that is either what `tags` are for (soft grouping) or what spawning a spec is for (hard consolidation). Nesting is where simple models go to die, and vane declines the invitation.

## The Lifecycle

Seven states, five of them lifted directly from years of forge-board practice, two added to name what that practice was missing.

- **inbox** — untriaged. Ideas, drive-by captures, material pulled from forge issues. Getting something into the inbox should cost nothing; that is the CLI's first job.
- **horizon** — triaged and deliberately at rest. This is the state the old boards lacked, and its absence is why vision notes and deferred work orders piled up as orphaned documents. A horizon item is not backlog that failed — its job is to sit still. Vision-register material lives here, possibly forever, legitimately.
- **researching** — being shaped: thinking underway, design sessions happening, a spec possibly being drawn from it.
- **building** — implementation in flight. The pipeline's artifacts (spec, work order, terminus reviews) carry the fine grain of execution; the item's state carries the one-glance weather.
- **evaluating** — built, and being lived with. Evaluation here means soak: running the thing in the practice to see whether it is actually what was wanted. It exits forward to done when it holds, or backward to researching (or out to dropped) when it doesn't.
- **done** — realized.
- **dropped** — declined.

Done and dropped are terminal states, not deletions. Marking an item done is a state transition; removing the file is a separate, deliberate curation act by the operator, probably batched. The tool never deletes. The board gets a visible shipped lane for as long as that's useful, the working tree never accumulates permanent residue, and git history remains the archaeology for anything removed — the same discipline the pipeline already applies to realized specs.

## Items and Specs: Spawn, Not Grow

A spec can be born from one or more items, so an item cannot *become* a spec — there is no single file to grow. The relation lives in links, not containment. When the design phase picks up a thread, the resulting spec records which items fed it, and those items take a `spec:` pointer and (typically) the researching state. The item remains the roadmap-grain record; the spec is the design-grain artifact; neither pretends to be the other.

## The Write Model

vane's tool is read-only by default and read-write by intent, and the write model is the load-bearing safety property of the whole design: **the working tree is the write buffer, and git is the judgment gate.** Every edit the tool makes — a state flip, a rank change, a body edit — is a file write into the working tree, and nothing more. The tool never commits, never pushes, never touches git at all. The uncommitted diff *is* the pending-review queue, and the operator (or sexton, at the operator's direction) decides what enters history.

This is the same safeguard the grimoire relies on — nothing settles into the durable record without a judgment gate — implemented mechanically by infrastructure that already exists. It is also the answer to the agent question: agents can participate in the roadmap freely, because participation produces reviewable diffs, not committed facts.

## Ranking

Lanes are ordered; priority is real. The rank is a **sparse** key — fractional or lexicographic, inserted between neighbors — so that reordering one item touches exactly one file. Dense ranks (1, 2, 3...) would turn every drag into a lane-wide renumbering: N file writes, N diff lines, noise in git and in agent context. Sparse keys keep every UI gesture commit-shaped and minimal. When gaps exhaust, rebalancing is a cheap, explicit operation, not a side effect.

This is the general principle showing through a specific field: because files are the truth, every gesture in the UI is an edit someone will review. Gestures should touch the minimum surface that expresses them.

Agents don't rank. An agent writing an item leaves `rank` absent; unranked items sort to a natural new-arrivals position in their lane, and rank is assigned at triage, by the operator. Rank is a judgment about the operator's energy and priorities — context an agent shouldn't guess at — and any rank an agent supplied would be re-ranked at triage anyway, so the field carries no information worth writing.

## The Forge Inbox

Private repos need no ingestion story — the operator is their only real user and simply writes items directly. Public repos collect issues from outside users, and the forge remains the right surface for those conversations: it's where the users are, and replies belong there.

Ingestion is therefore **one-way, pull-based, lossy by design, and in v1, manual**. When an issue turns out to be roadmap-shaped, the operator creates an inbox item carrying a `source:` provenance line pointing back at it. From that moment the item is the truth for intent and the issue remains the truth for conversation. No comment mirroring, no status sync-back, no webhooks, no credentials held by the tool. Closing the loop on the forge — a "this is on the roadmap" comment, closing the issue when the item ships — is the operator's courtesy, not the tool's obligation.

The `source:` field is the entire ingestion contract. Tooling to populate it (an ingest command shelling out to `gh` and gitea's equivalent) can arrive later without the convention changing shape.

## The Tool

A single Go binary, `vane`, operating on the repo it's invoked in.

**CLI.** The essential surface is capture: `vane` gets an item into the inbox at terminal speed, opening the operator's editor via a `VANE_EDITOR` / `EDITOR` cascade for the body. Listing and state transitions at the CLI are cheap to add and probably earn their place, but capture is the non-negotiable — the inbox only works if entry costs nothing.

**Web UI.** `vane serve` presents a localhost board: lanes rendered from states, items ordered by rank, drag to reorder or transition, click to view and edit, and capture — a new item created directly from the board, landing in the inbox lane as a fresh file. Capture in the UI matters for the same reason it matters at the CLI: the inbox only works if entry costs nothing, and when the board is already open, dropping to a terminal is the friction. Everything the UI does resolves to file writes under the write model above. Plain `localhost:port` in a browser for now; the desktop-native shape is a future concern (see deferrals).

**Implementation dispositions.** Go, with `github.com/michaelquigley/df/dl` for logging and `github.com/michaelquigley/df/dd` for frontmatter/YAML handling, per the standing convention. The domain model — items, states, ranks — stays free of both rendering and transport knowledge; the CLI renderer, the HTTP layer, and the web UI are all walkers over the same model. (See the seam census.)

## Scenarios

**Capture mid-flow.** Michael is deep in zrok work and an idea surfaces about a different project. He switches to that repo's directory, runs `vane`, types two lines into the editor that opens, closes it. An inbox item exists as an uncommitted file. Total cost: seconds. The idea is caught; the flow resumes.

**An agent proposes.** A Claude Code instance working in a repo notices that a deferred concern from a just-realized spec deserves roadmap presence. It writes `docs/future/roadmap/retry-semantics.md` with inbox state and a one-paragraph body — a plain file write, no MCP, no auth. The item shows up in Michael's next `git status` and on the board. He reads the diff, adjusts a line, commits it — or discards it. The agent participated; the judgment gate held.

**Triage.** Saturday morning, coffee, `vane serve`. Four inbox items from the week. One is noise — dropped. One is a real idea with no near-term energy — horizon. Two are alive — researching, ranked against what's already there by dragging. Five file edits sit in the working tree; one commit records the triage.

**A spec is born.** Three researching items in the frame repo turn out to be one piece of work. A design session produces a spec in `docs/future/`; the spec's frontmatter names the three items as sources; each item gains a `spec:` pointer. The board now tells the truth at a glance: three intents, one design, in motion.

**An outside user's idea ships.** A zrok user files a GitHub issue proposing a capability. It's roadmap-shaped, so Michael creates an inbox item with `source: github:openziti/zrok#NNN` and replies on the issue that it's under consideration. Months later the item moves through building and evaluating to done; Michael closes the issue with a pointer to the release. The conversation lived on the forge; the intent lived in the repo; neither system pretended to be the other.

## Seam Census

Boundary calls made in this design, recorded for downstream review (mercurius on this spec; terminus on the eventual code).

- **model / render** — **separate.** The item model feeds at least three renderers from day one: CLI output, the web UI, and raw markdown read by humans and agents. Multiple consumers meet at the model, so the boundary earns its cost. No rendering or formatting behavior on domain types.
- **model / transport** — **separate** (unconditional, per standing disposition). The model knows nothing of HTTP, JSON wire shapes, or the browser. The serve layer translates at the edge.
- **contract circumvention** — **the tool never touches git.** This is the load-bearing facade of the design: all persistence is working-tree file writes; commit, push, and history are the operator's jurisdiction exclusively. Any future feature that wants the tool to commit — however convenient — reaches around the judgment gate and should be treated as a design change, not an enhancement. Review should catch any diff that imports git plumbing.
- **error by tier** — bootstrap failures (unparseable item file, missing directory) fail fast and loud at startup or on load; the serve loop wraps, logs via `dl`, and continues on per-request failures; malformed frontmatter in a single item degrades that item's rendering (visibly flagged) rather than downing the board. Revisit if item-file corruption in practice turns out to deserve harder failure.
- **convention / tool ownership** — **the convention is primary; the tool is a reader.** Nothing about the file format may exist only in the tool's understanding. If vane's binary vanished, the roadmap remains fully legible and operable by hand. Revisit-condition: none foreseen; this is the thesis.

## Deferred (and Why)

- **Cross-repo aggregation.** A single board over every repo on disk (github clones, HQ gitea clones) is a genuinely useful future shape — but it drags in repo discovery, and the per-repo convention has to prove itself first. The convention is designed so aggregation is purely additive: a future reader walks N repos instead of one.
- **Forge ingestion tooling.** The `source:` field defines the contract now; the `gh`/gitea plumbing arrives only if manual item creation proves to be real friction. Issue volume today doesn't justify the auth surface.
- **dfw integration.** vane looks like the shape dfw was designed for, but coupling two unshipped projects is a known trap. v1 is plain localhost-in-a-browser; the desktop-native wrap is a later, separate move once both projects stand on their own.
- **Done-item reporting.** "What shipped last quarter" becomes a git-log question once done items are curated away. If that reporting need gets real, the answer is the tool reading history — a future shape, not a reason to keep corpses in the working tree.
- **Automated state sync with the pipeline.** Items in building could theoretically track pipeline events automatically. Declined for now: the operator flipping coarse states by hand is cheap, and automation here would couple vane to pipeline internals it has no business knowing.
