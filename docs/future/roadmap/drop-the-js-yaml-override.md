---
title: drop the js-yaml override
state: horizon
created: 2026-07-22
tags: [chore]
---

`ui/package.json` carries a scoped npm override forcing `js-yaml` to 4.3.0 under `@redocly/openapi-core` — redocly 1.34.17 pins the vulnerable 4.2.0 exactly (GHSA-52cp-r559-cp3m, quadratic CPU on YAML merge-key chains), and no fixed 1.x exists; the upstream fix lives in redocly 2.x, which `openapi-typescript` 7.x doesn't accept.

the override is dev-tooling only (`gen:api` codegen; nothing ships) and was verified byte-identical on regeneration. drop it when `openapi-typescript` moves to `@redocly/openapi-core` 2.x — check with `npm ls js-yaml` after any openapi-typescript major bump.
