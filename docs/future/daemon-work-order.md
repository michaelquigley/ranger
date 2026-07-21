# the daemon — work order

*Work order for [daemon.md](daemon.md), drafted 2026-07-21 against the code as of the roadmap-images landing. Five stages, each landing green (full Go suite, UI vitest, `make build`) and terminus-gated before the next begins.*

## ground truth the plan builds on

The server is stateless per request over a single `workspace.Workspace` (`internal/server/server.go` — "a workspace root and nothing else"); every handler rebuilds from a fresh `Load()`. The ogen contract (`internal/api/specs/ranger.yml`) mounts at `/api/v1` with no project dimension; `make generate` regenerates both the Go server and `ui/src/api/schema.d.ts`, and the SPA calls exclusively through `openapi-fetch` in `ui/src/api.ts` (one `client` bound to `/api/v1`, one exported function per operation). `ui.Middleware` routes `/api/*`, serves the embedded SPA with index-fallback; `/roadmap/` (project-unaware) serves the single root's assets via `server.Assets`. `cmd/ranger/serve.go` composes the mux; `discovered()` in `main.go` builds the workspace from cwd. dfw v0.1.0's `tray.DaemonApp` supplies `{AppID, Title, IconPNG, Listen, SpawnWindow, TrayItems}` — tray-only import, no CGO.

## stage 1 — config and the project set

New package `internal/config`:

- `Config{Projects []ProjectRef, Default string, Port int}`, `ProjectRef{Root string, Name string}`; dd-bound read of `~/.config/ranger/config.yaml` (path a function parameter for tests). Read-only — ranger never writes config.
- Normalization at load, fail-loud (bootstrap tier): an absent or empty `projects` list is a load error; `~` expansion via `os.UserHomeDir`; a root still relative after expansion and cleaning is a load error naming the entry (no resolution base exists — the daemon's cwd is arbitrary); an empty `Name` — omitted and explicitly blank are the same thing; plain strings, no presence machinery — defaults to `model.Slug(filepath.Base(root))` — the convention's own grammar, so names drop into URL segments verbatim with no encoding rule anywhere; a basename that slugifies to empty is a load error demanding a `name:`; a non-empty `name:` that isn't slug-shaped (`name != model.Slug(name)`) is a load error; duplicate names (post-slugification included) are a load error naming both roots; `Default` must name a configured project (absent → first entry); `Port` absent → 4114.

New type in `internal/server` (beside `Server`): `Projects` — project resolution over a config *source*, `func() (*config.Config, error)`, consulted fresh on every resolution. The daemon's source re-reads the config file per call, so an edit takes effect on the next request and a runtime parse failure surfaces as that request's plain error (spec: the daemon stays resident, heals on the next good save); serve's source returns its synthesized config constantly. `Resolve(name) (*workspace.Workspace, error)` and the index enumeration both go through the source. No cached config, no cached health, no background probing: availability is judged by a fresh `Load()` wherever it's asked, same as every other read in the system.

Tests: config parsing, defaulting, tilde expansion, empty/absent project list rejection, duplicate/default validation, and the name grammar (basename slugification, empty-slug rejection, non-slug override rejection, post-slug collision); project resolution through the source. No user-visible change this stage; serve still runs the old wiring.

## stage 2 — the contract turns project-scoped

The breaking stage; contract, server, and SPA move together so it lands green.

- `specs/ranger.yml`: every existing path gains the `/projects/{project}` prefix (`/projects/{project}/board`, `/projects/{project}/items/{filename}/state`, …) with a shared `project` path parameter; unknown project responds 404 `errorResponse`. One new operation: `GET /projects` → `{projects: [{name, available, error?}], default}` — availability judged by a fresh load at request time. The board schema's `project` field becomes the configured name (identical to today's basename unless overridden).
- `make generate` regenerates both sides; generated files are never hand-edited.
- `internal/server`: `Server` holds `*Projects` instead of one workspace. Every handler resolves `{project}` first — miss is the 404, hit proceeds exactly as today against that workspace. One exception to "exactly as today": `CreateItem` gains a fresh `Load()` preflight after project resolution and before `CreateDraft` — capture is currently the only mutation that never reads the repository first (`CreateDraft` opens with `MkdirAll`), and under a moved root it would silently recreate the roadmap in the dead tree; the preflight makes a degraded project refuse capture with its repository error, bytes untouched. The CLI's create-on-demand capture keeps its current behavior — in-repo, cwd-discovered, directory creation is the feature there. `freshBoard` takes the project. `NewError` and `asConflict` untouched — deliberately: path-valued diagnostics and recovery fields are the census's named exception (paths inform, names address).
- Assets: the mux mounts `/roadmap/{project}/…` — resolve the project name, then serve from that root's roadmap directory with `server.Assets` unchanged (every containment property intact; resolution addition only).
- `cmd/ranger/serve.go`: builds a synthesized one-project config from `discovered()` through the *same* normalization the file-backed loader runs — name = `model.Slug(filepath.Base(root))`, an empty slug failing plainly at startup — with port from `--port`; constructs the same `Projects` + mux the daemon will use (shared assembly helper extracted here), keeps the startup `Load()` fail-fast. One normalization helper, both callers, so serve stays a true single-project instance of the same server for every root accepted today.
- SPA mechanics (UX unchanged beyond a redirect): `ui/src/api.ts` becomes a project-bound factory — `makeApi(project)` returning the operation set with `project` pre-applied, so call sites don't thread the name. `App` reads the project from the URL (`/p/{name}`); at `/`, fetch `GET /projects` and navigate to the default. `markdown.ts`'s transform gains the project prefix (`/roadmap/{project}/…`) with containment as a resolved-path check, not a spelling scan: the candidate resolves through real URL semantics (`new URL(…, base)`) and the result must sit under `/roadmap/{project}/`, else inert — browsers normalize `../`, backslash separators, and percent-encoded dot segments alike before the request ever leaves, which no route-side test can see. Vite proxy already forwards both prefixes.
- Tests: server tests move to project-scoped calls plus 404-on-unknown; UI tests cover the URL parse and the project-scoped transform. Three acceptance checks prove the census rather than assuming it: (1) a contract-path census over the parsed spec — every operation except `GET /projects` lives under `/projects/{project}`, exactly once, no unscoped or double-scoped survivor; (2) one read and one mutation driven through the real HTTP stack — generated client through generated router to handler — proving the wire paths compose (a method-level test cannot catch a stale path); (3) asset-route cross-project assertions — `/roadmap/a/…` can only ever serve from project a's root, an unknown project 404s, and the transform-plus-normalization composition is tested end to end: a `../b/images/x.png` reference in project a's item body must render inert, never producing a request that serves project b's bytes — with backslash-separated (`..\b\…`) and percent-encoded (`%2e%2e`) spellings covered by the same test; (4) the degradation-and-heal story on one server instance — two projects, one healthy and one with a broken roadmap: the index reports one available and one flagged with its error, the broken board returns the repository error while the healthy board works, then the broken root is healed *on disk* and index and board recover on the next request with no rebuild — the executable form of the error-by-tier census entry, and the test that catches an accidental startup gate or cached health. The same test asserts capture against the broken project: repository error returned, filesystem left byte-for-byte unchanged (no recreated roadmap directory, no landed draft). A separate serve-startup fail-fast test keeps the two tiers visibly distinct.

## stage 3 — the selector

- Header dropdown between the mark and the search box, fed by `GET /projects`: available projects by name; unavailable ones present, flagged, carrying their diagnostic (tooltip + disabled or warn styling — explain, don't hide). Switching navigates to `/p/{name}`; all state downstream keys off the URL, so nothing else changes hands.
- The header and selector render from the project index, independent of board success — today's `App` returns the fatal panel *instead of* the page, and that inversion is the stage's real work: the repository-level error becomes a body-region state under a live header, shown for unavailable and unknown projects alike, so a broken default can never strand the operator away from healthy projects. (The pre-board `loading…`/fatal shape survives only for the project *index* itself failing — the daemon-level error tier, where there is genuinely nothing to select.)
- `document.title` already follows `board.project`; confirm it reads the configured name.
- Tests: selector rendering states and navigation, including the trap case — an unavailable default project plus a healthy sibling: the selector renders over the body-region error and navigates to the sibling. The pre-existing UI suite keeps passing throughout.

## stage 4 — the daemon

- Dependency: `github.com/michaelquigley/dfw` (tray subpackage only).
- `cmd/ranger/daemon.go`: `ranger daemon` — validates the config once at bootstrap (unreadable, invalid, or missing file: fail-fast with a message naming the path and the fix), then hands the server a file-backed config source that re-reads per request — no per-root load gate anywhere (degradation is the point) — assembles the shared mux, and hands `tray.Daemon` the `DaemonApp`: AppID `com.michaelquigley.ranger`, title `ranger`, the binoculars-mark PNG, `Listen` binding `127.0.0.1:{port}` (config, `--port` override). `SpawnWindow` stays unset — it is dfw's webview-window hook and ranger opens no webview, ever; the menu is one `TrayItem` labeled `open board` (lowercase, house chrome voice) invoking the browser helper, plus dfw's built-in quit. The board URL comes from the *bound listener's* address, never from a config re-read — the port is bootstrap-fixed (the spec's one carve-out from the fresh-config promise) and a later `port:` edit must not change where the tray points.
- Browser opening is a small helper: `xdg-open <board-url>` on Linux, `rundll32.exe url.dll,FileProtocolHandler <board-url>` on Windows (the entry point is part of the argv — `url.dll` alone does nothing), error surfaced plainly if the opener fails. The helper's argv construction is unit-tested per platform, since the Windows branch will never run under a green Linux build. No webview, no window tracking — multiple windows are the browser's business.
- Icon: commit a PNG rendered from `favicon.svg` (asset beside the command, `go:embed`), sized per dfw's tray expectations.
- Tests: config fail-fast paths, the URL/browser helper (per-platform argv), and port precedence — the config's port is used unless `--port` was explicitly supplied (cobra's `Changed`, not a zero-value check) — are unit-testable; a bootstrap-boundary test runs the pre-tray assembly with a valid config holding one broken root and asserts construction succeeds with the project subsequently reported unavailable — the stage 2 degradation test's bootstrap-side mirror, pinning that root health never creeps into bootstrap validation; the tray loop itself is verified by hand (no headless tray).

## stage 5 — close-out

- `docs/current/`: ui-and-serve.md reshaped for project-scoped routes, the selector, and the daemon (or a split into a daemon-focused doc if it reads better); workspace-and-cli.md gains the daemon command and the serve/daemon relationship; api.md follows the contract.
- `CHANGELOG.md`: FEATURE for the daemon + selector; CHANGE recording the breaking contract restructure.
- `README.md`: the daemon becomes the headline serving story, ad-hoc serve beside it.
- Roadmap: `desktop-integration` is realized — synthesis into docs/current, then deletion on Michael's direction. `cross-repo-aggregation` untouched on horizon. Spec and work order leave `docs/future/` per the realized-document discipline.

## integration notes and gotchas

- **The client factory is the one UI seam.** Every function in `api.ts` must learn the project; binding it once in `makeApi(project)` keeps `App.tsx`'s call sites unchanged except construction. `ItemModal`/`CaptureModal` receive the bound api (or the project) via props — small, mechanical, but touches their signatures.
- **`no_ui` headless build**: the daemon compiles and runs (tray needs no webview); the board URL serves the "wasn't built in" message. Acceptable; documented, not defended against.
- **Port collision** (daemon up + ad-hoc serve on 4114): bind error surfaces plainly; nobody retries or scans. Recorded in the spec.
- **`board.project` semantics** shift from "discovered basename" to "configured name" — slug-shaped everywhere, including serve's synthesized entry, so the value changes for any root whose basename isn't already a slug; `list`/CLI output unaffected (CLI never consults config).
- **Dev loop**: vite proxies `/api` and `/roadmap` to whichever of serve/daemon is running; the `/p/{name}` route needs no proxy (SPA-side).
- **Dependency surface**: dfw v0.1.0 and its tray transitives; no CGO, no webview headers. flo already carries it — known quantity.
- **Migration**: none. No data changes, no convention changes; serve users see identical behavior (plus a URL redirect). The config file is purely additive.

## critical files

| file | stages | nature |
| --- | --- | --- |
| `internal/config/` (new) | 1 | config load + validation |
| `internal/server/projects.go` (new) | 1, 2 | project set, resolution |
| `internal/api/specs/ranger.yml` | 2 | breaking restructure + index endpoint |
| `internal/server/server.go`, `handlers.go` | 2 | per-request project resolution |
| `internal/server/assets.go` | 2 | project-scoped mount (containment unchanged) |
| `cmd/ranger/serve.go` | 2 | synthesized config, shared assembly |
| `ui/src/api.ts` | 2 | project-bound client factory |
| `ui/src/App.tsx` | 2, 3 | URL project, redirect, selector |
| `ui/src/ItemModal.tsx`, `ui/src/CaptureModal.tsx` | 2 | receive the bound api via props |
| `ui/src/markdown.ts` | 2 | project-scoped asset prefix, resolved-path containment |
| `cmd/ranger/daemon.go` (new) | 4 | tray daemon entry |
| tray icon PNG (new, `go:embed`) | 4 | the binoculars mark for the tray |
| `docs/current/*`, `CHANGELOG.md`, `README.md` | 5 | record catches up |
