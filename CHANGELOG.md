# CHANGELOG

## Unreleased

CHANGE: the project is renamed vane -> ranger — the vision matured through use into a thinking tool for developing scouting intelligence over a product's territory, and the name follows. Module path `git.hq.quigley.com/products/ranger`, binary `ranger`, env var `RANGER_EDITOR` (EDITOR fallback unchanged), spec `specs/ranger.yml`, new binoculars mark. Earlier changelog entries keep the old name as history.

CHANGE: the lifecycle ends at evaluating — `done` and `dropped` are retired (a recorded design change from pilot experience). An item is a prompt: realized or declined prompts are deleted via the delete gesture, their information synthesized into the project first; git history keeps the archaeology, and the board never accumulates terminal residue. Existing `state: done`/`dropped` items surface as malformed-flagged cards in inbox for re-triage or deletion.

FEATURE: search — a header search box backed by `GET /search?q=`: case-insensitive substring match over item titles and bodies against a fresh disk read (no index, no store), debounced in the client, with the match set composing with tag/subsystem/milestone filters. Escape clears.

FEATURE: `subsystems:` — a second optional claimed list field naming the parts of the stack an item impacts, for combined projects with distinguishable parts. Rendered as outlined mono chips (visibly a different species from the solid genre tags), AND-filterable like tags and composing with tag and milestone filters; shown in the item modal's metadata block. Schema addition recorded in the grimoire's roadmap-convention note.

## v0.1.1

CHANGE: the board carries `project` (the discovered root's name), shown in the header beside the mark and in the browser tab title.

FEATURE: item deletion — the operator's curation act given a gesture (a recorded design change from v1's tool-never-deletes rule): `POST /items/{filename}/delete` removes the file and its order.yaml entries in one hash-guarded gesture, surfaced in the item modal behind an inline confirm. Agents still never delete items.

FEATURE: `milestone:` — a new optional claimed frontmatter field (release-train form, e.g. `v0.1.x`), flowing through the whole stack: document schema (duplicate/shape validated like every claimed field), API card, board display (small mono badge on cards), and single-select milestone filtering composing with tag filters. The grimoire's roadmap-convention note carries the schema addition.

CHANGE: the item modal's metadata row of chips became a structured metadata block — labeled rows (state, created, milestone, tags, source, file) in a bordered grid.

FEATURE: tag filtering on the board — clicking a card's tag chip narrows the board to items carrying every selected tag (forge-style AND), with active filters shown in the header for removal; lane counts and ranked boundaries follow the filtered view.

CHANGE: drop placement is anchor-based — the moved card lands directly beside the card it was dropped against in the lane's full order — which makes dragging work under active filters: hidden neighbors stay put, and a drop targeting an unranked card or empty space serializes as end-of-ranked-list.

CHANGE: board presentation tuning — the whole UI scales from the root font size (115%, one number in `style.css`, with px dimensions converted to rem and icons to em), lanes are 25% wider, cards no longer show log stamps on their faces (the item modal keeps them), and tag chips render 10% smaller.

## v0.1.0

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
