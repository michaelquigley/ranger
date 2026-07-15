---
title: the document layer
created: 2026-07-14
---

# the document layer

`internal/document` owns everything byte-shaped. One discipline governs it: dd/yaml.v3 read, hand-patched writes — no vane write ever passes through a YAML encoder, because a decode-and-encode cycle is exactly the reformat the surgical-edit commitment forbids. `gopkg.in/yaml.v3` is imported here and nowhere else, for the read-side node pass.

## the two-pass parse

Both document kinds parse the same way. Pass 1 decodes the whole YAML block into a `yaml.Node` AST — purely syntactic, so duplicate keys inside unknown material don't trip it; a failure here is malformed (item) or unreadable (order.yaml). Pass 2 works on that AST, never re-parsed text: top-level keys are walked, duplicates are detected among *claimed* keys only, and each claimed field's value node decodes into plain values — aliases resolve against the full tree, so a `tags:` or lane list aliasing an anchor defined under an ignored key reads exactly as any YAML reader would read it. Scalars decode as their source text (the claimed fields are all string-shaped, and source text is what shape validation must judge); `dd.Bind` validates each field's shape independently. Unknown material is skipped entirely — its duplicates and anchors are never disturbed, and it can never be promoted to malformed.

A hand line-scanner does the one job only it can: the key → line-range map for surgical write spans, trimming trailing blank and comment lines so a rewrite never swallows a neighbor's comment.

## items

`ParseItem` always returns a document — malformed is a verdict, not a blackout. The frontmatter fence splits by hand (`---` open, `---` or `...` close, body preserved verbatim); each claimed field decodes and validates independently, so an item with a valid `state: researching` and broken `tags:` stays in the researching lane, flagged, with title and created intact. A card's state is unreadable only when `state:` itself fails.

Malformed per the spec's table: unparseable frontmatter, a duplicate claimed key, a missing required field, a claimed field violating its shape (state outside the seven, dates not `YYYY-MM-DD`, wrong structural types). Diagnostics accumulate rather than short-circuit.

Patches: `SetState` and `SetTitle` replace the field's complete mapped line range — a block-scalar title's continuation lines go with it — preserving indentation and the key line's inline comment, touching nothing else. A duplicated or missing field refuses to patch (repair is a raw edit). Scalar values emit through one shared helper, plain when unambiguously safe, double-quoted otherwise, shared with order entry emission and (later) capture skeletons.

## order documents

`ParseOrder` returns an error for the repository-level tier: YAML syntax failure, a duplicate recognized lane key, a lane that isn't a list of filename strings (shape-validated through dd, like every claimed read), or a lane entry spanning multiple lines — the convention's entries are single lines, and an entry that spans more would strand bytes under every line-targeted op, so it fails loud at parse instead of corrupting on a later write. Duplicate or singular *unknown* keys are ignored per the claimed/unknown split. Unknown-lane blocks are classified prunable — removed on the next opportunistic prune so a misspelled lane heals — with one guard: when any surviving recognized node aliases an anchor in unknown territory, *every* unknown block is kept that write. The guard is deliberately coarse because unknown blocks may alias each other: either all of them go (their cross-references die together) or none do, so a prune can never orphan an alias and render our own write unreadable.

Surgical ops, each returning new bytes: `Prune` (prunable entries and unknown blocks; the one sanctioned multi-line side effect, applied only when the file is already being written), `RewriteLane` (one lane's active entries; inert lines stay in place), `RemoveEntry`, `InsertEntry` (position indexes the ranked list only, skipping inert lines), `ReplaceFilename` (every retained occurrence across all lanes, active and inert alike, positions and inline comments preserved), `NewOrder` (the first-ever ranking). A flow-style lane (`building: [a.md, b.md]`) materializes as block form when an op targets it; everything else survives byte-for-byte.

## guarded writes

`Hash` is SHA-256 lowercase hex — the guard token everywhere. `CompareAndWrite` compares the disk against the hash *carried by the caller* (never a fresh self-comparison) and refuses on mismatch with a typed `ConflictError`; `VersionAbsent` is a legal expected version that creates `O_CREATE|O_EXCL`, so a racing creator wins. `FinalizeLink` is the no-clobber funnel for capture and rename: link then remove, refusing with a typed `CollisionError` carrying both paths while the temp survives. Best-effort detection, no locks — the git gate is the real net.

The fixture suite pins all of it: hand-formatted items and order files with comments, unusual spacing, unknown fields, block/folded scalars, and flow lanes; every patch asserts the exact full-file output, so any byte moved beyond the expressing lines fails the test.
