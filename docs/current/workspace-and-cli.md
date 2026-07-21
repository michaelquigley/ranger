---
title: the workspace and the CLI
created: 2026-07-14
---

# the workspace and the CLI

`internal/workspace` composes document operations into the spec's gestures against a discovered root. It is stateless — every `Load` is a fresh read, because the working tree is a shared buffer and other writers never stop. `cmd/ranger` is a thin cobra surface over it: the root command is capture, plus `list`, `state`, and `version`.

## root discovery

`DiscoverRoot` is the single upward walk: at each ancestor, `docs/future/roadmap/` claims the root; failing that, any entry named `.git` — file or directory, checked with `Lstat` only, never opened — claims it and walls the walk. Exhaustion falls back to the start directory. Capture from three directories deep lands in the repo's roadmap; capture in a nested repo lands in that repo, never the enclosing one.

## load and the snapshot

`Load` enumerates `roadmap/*.md` flat — skipping directories, non-markdown, and `.capture-` temps — and reads `order.yaml`. Error tiers per the spec: a missing or unreadable roadmap directory and an unreadable `order.yaml` are repository-level errors; a single bad item degrades to a flagged card. The snapshot carries each item's raw bytes, guard hash, and parsed document, plus the order document and its version (absence is the `absent` sentinel). `Cards()` classifies everything for `ComputeBoard`, attaching the malformed and filename-mismatch flags.

## the gesture discipline

Every gesture follows one shape: load fresh, verify every affected guard token against that fresh read (a mismatch is a typed conflict — reload, don't retry), compute the minimal bytes that express the gesture, then write — item first, order second, any partial failure reported plainly. Opportunistic pruning rides along only when `order.yaml` is already being written.

- **Capture is two operations**, because the CLI's editor sits between them and may rewrite anything, the title included. `CreateDraft` writes a `.capture-` temp into `roadmap/` (creating the directory on demand) with the skeleton — title from the argument or bare, `state: inbox`, `created:` today. `FinalizeDraft` rereads the saved bytes, recovers the title from them, derives the slug, and no-clobber links the unchanged bytes into place. Four explicit outcomes: finalized, empty title (canceled, temp kept), empty slug (temp kept, rename by hand), collision (temp kept, both paths reported). The temp survives every non-finalized outcome, inside the working tree, never outside the judgment gate.
- **Transition** patches the state line; a ranked item's old-lane entry is removed in the same gesture — leaving a lane always costs your place in it — and a position moves the entry instead (transition-and-place). Position indexes the destination's ranked list only. An absent `order.yaml` is created no-clobber when a placement demands it.
- **Reorder** rewrites one lane's ranked prefix, moving existing entry lines byte-for-byte.
- **Retitle** patches the title, renames to the new slug no-clobber, and replaces the old filename in every retained order occurrence — active and inert alike, positions preserved, so a rank held through a transient parse failure survives the rename. A title whose slug is empty patches in place: the existing filename becomes the hand-picked name such titles carry, rank intact. **RenameToSlug** is the one-gesture mismatch repair, refusing when there is no slug to repair toward.
- **Delete** (design change 2026-07-16 — v1 shipped with no tool deletion; this is the operator's curation act given a gesture) removes the item file and every one of its order.yaml entries in one hash-guarded gesture — a deleted file's entries are prunable unconditionally. Under the write model a tracked file's deletion is a reviewable diff; a never-committed file has no net, so the UI confirms first.
- **SaveContent** lands the operator's bytes verbatim. A save that changes state is a transition made through ranger: compared in effective lanes (an unreadable old state is inbox), departing a lane the card was actively ranked in costs that rank, and entries are retained only when the *new* state is unreadable. Renames are never a side effect of a save — a changed title surfaces as the mismatch flag.

The composite-gesture suite asserts each of these against the whole tree: exactly the files that express the gesture change, to exactly the expected bytes, and every other byte survives.

## the CLI

- `ranger [title words...]` — capture. Editor cascade `RANGER_EDITOR` → `EDITOR`; neither set is an error naming the fix. Exit paths print the finalize outcome, the temp path always surviving cancellation.
- `ranger list` — lanes in lifecycle order, ranked prefix numbered, unranked tail dashed, flags marked inline. A plain renderer over the same `ComputeBoard` output the UI will consume.
- `ranger state <filename|slug> <state>` — the transition gesture from the terminal (`.md` optional). No placement; the card lands unranked in its new lane.
- `ranger serve` — the ad-hoc localhost board over the discovered root; `ranger daemon` — the tray-resident board over every configured root ([daemon.md](daemon.md)). Serve is the daemon's single-project degenerate case: one server implementation, two entry commands.
- `ranger version` — build info. `-v/--verbose` re-inits `dl` at debug.

The capture, `list`, and `state` gestures remain cwd-discovery in-repo gestures and never consult the daemon's config — the CLI's create-on-demand capture keeps its directory-creating behavior, because in-repo, discovery *is* the addressing.

ranger's own repo now dogfoods the convention: `docs/future/roadmap/` holds the spec's still-live deferred concerns as horizon items, captured and triaged entirely through the binary.
