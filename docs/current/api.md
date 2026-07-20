---
title: the API
created: 2026-07-14
---

# the API

Contract-first: `internal/api/specs/ranger.yml` (OpenAPI 3.0.3) is the single source of truth, with the ogen-generated server code (`ogen@v1.20.3`, `--clean`) committed beside it in `internal/api`. Hand-written handlers live in `internal/server`, implementing the generated `Handler` interface flo-style: constructor-injected workspace, every load fresh from disk — the server holds no snapshot, because files are the truth and other writers never stop. Reads load once; a mutation loads twice, deliberately: once inside the workspace gesture as its guard preflight, and once after the write to build the response board from disk truth rather than from what the mutation thinks it did. Model↔wire translation happens at this edge only. No auth; the serve binding (localhost, `/api/v1`) arrives with stage 5.

## the surface

- `GET /board` — lanes in lifecycle order; each card carries filename, title (empty when unreadable — show the filename), its readable state and created date when they exist, tags/subsystems/milestone/source/log, flags, and its content `hash`; each lane carries `rankedCount`, the boundary between the ranked prefix and the computed unranked tail. The board carries `orderVersion` — order.yaml's hash or the `"absent"` sentinel, because absence is a version.
- `GET /search?q=` — case-insensitive substring search over titles and bodies, computed against a fresh disk read per query; returns matching filenames.
- `GET /items/{filename}` — raw `content` + parsed card + `hash`; 404 for a name that doesn't exist.
- `POST /items` — capture into inbox. Empty and empty-slug titles are prevalidated to a typed 400 with no draft file written, so the form keeps its content and the tree stays clean. A slug collision is a 409 carrying the preserved `.capture-` temp path.
- `PUT /items/{filename}/content` — raw save; a state-changing save runs the ranked-transition cleanup.
- `POST /items/{filename}/state` — transition, or transition-and-place with `position`, which indexes the destination lane's ranked list only.
- `PUT /order/{lane}` — `filenames` is only the resulting ranked prefix, never the whole displayed lane.
- `POST /items/{filename}/retitle`, `POST /items/{filename}/rename-to-slug` — the rename gestures; both return the landing filename.
- `POST /items/{filename}/delete` — the operator's curation gesture (design change 2026-07-16): removes the file and its order.yaml entries in one hash-guarded gesture.

Every mutation carries `expectedHash` and/or `expectedOrderVersion` — the order version is required on *every* gesture that can touch order.yaml, never conditional on whether the item happens to be ranked, so a ranked item's two-file preflight and an unranked item's one-file gesture run the same contract. Every mutation success returns a fresh board, rebuilt from disk truth after the write.

## the 409 family

The typed conflict carries a machine-readable `reason`:

- `item_conflict` / `order_conflict` — a guard refusal: the view went stale; the client reloads. Split by which file's guard tripped.
- `slug_collision` — a no-clobber refusal, carrying structured recovery paths: `tempPath` (capture: the preserved draft), `sourcePath`/`destPath` (retitle, rename-to-slug) — so the collision affordance the workspace guarantees survives to the browser surface.

Refusals that are neither conflicts nor server faults — rename-to-slug against an unreadable or empty-slug title — are typed 400s. Everything else, including partial two-file failure, lands in the default error response with the server's message verbatim.

## tests

`internal/server` tests run against temp-dir workspaces and pin the guard wire semantics: stale item hash → 409 `item_conflict`; stale order version → 409 `order_conflict`; expected-absent against a present order.yaml refuses while the same expectation against a genuinely absent one creates it; partial two-file failure (induced with a read-only order.yaml) surfaces the plain "…but order.yaml was not updated" report; capture and retitle collisions carry their recovery paths and the preserved draft exists at the reported path; mutations return boards that reflect the write. An `httptest` round-trip proves the generated routing serves the contract.
