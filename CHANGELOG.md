# CHANGELOG

## Unreleased

FEATURE: `internal/workspace` — root discovery (roadmap dir, `.git` wall, start-dir fallback), fresh-read snapshots with spec error tiers, and the composite gestures: two-phase capture (draft + no-clobber finalize with four explicit outcomes), transition and transition-and-place with ranked-entry cleanup, reorder, retitle/rename-to-slug with in-place order occurrence replacement, and verbatim content save with effective-lane transition discipline. Every gesture preflights its guard hashes and prunes opportunistically only on order writes; a whole-tree diff suite pins each gesture to exactly its expressing files and lines.

FEATURE: `cmd/vane` — the CLI: `vane [title]` capture through the `VANE_EDITOR`/`EDITOR` cascade, `vane list` (lane-grouped board with ranks and flags), `vane state` transitions, and `vane version`; `-v` re-inits `dl` at debug.

FEATURE: `internal/document` — the byte-shaped layer: two-pass item and order.yaml parsing (yaml.v3 node AST + per-field dd shape validation, claimed/unknown key split, alias resolution), malformed-as-verdict classification, SHA-256 content hashes, surgical line patches for items (`SetState`, `SetTitle`) and order documents (prune, lane rewrite, entry remove/insert, filename replace, fresh-file emission), and guarded writes (`CompareAndWrite` with the absent sentinel, no-clobber `FinalizeLink`). A round-trip fixture suite asserts every patch touches only its expressing lines.

FEATURE: `internal/model` — the pure domain layer: the seven lifecycle states with canonical lane order, the ASCII-mechanical slug rule, card flags (malformed, filename-mismatch), and `ComputeBoard`, implementing the spec's ordering semantics — discard-then-first-occurrence order evaluation, effective-lane handling for unreadable states, inert/prunable entry dispositions, and the created/filename unranked tail sort.
