---
title: the ranger v2 convention
created: 2026-11-24
---

# the ranger v2 convention

Ranger v2 reads items directly under `docs/future/`. A project adopts the convention by writing `docs/future/.ranger/version.yaml` by hand.

```yaml
version: 2
```

A `.ranger/` directory at that location requires a readable version marker with value `2`. With that marker in place, `docs/future/roadmap/` is treated like any other item directory. Without `docs/future/.ranger/`, the project uses v1.

Ranger follows no symlink in a v2 project. A symlink at `docs/`, `docs/future/`, `.ranger/` or `version.yaml` is a project error that names the link. A symlinked view file fails its view. Ranger skips symlinked item directories and treats symlinked files and directories inside an item as missing.

## Items

Capture by creating `docs/future/<slug>/index.md`. Make the slug from the title by lowercasing ASCII letters, keeping ASCII letters, digits, spaces and hyphens, turning spaces into hyphens, collapsing hyphen runs and trimming the ends. Pick a directory name by hand if the title produces an empty slug. Use the directory name in commands, view files and links. After a title edit, the directory name still works and ranger flags a mismatch.

```markdown
---
title: Fullscreen in the client
state: pending
created: 2026-10-15
since: 2026-10-15
kind: story
requires:
  - dfw:window-fullscreen
---

Make the fullscreen action work from the client.
```

`title`, `state` and `created` are required. States are `inbox`, `horizon`, `pending`, `building` and `evaluating`. Set `since` to today's date when you change the state. Write dates as `YYYY-MM-DD`. The optional `kind` labels are `epic`, `story`, `defect` and `punchlist`. The v1 fields `tags`, `subsystems`, `source`, `milestone` and `log` carry over. Tags and subsystems are string lists, source and milestone are strings, and log is a list of `stamp` dates and `note` strings. Item fields validate independently. An unreadable state puts the item in Inbox.

Use `includes` for items gathered here, `requires` for dependencies and `blocked_on` for current blockers. Each is a list of directory names. Use `project:slug` for a reference into another project. Ranger shows that reference as written. Write requests for the operator as a `needs` list with one request per entry.

The prompt ends at the first `##` heading. Ranger counts open `[ ]` and checked `[x]` list entries in the prompt. To break out a checklist finding, make an item, add it to the queue as Pending, and link to it from the checked line.

Ranger reads inline Markdown links throughout `index.md`, including in supporting sections. Write `[Fullscreen in the client](../fullscreen-in-the-client/index.md)` to link to a sibling item. A relative path into another visible directory under `docs/future/` identifies that directory's item, even if the link names an attachment there. A missing directory raises a flag on the referring item. An existing directory without `index.md` is ignored. Ranger skips fenced code, inline code and reference-style links. Write the link where the relationship is asserted. Ranger derives the incoming links.

## Roadmap and queue

`.ranger/roadmap.yaml` and `.ranger/queue.yaml` are independent lists of item directory names. List each slug on its own YAML sequence line.

```yaml
- look-and-feel-pass
- fullscreen-in-the-client
```

An absent, empty or comment-only file represents an empty list. Ranger reads bracketed lists as well. If a view fails to read, the other view and item details remain available.

Inbox appears in both views, sorted by `created` with undated items last. The other lanes show only members of that view. Horizon and Pending use the list order. Building and Evaluating use `since`, newest first with undated items last. Dates tie by directory name. The first occurrence of each slug sets its position. Returning an item to Inbox leaves its lines in the view files so a later triage can reuse their positions.

## Designs, plans and stages

Keep the design in `design.md` and the plan's shared setup in `plan.md`. The plan's frontmatter is optional. When posting a Mercurius verdict, include its session. List stages in execution order using the filenames without `.md`.

```yaml
---
mercurius: ready_to_build s_GtmwEX60upbB
stages:
  - file-backed-score-calls
---
```

Write the stage's instructions in `stages/<name>.md`. Put the stage facts in its frontmatter. The implementer posts `status` as `started` or `built` and posts `terminus` with the verdict and run ID. The `needs` list works as it does on an item. The design agent adds an `accepted` date after the operator accepts the stage.

```yaml
---
title: File-backed complete-score calls
status: built
terminus: clean rehearsal-run
needs:
  - Audition the create and replace calls.
---
```

For an item in Building without a plan, ranger shows whether it has a design. A plan with at least one listed stage is complete when all its listed stages are accepted. Once a listed stage has a status or acceptance, the first unaccepted stage is current and supplies its Terminus verdict if posted. Missing or unreadable stage files hold their place in the list. Before any listed stage starts, ranger shows the plan's Mercurius verdict. A verdict beginning with `ready_to_build` means ready to hand off. A plan with an empty stage list says no stages listed.

Needs you collects item and stage requests and built stages awaiting acceptance. Flags report malformed items, title slug mismatches, missing local link targets, missing items named by view lines, and items past Inbox in neither readable view. Flags also cover missing, unlisted and unreadable stage files, unreadable plans, and items still Building after all stages are accepted.

## Read from the CLI

From inside the project, run `ranger roadmap`, `ranger queue`, `ranger needs`, `ranger show <slug>` or `ranger flags`. Each command reads the files fresh. Use the editor for v2 capture, state changes and view edits. In a v1 project, quote a multiword capture title that begins with a command word, for example `ranger "queue follow-up"`. A quoted reserved word on its own still invokes the command.

## Trial project AGENTS.md paragraph

This project uses the [Ranger v2 convention](https://github.com/michaelquigley/ranger/blob/main/docs/current/v2-convention.md). Capture work in `docs/future/<slug>/index.md` with a title, `state: inbox`, and today's `created` date. Design agents may classify, link, and move items to Horizon or Pending in the queue, and may order the queue. When breaking out a checklist finding, add its item to the queue as Pending and link to it from the checked line. Moves into Building or Evaluating, retirement, and edits to `docs/future/.ranger/roadmap.yaml` follow the operator's direction. Implementers post facts in their own stage file. Make surgical file edits and leave the working tree for the operator to review and commit.
