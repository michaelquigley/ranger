# Ranger v2 work order

status: reviewed and ready to build. Mercurius session `s_Fm1UH1qg9BAU` reviewed both documents over six rounds and returned `ready_to_build` on 2026-11-24. Stage 1 implementation is in the working tree for operator review.

Paired with the [spec](ranger-v2.md). Its "The first design" section is the authority on behavior. This document holds the mechanism: where the code goes, in what order, and what each stage has to prove.

## delivery

Six stages, landed in order. Each stage ends with `make test` green, terminus `clean` (every finding resolved or vetoed by Michael), its behavior synthesized into `docs/current/`, and a `CHANGELOG.md` entry under `## Unreleased`. Then it stops for Michael's review. The operator commits.

The posture is the project's: prototype-first. The format can change during the trial, so build the plain thing. Do not add corner-case hardening the spec does not ask for.

The trial can begin once stage 3 lands. A hand-migrated project is then fully readable in the UI and the CLI, and agents make every edit by hand, which is the premise of the model. Stages 4 and 5 add the operator's gestures while the trial runs.

One precondition sits outside this repository. Stage 2 widens the status read's scope for v2 projects, and the terminus quality `projects/ranger/no-git-integration` used to sanction it for the roadmap directory only. The quality was amended on 2026-11-24 to allow `docs/future/` for a v2 project from the one invocation site. That amendment has to be committed in terminus-canon before stage 2 is reviewed.

## grounding

What exists, and what v2 builds beside.

- `internal/model` is the pure v1 domain: five states including `researching`, `Slug`, the flags, and `ComputeBoard`. `internal/slug` owns the shared slug rule. `model.Slug` is the compatibility entry point used by v1, `internal/config` and the server.
- `internal/document` is the only package that imports `gopkg.in/yaml.v3`, which the terminus quality `surgical-writes` requires. Its two-pass parse (`parseMapping`, `scanSpans`, `nodeToPlain`), its scalar emission (`emitScalar`) and its span patch are unexported. `Hash`, `CompareAndWrite`, `FinalizeLink`, `VersionAbsent`, `ConflictError` and `CollisionError` are exported.
- `internal/workspace` holds v1's load and gestures against `docs/future/roadmap/`, `DiscoverRoot`, and the program's one `git status` invocation in `gitstatus.go`.
- `internal/server` resolves a project name to a v1 workspace (`Projects.Resolve`), judges availability by a v1 load (`Projects.Index`), and translates model to wire in its handlers. `buildMux` in `cmd/ranger/serve.go` mounts the API, the `/roadmap/` asset route and the SPA.
- `internal/api/specs/ranger.yml` is the contract. `projectStatus` has no convention field. `census_test.go` requires every operation except `GET /projects` to sit under `/projects/{project}`.
- `ui/src`: `App` chooses `ProjectBoard` or `DefaultRedirect` from the URL. There is no router. `ProjectSelector` is a helper inside `App.tsx`. Tests are vitest over pure functions; there are no component tests.
- `cmd/ranger`: the root command is capture, beside `list`, `state`, `serve`, `daemon`, `desktop` and `version`. `discovered()` wraps `DiscoverRoot`.

## shared mechanism

**Packages.** v2 lives in `internal/v2/document`, `internal/v2/model` and `internal/v2/workspace`. `internal/document` gains one new file that exports the convention-neutral primitives v2 needs: the two-pass field parse with write spans, a top-level sequence parse with line numbers, scalar emission, and span replacement. No existing function in `internal/document` changes, and its byte-exact fixture suite passing untouched is the proof. v2 packages never import `yaml.v3`. The slug rule moves to a leaf package of its own. `model.Slug` stays as a pass-through, so no v1 call site or test changes, and `internal/v2/model` imports nothing from `internal/model`.

**Detection.** When `docs/future/.ranger/version.yaml` exists and reads `version: 2` the project is v2. Without a `docs/future/.ranger/` directory it is v1, including the no-roadmap-yet case v1 already handles. A `.ranger/` directory with no `version.yaml`, a `version.yaml` that cannot be read, and any version other than 2 are each a project error naming the file. `version.yaml` wins: in a v2 project a `docs/future/roadmap/` directory is an item when it holds an `index.md`, and is otherwise left alone like anything else there. Ranger never writes `version.yaml`. One function answers this for the CLI and the server. `DiscoverRoot` gains `docs/future/.ranger/` as a claim beside `docs/future/roadmap/`. In a tree without that directory it walks exactly as before, and the `.git` check stays `Lstat` only.

**Reading a project.** Items are the directories directly under `docs/future/` whose name does not begin with `.` and that hold an `index.md`. An item's key is its directory name: views, links and every lookup use it. The slug of the title is used at capture and for the mismatch flag, and nowhere else. For each, read `index.md` (raw bytes, hash, parsed), whether `design.md` exists, `plan.md` if present, the stage files, and the names of its other files. Then read `.ranger/roadmap.yaml` and `.ranger/queue.yaml`. A view file holds one slug per line. A flow-style list is read as well, and becomes one per line when a write targets the file, as v1 does for a flow lane. An absent view file is an empty list with the absent version, and so is an empty or comment-only one with its own version. An unreadable one is that view's error and nothing else's. Every load is a fresh read, and a load writes nothing.

**Symlinks.** The v2 reader follows none. Before it opens or lists anything it checks each project-relative component with `Lstat`, as the asset handler does. A symlink at `docs`, `docs/future`, `.ranger` or `version.yaml` is a project error naming it, and a symlinked view file is that view's error. A symlinked directory under `docs/future` is not an item. A symlinked `index.md`, `design.md`, `plan.md`, `stages` directory or stage file counts as missing. `DiscoverRoot` tests the `.ranger` claim with `Lstat`, so a symlinked one still claims the root and detection reports it. v1's reader does not change. The check is best-effort against the tree as found: a path swapped for a symlink between the check and the open is out of scope, root discovery's existence probes are not reads, and the asset route keeps checking from the directory it serves.

**Item fields.** The claimed keys are v1's plus `since`, `kind`, `includes`, `requires`, `blocked_on` and `needs`. `title`, `state` and `created` stay required. Each field validates independently, as in v1: a bad `kind` is a malformed flag and the item keeps its lane. An unreadable `state` is Inbox. A link entry is `slug` or `project:slug`, split at the first colon.

**The body.** A line scan, with no markdown library. Outside fenced code blocks, a list line starting `[ ]` or `[x]` is a checkbox. Checkboxes are counted in the prompt only, above the first `##` heading. Links are read across the whole body: an inline link with a relative target is judged by the shape of its path. Resolve it lexically from the item's directory. When it lands at `docs/future/<name>/…`, with `<name>` not the item's own directory and not dot-prefixed, the target slug is `<name>`, derived before any existence check. An item named `<name>` makes it a link. No directory named `<name>` makes it a missing target. A directory that is not an item makes it nothing. Reference-style links are not read. Only `index.md` is scanned in this work order.

**Plans and stages.** `plan.md` frontmatter is optional and claims `mercurius` and `stages`. A stage is `stages/<name>.md`, claiming `title`, `status`, `terminus`, `needs` and `accepted`. A listed stage with no file, a stage file that is not listed, and a plan or stage file that does not parse are each a flag on the item.

**The model.** Pure functions over the loaded project.

- **Views.** Lanes run Inbox, Horizon, Pending, Building, Evaluating. Inbox holds every `inbox` item in both views, whether or not a view file lists it, by `created` then slug. Horizon and Pending hold the view's members in list order. Building and Evaluating hold the view's members by `since`, newest first, undated last, then slug. The first occurrence of a slug in a view file wins.
- **Links.** Each item's links out, typed and from the body, and its links in.
- **Where an item is in Building, and its verdict.** Progress follows the ordered `stages` list, not the files found. A listed stage that is missing or does not parse holds its position as not started and not accepted. Every verdict carries its broker's name. With a `plan.md`, apply these in order and stop at the first that holds:
    1. **Complete.** The plan lists at least one stage and every listed stage carries a valid `accepted`. The row reads all stages accepted, with no verdict.
    2. **Started.** Any listed stage carries a `status` or an `accepted`. The row reads "stage k of n" for the first stage without `accepted`, with its status and its `terminus` value, empty when it has posted none.
    3. **The plan itself.** Plan in review until `mercurius` begins `ready_to_build`, then ready to hand off, or no stages listed when the list is empty. The verdict is the `mercurius` value, empty when none is posted.

    Without a plan: designing when `design.md` exists, otherwise no design yet.
- **Needs you.** Every `needs` line on an item or a stage, plus every stage that is `built` without `accepted`.
- **Flags.** The spec's list. A flag's owner is an item, or a view line named by its view file and slug. The in-neither-view flag needs both views readable: an absent view file is an empty view, and an unreadable one is unknown. A missing link target, typed or from the body, is flagged on the item that wrote the link. A `project:slug` link is never flagged.

## stage 1: read the files

- The new export file in `internal/document`.
- `internal/v2/document`: the item, plan, stage and view parsers, and the body scan.
- `internal/v2/model`.
- `internal/v2/workspace`: detection, discovery and load.
- The CLI reads: `ranger roadmap`, `ranger queue`, `ranger needs`, `ranger show <item>` and `ranger flags`, each a plain renderer over the model. In a v2 project `list` names `roadmap` and `queue` and stops, and capture and `state` say they are not built yet. In a v1 project the five new commands name the convention and stop. Registering them reserves their names as first words in every project, which the spec records as the one v1 change. A quoted multiword title reaches capture as one argument. A quoted reserved word on its own still invokes the command.
- `docs/current/v2-convention.md`: the convention as built, written for an agent with a text editor, ending with the paragraph a trial project's `AGENTS.md` carries. It shows `version.yaml`, and the one link form ranger reads, with an example.
- Fixtures: the spec's two rehearsals as testdata trees, the BAAB item with its plan and one stage and the Screeningroom punchlist with its broken-out item.

Acceptance:

- Model tests pin the view ordering in every lane, links in both directions, Needs you, each flag, and each position within Building.
- A plan that lists no stages is not complete and raises no still-Building flag. A plan whose listed stage has no file is not complete either.
- A complete plan with no `mercurius` line reads as all stages accepted. A plan whose unstarted stage is listed ahead of an accepted one reads that unstarted stage as current with no verdict, not the plan's.
- A two-stage plan pins the queue's verdict at each step: the plan's verdict before stage 1 starts, stage 1's while it is current, and none once stage 1 is accepted and stage 2 has not started.
- Every CLI read leaves the fixture tree byte-identical, checked by a whole-tree hash before and after.
- An unreadable `queue.yaml` still yields the roadmap view, and the reverse. An item listed only in the unreadable view is not flagged as in neither view.
- The backticked name of a real item produces no link. A link inside a fenced block produces none, and neither does a reference-style link to a real item.
- A `project:slug` link is shown and never flagged.
- A title-only edit leaves the item in its view and its incoming links resolved, with the mismatch flag raised.
- A checkbox under a `##` heading in the Screeningroom fixture does not change the count.
- One fixture removes an item that a body link, a typed link and a view file all name, and asserts each flag: two on the referring item and one on the view line.
- v1's suites pass unmodified, and a v1 fixture project behaves as before under every v1 command.
- In a v1 project, a quoted multiword title that begins with a new command word is captured, and the same words unquoted reach the command.
- A fixture holding only `version.yaml` and Inbox items, with no view files, is read as v2. A v2 fixture with an item named `roadmap` is read as v2, with that item in its lane. A `.ranger/` directory without `version.yaml` errors, naming the file.
- Symlinks are refused, and the fixture is hostile: a symlinked item directory, a symlinked `index.md` and a symlinked view file each point at real content elsewhere, and none of that content or its names reaches any CLI output. A symlinked `.ranger/` or `version.yaml` is a project error naming it.

## stage 2: serve the reads

- Contract: `convention` on `projectStatus`. `GET /projects/{project}/v2/project` returns one payload: both views as lanes of slugs, an item summary carrying everything the spec's Roadmap card and Queue row show, Needs you, the flags, the two view versions, and the git marks. `GET /projects/{project}/v2/items/{slug}` returns the raw `index.md`, its hash, the summary, links out and in, the plan with each stage's facts, and the item's file list. Regenerate both sides with `make generate`.
- Server: resolve a project to its convention. The index judges availability with the right loader, and `ranger serve` fails fast with it. A v1 operation against a v2 project, and the reverse, is a typed 400 naming the convention.
- `/future/{project}/…` serves `docs/future/` read-only through the existing `Assets` handler, its confinement unchanged. Dot-prefixed segments are refused, so `.ranger/` is never served.
- Git: move the one `git status` invocation into a shared function that takes the scope. v1's `GitStatus` keeps its signature and its results, and its body calls the shared function with the roadmap directory. v2 calls it with `docs/future` and reports, per item and per view file, unstaged changes, staged changes, and new. A path maps to the item whose directory holds it, and the marks accumulate across every path under the item. An item is new when its `index.md` is untracked, and any other untracked file under it is an unstaged change. A rename is attributed to its destination path only, as v1's parser reads it. Unknown stays absent, never clean. The program still has exactly one `exec` site, and it runs once per project per request that reports status, as v1 does. Both views read that one result. The project index's dirty verdict for a v2 project is true when that result holds any entry under `docs/future/`, from the same single call per project per index request, so a deleted item still marks the project. v1's verdict and scope are unchanged. Reword the git rule in this repository's `AGENTS.md` for the v2 scope.
- Error tiers as the seam census records them.

Acceptance:

- Server tests through project-scoped calls show the payload reflecting the fixture.
- With one view unreadable, that view carries its error and the other view and every item are intact. A bad item is flagged, never a failed project.
- Wrong-convention calls refuse. The census test still holds.
- The asset route serves neither `.ranger/` nor another project's bytes.
- A git-backed fixture shows each mark, including a tracked file changed and never staged, and a git-less one shows none. An untracked file added to a committed item reads as an unstaged change, not as new. One item with a staged `index.md` and an unstaged `plan.md` shows both marks, and so does an item whose `index.md` is edited again after staging. After a tracked item is deleted, the index marks the project dirty. The status call runs once for a v2 project payload, and once per project for an index request over a v1 and a v2 project. v1's dirty tests pass unmodified.

## stage 3: show the reads

- `App` reads the index, then renders `ProjectBoard`, unchanged, for a v1 project and the v2 shell for a v2 project. `ProjectSelector` moves to its own module so both use it.
- The v2 shell carries the five surfaces of the spec's table, read-only. The surface is in the URL with no router library: `/p/{name}` is the roadmap, then `/p/{name}/queue`, `/needs`, `/flags` and `/i/{slug}`.
- The item page renders the prompt and its supporting sections as v1's modal does, then the fields, links out and in, the plan with stage facts, the design, plan and stage text fetched from `/future/`, the file list, and the flags. Relative URLs in each rendered file resolve from that file's own directory: `index.md`, `design.md` and `plan.md` against `/future/{project}/{slug}/`, and a stage file against `/future/{project}/{slug}/stages/`. The containment check is `makeRoadmapUrl`'s, confined to the project.
- The live poll is v1's: two seconds, visibility-gated, identity-preserving. An open item page re-reads its detail and the design, plan and stage text it shows on every tick, because those files change while `index.md` does not. It fetches that text with the browser cache bypassed (`cache: "no-store"`), since the asset route sends no cache directive.
- Marks for needs-you, flagged and git.

Placement here is a starting position. Expect a hands-on loop.

Acceptance:

- Vitest covers the new pure helpers.
- A smoke run shows both fixtures on every surface. It then edits only the body of a stage file that predates the run, so a cached copy would show, and sees the open item page update. A stage's link to `../plan.md` opens the plan.
- One daemon serves a v1 project and a v2 project side by side, the v1 board unchanged.

## stage 4: gestures below the UI

Each gesture follows v1's shape: load fresh, verify the guards, compute the minimal bytes, write, and report a partial failure plainly.

- **Capture.** `<slug>/index.md` with v1's skeleton, no-clobber on the directory. CLI capture drafts a `.capture-` temp in `.ranger/` and finalizes as v1 does, with the same four outcomes.
- **State.** Patch `state`. Patch `since` to today, or insert it directly after `state` when absent. One file, and never a view file, so a move back to Inbox leaves the item's lines in the view files. Setting the state an item already has writes nothing.
- **View edits.** Add at a position, move, and remove. Remove takes out every line that names the slug. A position is an anchor: before a named slug, after one, or the end. A move within one lane then never reorders another lane's members. One line moves, and comments and every other line survive. An absent view file is created under the absent version.
- **State and place.** A drag across lanes that also places writes the item first and the view second. A drag out of Inbox always places: the item joins the view it was dragged in at the drop position, and a line that view already holds for it moves there. When the destination lane takes no position, as Building and Evaluating do not, a line the view already holds stays where it is and a new one is appended at the end.
- **Save content.** `index.md` verbatim under its hash, with no side effects.
- **Delete.** Remove the item's directory and its lines in both view files, under the `index.md` hash and both view versions. Other items' links are left alone.
- **CLI.** Capture and `ranger state <slug> <state>` work in a v2 project.
- **API.** A mutation for each gesture, carrying the expected hash and view versions and returning the fresh project payload. Conflicts are typed as in v1, with a reason for a view conflict.

Acceptance:

- Composite-gesture tests run against the whole tree: exact bytes for the changed files, everything else byte-identical.
- A state change leaves both view files byte-identical, a move back to Inbox included. Repeating an item's current state on a later day leaves the whole tree byte-identical.
- A drag out of Inbox adds one line to the view it was dragged in and leaves the other view file byte-identical.
- A move changes one line's position and nothing else.
- Removing a view's last entry leaves a file with no entries and its comments intact. A write to a flow-style view leaves it one slug per line.
- A stale hash or version refuses with nothing written.
- Deleting an item that holds a design, a plan, stages and an image removes all of it and nothing outside it.
- No write path passes through a YAML encoder.

## stage 5: gestures in the UI

- Capture, reusing v1's modal.
- Drag on the roadmap board and the queue: across lanes changes state, out of Inbox also places, and within Horizon or Pending re-ranks. A card dropped into Building or Evaluating shows where the lane's sort puts it; positional drop cues appear only in Horizon and Pending.
- Add to and remove from a view, as a control on the card, the row and the item page.
- Raw edit of `index.md` on the item page. An open raw edit suspends the page's re-read, as v1's modal does.
- Delete behind a confirm that names the file count and the risk for uncommitted work.
- Optimistic apply and conflict handling as v1.

Acceptance:

- Vitest covers the drop arithmetic for flat lists.
- A smoke run exercises each gesture against a fixture and checks the resulting diff.

## stage 6: close-out

- `docs/current/` reads as one account of both conventions. Update the README and the changelog.
- Canon review: read ranger's terminus qualities against the code that landed and report drift in the journal. Do not edit the canon.
- Roadmap: list for Michael the cards this arc answers (`linking`, `issues-or-punchlist`, `batches-slices-and-follow-ups`, `specwork-order-management`, `field-notes-from-megas-roadmap`) and the ones it leaves. Deleting them is his.
- Re-synthesize the spec's deferred list into roadmap items. The spec and this work order are then removed, on Michael's word.

## not in this work order

- Migrating any project. That is a curation session in the trial project, after stage 3, on Michael's word.
- Filters, saved filters and search on the v2 surfaces.
- Everything in the spec's "Deferred, and why".
