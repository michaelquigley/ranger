# CHANGELOG

## Unreleased

CHANGE: the board UI moved to a light theme with the house fonts (Source Serif 4, Source Code Pro, bundled — no CDN), Material Design icon buttons (inlined SVG), and a map-mark logo/favicon.

CHANGE: the item view is now a centered fixed-size modal rendering the body as markdown (react-markdown + GFM), with meta pills, click-the-title retitling, and raw-bytes editing behind an explicit edit mode; the capture modal matches its dimensions.

CHANGE: drags are fully telegraphed — a floating card overlay follows the pointer, cross-lane drags live-open a slot in the destination, and drops apply optimistically with the server's fresh board replacing the preview; gestures always compute against the pre-drag server-truth snapshot, and failed drops restore it.

FEATURE: cards display their tags as colored label chips, sorted for display: the house vocabulary (defect, documentation, enhancement, epic, feature, spike, story, product labels) carries fixed colors matched to the practice's forge boards, and unlisted tags derive stable colors from their text.

FIX: within-lane downward drags landed one position short (dnd-kit's drop index already encodes the final slot); the reorder math moved to `ui/src/reorder.ts` with a vitest table pinning the semantics.

FEATURE: GitHub CI following the terminus shape — vet/test (Go suite, UI vitest, headless-tag compile) on every push and PR, a stamped linux-amd64 build using push's `ci/ldflags.sh` with a verify-stamp gate, and a drafted GitHub release with the tarball (binary + CHANGELOG + docs/current) on `v*` tags.

CHANGE: versioning follows the push pattern (`github.com/michaelquigley/push/build`): figlet `vane version` with ldflags-stamped build detail, `v0.1.x` dev base, and a `make push` depot-vendoring target.

FEATURE: `vane serve` + `ui/` — the localhost board: cobra `serve` (default port 4114, `127.0.0.1` only, startup fail-fast, graceful shutdown) serving the embedded Vite/TypeScript/React 19 board over the ogen API at `/api/v1`, with a `no_ui` build tag for headless binaries. The board renders seven lanes with flag badges, log stamps, and ranked/unranked boundaries; dnd-kit drags express reorder (ranked-prefix PUT, tail drops snap to the boundary) and cross-lane transition-and-place; the item panel edits raw bytes with retitle/rename-to-slug gestures; capture keeps its content through every refusal and shows slug-collision recovery paths. Client types generate from the same OpenAPI spec via openapi-typescript/openapi-fetch; the Makefile follows the archive repo's shape (`build` = npm + `go install`, plus `generate` and `headless` targets).

FEATURE: `internal/api` + `internal/server` — the contract-first API: OpenAPI 3.0.3 spec with committed ogen-generated server code, and handlers implementing the generated interface over a fresh `workspace.Load` per request. Board reads deliver per-card hashes, per-lane `rankedCount`, and `orderVersion` (absence is a version); every mutation carries the guards back and returns a fresh board; the typed 409 family splits `item_conflict`/`order_conflict` from `slug_collision`, which carries structured recovery paths (preserved capture draft, rename source/destination). Guard wire semantics pinned by tests against temp-dir workspaces.

FEATURE: `internal/workspace` — root discovery (roadmap dir, `.git` wall, start-dir fallback), fresh-read snapshots with spec error tiers, and the composite gestures: two-phase capture (draft + no-clobber finalize with four explicit outcomes), transition and transition-and-place with ranked-entry cleanup, reorder, retitle/rename-to-slug with in-place order occurrence replacement, and verbatim content save with effective-lane transition discipline. Every gesture preflights its guard hashes and prunes opportunistically only on order writes; a whole-tree diff suite pins each gesture to exactly its expressing files and lines.

FEATURE: `cmd/vane` — the CLI: `vane [title]` capture through the `VANE_EDITOR`/`EDITOR` cascade, `vane list` (lane-grouped board with ranks and flags), `vane state` transitions, and `vane version`; `-v` re-inits `dl` at debug.

FEATURE: `internal/document` — the byte-shaped layer: two-pass item and order.yaml parsing (yaml.v3 node AST + per-field dd shape validation, claimed/unknown key split, alias resolution), malformed-as-verdict classification, SHA-256 content hashes, surgical line patches for items (`SetState`, `SetTitle`) and order documents (prune, lane rewrite, entry remove/insert, filename replace, fresh-file emission), and guarded writes (`CompareAndWrite` with the absent sentinel, no-clobber `FinalizeLink`). A round-trip fixture suite asserts every patch touches only its expressing lines.

FEATURE: `internal/model` — the pure domain layer: the seven lifecycle states with canonical lane order, the ASCII-mechanical slug rule, card flags (malformed, filename-mismatch), and `ComputeBoard`, implementing the spec's ordering semantics — discard-then-first-occurrence order evaluation, effective-lane handling for unreadable states, inert/prunable entry dispositions, and the created/filename unranked tail sort.
