# ranger

A roadmap that lives in your repository as plain markdown files, plus a Go reader tool — CLI capture and a localhost board — for working with it. The convention is the product; the tool is a reader. One binary, no daemon, no database, and it never touches git.

## The idea

Old-school development had stories. Agentic development has prompts... items of work that get designed, sharpened, and eventually handed to a coding session for execution. ranger is a thinking tool for developing that kind of scouting intelligence over a product's territory: every item on the board is a prompt at some stage of development, from a captured stray thought to a piece of work ready to run, and the operator is the ranger.

The roadmap is files in the working tree because that's where both kinds of collaborator already work. Humans edit items in any editor; agents read and write them with ordinary file operations; review happens in the diff, like everything else. There is no sync, no export, and no second source of truth — git is the judgment gate, and `git diff` before commit is the review surface for every change the tool makes.

## The convention

A roadmap is a directory: `docs/future/roadmap/`, one markdown file per item. An item is YAML frontmatter over a free markdown body:

```markdown
---
title: live board reload
state: researching
created: 2026-07-15
tags: [enhancement, spike]
milestone: v0.1.x
---

the current board implementation surfaces changes on a manual browser
refresh; let's look at upgrading the board implementation to poll or
otherwise self-refresh.
```

`title`, `state`, and `created` are required; `tags`, `subsystems`, and `milestone` are optional. Anything else in the frontmatter is unknown material — ranger skips it, preserves it byte-for-byte, and never promotes it to an error, so items can carry whatever extra metadata your practice wants.

The filename is the slug of the title (`live board reload` → `live-board-reload.md`), by a mechanical ASCII rule. A file whose name has drifted from its title gets a flag on the board and a one-click repair, not a silent rename.

`state` is one of five lifecycle lanes: **inbox → horizon → researching → building → evaluating**. The lifecycle ends at evaluating — there are no `done` or `dropped` lanes. A realized prompt doesn't park in a trophy lane; it gets deleted, its information synthesized into the project it produced, and the board never accumulates terminal residue.

Ranking is a separate file, `order.yaml`, holding each lane's ranked prefix:

```yaml
horizon:
  - cross-repo-aggregation.md
  - richer-log.md
researching:
  - live-board-reload.md
```

Files listed rank in that order at the top of their lane; everything else in the lane is the unranked tail below a visible boundary, sorted by `created` then filename. An item that fails its schema doesn't vanish — it degrades to a flagged card in its lane (or inbox, when the state itself is unreadable) with the diagnostic attached.

## The tool

Capture is the root command. `ranger retry semantics v2` opens your editor (`RANGER_EDITOR`, falling back to `EDITOR`) on a skeleton with the title filled in and `state: inbox`; save, and the item lands in the roadmap under its slug. Cancel, and the draft temp survives inside the working tree — nothing is lost outside the judgment gate.

`ranger serve` hosts the board on `http://127.0.0.1:4114` — five lanes in lifecycle order, embedded in the binary and fully offline (fonts, icons, everything). Dragging a card between lanes patches its `state:` line; dragging within a lane rewrites the ranked prefix; drops are anchor-placed, so reordering works correctly even under active filters. The board filters by tag, subsystem, and milestone, searches titles and bodies against a fresh disk read, and opens each item in a modal: rendered markdown, in-place retitle, raw-bytes edit, and deletion behind a confirm. Freshness is a manual browser refresh — every request re-reads the disk, so there is nothing to go stale.

`ranger list` prints the same board in the terminal, ranked prefix numbered and flags inline; `ranger state <filename|slug> <state>` transitions an item from the command line.

## Load-bearing rules

- **The tool never touches git.** No git imports anywhere; all persistence is working-tree file writes, and the operator owns history.
- **Parsed reads, hand-patched writes.** ranger reads through a YAML parser but writes by surgical byte-level line patching — never a decode-and-encode cycle. Your comments, spacing, and unknown fields survive every gesture.
- **Guarded gestures.** Every write verifies the file's hash against the snapshot the gesture was computed from; a mismatch is a conflict and a reload, never a blind retry.
- **The convention is primary.** Nothing about the format exists only in the tool's understanding — the files are fully legible, and editable, without ranger present.

## Quick start

```sh
go install github.com/michaelquigley/ranger/cmd/ranger@latest
```

Then, from anywhere inside a repository:

```sh
# capture a prompt
ranger retry semantics v2

# the board
ranger serve

# terminal views
ranger list
ranger state retry-semantics-v2 horizon
```

Root discovery is an upward walk: the nearest ancestor with `docs/future/roadmap/` — or, failing that, a `.git` entry — claims the root, so capture from three directories deep lands in the right roadmap, and a nested repo never leaks items into its enclosing one.

## Status

Prototype-first, v0.1.x. The convention is validating through pilot use on real projects — including this one: `docs/future/roadmap/` here is ranger's own roadmap, captured and triaged entirely through the binary. Hardening comes after the model proves out.

## Development

```sh
make test    # go test ./... + go vet; the UI's vitest suite runs via `npm test` in ui/
make build   # UI build (embedded via go:embed) + go install
```

Architecture lives in [`docs/current/`](docs/current/); working memory in [`docs/journal/`](docs/journal/).
