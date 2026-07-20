# AGENTS.md — ranger

ranger is a roadmap-as-files convention plus a Go reader tool (CLI capture + localhost board). The convention is the product; the tool is a reader. Repo: `git.hq.quigley.com/products/ranger`.

## How to arrive

1. **`docs/journal/`** — agent memory, dated entries, newest first. Read the recent entries before anything else; write freely as you work.
2. **`docs/current/`** — built behavior: the domain model, the document layer, workspace + CLI, the API, and the UI/serve surface. This is the authority on what exists.
3. **`docs/future/roadmap/`** — ranger's own roadmap, kept in ranger's own convention (dogfooding). Deferred concerns live here as items; agents may write items but never touch `order.yaml`. The agent-facing convention spec lives in the grimoire at `software/conventions/roadmap-convention.md`.
4. The originating spec and work order were realized and removed 2026-07-15 (v1 complete, all five stages terminus-gated); git history keeps the archaeology. Their settled normative rules — slug algorithm, malformed semantics, `order.yaml` evaluation order, the hash guard, root discovery — now live in the code, its tests, and `docs/current/`.

## Posture

Prototype-first. v1 exists to validate the model and the approach; hardening comes after the model proves out. Corner-case defensiveness was explicitly vetoed in review (see the `settled_decisions` guards in `mercurius.yaml` for the genre) — do not re-introduce it in code, and do not gold-plate. The write model plus the git judgment gate is the accepted safety net.

## Load-bearing rules

- **The tool never touches git.** No git imports anywhere — `.git` appears only as a filename string in root discovery. A git import is an automatic review finding.
- **dd reads, hand-patched writes.** `df/dd` is the read/validate path only; every write is surgical byte-level line patching. A write path through any YAML encoder is a review finding.
- **The convention is primary.** Nothing about the file format may exist only in the tool's understanding.
- **The tool never commits or pushes.** All persistence is working-tree file writes; the operator owns history. Deletion (design change 2026-07-16) exists as an explicit, hash-guarded, confirmed operator gesture — v1's tool-never-deletes rule assigned removal to the operator's curation, and this is that act given a button. Agents still never delete items.

## Process

- v1's staged work order is realized; ongoing work is ad-hoc improvement or new pipeline arcs for anything architectural. Substantive changes are still gated by **terminus**: run it, resolve or get an explicit veto on every finding, and bring the work to Michael `clean`. Findings vetoes are Michael's call, recorded in conversation and journal. Terminus judges convention qualities, not behavior — clean is necessary, not sufficient.
- As behavior lands, synthesize it into `docs/current/` and add a `CHANGELOG.md` entry under `## Unreleased` (`FEATURE`/`CHANGE`/`FIX` prose, in-house format — not Keep a Changelog).
- Run `unfurl -i <file>` on any markdown you author or edit, unconditionally.
- Go conventions: `df/dl` for logging, `df/dd` for YAML/JSON binding. Reference implementation for the ogen/embed/UI wiring: `flo` in `../archive`.
- Realized specs and work orders are removed from `docs/future/` (git history keeps the archaeology); still-live deferred concerns get re-synthesized first — for ranger, that means roadmap items.
