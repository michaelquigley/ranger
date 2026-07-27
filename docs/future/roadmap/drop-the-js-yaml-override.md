---
title: drop the js-yaml override
state: horizon
created: 2026-07-22
tags: [chore]
---

`ui/package.json` carries scoped npm overrides under `@redocly/openapi-core` — now two of them, same root cause both times (redocly 1.x exact-pins vulnerable transitives, no fixed 1.x exists, the upstream fix lives in redocly 2.x, which `openapi-typescript` 7.x doesn't accept):

- `js-yaml` → 4.3.0 (2026-07-22; GHSA-52cp-r559-cp3m, quadratic CPU on YAML merge-key chains; redocly pins 4.2.0 exactly).
- `minimatch.brace-expansion` → 5.0.8 (2026-07-27; GHSA-mh99-v99m-4gvg, DoS via unbounded expansion; the only patched release — no 2.x backport exists). known accepted edge: brace-expansion 5.x exports `{ expand }` named while minimatch 5.1.9 calls the module as a function, so a *brace-containing* pattern through redocly's minimatch would throw "expand is not a function" — loudly, never silently wrong. our codegen path (a single literal file path, no redocly config in the repo) never constructs one; verified byte-identical on regeneration.

both are dev-tooling only (`gen:api` codegen; nothing ships). drop the whole `overrides` block when `openapi-typescript` moves to `@redocly/openapi-core` 2.x — check with `npm ls js-yaml brace-expansion` after any openapi-typescript major bump. a brace-expansion 2.x backport (the previous advisory got them) would independently let the brace-expansion override drop to a same-major bump.
