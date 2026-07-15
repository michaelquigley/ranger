import { describe, expect, it } from "vitest";
import type { Lane } from "./api";
import { rankedAfterMove } from "./reorder";

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

describe("rankedAfterMove", () => {
  // the regression that shipped: dnd-kit's `to` already encodes the final
  // slot, so the single most common gesture — dragging a card down one
  // slot — must produce a changed list, not a silent no-op.
  it("moves a card down one slot", () => {
    expect(rankedAfterMove(lane(3, "a", "b", "c"), 0, 1)).toEqual(["b", "a", "c"]);
  });

  it("moves a card down several slots", () => {
    expect(rankedAfterMove(lane(3, "a", "b", "c"), 0, 2)).toEqual(["b", "c", "a"]);
  });

  it("moves a card up", () => {
    expect(rankedAfterMove(lane(3, "a", "b", "c"), 2, 0)).toEqual(["c", "a", "b"]);
  });

  it("returns null for a drop back into place", () => {
    expect(rankedAfterMove(lane(3, "a", "b", "c"), 1, 1)).toBeNull();
  });

  it("ranks an unranked card alone at the drop position", () => {
    expect(rankedAfterMove(lane(2, "a", "b", "u1", "u2"), 2, 0)).toEqual(["u1", "a", "b"]);
  });

  it("snaps a ranked card's tail-region drop to end-of-ranked-list", () => {
    expect(rankedAfterMove(lane(2, "a", "b", "u1", "u2"), 0, 3)).toEqual(["b", "a"]);
  });

  it("ranks an unranked card dragged within the tail at the boundary", () => {
    expect(rankedAfterMove(lane(2, "a", "b", "u1", "u2"), 3, 2)).toEqual(["a", "b", "u2"]);
  });
});
