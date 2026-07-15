---
title: the UI and serve
created: 2026-07-15
---

# the UI and serve

`vane serve --port N` (default 4114) presents the localhost board: bound to `127.0.0.1` only, repository-level fail-fast at startup (missing/unreadable roadmap directory, unreadable order.yaml), graceful shutdown on signal. The ogen server mounts at `/api/v1`; a small middleware routes `/api/*` there and serves the embedded SPA for everything else, falling back to `index.html`. `-tags no_ui` builds a headless binary whose middleware serves the API and says plainly that the board wasn't built in.

`ui/` follows the flo pattern: Vite + TypeScript + React 19, built into `dist/` and embedded via `go:embed all:dist` (builds not committed; `make build` runs the frontend first). The TypeScript client is generated from the same contract the server is — `npm run gen:api` runs `openapi-typescript` against `internal/api/specs/vane.yml` into `src/api/schema.d.ts`, and all calls go through `openapi-fetch` in a thin `src/api.ts` — so the server contract and the client types cannot diverge. The Vite dev server proxies `/api` to a running `vane serve` for the hot-reload loop.

## the surface

- **Board** — seven lanes in lifecycle order. Cards show the title (or the filename when the title is unreadable), flag badges with the diagnostic on hover, and log stamps as compact `stamp — note` lines in file order. A dashed rule marks each lane's ranked/unranked boundary. Manual browser refresh is the freshness contract; no polling.
- **Drag** — `@dnd-kit/core` + sortable. A within-lane drop computes the lane's resulting ranked prefix and PUTs it to `/order/{lane}`: dragging an unranked card ranks that card alone, untouched neighbors stay unranked, and a drop anywhere in the unranked tail serializes as end-of-ranked-list — between-ness among unplaced cards is not expressible without placing them. A cross-lane drop is transition-and-place, position clamped to the destination's ranked list.
- **Item view/edit** — click opens a panel with the raw file bytes in a textarea; save is `PUT …/content` with the expected hash. Retitle and rename-to-slug are explicit buttons; rename-to-slug appears only on a mismatch-flagged card.
- **Capture** — a board-level button, title + optional body, lands in inbox.
- **Conflicts** — every mutation resolves through one outcome path. `item_conflict`/`order_conflict`: a dismissable notice ("changed on disk — reloaded") and a board refetch. `slug_collision`: the capture modal and item panel keep their content on screen and show the structured recovery paths — the preserved `.capture-` draft for capture, source and colliding destination for the renames. Validation refusals show inline with nothing written. Partial two-file failure and any other fault surface the server's message verbatim. Every successful mutation repaints from the fresh board the server returned — disk truth, never what the client thinks it did.

## build

`make build` = the frontend (npm install + build) then `go install ./...`, per the archive repo's Makefile shape; `make headless` installs with `no_ui` and no frontend step; `make generate` regenerates both sides of the contract (ogen server code, TypeScript schema) from `specs/vane.yml`; `make test` runs the suite uncached plus `go vet`; `make clean` removes the installed binary, `dist/`, and `node_modules/`.
