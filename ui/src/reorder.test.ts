import { describe, expect, it } from "vitest";
import type { Lane } from "./api";
import { anchorFor, positionAfterDrop, rankedAfterDrop } from "./reorder";

function lane(rankedCount: number, ...filenames: string[]): Lane {
  return {
    state: "researching",
    rankedCount,
    cards: filenames.map((filename) => ({
      filename,
      title: filename,
      flags: [],
      hash: "x",
    })),
  };
}

describe("anchorFor", () => {
  it("moving down anchors after the target", () => {
    expect(anchorFor(["a", "b", "c"], "a", "b")).toEqual({ filename: "b", after: true });
  });

  it("moving up anchors before the target", () => {
    expect(anchorFor(["a", "b", "c"], "c", "a")).toEqual({ filename: "a", after: false });
  });

  it("a lane-container drop has no anchor (tail)", () => {
    expect(anchorFor(["a", "b"], "a", "lane:researching")).toBeNull();
  });

  it("dropping on itself falls back to the visible neighbor above", () => {
    expect(anchorFor(["a", "x", "b"], "x", "x")).toEqual({ filename: "a", after: true });
  });

  it("dropping on itself at the top falls back to the neighbor below", () => {
    expect(anchorFor(["x", "a"], "x", "x")).toEqual({ filename: "a", after: false });
  });
});

describe("rankedAfterDrop", () => {
  it("moves a card down one slot", () => {
    expect(rankedAfterDrop(lane(3, "a", "b", "c"), "a", { filename: "b", after: true })).toEqual(["b", "a", "c"]);
  });

  it("moves a card up", () => {
    expect(rankedAfterDrop(lane(3, "a", "b", "c"), "c", { filename: "a", after: false })).toEqual(["c", "a", "b"]);
  });

  it("returns null for a drop back into place", () => {
    expect(rankedAfterDrop(lane(3, "a", "b", "c"), "b", { filename: "a", after: true })).toBeNull();
  });

  it("ranks an unranked card beside the anchor", () => {
    expect(rankedAfterDrop(lane(2, "a", "b", "u1", "u2"), "u1", { filename: "a", after: false })).toEqual([
      "u1",
      "a",
      "b",
    ]);
  });

  it("an unranked anchor is a tail drop: end-of-ranked-list", () => {
    expect(rankedAfterDrop(lane(2, "a", "b", "u1", "u2"), "a", { filename: "u1", after: true })).toEqual(["b", "a"]);
  });

  it("a nil anchor is a tail drop", () => {
    expect(rankedAfterDrop(lane(2, "a", "b", "u1", "u2"), "u2", null)).toEqual(["a", "b", "u2"]);
  });

  // the filtered-board case: b is hidden by a filter; dropping c below the
  // visible a lands it directly below a in the full order, above hidden b.
  it("lands directly beside the anchor across hidden neighbors", () => {
    expect(rankedAfterDrop(lane(3, "a", "b", "c"), "c", { filename: "a", after: true })).toEqual(["a", "c", "b"]);
  });
});

describe("positionAfterDrop", () => {
  it("after a ranked anchor", () => {
    expect(positionAfterDrop(lane(2, "a", "b", "u1"), { filename: "a", after: true })).toBe(1);
  });

  it("before a ranked anchor", () => {
    expect(positionAfterDrop(lane(2, "a", "b", "u1"), { filename: "b", after: false })).toBe(1);
  });

  it("nil anchor is end-of-ranked-list", () => {
    expect(positionAfterDrop(lane(2, "a", "b", "u1"), null)).toBe(2);
  });

  it("unranked anchor is end-of-ranked-list", () => {
    expect(positionAfterDrop(lane(2, "a", "b", "u1"), { filename: "u1", after: true })).toBe(2);
  });
});
