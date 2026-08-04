import { describe, expect, it } from "vitest";
import type { Board, Card } from "./api";
import {
  emptyFilters,
  filterBoard,
  isApplied,
  isFiltering,
  recognizeNo,
  toFilters,
  toSaved,
  toggleMilestone,
  toggleNo,
  toggleSubsystem,
  toggleTag,
  upsert,
  withoutName,
} from "./filter";

function card(filename: string, extra: Partial<Card> = {}): Card {
  return { filename, title: filename, flags: [], hash: "x", ...extra };
}

function board(rankedCount: number, ...cards: Card[]): Board {
  return {
    project: "p",
    orderVersion: "v",
    lanes: [{ state: "researching", rankedCount, cards }],
  };
}

function shown(b: Board): string[] {
  return b.lanes[0].cards.map((c) => c.filename);
}

describe("toggleTag", () => {
  it("adds and removes an include", () => {
    const on = toggleTag(emptyFilters, "feature", false);
    expect(on.tags).toEqual(["feature"]);
    expect(toggleTag(on, "feature", false).tags).toEqual([]);
  });

  it("adds and removes an exclude", () => {
    const on = toggleTag(emptyFilters, "feature", true);
    expect(on.notTags).toEqual(["feature"]);
    expect(toggleTag(on, "feature", true).notTags).toEqual([]);
  });

  it("flips polarity rather than stacking", () => {
    const included = toggleTag(emptyFilters, "feature", false);
    const flipped = toggleTag(included, "feature", true);
    expect(flipped.tags).toEqual([]);
    expect(flipped.notTags).toEqual(["feature"]);
    const back = toggleTag(flipped, "feature", false);
    expect(back.tags).toEqual(["feature"]);
    expect(back.notTags).toEqual([]);
  });
});

describe("toggleSubsystem", () => {
  it("flips polarity rather than stacking", () => {
    const excluded = toggleSubsystem(emptyFilters, "flo", true);
    const flipped = toggleSubsystem(excluded, "flo", false);
    expect(flipped.subsystems).toEqual(["flo"]);
    expect(flipped.notSubsystems).toEqual([]);
  });
});

describe("toggleMilestone", () => {
  it("sets and clears each polarity", () => {
    const inc = toggleMilestone(emptyFilters, "v0.1.x", false);
    expect(inc.milestone).toBe("v0.1.x");
    expect(toggleMilestone(inc, "v0.1.x", false).milestone).toBeNull();
    const exc = toggleMilestone(emptyFilters, "v0.1.x", true);
    expect(exc.notMilestone).toBe("v0.1.x");
    expect(toggleMilestone(exc, "v0.1.x", true).notMilestone).toBeNull();
  });

  it("flips polarity on the same value", () => {
    const inc = toggleMilestone(emptyFilters, "v0.1.x", false);
    const flipped = toggleMilestone(inc, "v0.1.x", true);
    expect(flipped.milestone).toBeNull();
    expect(flipped.notMilestone).toBe("v0.1.x");
  });

  it("stays single-slot across polarities: a new value replaces the whole slot", () => {
    const inc = toggleMilestone(emptyFilters, "v0.1.x", false);
    const exc = toggleMilestone(inc, "v0.2.x", true);
    expect(exc.milestone).toBeNull();
    expect(exc.notMilestone).toBe("v0.2.x");
  });
});

describe("isFiltering", () => {
  it("is false only when every dimension is empty", () => {
    expect(isFiltering(emptyFilters)).toBe(false);
    expect(isFiltering(toggleTag(emptyFilters, "feature", false))).toBe(true);
    expect(isFiltering(toggleTag(emptyFilters, "feature", true))).toBe(true);
    expect(isFiltering(toggleSubsystem(emptyFilters, "flo", true))).toBe(true);
    expect(isFiltering(toggleMilestone(emptyFilters, "v0.1.x", true))).toBe(true);
  });
});

describe("absence filters", () => {
  it("toggleNo displaces the dimension's values whole", () => {
    const f = toggleTag(toggleTag(emptyFilters, "feature", false), "spike", true);
    const no = toggleNo(f, "tags");
    expect(no).toEqual({ ...emptyFilters, noTags: true });
    expect(isFiltering(no)).toBe(true);
  });

  it("a value click displaces absence back", () => {
    const no = toggleNo(emptyFilters, "milestone");
    expect(toggleMilestone(no, "v0.1.x", false)).toEqual({ ...emptyFilters, milestone: "v0.1.x" });
    expect(toggleSubsystem(toggleNo(emptyFilters, "subsystems"), "flo", false)).toEqual({
      ...emptyFilters,
      subsystems: ["flo"],
    });
  });

  it("recognizeNo plucks complete tokens and keeps the rest as search", () => {
    expect(recognizeNo("no:milestone")).toEqual({ dims: ["milestone"], rest: "" });
    expect(recognizeNo("auth no:tags flow")).toEqual({ dims: ["tags"], rest: "auth flow" });
    expect(recognizeNo("no:tags no:subsystems")).toEqual({ dims: ["tags", "subsystems"], rest: "" });
  });

  it("recognizeNo leaves non-tokens untouched, spacing included", () => {
    expect(recognizeNo("no:mile")).toEqual({ dims: [], rest: "no:mile" });
    expect(recognizeNo("kno:tags")).toEqual({ dims: [], rest: "kno:tags" });
    expect(recognizeNo("  auth  flow ")).toEqual({ dims: [], rest: "  auth  flow " });
  });
});

describe("saved filters", () => {
  it("toSaved omits empty dimensions", () => {
    const f = toggleTag(toggleMilestone(emptyFilters, "v0.1.x", true), "spike", true);
    expect(toSaved("quiet", f)).toEqual({ name: "quiet", notTags: ["spike"], notMilestone: "v0.1.x" });
  });

  it("toSaved and toFilters round trip", () => {
    const f = toggleSubsystem(toggleTag(emptyFilters, "feature", false), "flo", true);
    expect(toFilters(toSaved("x", f))).toEqual(f);
  });

  it("an all-empty saved filter applies as no filter", () => {
    expect(toFilters({ name: "empty" })).toEqual(emptyFilters);
    expect(isFiltering(toFilters({ name: "empty" }))).toBe(false);
  });

  it("upsert replaces a same-named filter in place", () => {
    const list = [{ name: "a", tags: ["feature"] }, { name: "b" }];
    const next = upsert(list, { name: "a", tags: ["defect"] });
    expect(next).toEqual([{ name: "a", tags: ["defect"] }, { name: "b" }]);
  });

  it("upsert appends a new name at the end", () => {
    expect(upsert([{ name: "a" }], { name: "b" })).toEqual([{ name: "a" }, { name: "b" }]);
  });

  it("withoutName removes only the named filter", () => {
    expect(withoutName([{ name: "a" }, { name: "b" }], "a")).toEqual([{ name: "b" }]);
  });

  it("isApplied matches the applied filter and ignores member order", () => {
    const s = { name: "x", tags: ["a", "b"], milestone: "v0.1.x" };
    expect(isApplied(toFilters(s), s)).toBe(true);
    expect(isApplied({ ...emptyFilters, tags: ["b", "a"], milestone: "v0.1.x" }, s)).toBe(true);
  });

  it("isApplied drops when the filter is edited past the saved shape", () => {
    const s = { name: "x", tags: ["a"] };
    const applied = toFilters(s);
    expect(isApplied(toggleTag(applied, "b", false), s)).toBe(false);
    expect(isApplied(toggleTag(applied, "a", false), s)).toBe(false);
    expect(isApplied(toggleTag(applied, "a", true), s)).toBe(false);
  });

  it("absence flags round trip through toSaved and toFilters", () => {
    const f = toggleNo(toggleNo(emptyFilters, "tags"), "milestone");
    expect(toSaved("untriaged", f)).toEqual({ name: "untriaged", noTags: true, noMilestone: true });
    expect(toFilters(toSaved("untriaged", f))).toEqual(f);
    expect(isApplied(f, toSaved("untriaged", f))).toBe(true);
  });
});

describe("filterBoard", () => {
  it("drops cards carrying an excluded tag", () => {
    const b = board(0, card("a", { tags: ["feature", "spike"] }), card("b", { tags: ["defect"] }), card("c"));
    const f = toggleTag(emptyFilters, "spike", true);
    expect(shown(filterBoard(b, f, null))).toEqual(["b", "c"]);
  });

  it("composes includes with excludes", () => {
    const b = board(0, card("a", { tags: ["feature", "spike"] }), card("b", { tags: ["feature"] }), card("c", { tags: ["defect"] }));
    const f = toggleTag(toggleTag(emptyFilters, "feature", false), "spike", true);
    expect(shown(filterBoard(b, f, null))).toEqual(["b"]);
  });

  it("drops cards carrying an excluded subsystem", () => {
    const b = board(0, card("a", { subsystems: ["flo"] }), card("b", { subsystems: ["reef"] }), card("c"));
    const f = toggleSubsystem(emptyFilters, "flo", true);
    expect(shown(filterBoard(b, f, null))).toEqual(["b", "c"]);
  });

  it("excluding a milestone keeps unmilestoned cards", () => {
    const b = board(0, card("a", { milestone: "v0.1.x" }), card("b", { milestone: "v0.2.x" }), card("c"));
    const f = toggleMilestone(emptyFilters, "v0.1.x", true);
    expect(shown(filterBoard(b, f, null))).toEqual(["b", "c"]);
  });

  it("composes with the search match set", () => {
    const b = board(0, card("a"), card("b", { tags: ["spike"] }), card("c"));
    const f = toggleTag(emptyFilters, "spike", true);
    expect(shown(filterBoard(b, f, new Set(["a", "b"])))).toEqual(["a"]);
  });

  it("shrinks rankedCount to the surviving ranked prefix", () => {
    const b = board(2, card("a", { tags: ["spike"] }), card("b"), card("c"));
    const f = toggleTag(emptyFilters, "spike", true);
    const filtered = filterBoard(b, f, null);
    expect(shown(filtered)).toEqual(["b", "c"]);
    expect(filtered.lanes[0].rankedCount).toBe(1);
  });

  it("absence keeps only cards where the dimension is unspecified", () => {
    const b = board(
      0,
      card("a", { tags: ["feature"], subsystems: ["flo"] }),
      card("b", { milestone: "v0.1.x" }),
      card("c"),
    );
    expect(shown(filterBoard(b, { ...emptyFilters, noTags: true }, null))).toEqual(["b", "c"]);
    expect(shown(filterBoard(b, { ...emptyFilters, noSubsystems: true }, null))).toEqual(["b", "c"]);
    expect(shown(filterBoard(b, { ...emptyFilters, noMilestone: true }, null))).toEqual(["a", "c"]);
  });
});
