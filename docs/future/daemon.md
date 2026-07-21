# the daemon

*Spec, drafted 2026-07-21 from the design session realizing the `desktop-integration` roadmap item. Settled through conversation with Michael and converged through a six-round mercurius review (session `s_1YliYZO9qRJS`).*

ranger's pilot has been one repo at a time: `cd` into a project, `ranger serve`, open localhost. That shape made sense while the convention proved itself in a single repo. The practice it serves doesn't work that way — the operator ranges across a handful of products in a day, and the roadmaps are the scouting intelligence for all of that territory at once. The tool should sit where the practice sits: ambient, resident, one gesture from any board.

This spec makes ranger a tray-resident daemon that knows every root the operator cares about, serves all of them from one process, and opens browser windows on request. The convention does not move: the daemon is still just a reader over files in working trees, it still never touches git, and every write is still the same guarded, surgical gesture against whichever project the operator is working. The daemon changes where the reader lives, not what it is.

## the config

The daemon reads `~/.config/ranger/config.yaml` (dd-bound, read-only — ranger never writes its own config; the operator's editor is the config surface):

```yaml
projects:
  - root: ~/Repos/q/products/ranger
  - root: ~/Repos/q/products/archive
default: ranger
port: 4114
```

Each entry names a repository root — the directory whose `docs/future/roadmap/` is the project's roadmap. A project's name is the basename of its root passed through the slug rule below, overridable with an explicit `name:` beside the root for the rare collision. `default:` names the project the bare board URL lands on (absent, the first entry is the default). `port:` is the listen port (absent, 4114; a `--port` flag overrides). `projects:` must hold at least one entry — a daemon with nothing to serve is a config error, not a running tray. Paths may use `~`, and after expansion a root must be absolute — a relative root is a load error, because a tray daemon's working directory is whatever the desktop session felt like, and a path that resolves differently by cwd is the wrong tree waiting to be selected silently.

Names are slug-shaped — the same ASCII-mechanical rule item filenames already live by (`a`–`z`, `0`–`9`, hyphens). A basename that isn't already a slug passes through the slug rule (`My Repo` → `my-repo`); a basename that slugifies to nothing, an explicit `name:` that isn't slug-shaped, or two entries arriving at the same name is a config load error naming the fix. The payoff is that a project name drops into `/p/{name}`, the API's `{project}` segment, and `/roadmap/{name}/…` verbatim — no encoding rule exists anywhere in the system because nothing ever needs encoding, and the convention's one naming vocabulary does one more job.

There is no discovery, no registration gesture, and no auto-detection. The config is a hand-edited file — explicit, legible, judgment-gated by the same editor-and-eyes discipline as everything else in the practice.

The daemon reads the config fresh on every request — the same fresh-from-disk discipline as every other read in the system — so an edit takes effect on the next request, no restart. A mid-flight edit that breaks the file doesn't kill the daemon: requests fail plainly with the parse error until the next good save heals it. Bootstrap keeps its fail-fast, so the daemon never *starts* on a broken config. The one carve-out is `port:` — the listener binds once at bootstrap, so a port change requires a restart; `projects:` and `default:` are the live surface.

## the daemon

`ranger daemon` starts the resident process: dfw's tray mode (`dfw/tray` — deliberately not the webview, so the binary keeps building without CGO or native webview headers), holding the HTTP server on `127.0.0.1:{port}` with the tray icon carrying the binoculars mark.

The tray menu is minimal: **open board** — which opens the operator's default browser at the board URL — and **quit**. Opening is not stateful window management; it fires the URL and the browser does the rest. Multiple windows are expected and free: every window is just a browser tab pointed at a project URL, and the project selector in each window navigates independently. The daemon neither knows nor cares how many windows exist.

The daemon fail-fasts only on its own bootstrap: an unreadable config, an unbindable port. Project roots are not bootstrap — see failure posture below.

## serve, unchanged in spirit

`ranger serve` survives as the ad-hoc, zero-config path: run it anywhere inside a repo and get that one project's board, exactly as today. Under the covers it becomes the degenerate case of the daemon's server — a synthesized one-project config from the discovered root, no tray, foreground process, fail-fast on a bad repository at startup exactly as it does now. One server implementation, two entry commands, no divergence to maintain.

Port collision between a running daemon and an ad-hoc serve is handled by nobody: the bind fails, the error says so plainly, the operator picks another port. No single-instance enforcement, no port scanning — dfw deliberately omits this and so do we.

## the API

The contract becomes project-scoped: `/api/v1/projects/{project}/…` for everything that today lives at `/api/v1/…`, plus one new endpoint:

- `GET /api/v1/projects` — the project index: each configured project's name and availability (`ok`, or an error string for a root that failed its load), plus which project is the default. This is what the selector renders, and what the bare `/` consults to redirect.

Project *names* address everything on the wire; filesystem roots stay in the config and route nothing. Paths may still *inform*: error diagnostics and the collision-recovery fields (`tempPath`, `sourcePath`, `destPath`) describe the operator's own disk and stay absolute, because the operator's next move is opening that file — flo carves exactly this exception for `localPath`. Addressing by name is what keeps the wire contract stable when a root moves on disk; informative paths are what keep an error actionable.

The asset route gains the same scoping: `/roadmap/{project}/…` serves the named project's roadmap files, with every containment property the route has today (root-confined, no symlink or dot-prefixed component, categorically no git metadata). The modal's URL transform carries the current project into the prefix, and its containment guarantee is defined *after* browser normalization, never by inspecting spellings: the transformed reference must resolve — real URL semantics — to a path under `/roadmap/{project}/`, or it renders inert (flo's document-view precedent). Parent traversal in any encoding — `../`, backslash separators, percent-encoded dot segments — normalizes across the project boundary before the request ever leaves and would silently misattribute another project's content; one resolved-path prefix check absorbs every spelling at once.

This is a breaking change to `specs/ranger.yml`, regenerated on both sides (ogen server, TypeScript schema) so the contract and the client cannot diverge. v0.1.x absorbs breaks like this by design.

## the UI

The project lives in the URL — `/p/{name}` — so windows are independently addressable, bookmarkable, and multiple windows on different projects need no coordination. The bare `/` redirects to the default project. No router library: the SPA reads the path segment at load and the selector navigates.

The selector itself is a dropdown in the header, beside the mark and project name — flo's practice-selector pattern. It renders the project index: available projects by name, unavailable ones present but flagged with their diagnostic rather than hidden (the same explain-don't-hide posture the board takes toward malformed items). The header and selector render from the project index, independent of any one board's success — a failed project shows its error in the board region *under* a live selector, never as a page-replacing panel, so the healthy projects stay one click away no matter which project the URL landed on. Switching projects is a navigation, not a state mutation; everything downstream — board fetch, search, capture, asset URLs — keys off the URL's project.

Capture from the board captures into the window's current project. The CLI gestures (`ranger` capture, `list`, `state`) are untouched: they remain cwd-discovery in-repo gestures and do not consult the config at all.

## failure posture

Error by tier, recorded deliberately:

- **Daemon bootstrap** — unreadable config, unbindable port: fail-fast at startup. These are the daemon's own preconditions.
- **Config under a running daemon** — an edit that breaks the file: the daemon stays resident, requests fail plainly with the parse error, and the next good save heals on the next request. Never a crash, never a cached last-good masquerading as current.
- **Project roots under the daemon** — a missing or unreadable roadmap directory, an unreadable order.yaml: *not* fatal, not at startup and not later. The project degrades — flagged in the index with its error, its board requests returning the repository-level error plainly — and recovers on the next request after the root heals, because every load is a fresh read and the daemon holds no snapshot. Every gesture against a degraded project — capture included — refuses with that error before any byte is written; a degraded project is read-broken *and* write-refused, never a tree that mutations quietly recreate. A daemon with six roots never dies because one repo moved.
- **Ad-hoc serve** — fail-fast on a bad repository at startup, as today. One root is the whole point of the process; degradation has nothing to degrade to.
- **Items within a healthy project** — unchanged: a bad item is a flagged card, never a failed board.

## seam census

- **model / transport** — *separate, with a named exception.* Project names address everything on the wire; roots stay config-private and route nothing. Path-valued *diagnostics* — load-error messages, guard-conflict messages, collision recovery fields — remain absolute: they describe the operator's own machine, and their value is being openable (flo's `localPath` precedent). Why: names keep the contract stable across disk moves; absolute diagnostics keep errors actionable. Revisit: if ranger ever serves beyond the operator's own machine.
- **contract circumvention** — *enforce.* The ogen contract remains the SPA's one door for API calls; the asset route (`/roadmap/{project}/…`) is the one deliberate non-contract surface, static-file-shaped, established and reviewed 2026-07-20 — its existence is a recorded decision, not a bypass. The tray reaches nothing directly: it opens URLs, full stop.
- **error by tier** — *recorded above.* Daemon degrades per-project; serve and daemon-bootstrap fail fast; items flag. A failure handled at the wrong tier is a review finding.
- **model / render** — *not live.* The board computation and domain layer are untouched by this work.

## scenarios

Morning: the daemon is in the tray from login. Michael clicks **open board**, lands on the default project, flips the selector to `archive`, triages two items, drags one to building. A second window opens on `ranger` for a capture mid-thought. Neither window knows about the other; both are just URLs.

A root moves: `~/Repos/q/products/anpheq` gets renamed during a reorganization. The daemon's selector shows `anpheq — roadmap directory not found`; every other project works untouched. Michael fixes the config line in his editor; the next request reads the new path. No restart.

Ad-hoc: a scratch repo not worth a config entry. `cd` in, `ranger serve --port 4200`, the board opens on that one project, ctrl-C when done. The daemon never knew.

## deferred (and why)

- **Cross-repo aggregation** — one board over every project — stays on horizon as its own item. The selector is deliberately additive toward it: a future aggregate view is another entry in the same project-scoped world, not a rework.
- **Root discovery / registration gestures** — the config file is enough until the project list churns often enough to make hand-editing a felt cost. It hasn't yet.
- **Single-instance enforcement** — dfw omits it on principle; the bind error is honest and the operator is one person.
- **Webview windows** — the browser is the window manager. Revisit only if browser chrome becomes a felt friction, which would reopen the CGO cost dfw's split exists to avoid.
- **macOS** — dfw defers it; ranger inherits the deferral.
- **Live board reload** — separate roadmap item (`live-board-reload`, researching); the manual-refresh contract is unchanged by this spec, and a future reload mechanism slots into the project-scoped world unchanged.
- **Tray-menu per-project entries** — the menu opens *the board*; the selector owns project choice. Revisit if tray-to-specific-project turns out to be a real reach-for.
