import { describe, expect, it } from "vitest";
import type { Board, Card } from "./api";
import { emptyFilters, filterBoard, isFiltering, toggleMilestone, toggleSubsystem, toggleTag } from "./filter";

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
});
