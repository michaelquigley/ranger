# AGENTS.md — ranger

ranger is a roadmap-as-files convention plus a Go reader tool (CLI capture + localhost board). The convention is the product; the tool is a reader. Repo: `github.com/michaelquigley/ranger`.

## How to arrive

1. **`docs/journal/`** — agent memory, dated entries, newest first. Read the recent entries before anything else; write freely as you work.
2. **`docs/current/`** — built behavior: the domain model, the document layer, workspace + CLI, the API, and the UI/serve surface. This is the authority on what exists.
3. **`docs/future/roadmap/`** — ranger's own roadmap, kept in ranger's own convention (dogfooding). Deferred concerns live here as items; agents may write items but never touch `.ranger/order.yaml`. The agent-facing convention spec lives in the grimoire at `software/conventions/roadmap-convention.md`.
4. The originating spec and work order were realized and removed 2026-07-15 (v1 complete, all five stages terminus-gated); git history keeps the archaeology. Their settled normative rules — slug algorithm, malformed semantics, `order.yaml` evaluation order, the hash guard, root discovery — now live in the code, its tests, and `docs/current/`.

## Posture

Prototype-first. v1 exists to validate the model and the approach; hardening comes after the model proves out. Corner-case defensiveness was explicitly vetoed in review (see the `settled_decisions` guards in `mercurius.yaml` for the genre) — do not re-introduce it in code, and do not gold-plate. The write model plus the git judgment gate is the accepted safety net.

## Load-bearing rules

- **The tool never writes git.** Read-only status inspection is the one allowed touch (a 2026-07-21 design change; the rule was previously never-touches): a single `git status` shell-out scoped to the roadmap directory, degrading to unknown wherever git can't answer. No git library imports anywhere, nothing opens `.git` internals — the name stays a filename string in root discovery — and any invocation that alters git state (add, commit, push, index, refs, hooks, config) is an automatic review finding.
- **dd reads, hand-patched writes.** `df/dd` is the read/validate path only; every write is surgical byte-level line patching. A write path through any YAML encoder is a review finding.
- **The convention is primary.** Nothing about the file format may exist only in the tool's understanding.
- **The tool never commits or pushes.** All persistence is working-tree file writes; the operator owns history. Deletion (design change 2026-07-16) exists as an explicit, hash-guarded, confirmed operator gesture — v1's tool-never-deletes rule assigned removal to the operator's curation, and this is that act given a button. Agents still never delete items.

## Process

- v1's staged work order is realized; ongoing work is ad-hoc improvement or new pipeline arcs for anything architectural. Substantive changes are still gated by **terminus**: run it, resolve or get an explicit veto on every finding, and bring the work to Michael `clean`. Findings vetoes are Michael's call, recorded in conversation and journal. Terminus judges convention qualities, not behavior — clean is necessary, not sufficient.
- As behavior lands, synthesize it into `docs/current/` and add a `CHANGELOG.md` entry under `## Unreleased` (`FEATURE`/`CHANGE`/`FIX` prose, in-house format — not Keep a Changelog).
- Run `unfurl -i <file>` on any markdown you author or edit, unconditionally.
- Go conventions: `df/dl` for logging, `df/dd` for YAML/JSON binding. Reference implementation for the ogen/embed/UI wiring: `flo` in `../archive`.
- Realized specs and work orders are removed from `docs/future/` (git history keeps the archaeology); still-live deferred concerns get re-synthesized first — for ranger, that means roadmap items.

## Project memory

Durable knowledge about this project lives in `docs/journal/`, dated files `docs/journal/YYYY-MM-DD.md`. This is project memory; it does not go in harness-local storage (`.claude/` or equivalent), where it's invisible to every other harness and collaborator and dies with the host. Concretely: do not write to your harness's memory directory or memory tool for this project — even when the harness presents it as the default place for durable knowledge. That tool is the silo this convention exists to replace; the journal is the only durable home.

On arrival, read the most recent entries to pick up where the last session left off, before you start changing things. Treat them as prior-session context, not verified truth — if an entry conflicts with the code or a `docs/current/` doc, the code wins.

Write the smallest entry that carries the session's durable insight, and nothing more. The test for every line: *would a competent agent get this wrong, or waste time rediscovering it, working from the tree alone?* If it's recoverable by reading the code, the diff, `docs/current/`, or git history, leave it out.

That filter keeps four kinds of thing and discards the rest:

- **Decisions whose rationale isn't visible in the result** — why a value was chosen, what a line guards against, why something that looks like dead code or a no-op is load-bearing.
- **Deliberate non-actions** — a change you considered and chose not to make, so the next agent doesn't "fix" it. An unchanged file leaves no trace in a diff.
- **Couplings that span files** — two places that must move together, an ordering that matters, an assumption one file makes about another.
- **Live state** — what's unverified, unfinished, or waiting on something external.

Skip change inventories, restatements of the diff, and play-by-play of how you worked. There's no write-time approval gate; Michael reviews on commit. Append to the day's file if it exists, and write the few lines you'd want the next agent to read — honest and self-contained.

## Commits

The operator commits; agents don't. Never run `git commit` or `git push` in this repo. Finish the edit, leave the change in the working tree (staged is fine), report what changed, and hand off — the uncommitted diff is the review queue and the commit is the operator's act of acceptance. Approval of a change is not direction to commit; only an explicit instruction to commit is, and only for that commit.
