---
title: the daemon
created: 2026-07-21
---

# the daemon

`ranger daemon` is the resident serving story: a tray process that knows every root the operator cares about, serves all of them from one HTTP server on `127.0.0.1:{port}`, and opens browser windows on request. The convention does not move — the daemon is still just a reader over files in working trees, it never touches git, and every write is the same guarded, surgical gesture against whichever project the operator is working. The daemon changes where the reader lives, not what it is.

## the config

The daemon reads `~/.config/ranger/config.yaml` (dd-bound, read-only — ranger never writes its own config; the operator's editor is the config surface):

```yaml
projects:
  - root: ~/Repos/q/products/ranger
  - root: ~/Repos/q/products/archive
default: ranger
port: 4114
```

Each entry names a repository root — the directory whose `docs/future/roadmap/` is the project's roadmap. A project's name defaults to the slug of its root's basename (`My Repo` → `my-repo`, the same ASCII-mechanical rule item filenames live by), overridable with an explicit `name:` for the rare collision. Because names are slug-shaped, they drop into `/p/{name}`, the API's `{project}` segment, and `/roadmap/{name}/…` verbatim — no encoding rule exists anywhere in the system. `default:` names the project the bare board URL lands on (absent, the first entry); `port:` is the listen port (absent, 4114; an explicit `--port` flag wins — cobra's `Changed`, not a zero-value check).

Normalization is fail-loud at load: an absent or empty `projects` list, a root still relative after `~` expansion (a tray daemon's cwd is whatever the desktop session felt like), a basename that slugifies to nothing, a `name:` that isn't slug-shaped, duplicate names, or a `default:` naming no configured project — each is a load error naming the fix. There is no discovery, no registration gesture, no auto-detection: the config is a hand-edited file, judgment-gated by the same editor-and-eyes discipline as everything else in the practice.

The daemon reads the config fresh on every request — the same fresh-from-disk discipline as every other read in the system — so an edit takes effect on the next request, no restart. A mid-flight edit that breaks the file doesn't kill the daemon: requests fail plainly with the parse error until the next good save heals it. The one carve-out is `port:` — the listener binds once at bootstrap, so a port change requires a restart, and the tray's board URL comes from the bound listener's address, never from a config re-read.

## the tray

dfw's tray mode (`dfw/tray` — deliberately not the webview, so the binary keeps building without CGO or native webview headers), with the binoculars mark as the icon — a tray-simplified variant of the favicon drawing (`cmd/ranger/tray-icon.svg` beside its rendered `tray-icon.png`, `go:embed`ded): light monochrome to sit with the panel's symbolic icons, hinge ring and highlight arcs dropped since they melt below 24px, rendered at 128px so the panel's downscale stays smooth. The menu is minimal: **open board** — the operator's default browser at the board URL (`xdg-open` on linux; `rundll32.exe url.dll,FileProtocolHandler` on windows, where the entry point is part of the argv) — and dfw's built-in quit. Opening is not window management: the daemon fires the URL and the browser does the rest. Multiple windows are expected and free — every window is a browser tab pointed at a project URL, navigating independently through the selector; the daemon neither knows nor cares how many exist.

## error by tier

Recorded deliberately; a failure handled at the wrong tier is a review finding.

- **Daemon bootstrap** — an unreadable, invalid, or missing config, an unbindable port: fail-fast at startup, the message naming the path and the fix. These are the daemon's own preconditions.
- **Config under a running daemon** — an edit that breaks the file: the daemon stays resident, requests fail plainly with the parse error, the next good save heals on the next request. Never a crash, never a cached last-good masquerading as current.
- **Project roots** — a missing or unreadable roadmap directory, an unreadable order.yaml: *not* fatal, not at startup (root health never gates bootstrap) and not later. The project degrades — flagged in the index with its error, board requests returning the repository error plainly — and recovers on the next request after the root heals, because every load is a fresh read and the daemon holds no snapshot. Every gesture against a degraded project refuses before any byte is written — capture included, via a fresh-load preflight, since capture is otherwise the one mutation that never reads the repository first and would silently recreate the roadmap in a dead tree. A daemon with six roots never dies because one repo moved.
- **Ad-hoc serve** — fail-fast on a bad repository at startup: one root is the whole point of the process; degradation has nothing to degrade to.
- **Items within a healthy project** — unchanged: a bad item is a flagged card, never a failed board.

## serve, the degenerate case

`ranger serve` survives as the ad-hoc, zero-config path: run it anywhere inside a repo and get that one project's board. Under the covers it is the daemon's server with a synthesized one-project config from the discovered root — built through the *same* normalization the file-backed loader runs, so the name grammar holds — no tray, foreground, fail-fast at startup. One server implementation, two entry commands, no divergence to maintain.

Port collision between a running daemon and an ad-hoc serve is handled by nobody: the bind fails, the error says so plainly, the operator picks another port. No single-instance enforcement, no port scanning.

## deliberately not

- **Webview windows** — the browser is the window manager; dfw's `SpawnWindow` hook stays unset, keeping the CGO cost dfw's tray/webview split exists to avoid. Revisit only if browser chrome becomes a felt friction.
- **Root discovery / registration gestures** — the config file is enough until the project list churns often enough to make hand-editing a felt cost. It hasn't yet.
- **Single-instance enforcement** — dfw omits it on principle; the bind error is honest and the operator is one person.
- **Tray-menu per-project entries** — the menu opens *the board*; the selector owns project choice. Revisit if tray-to-specific-project turns out to be a real reach-for.
- **macOS** — dfw defers it; ranger inherits the deferral.
