---
title: vane — work order
status: draft
created: 2026-07-14
spec: docs/future/vane-spec.md
---

# vane — work order

Implementation-shaped translation of [vane-spec](vane-spec.md) into concrete code changes, slicing, and integration points. The spec survived a seven-round spec-only mercurius arc (s_u4Myqkujv8zn) and its convention-level rules are settled; this document does not restate them. Where the spec is normative — slug algorithm, malformed semantics, order.yaml evaluation order, the hash guard, root discovery — the spec is the authority and this work order only maps those rules onto packages, functions, and tests. Both documents converged through a joint mercurius arc (s_8eX1DIbb8BLr, six rounds, operator-called at the noise floor; synopsis under `.mercurius/`).

## Planning decisions

Settled with Michael during work-order drafting (2026-07-14):

1. **dd reads, hand-patched writes.** `df/dd` (yaml.v3 underneath) is the read/validate path: frontmatter mappings and `order.yaml` bind into typed structs through it. dd's writer is decode-and-encode — sorted keys, no comment preservation — which is exactly the reformat the surgical-edit requirement forbids, so **no vane write ever passes through dd**. The document layer owns byte-level line patching for every mutation. This is the `model / document persistence` seam made concrete.
2. **CLI surface: capture, serve, list, state, version.** Capture and serve are the non-negotiables; `list` and `state` land as part of stage 3 (they fall out of the same workspace gestures the UI needs) and are cuttable if they drag.
3. **UI and API per the flo pattern.** Vite + TypeScript + React (React 19) embedded via `go:embed`, with a contract-first OpenAPI 3.0.3 spec and ogen-generated server, `openapi-typescript`/`openapi-fetch` client. Reference implementation: `archive/flo` (`flo/service/api` for the generated layer, `flo/studio/server` for handlers implementing the generated interface, `flo/studio/ui` for the embed/middleware/client shape).
4. **No config files in v1.** No config cascade, no `vane.yaml`, no `~/.config/vane/`. Configuration surface is: `--port` on `serve`, `-v/--verbose` persistent flag, and the `VANE_EDITOR` → `EDITOR` cascade. A recorded deviation from the base stack — there is nothing to configure in a files-are-truth tool yet; the cascade can arrive when settings accumulate.

## Repository shape

Greenfield. Module path `git.hq.quigley.com/products/vane`, Go current stable. `gopkg.in/yaml.v3` is imported directly in `internal/document` only, for one job: the node-level whole-document syntax pass (read side; no vane write ever passes through any YAML encoder).

```
cmd/vane/main.go        cobra entry: root = capture, subcommands serve/list/state/version
internal/model/         pure domain: states, slug, card flags, board ordering computation
internal/document/      item + order.yaml documents: parse, validate, hash, surgical patch
internal/workspace/     filesystem: root discovery, enumeration, capture lifecycle, gestures
internal/api/           specs/vane.yml + committed ogen-generated server code
internal/server/        handlers implementing api.Handler; model↔wire translation at the edge
ui/                     Vite/TS/React app: embed.go, middleware.go, src/, dist/ (committed builds not required; Makefile builds before go build)
```

Package-to-seam mapping, so terminus can adjudicate the census entries:

| census entry | enforced where |
| --- | --- |
| model / render | `internal/model` has no formatting or rendering; CLI renderer in `cmd/vane`, web renderer in `ui/` |
| model / transport | `internal/model` imports nothing HTTP/JSON-wire; `internal/server` translates |
| model / document persistence | `internal/model` never sees raw bytes, paths, hashes, or frontmatter; only `internal/document` parses/patches |
| filesystem / board freshness | `internal/server` holds no snapshot; every handler call rebuilds via `internal/workspace` |
| contract circumvention | no git imports anywhere; `.git` appears only as a string literal in the root-discovery filename test |
| error by tier | `workspace.Load` distinguishes repository-level errors (fatal at serve startup) from per-item degradation (flagged cards) |

## internal/model

Pure types and pure functions. No I/O, no dd, no bytes.

- `State` — string enum for the seven states plus the canonical lane order (inbox, horizon, researching, building, evaluating, done, dropped). `ParseState` returns invalid for anything else.
- `Slug(title string) string` — the spec's ASCII-mechanical rule, implemented over code points exactly as written: map `A`–`Z`→`a`–`z`; keep `a`–`z`, `0`–`9`, space, hyphen; discard every other code point; spaces→hyphens; collapse hyphen runs; trim. Tests carry the spec's normative vectors (`Retry Semantics (v2)` → `retry-semantics-v2`, `naïve K-scale` → `nave-scale` with the Kelvin-sign discard asserted) plus empty-reduction cases.
- `Flag` — enum of card-flag conditions: `FlagMalformed` (with a diagnostic message) and `FlagFilenameMismatch` (filename ≠ slug(title) for an item with a readable title **whose slug is non-empty** — a title that reduces to nothing legitimately carries a hand-picked filename per the spec, renders clean, and is never flagged). Test: an item whose title is entirely discarded code points, under any filename, is unflagged.
- `CardInput` — the classification the ordering computation consumes, produced upstream by document/workspace: filename, readable-state (valid `State` or unreadable), `created` (valid date or absent), title, flags. The **effective lane** of an unreadable-state card is inbox, for all ordering purposes.
- `ComputeBoard(cards []CardInput, order map[State][]string) Board` — implements the spec's ordering semantics as one pure function: discard invalid order entries first (nonexistent file; unknown lane key handled at parse; lane-mismatch only with positive evidence — the item exists and has a valid, *different* readable state; entries for existing-but-unreadable files are retained, and retained entries are **inert only outside the card's effective lane** — since an unreadable-state card's effective lane is inbox for all ordering purposes, an inbox entry for it participates normally in ranking and first-occurrence resolution, while its retained entries in other lanes are inert: held on disk, participating in no duplicate resolution, shadowing nothing), then first-occurrence-wins among survivors, then the unranked tail sorted `created` ascending / filename ascending, with undated cards after every dated card, by filename. The function also reports, per lane, which retained entries are inert and which entries are prunable — the workspace uses that report for opportunistic pruning on its next order write; the model never decides *when* to prune.

The determinism cases from the review arc become table tests: stale line never shadows a valid one; duplicate filename first-wins after discard; transient-unreadable retains rank; malformed-created sorts after dated unranked; an inbox order entry for an unreadable-state card actively ranks it in inbox while that card's entries in other lanes stay inert.

## internal/document

The document layer owns everything byte-shaped. Two document kinds, one discipline: dd reads, hand-patched writes, every mutation surgical.

Parsing is **two passes** for both document kinds, both over the same tree. Pass 1 is whole-document syntax validation: decode the entire YAML block into a `yaml.Node` AST — purely syntactic, so permitted duplicate keys don't trip it (no map is constructed) — and a failure here is malformed (item) or unreadable (order.yaml), per the spec's "frontmatter that doesn't parse" and "unparseable order.yaml" tiers. Pass 2 works **on that AST, never on re-parsed text**: walk the document mapping's key nodes (duplicate detection among *claimed* keys — item fields, state-name lanes — happens right here), decode each claimed field's value node into plain values — aliases resolve against the full tree, so a claimed field referencing an anchor defined under an ignored key is valid, exactly as any normal YAML reader would read it — then assemble the claimed-only map and dd-bind it for shape validation. Unknown material is skipped, so its permitted duplicates and its anchors are never disturbed. The line scanner's job narrows to the one thing only it can do: the key → line-range map for surgical write spans. Fixtures: broken syntax confined inside an unknown field/block is malformed; valid duplicate unknown keys are not; an item `tags:` and an order lane each aliasing an anchor defined under an ignored key are valid.

**Item document.**

- `ParseItem(raw []byte) ItemDoc` — split the frontmatter fence by hand (opening `---` line, closing `---`/`...` line, body is everything after, preserved verbatim); run the two-pass parse above; build the line map for write spans; classify malformed per the spec's table (unparseable frontmatter, duplicate claimed key, missing required field, claimed field violating its shape). Duplicate detection applies to claimed keys only — the spec accepts and ignores unknown fields, and a duplicate inside someone else's extension data must not flag the item; unknown material never reaches the dd bind, so it can never be promoted to malformed. Unknown fields exist only in the line map and the preserved bytes, not in the typed schema. Tests: a duplicate claimed key flags the item; a duplicate unknown key does not.
- **Malformed is a verdict, not a blackout.** Each claimed field's value node is decoded and validated independently; the malformed flag is computed *from* the per-field results and never erases them. `CardInput` carries whichever of state, title, and created were individually readable — so an item with a valid `state: researching` and a broken `tags:` stays in the researching lane, flagged, with its dated sort position and its positive-evidence standing for order pruning intact. A card falls to inbox only when `state:` itself is unreadable. Tests: an invalid sibling field never erases a valid state, title, or created.
- `Hash(raw []byte) string` — SHA-256, lowercase hex. This is the guard token everywhere: board payloads, mutation requests, CLI gestures.
- Surgical patches, each returning new bytes and touching only the lines that express the gesture. A scalar mutation replaces the field's complete mapped **line range** from the line map — a valid title may be a block scalar spanning several lines, and a single-line rewrite would strand its continuation lines — preserving all surrounding bytes and any inline comment on the key line:
  - `SetState(doc, state)` — replace the `state:` field range in place (same indentation, no requoting of anything else).
  - `SetTitle(doc, title)` — replace the `title:` field range (plain scalar when the value is YAML-safe, double-quoted otherwise).
  - That YAML-safe-or-quote decision lives in **one scalar-emission helper**, shared by every path that hand-emits a scalar — title patches, capture skeletons, order entry lines — because writes never pass through a YAML encoder and each emission path is a chance to write malformed YAML ourselves. Fixtures: a capture title containing `: ` and ` #`, and an order entry for a hand-picked filename beginning with `#`, all reparse cleanly.
  - Full-content replace — the UI's raw-edit gesture; the operator's own bytes land verbatim, no normalization.
- v1 has **no log-append gesture**: log stamps are written by hands and design agents, not by vane. The schema parses `log:` for display only.

**Order document.**

- `ParseOrder(raw []byte) OrderDoc` — the two-pass parse above (recognized state-name lanes are the claimed keys), plus the line map of lane blocks and entry lines for write spans. The duplicate-lane-key rule follows the same claimed/unknown split as items: a duplicate *state-name* lane key marks the document **unreadable** — the repository-level fail-fast tier, detected on the AST key-node walk — while duplicate or singular *unknown* keys are ignored, per the spec's "a key that isn't a state name: ignored." A YAML syntax failure is likewise unreadable. Unknown-key lines are carried in the line map and excluded from ordering — and, per the spec's discard-then-prune rule, whole unknown-lane blocks (key line plus entry lines) are classified **prunable**: they're removed on the next opportunistic prune like every other discarded reference, so a misspelled lane heals instead of lingering as something a human reads and the tool ignores. One precondition guards the prune: an unknown block is **not prunable while any surviving recognized node aliases an anchor defined inside it** — pruning the definition would leave the alias dangling and our own write would render the file unreadable, the one thing a writer must never do; the AST makes the dependency check cheap. Unrelated comments and recognized lanes keep the byte-for-byte guarantee. Tests: duplicate `researching:` keys fail fast; duplicate unknown keys don't; an unknown lane block disappears on the next order write with surrounding comments intact; an unknown block defining an anchor aliased by recognized lanes survives an unrelated prune and the file stays readable.
- `OrderVersion` — the hash string, or the distinguished **absent** sentinel when no file exists. Absence is a version.
- Surgical ops: rewrite one lane's entry lines (reorder), remove one entry line, replace one filename in place (retitle), insert an entry at a position (transition-and-place), create the file fresh (first-ever ranking). Opportunistic pruning is the one sanctioned multi-line side effect, applied only when the file is already being written and only to what's classified prunable — discarded entries and unknown-lane blocks alike. Comments and recognized-lane lines otherwise survive every op byte-for-byte.

**Guarded writes.**

- `CompareAndWrite(path, expectedHash, newBytes)` — read disk, compare against the expected hash *carried by the caller* (never a fresh self-comparison), refuse on mismatch, else write. For expected-absent: create with `O_CREATE|O_EXCL` so a racing creator wins and we refuse. Best-effort detection per the spec — no locks, no fsync ceremony; the git gate is the real net.
- No-clobber finalize for capture/rename: `os.Link(src, dst)` then remove src — atomic refuse-if-exists on POSIX; report both paths on collision, temp survives.

**Round-trip tests (spec-mandated), two layers.** Stage 2, primitives: hand-formatted fixtures — item files with comments, unusual-but-valid spacing, unknown fields, quoted and unquoted scalars, **block and folded multiline scalars**; `order.yaml` fixtures with comments and unknown lane keys. Every patch asserts the full-file diff touches only the expressing lines. Stage 3, compositions: full-tree diff tests for every composite workspace gesture — transition, transition-and-place, retitle, rename-to-slug, capture finalize, state-changing save, **body-only save** (the spec's named body-edit commitment: only body bytes change, frontmatter and sidecar untouched), opportunistic pruning — including ranked and malformed/inert cases, each asserting the gesture changes only the files and lines that express it. Surgical primitives don't prove surgical gestures; both layers are the reviewable form of the commitment.

## internal/workspace

The filesystem layer: composes document ops into the spec's gestures against a discovered root. Stateless — every `Load` is a fresh read.

- `DiscoverRoot(startDir)` — the single upward walk: at each ancestor, `docs/future/roadmap/` claims the root; else any entry named `.git` (file **or** directory, `Lstat` only, never opened) claims it and walls the walk; exhaustion falls back to the start directory. Tests: nested repo inside a repo with a roadmap, worktree file-form `.git`, no markers at all.
- `Load(root) (Snapshot, error)` — enumerate `roadmap/*.md` (flat; skip the `.capture-` prefix, skip directories), parse each into a `CardInput` + raw/hash, read `order.yaml` (or absent). Error tiers: roadmap directory missing/unreadable and order.yaml unreadable are repository-level errors (serve refuses to start; per-request they become 5xx); any single bad item degrades to a flagged card in the snapshot.
- Gestures, all preflighting every affected hash before the first write, reporting partial failure plainly:
  - Capture is **two operations**, because the CLI's editor sits between them and may rewrite anything — including the title:
    - `CreateDraft(title, body)` — temp `.capture-<rand>.md` in `roadmap/` (creating the directory on demand), skeleton `title:`/`state: inbox`/`created: <process-local date>` + body; returns the temp path.
    - `FinalizeDraft(tempPath)` — rereads the **saved** temp bytes, recovers the title from those bytes, derives the slug, and no-clobber-links the *unchanged* bytes into place. Four explicit outcomes: finalized (returns filename); empty title — cancel, temp kept, path returned; non-empty title with empty slug — temp kept, path + rename-by-hand instruction returned; collision — temp kept, both paths returned. The temp file survives every non-finalized outcome.
    - The CLI runs create → editor → finalize; UI capture writes its draft from the form fields and runs the same finalize — one no-clobber, exact-bytes funnel for every path into `roadmap/`. Test: a title edited in the editor lands under the edited title's slug, hand-formatted frontmatter surviving byte-for-byte.
  - `Transition(filename, state, expectedHash, order…)` — surgical state patch; if the item is ranked in its old lane, remove that entry in the same gesture (two-file preflight); transition-and-place *moves* the entry to the destination position.
  - `Reorder(lane, filenames, expectedOrderVersion)` — one-lane order rewrite + opportunistic prune.
  - `Retitle(filename, newTitle, …)` / `RenameToSlug(filename, …)` — title line patch + no-clobber rename + in-place `order.yaml` replacement of the old filename in **every retained occurrence, active and inert alike, positions preserved** — retained entries exist so priority survives a transient parse failure, and a rename that missed the inert ones would leave them pointing at a nonexistent file, the one condition that prunes unconditionally. Occurrences already classified discardable prune under the normal opportunistic rule. Test: rename a malformed ranked card, repair its state, priority survives. `RenameToSlug` is the one-gesture repair for the mismatch flag, and refuses when the title's slug is empty — there is no destination to repair toward, and no flag to repair (see the Flag rule). `Retitle` to a title whose slug is empty patches the title and **preserves the existing filename and rank** — no rename, no order.yaml touch; the old filename simply becomes the hand-picked name the spec permits for such titles, which the Flag rule leaves unflagged. Composition test: retitle a ranked item to an all-discarded-code-points title — only the title line changes, filename and rank intact.
  - `SaveContent(filename, content, expectedHash, expectedOrderVersion)` — raw replace, but **a raw save that changes `state:` is a transition made through vane** and gets the transition's discipline. The comparison runs in **effective lanes**, not raw validity: the old effective lane (an unreadable old state is inbox, per the spec's for-all-ordering-purposes rule) versus the new valid lane; when they differ on a card actively ranked in the old lane, the standard ranked-transition cleanup applies (old-lane entry removed, two files preflighted) — so repairing an unreadable state out of inbox costs the inbox rank, exactly like any other departure. Entries are retained only when the *new* state is unreadable — still not positive evidence, same rule as pruning. Tests: valid→valid transition on a ranked card; unreadable-ranked-in-inbox repaired to a valid other lane removes the inbox entry. No placement: the card lands unranked in its new lane, like any transition-without-place. If the saved content's title now slugs differently, the server surfaces that as the mismatch flag rather than auto-renaming — renames stay explicit gestures, never side effects of a text save.

## The API contract

`internal/api/specs/vane.yml`, OpenAPI 3.0.3, generated with `//go:generate go run github.com/ogen-go/ogen/cmd/ogen@v1.20.3 --target . --clean specs/vane.yml`, mounted at `/api/v1`, generated code committed. No auth, localhost only. Every read that paints state carries hashes; every mutation carries them back; every mutation's success response returns a fresh board so the client repaints from disk truth. The guard failure is a typed `409` carrying a machine-readable reason. `item_conflict` and `order_conflict` mean the view went stale — the client reloads. `slug_collision` means a no-clobber refusal and carries **structured recovery paths** in the response body: the preserved `.capture-` temp path for capture, source and colliding destination for retitle/rename-to-slug — so the collision affordance the workspace layer guarantees survives to the browser surface (see The UI, Conflicts).

| operation | shape |
| --- | --- |
| `GET /board` | lanes in lifecycle order, each with its cards (filename/title/state/created/tags/source/log/flags/`hash`, no body) and a `rankedCount` — the first `rankedCount` cards are the lane's ranked prefix, the rest the computed unranked tail; plus `orderVersion` (hash or `"absent"`) |
| `GET /items/{filename}` | raw `content` + parsed card + `hash` |
| `POST /items` | `{title, body?}` → capture into inbox; `201 {filename}`, typed `400` validation error for an empty or empty-slug title (prevalidated — no draft file is created; the form retains its content client-side, so nothing is lost), or `409 slug_collision`. The temp-preserving `FinalizeDraft` outcomes belong to the CLI's editor flow alone |
| `PUT /items/{filename}/content` | `{content, expectedHash, expectedOrderVersion}` raw save; a state-changing save runs the ranked-transition cleanup |
| `POST /items/{filename}/state` | `{state, expectedHash, expectedOrderVersion, position?}` transition / transition-and-place |
| `PUT /order/{lane}` | `{filenames[], expectedVersion}` — `filenames` is **only the resulting ranked prefix**, never the whole displayed lane; cards absent from it stay unranked |
| `POST /items/{filename}/retitle` | `{title, expectedHash, expectedOrderVersion}` retitle + rename |
| `POST /items/{filename}/rename-to-slug` | `{expectedHash, expectedOrderVersion}` mismatch repair |

`expectedOrderVersion` is **required on every gesture that can touch `order.yaml`** — state, content save, retitle, rename-to-slug — never conditional on whether the item happens to be ranked. The board payload delivers `orderVersion` on every read (absence is the `"absent"` sentinel, a legal version), so the client always has it to carry back; the server validates it before any write. This kills the conditional-invariant class outright: a ranked item's two-file preflight and an unranked item's one-file gesture run the same contract, and the guard is never compared against a fresh read.

`position` indexes the destination lane's **ranked list only** (0…len of the lane's `order.yaml` entries after discard). It never indexes the displayed card sequence: a drop anywhere in the unranked-tail region serializes as end-of-ranked-list, because between-ness among unplaced cards is not expressible without placing them — and placing cards the operator didn't touch would smuggle extra priority judgments into a one-card gesture. The write stays minimal: one inserted (or moved) entry. The client learns the boundary from each lane's `rankedCount`; it never infers it. Tests: position 0, mid-list, len, and a tail-region drop all produce exactly one entry-line change; dragging an unranked card within its lane ranks that card alone — untouched neighbors stay unranked.

`internal/server` implements the generated `Handler` interface flo-style: constructor-injected workspace root, one `workspace.Load` per request, translation between model types and wire types at this edge only.

## The UI

`ui/` per flo/studio: Vite + TS + React 19, `//go:embed all:dist` + SPA-fallback middleware routing `/api/*` to the ogen server, `gen:api` script running `openapi-typescript` against `internal/api/specs/vane.yml` into `src/api/schema.d.ts`, calls through `openapi-fetch` in a thin `api.ts`. Vite dev server proxies `/api` for the hot-reload loop.

v1 surface, deliberately small:

- **Board** — seven lanes in lifecycle order, cards showing title (or filename when the title is unreadable), flag badges with the diagnostic on hover, and the item's log stamps as compact `stamp — note` lines in file order — the spec's human-visible provenance, on the card where it promises it. Manual browser refresh is the freshness contract; no polling.
- **Drag** — `@dnd-kit/core` + sortable: within-lane drop → `PUT /order/{lane}`; cross-lane drop → state gesture with position (transition-and-place). Position follows the ranked-list-only rule (see the API contract): while dragging over a lane's unranked tail, the drop indicator **snaps to the ranked/unranked boundary**, so the board never shows a landing spot the sidecar can't express. The only new frontend dependency beyond the flo baseline.
- **Item view/edit** — click opens a panel with the raw file content in a textarea; save is `PUT …/content` with the expected hash. Retitle and rename-to-slug are explicit buttons wired to their gestures.
- **Capture** — a board-level button, title + optional body, lands in inbox.
- **Conflicts** — the 409 family splits by what the operator needs. `item_conflict`/`order_conflict`: a plain notice ("changed on disk — reloaded") and a board refetch — reload genuinely is the answer. `slug_collision` is different: the response carries structured paths (the preserved `.capture-` temp for capture; source and colliding destination for retitle/rename-to-slug), and the UI shows them — the capture modal keeps its content on screen with both paths displayed, so the operator can retitle and retry knowing exactly where the preserved draft lives. Nothing typed is ever lost *or invisible*. Partial two-file failure surfaces the server's message verbatim.

## The CLI

- `vane [title words...]` — capture. Args joined with spaces form the initial title. Editor cascade `VANE_EDITOR` → `EDITOR`; neither set is an error naming the fix, not a guess at a default. Exit paths per the spec: slug rename, empty-title cancel (print temp path), empty-slug instruction.
- `vane serve --port N` — default port **4114**; binds `127.0.0.1` only; fail-fast tiers at startup; graceful shutdown on signal.
- `vane list` — lane-grouped, board order, flags marked; a plain renderer over the same `ComputeBoard` output the UI consumes.
- `vane state <filename|slug> <state>` — the transition gesture from the terminal (accepts the filename with or without `.md`).
- `vane version` — build info.

`-v/--verbose` re-inits `dl` at debug, per the base stack. Root command wiring in `cmd/vane/main.go`, constructors only, no globals.

## Stages

Each stage lands terminus-gated (`clean` before Michael reviews), with `docs/current/` synthesis and a CHANGELOG entry as it lands. The order runs convention-out: the file convention must be fully operable headless before any pixel exists — that's the thesis, expressed as slicing.

1. **Scaffold + model.** go.mod, Makefile skeleton, CHANGELOG.md, `internal/model` complete with slug vectors and the `ComputeBoard` determinism table tests. Exit: the entire ordering/flag semantics of the spec passes as pure-function tests.
2. **Document layer.** Item + order documents, malformed classification, hashes, all surgical patches, guarded writes, and the spec-mandated round-trip fixture suite. Exit: every gesture provably touches only its expressing lines.
3. **Workspace + CLI.** Root discovery, enumeration/Load, all gestures, capture lifecycle; `vane`, `vane list`, `vane state` working end-to-end. Exit: the composite-gesture tree-diff suite passes (see round-trip tests, layer two) and the triage scenario is performable entirely from the terminal against a real repo; vane's own repo gets `docs/future/roadmap/` and starts dogfooding.
4. **API.** `specs/vane.yml`, committed ogen output, `internal/server` handlers, tests against temp-dir workspaces covering the guard wire semantics: stale item hash → 409, stale order version → 409, expected-absent vs. racing creation, two-file preflight partial-failure reporting.
5. **UI + serve.** React board, drag gestures, edit/capture/conflict flows, embed + middleware, Makefile `frontend`/`generate`/`build` targets (`no_ui` headless tag per the base stack). Exit: the Saturday-morning triage scenario runs in the browser; manual soak against vane's own roadmap.

## Dependencies

Go: `spf13/cobra`, `michaelquigley/df` (`dl`, `dd`), `gopkg.in/yaml.v3` (direct, `internal/document` only, node-level syntax pass), `ogen-go/ogen` (generate-time pin @v1.20.3 + runtime libs). Stdlib for SHA-256, HTTP serving under ogen, filesystem. **No** sqlite, no config machinery, no git libraries — a git import anywhere is a terminus finding by census.

JS: react/react-dom 19, vite, typescript, `@dnd-kit/core`/`@dnd-kit/sortable`, `openapi-fetch`; dev: `openapi-typescript`.

## Deliberately not in this work order

- **Polling / live reload, forge ingestion tooling, cross-repo aggregation, dfw wrap, done-item reporting** — deferred by the spec with rationale; nothing here pre-builds toward them beyond the additive shapes the spec already chose.
- **Log-append gestures and any machine semantics over log notes** — v1 renders `log:`, never writes it.
- **Config cascade** — see planning decision 4; revisit when a second real setting appears.
