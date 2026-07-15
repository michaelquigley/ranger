# CHANGELOG

## Unreleased

FEATURE: `internal/document` — the byte-shaped layer: two-pass item and order.yaml parsing (yaml.v3 node AST + per-field dd shape validation, claimed/unknown key split, alias resolution), malformed-as-verdict classification, SHA-256 content hashes, surgical line patches for items (`SetState`, `SetTitle`) and order documents (prune, lane rewrite, entry remove/insert, filename replace, fresh-file emission), and guarded writes (`CompareAndWrite` with the absent sentinel, no-clobber `FinalizeLink`). A round-trip fixture suite asserts every patch touches only its expressing lines.

FEATURE: `internal/model` — the pure domain layer: the seven lifecycle states with canonical lane order, the ASCII-mechanical slug rule, card flags (malformed, filename-mismatch), and `ComputeBoard`, implementing the spec's ordering semantics — discard-then-first-occurrence order evaluation, effective-lane handling for unreadable states, inert/prunable entry dispositions, and the created/filename unranked tail sort.
