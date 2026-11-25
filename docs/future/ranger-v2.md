# Ranger v2

Vision and first design, in draft.

## Why grow Ranger

Ranger began with a useful proposition. A sentence in a Markdown file can be enough to carry an idea from capture through evaluation. An agent with a text editor can participate. The operator can open a board, arrange the work, and use Git to decide what to keep.

As projects grew, a single board position came to carry several meanings. An epic sat alongside its own stories. Lane order stood in for both priority and dependency. An agent could finish a stage while the card still said researching. A studio audition could be the only thing holding up acceptance, and the place to discover that was yesterday’s conversation.

Ranger v2 should let the operator steer a project and let design agents help turn that direction into work. Someone returning to a project should be able to find the reasoning behind an item, the next piece that can be handed off, and the decisions waiting on them. They should also be able to capture another sentence and get back to what they were doing.

## The files are the working model

An independently tracked item lives at `docs/future/<slug>/index.md`. Its frontmatter carries the facts the tool needs to read. Its body carries the prompt in the author’s words and whatever context helps someone act on it. A new capture can consist of that file alone. The directory gives a defect somewhere to keep a screenshot, a punchlist somewhere to keep a feedback document, and a substantial piece of work somewhere to grow a design and plan.

```text
docs/future/
  score-creation-from-a-local-file/
    index.md
    design.md
    plan.md
    stages/
      file-backed-score-calls.md
```

Supporting files have paths that people and agents can open directly. The convention must be small enough to teach from a project’s `AGENTS.md`, with examples that work in an ordinary editor. A board gesture, a CLI operation, and a hand edit should produce changes the next reader can understand from the files.

`docs/current/` remains the published account of built behavior. `docs/journal/` remains the chronological scratchpad for observations, decisions, and context worth recovering. An item can point to the journal entry that explains a decision. Completed work leaves `docs/future/`; Git preserves the committed history.

## The roadmap and the work queue

**One collection of items, one lifecycle state per item, two independently curated views.** This is the central choice to test in practice.

The **roadmap** is a board for steering. The operator chooses which outcomes to keep in view and orders them. The **work queue** is a denser table for the operator and design agent to curate together. It can show the item, its next stage, the latest posted broker result, what it depends on, and whose action is needed.

A shared **Inbox** accepts unclassified captures. At triage, an item can go on the roadmap, in the queue, or in both. Each view owns its membership and order. Both use the same lanes:

| Lane | Judgment it records |
| --- | --- |
| Horizon | Worth keeping in consideration. Its position expresses a preference while the choice remains open. |
| Pending | Chosen to take up and awaiting pickup. |
| Building | Work is underway. |
| Evaluating | The result is in use and deserves deliberate observation or acceptance. |

Adding an item to a view and moving it to Pending are different decisions. A story can sit first in the queue and third on the roadmap. Moving it to Building changes its state in both places. Changing its rank or removing it from the roadmap leaves its queue placement alone. This should let an operator revise the story they want the roadmap to tell without disturbing a live handoff.

## Kinds, relationships, and changing shape

An item can begin without a kind. We’ll try one optional kind once its purpose is clear.

- An **epic** expresses a broad outcome. A **story** expresses a focused change that can be designed and accepted in its own right. The distinction is scope. Either can carry a design and plan and appear in either view.
- A **defect** begins with an observation about existing behavior. It can keep its identity through investigation, implementation, and a real-world check.
- A **punchlist** collects findings from a pass or session. The whole list can be handed to an agent. A finding can start as a line and become a linked item when someone needs to prioritize, design, or accept it independently.

Reclassification is an ordinary edit. A split creates new items when pieces need independent attention. A fold brings work together. The curation review should show what became of each live reference. Titles and filenames can evolve together. A rename should repair references in the project’s live `docs/future/` tree and surface incoming references from other visible projects for their owners to resolve.

Any item can name other items through **ordered, typed links**. The order says how the author wants the work read or considered. The type says what the relationship means. For example, an epic can include two stories in a preferred order while both proceed in parallel. A `requires` link expresses reliance on another item’s outcome. A `blocked_on` link expresses a present obstacle. Ranger should be able to walk these links in either direction and show which work a decision would unblock.

Dependencies target items. Stages have direct links for handoffs. A design agent can clear a blocker when the relevant decision lands, even if the blocking item stays open. Retiring completed work can clear its live blocker references in the same reviewed change. A missing file after a hand edit deserves inspection: the work may have moved or acquired a successor. Cross-project links identify the project independently of its checkout path. A temporarily unavailable checkout gives no evidence that its work has finished.

## From design to acceptance

Point a design agent at the item and say “let’s work on this.” The **design** records what should be built, how it should behave, and why. The **plan** records how to deliver it. These are the preferred names for the artifacts previously called a *spec* and a *work order*. Substantial work gets both. For smaller work, the index can carry the design intent and the plan can carry the implementation detail.

The plan owns the shared setup and an ordered list of links to **stages**, each in its own Markdown file. A story invites a design decision. A stage describes work already decided, with enough context and acceptance criteria for an implementation agent to take it. Mercurius reviews the design and plan together, including the stages. If implementation reopens a design decision, revise the affected plan and take the decision back through review.

The operator initiates handoffs. Starting a plan normally means intending to run it to completion. The operator can give other work to another agent in parallel. During a stage, the implementation agent posts the status and runs Terminus to `clean`. The broker result, any requested operator action, and acceptance are distinct facts. An agent can request a decision while continuing other work. A project-wide **Needs you** view should collect those requests from the items and stages where they were made. Posted updates are sufficient for the first version.

The operator reviews the code in their IDE and review tools, makes the commit, and tells the design agent in Mega that the stage is accepted. The design agent records acceptance and prepares the handoff prompt for the next stage. That update can legitimately appear in the working tree after the implementation commit. Accepted stages remain readable until the entire plan closes. At closeout the operator authorizes retirement, the published docs take up the built behavior, and the completed plan and stages leave together. The originating item can remain in Evaluating for as long as there is a real question to answer in use.

## Collaborative curation and Git

Design agents classify, link, group, and order routine work. The operator steers the roadmap. An agent can carry out a requested roadmap reorder or propose a change of direction in conversation. The review surface has to make the resulting edits comprehensible. Moving a story from sixth to second should read as one move. Breaking a finding out of a punchlist should show the original observation, the new item, and the link between them.

Ranger should be able to read the committed, staged, and working-tree versions of `docs/future/` so it can answer both “What am I about to commit?” and “What changed after I staged it?” This extends v1’s scoped Git status inspection. Staging, committing, and pushing belong to the operator. Meaningful capture, reshaping, review handoff, acceptance, and retirement are natural commit boundaries. The operator can group related changes. A capture that should survive in project history must reach a commit before it retires.

For early parallel work, we would try one coordination checkout per project. An operator-initiated handoff identifies the item, stage, agent, and worktree there. A single queue can then show where two implementations stand before either branch returns. The assignment record and its exact format should be tested with real parallel handoffs. Conversations stay in Mega. A later link from an item to its design conversation would make the switch easier.

## Two rehearsals

### BAAB · receiving a studio handoff

In BAAB’s November 16 local-file score work, a composition agent had generated twelve score files and retyped roughly 60 KB of JSON into tool calls. The existing card explains the need. The design and one-stage plan explain the change. The journal records that Mercurius cleared the plan and that the implementation was built and awaiting operator review and an audition. A returning operator had to read across those documents to assemble the handoff.

In the rehearsal, the operator opens the item from the queue, inspects its state, and asks the design agent **in Mega** for a short studio brief. Create a score from a project file, replace it from a second file, verify undo and redo, and listen. The operator can report “everything worked perfectly” and supply a feedback document. The agent records the audition once, clears that attention request, and reads the feedback to determine what deserves follow-up. A review verdict appears when an agent posts it.

The operator can then review the code, commit, and tell the design agent to proceed. Here the plan has only one stage. BAAB’s own acceptance instructions call for retiring the design and plan while keeping the card. This example exercises the distinction between finishing a plan, closing an audition request, and retiring an item. The audition response and acceptance sequence above are rehearsal outcomes. The historical snapshot comes from BAAB’s `docs/journal/2026-11-16.md` and its `agent-mcp-score-creation-from-a-local-file` card and work order.

### Screeningroom · catching findings from the couch

Screeningroom’s October 15 couch pass found that marking an episode finished sent the page to the top of a long list. It also surfaced a fullscreen change that involved the separate dfw project. The journal records both the experience and the resulting work in `docs/journal/2026-10-15.md`, under “The look-and-feel pass.”

In the rehearsal, the operator tells the design agent in Mega to expect a stream of findings. “Marking an episode finished loses my place” is enough. The agent captures the observation, groups follow-up remarks, and keeps the fullscreen finding distinct. “That’s it for now” hands curation to the agent. It organizes the punchlist, links existing items, breaks out anything requiring independent work, and reports back briefly. The operator can later say “let’s address the punchlist.” Checkboxes and ordinary links in `index.md` carry this case.

## Adoption and questions for practice

Migration is a curation session for each project. The operator and design agent read its live cards, plans, journal, and existing order. They decide what the old placements meant, which items belong in each view, and what relationships are real. The agent prepares coherent moves and repairs links. The operator reviews and commits. Legacy vocabulary stays intelligible in old documents and handoffs. The grimoire should eventually teach *design* and *plan* alongside their legacy synonyms.

The walkthroughs give the first version a few concrete tests. Can an agent with only a text editor capture a finding and later turn it into independent work? Can an operator review a queue reorder made alongside their own staged edits? Can they open one item and discover precisely what requires their attention? Does one shared state continue to tell the truth when an item matters to both views?

Those tests guide the first design, which follows. It draws the metadata schema, a starting relationship vocabulary, and the storage for the two views. The handoff record for parallel work and broker integration beyond posted results remain open. Richer punchlist sidecars, automated assignment, and a historical timeline can earn their place when there is work the simpler convention struggles to express.

## The first design

This part records what has been decided so far. It is drawn against the two rehearsals, and all of it can change during the trial.

### Recognition and layout

A project uses v2 when `docs/future/.ranger/version.yaml` exists. The file records the convention version of the folder, and holds one line: `version: 2`. It is written by hand when the project moves to v2 and committed with everything else, so a fresh checkout reads the same way. Ranger never writes it. A `.ranger/` directory with no `version.yaml`, or with one ranger cannot read as version 2, is an error that names the file.

A project uses v1 when `docs/future/roadmap/` exists and there is no `version.yaml`. When `version.yaml` is present the project is v2 whatever else `docs/future/` holds, so a v2 item can be named `roadmap` like any other. A project is one or the other. Ranger reads each with its own code, and the v1 reader does not change.

An item is a directory under `docs/future/` that holds an `index.md`. The directory name is the slug of the item's title, by the v1 slug rule. Everything else in `docs/future/` is left alone.

The directory name is the item's name everywhere: in the view files, in links, and on the command line. When a title edit leaves the two out of step, the item keeps its directory name and its place, and the mismatch is flagged.

Ranger follows no symlink in a v2 project. A symlink on the way to the project's files, at `docs/`, `docs/future/`, `.ranger/` or `version.yaml`, is a project error that names it, and a symlinked view file is that view's error. Below that, a symlinked directory is not an item, and a symlinked file inside an item counts as missing, so the existing flags report it.

```text
docs/future/
  .ranger/
    version.yaml
    roadmap.yaml
    queue.yaml
  agent-mcp-score-creation-from-a-local-file/
    index.md
    design.md
    plan.md
    stages/
      file-backed-score-calls.md
```

`design.md`, `plan.md` and `stages/` appear when the work needs them. Any other file in the directory is the item's own, such as a screenshot or a feedback document.

### The item

`index.md` keeps v1's frontmatter and its body rule: everything above the first `##` heading is the prompt. The states are now `inbox`, `horizon`, `pending`, `building` and `evaluating`, and the state is written here and nowhere else.

```yaml
---
title: agent MCP score creation from a local file
kind: story
state: building
since: 2026-11-16
created: 2026-11-15
tags: [spike]
source: ../baabhive/sessions/2026-11-15-velvet-brick-spine/README.md
---
```

v2 adds these fields. All are optional.

| Field | Meaning |
| --- | --- |
| `since` | The date the item entered its current state. Whoever changes `state` changes `since`. |
| `kind` | `epic`, `story`, `defect` or `punchlist`. |
| `includes` | The items this one gathers, in order, as an epic gathers its stories. |
| `requires` | Items whose outcome this one relies on. |
| `blocked_on` | Items standing in the way now. |
| `needs` | Requests waiting on the operator, one line each. |

A kind is a label. It changes what a card or row says, and nothing else. Ranger counts the prompt's checkboxes and shows `includes` on any item, whatever its kind.

Pending means chosen and not started. Building runs from the start of the work through design, plan, review and stages, and ends when the last stage is accepted. The queue says where in that an item is, for example no design yet, designing, plan in review, or stage 2 of 5.

### Links

A link is written on one side. An epic lists its stories under `includes`, and the story says nothing about the epic. Ranger walks links in both directions.

An item in the same project is named by its slug. An item in another project is `project:slug`, where the project is the slug of that repository's directory name.

```yaml
kind: epic
includes:
  - first-story
  - second-story
requires:
  - dfw:window-fullscreen
blocked_on:
  - choose-a-tagline
```

Ranger also reads inline Markdown links in an item's body, written `[text](../other-item/index.md)`, when they point into another item's directory. Reference-style links and bare URLs are prose, and ranger does not follow them. A link counts by the shape of its path: one that leads to `docs/future/<name>/` is a link to the item `<name>`. When no such directory exists, the target is missing. When the directory exists and is not an item, ranger ignores the link. It uses these links the same way as the typed ones: to walk backward, and to mark a link whose target is gone. To point at another item, write a link. A backticked name is prose, and ranger does not follow it.

### The two views

`.ranger/roadmap.yaml` and `.ranger/queue.yaml` have the same shape, a flat list of item slugs.

```yaml
- agent-mcp-score-creation-from-a-local-file
- agent-mcp-compressor-parameters
- agent-mcp-node-creation-position
```

The list is written one slug per line. Ranger also reads a bracketed list, and rewrites it one per line the first time it edits the file. A view with nothing in it is an empty file, or one that holds only comments.

Being in the list is membership in that view. Position is rank. The lane comes from the item's own `state`, so a state change is one edit to `index.md` and shows in both views at once. A view file changes only when someone adds, removes or re-ranks an item.

Position counts in Horizon and Pending. In Horizon it is a preference while the choice is open. In Pending it is the order to take things up. In Building and Evaluating ranger sorts by `since`, newest first.

While an item is in `inbox` the view files do not matter: both views show it, listed or not. Triage sets its state and adds it to one file or both. An item moved back to the Inbox keeps its lines in the view files, so if it is triaged again it returns where it was.

The cost of the flat list is that pending order cannot be read from a view file alone. It takes the items' states as well, or ranger.

### Plans and stages

`plan.md` carries the shared setup as its body. Its frontmatter holds the review result and the stages in order.

```yaml
---
mercurius: ready_to_build s_GtmwEX60upbB
stages:
  - file-backed-score-calls
---
```

A stage's number is its position in that list. Its filename never carries the number, so reordering stages renames nothing.

A plan is complete when it lists at least one stage and every listed stage is accepted. A listed stage whose file is missing or cannot be read counts as not accepted.

A stage file holds the stage's text as its body and four facts in its frontmatter.

```yaml
---
title: file-backed complete-score calls
status: built
terminus: clean <run id>
needs:
  - "audition: create from a project score, replace from a ~/Sandbox score, undo and redo, listen"
---
```

| Field | Written by | Meaning |
| --- | --- | --- |
| `status` | the implementer | `started` or `built`. Absent before the handoff. |
| `terminus` | the implementer | The latest verdict and its run. |
| `needs` | whoever asks | Requests waiting on the operator. A line is deleted when it is answered. |
| `accepted` | the design agent | The date the operator accepted the stage. |

`mercurius` on the plan works like `terminus` on a stage. It is the latest verdict and its session, posted by the planning agent.

The queue says where a plan stands by asking three questions in order. Is the plan complete? Then the row says every stage is accepted. Has any stage started? Then the row shows the current stage, the first one not yet accepted, with its `terminus` verdict, blank until one is posted. Otherwise the row shows the plan itself with its `mercurius` verdict, where `ready_to_build` means ready to hand off. Each verdict is labeled with its broker.

All of these are ordinary tracked edits. Acceptance is recorded after the operator commits the stage, so that edit rides with the next commit. The one fact kept out of the tracked files is which agent holds which worktree, because it names a path on one machine. It belongs to the parallel-work trial.

**Needs you** is every `needs` line in the project, plus every stage that is `built` and not yet `accepted`.

### Punchlists

A punchlist is an item whose prompt is a checklist.

```markdown
---
title: look-and-feel pass
kind: punchlist
state: building
since: 2026-10-15
created: 2026-10-15
---

Michael's pass over the interface from couch use. Take the findings one at a time.

- [ ] marking an episode finished sends a long list back to the top
    - home and browse do the same, unnoticed on short pages
- [ ] the type and sizes are the interface's own, and stack on the TV's 300% zoom
- [x] an unwatched movie still shows a Resume card: [mark unwatched leaves a resume card](../mark-unwatched-leaves-a-resume-card/index.md)
- [x] no fullscreen, and F11 does nothing: [fullscreen in the client](../fullscreen-in-the-client/index.md)
```

Each finding is a line, and follow-up remarks go under it. When another item carries a finding, whether newly broken out or already in the project, the line links to that item and its box is checked. The unchecked lines are what "address the punchlist" means.

Ranger counts the boxes in the prompt, above the first `##` heading, so the queue can say "2 open of 4". A checklist in a supporting section is not counted. Beyond links and checkboxes it reads nothing in the body.

A finding broken out is an ordinary item. The agent adds it to the queue as Pending.

```yaml
---
title: fullscreen in the client
kind: story
state: pending
since: 2026-10-15
created: 2026-10-15
requires:
  - dfw:window-fullscreen
---
```

### Flags

Ranger raises a flag when the files disagree. It changes nothing. A flag belongs to an item, or to a line in a view file.

- Every stage is accepted and the item is still Building.
- A link names an item that is not there. The flag is on the item that wrote the link. A link into another project is shown as written and is not flagged.
- A line in a view file names an item that is not there. The flag is on that line.
- An item past the inbox is in neither view. This is raised only when both view files could be read.
- The directory name is not the slug of the title.
- The plan lists a stage that has no file, or a stage file is not in the plan.
- A plan or stage file cannot be read.

v1's malformed flag carries over. One list shows every flag in a project, and the CLI prints the same list for an agent.

### Who writes what

Anyone captures. Design agents classify, link, break findings out, and order the queue. The roadmap's membership and order are the operator's, and an agent edits `roadmap.yaml` only when asked. Implementers post to their own stage. The operator commits.

On its own, an agent takes an item out of the Inbox into the queue as Horizon or Pending, and orders the queue. Moves to Building or Evaluating, retiring an item, and anything on the roadmap follow the operator's word, and the agent writes them down.

### Surfaces

For the trial, ranger's own UI and CLI carry every surface. Mega reads the same files later. This keeps ranger working without Mega.

| Surface | Shows |
| --- | --- |
| Roadmap | A board of the items in `roadmap.yaml` plus the Inbox, in lanes. Each card has its title, kind, tags, a mark when it needs you, and a mark when it is flagged. |
| Queue | A table of the items in `queue.yaml` plus the Inbox, grouped by lane. Each row has the item, its kind, the next stage and its status, the verdict for where the item is, what it requires or is blocked on, and what needs you. |
| Needs you | Every `needs` line and every stage built and not accepted, each with its item. |
| Item page | The prompt, its fields, links out and links in, the plan with each stage's facts, the files in its directory, and its flags. |
| Flags | Every flag and why: the flagged items, and the view lines that name no item. |

This is content, not layout. Placement is worked out in use.

The CLI prints the same five for an agent: `ranger roadmap`, `ranger queue`, `ranger needs`, `ranger show <item>` and `ranger flags`.

These five names become reserved first words on the command line in every project, v1 included, as v1's own command names already are. `ranger queue follow-up` no longer captures a card. Quote a multiword title to capture it, as in `ranger "queue follow-up"`. A quoted reserved word on its own still invokes the command. This is the one place v1's behavior changes.

### Gestures

Ranger writes for the operator's own gestures and nothing else.

- **Capture:** a new item in the Inbox.
- **Drag a card or row:** change state or re-rank. A card dragged out of the Inbox joins the view it was dragged in, at the drop position. Building and Evaluating take no position, so there a new line goes to the end of the list and a line already in the file stays put.
- **Place:** add an item to a view, or remove it from one.
- **Edit:** an item's `index.md` as raw text.
- **Delete:** remove an item.

Every write changes only the lines that express it, as in v1. The CLI keeps v1's capture and `ranger state`, which now sets `since` as well. `ranger state` changes one file. An item it takes out of the Inbox is in no view until someone adds it, and the flag says so.

Everything else is an agent's hand edit: kinds, links, stage facts, `needs`, and breaking a finding out.

Rename is not in the first build. It has to repair links across files, so it will be a command that the operator and agents both use. Until then a rename is a hand edit, and the flags catch a missed link.

### Reshaping and retiring

Renaming, splitting, folding and retiring are hand edits in the first build. Whoever removes or renames an item fixes what points at it in the same change, and the flags show what was missed.

Nothing is left behind to mark a removed item. The journal and the commit say what became of it. Ranger's delete removes the item's directory and its lines in the view files. It leaves other items' links alone, so they are flagged.

### Git

Ranger runs one read-only `git status`, as v1 does, widened to `docs/future/`. It marks each item and view file with what git reports: changes not yet staged, staged changes, or both. The marks add up across an item's files. The project selector's mark covers everything under `docs/future/`, a deleted item included. An item is new when git has never seen its `index.md`. Any other new file in an item is an unstaged change to it.

Ranger reads no file contents from git, and the first build has no review surface. The operator reviews in the IDE. The file shapes are chosen so a plain diff reads well: a re-rank is one moved line, and a state change is two lines in one file.

### Seam census

The boundary calls for the code, recorded so review can check against them.

- **v1 and v2 code: separate.** v2 is new packages beside v1's, and v1's packages do not change behavior. The one change a v1 project sees is the five reserved command names, recorded under Surfaces. The slug rule moves to a leaf package of its own that both conventions use, with v1's entry point kept. v2 uses the hash guard, the guarded write and the frontmatter parse from `internal/document`, where they live. There is one CLI, one server, one API contract and one UI, each telling a v1 project from a v2 project. v1 has to stay in service unchanged through the trial, and retiring it should be a deletion.
- **Model and rendering: separate.** Views, links, Needs you and flags are computed with no files and no rendering. Two renderers read them, the CLI and the UI.
- **Model and transport: separate.** Wire types exist only at the server's edge.
- **Reads and writes.** Reads go through `dd`. Writes are line patches, never a YAML encoder. The reviewable diff is the safety property.
- **State.** An item's state is written only in `index.md`. Nothing in v2 stores a lane anywhere else, so there is nothing to reconcile. Revisit if the trial shows people need to read pending order from a view file alone.
- **No unprompted writes.** Ranger marks and never fixes. v1's one exception, the `order.yaml` migration, has no counterpart.
- **The body.** Ranger reads inline links and checkboxes there, and nothing else. Revisit if a kind needs more structure than a checklist.
- **Git.** One read-only `git status`, scoped to `docs/future/`. Revisit if the trial shows the IDE is not enough for review.
- **Symlinks.** The v2 reader refuses the symlinks it finds where it reads. It is a plain refusal, checked when read, and not a guarantee against a tree that changes underneath it. v1's reader is unchanged.
- **Errors by tier.** An unreadable view file fails its own view, and the other view still shows. This differs from v1, where an unreadable `order.yaml` fails the project; agents now hand-edit these files, so a slip should cost less. A bad `index.md` is a flagged item, never a failed project. A bad plan or stage file is a flag on its item.

### The trial

BAAB and Screeningroom are the likely first projects. v1 stays in service until v2 has soaked.

The format can change during the trial. When it does, the agent re-edits the trial projects, so there is no migration tooling. The v2 convention is documented in this repository until it has soaked. A trial project's `AGENTS.md` carries the v2 paragraph, and the grimoire's convention note stays v1.

Moving a project to v2 is a curation session, as described under adoption. The session writes `version.yaml`. The `epic`, `story` and `defect` tags become the kind. `feature` and `enhancement` become `story`. `spike` and `documentation` stay tags.

One thing to watch in use: whether queue order ever deliberately differs from the roadmap's order with each epic's stories slotted in. If it never does, ranger could compute the queue.

### Deferred, and why

- **Reading the committed and staged versions, and a review surface.** The file shapes make a plain diff readable, and Mega's change digestion may be the better home. It waits until the trial shows the IDE is not enough.
- **A rename command that repairs links.** It is the first gesture to add after the first build. Until then the flags catch a missed link.
- **Links into other projects.** The first build shows a `project:slug` link as written and does not follow it. Following it means reading a second project, possibly in the other convention.
- **Parallel work.** The coordination checkout and the record of which agent holds which worktree need real parallel handoffs to test.
- **Computing the queue.** Whether the hand-kept queue earns its place is a question for the trial.
- **Broker integration beyond posted results.** A posted verdict line is enough to test the model. Mega plans to watch the brokers directly.
- **Filters, saved filters and search on the v2 surfaces.** v1 keeps them. They wait until the trial shows which lenses the two views still need.
- **Editing design, plan and stage files in ranger.** Ranger shows them. They are edited in the IDE, where the operator already works.
- **Links from an item to its design conversation in Mega.**
- **Richer punchlist sidecars, automated assignment, and a historical timeline.** Each can earn its place when there is work the simpler convention struggles to express.
- **Retiring v1 and rewriting the grimoire's convention note.** After v2 has soaked.
