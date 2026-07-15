import type { Lane } from "./api";

// rankedAfterMove computes the lane's resulting ranked prefix for a
// within-lane drop: the moved card ranks at the drop position, snapped to
// the ranked/unranked boundary — a drop in the tail region is
// end-of-ranked-list, because between-ness among unplaced cards is not
// expressible without placing them. untouched neighbors stay unranked.
//
// `to` follows dnd-kit sortable's arrayMove convention: it is the moved
// card's final display index, already accounting for its own removal — no
// direction adjustment.
export function rankedAfterMove(lane: Lane, from: number, to: number): string[] | null {
  const moved = lane.cards[from].filename;
  const ranked = lane.cards.slice(0, lane.rankedCount).map((c) => c.filename);
  const without = ranked.filter((f) => f !== moved);
  const insertAt = Math.min(to, without.length);
  without.splice(insertAt, 0, moved);
  if (without.length === ranked.length && without.every((f, i) => f === ranked[i])) {
    return null;
  }
  return without;
}
