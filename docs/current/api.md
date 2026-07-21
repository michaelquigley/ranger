---
title: the API
created: 2026-07-14
---

# the API

Contract-first: `internal/api/specs/ranger.yml` (OpenAPI 3.0.3) is the single source of truth, with the ogen-generated server code (`--clean`) committed beside it in `internal/api`. Hand-written handlers live in `internal/server`, implementing the generated `Handler` interface flo-style: constructor-injected project set, every load fresh from disk — the server holds no snapshot, because files are the truth and other writers never stop. Reads load once; a mutation loads twice, deliberately: once inside the workspace gesture as its guard preflight, and once after the write to build the response board from disk truth rather than from what the mutation thinks it did. Model↔wire translation happens at this edge only. No auth; the binding is localhost-only under both serve and the daemon.

The contract is project-scoped: every operation except the index lives under `/projects/{project}`, and every handler resolves the project name first — through a config source consulted fresh per call — before proceeding against that workspace. An unknown name is a typed 404; a config-source failure is the request's plain error. Project *names* address everything on the wire; filesystem roots stay in the config and route nothing. Paths may still *inform*: error diagnostics and the collision-recovery fields (`tempPath`, `sourcePath`, `destPath`) describe the operator's own disk and stay absolute, because the operator's next move is opening that file. Addressing by name keeps the wire contract stable when a root moves on disk; informative paths keep an error actionable.

## the surface

- `GET /projects` — the project index: each configured project's name and availability (`available` plus an `error` diagnostic for a root that failed its load, judged by a fresh load at request time), and which project is the default. This is what the selector renders, and what the bare `/` consults to redirect.
- `GET /projects/{project}/board` — lanes in lifecycle order; each card carries filename, title (empty when unreadable — show the filename), its readable state and created date when they exist, tags/subsystems/milestone/source/log, flags, and its content `hash`; each lane carries `rankedCount`, the boundary between the ranked prefix and the computed unranked tail. The board carries `orderVersion` — order.yaml's hash or the `"absent"` sentinel, because absence is a version — and `project`, the configured name (slug-shaped everywhere, serve's synthesized entry included).
- `GET /projects/{project}/search?q=` — case-insensitive substring search over titles and bodies, computed against a fresh disk read per query; returns matching filenames.
- `GET /projects/{project}/items/{filename}` — raw `content` + parsed card + `hash`; 404 for a name that doesn't exist.
- `POST /projects/{project}/items` — capture into inbox. Empty and empty-slug titles are prevalidated to a typed 400 with no draft file written. A fresh-load preflight runs between project resolution and the draft write, so a degraded project refuses capture with its repository error, bytes untouched. A slug collision is a 409 carrying the preserved `.capture-` temp path.
- `PUT /projects/{project}/items/{filename}/content` — raw save; a state-changing save runs the ranked-transition cleanup.
- `POST /projects/{project}/items/{filename}/state` — transition, or transition-and-place with `position`, which indexes the destination lane's ranked list only.
- `PUT /projects/{project}/order/{lane}` — `filenames` is only the resulting ranked prefix, never the whole displayed lane.
- `POST /projects/{project}/items/{filename}/retitle`, `POST /projects/{project}/items/{filename}/rename-to-slug` — the rename gestures; both return the landing filename.
- `POST /projects/{project}/items/{filename}/delete` — the operator's curation gesture: removes the file and its order.yaml entries in one hash-guarded gesture.

Every mutation carries `expectedHash` and/or `expectedOrderVersion` — the order version is required on *every* gesture that can touch order.yaml, never conditional on whether the item happens to be ranked. Every mutation success returns a fresh board, rebuilt from disk truth after the write.

## the error family

The typed conflict (409) carries a machine-readable `reason`:

- `item_conflict` / `order_conflict` — a guard refusal: the view went stale; the client reloads. Split by which file's guard tripped.
- `slug_collision` — a no-clobber refusal, carrying structured recovery paths: `tempPath` (capture: the preserved draft), `sourcePath`/`destPath` (retitle, rename-to-slug) — so the collision affordance the workspace guarantees survives to the browser surface.

An unknown `{project}` is a 404 `errorResponse` on every scoped operation. Refusals that are neither conflicts nor server faults — rename-to-slug against an unreadable or empty-slug title — are typed 400s. Everything else, including partial two-file failure and repository-level errors on a degraded project, lands in the default error response with the server's message verbatim.

## tests

`internal/server` tests run against temp-dir workspaces through project-scoped calls and pin the guard wire semantics: stale item hash → 409 `item_conflict`; stale order version → 409 `order_conflict`; expected-absent against a present order.yaml refuses while the same expectation against a genuinely absent one creates it; partial two-file failure surfaces the plain "…but order.yaml was not updated" report; capture and retitle collisions carry their recovery paths; mutations return boards that reflect the write; an unknown project 404s.

Four acceptance checks prove the project scoping rather than assuming it: a contract-path census over the parsed spec (every operation except `GET /projects` under `/projects/{project}`, exactly once); one read and one mutation driven through the generated client and generated router, proving the wire paths compose; cross-project asset assertions (`/roadmap/a/…` can only serve project a's bytes, unknown projects 404, traversal spellings never cross); and the degradation-and-heal story on one server instance — index flags the broken root with its error, its board returns the repository error while the healthy board works, capture refuses byte-for-byte untouched, and an on-disk heal recovers index and board on the next request with no rebuild.
