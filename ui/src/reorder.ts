import type { Lane } from "./api";

// drop placement is anchor-based: the moved card lands directly beside the
// card the operator dropped it against, in the lane's full order. anchor
// semantics survive filtered views — hidden neighbors don't shift the
// landing. a nil anchor, or an unranked one, is a tail-region drop and
// serializes as end-of-ranked-list, because between-ness among unplaced
// cards is not expressible without placing them.
export type DropAnchor = { filename: string; after: boolean } | null;

// anchorFor derives the drop anchor from the displayed lane (which may be a
// filtered subset): the card whose display slot the drop targeted, with
// direction from the moved card's own display position. when the drop
// resolves onto the moved card itself (common after a cross-lane
// re-parent), the anchor falls back to its visible neighbors.
export function anchorFor(displayed: string[], moved: string, overId: string): DropAnchor {
  if (overId.startsWith("lane:")) {
    return null;
  }
  const from = displayed.indexOf(moved);
  if (overId === moved) {
    if (from > 0) return { filename: displayed[from - 1], after: true };
    if (from >= 0 && from + 1 < displayed.length) return { filename: displayed[from + 1], after: false };
    return null;
  }
  const to = displayed.indexOf(overId);
  if (to < 0) {
    return null;
  }
  if (from >= 0 && from < to) {
    return { filename: overId, after: true };
  }
  return { filename: overId, after: false };
}

// rankedAfterDrop computes the lane's resulting ranked prefix for a
// within-lane drop of moved beside anchor. returns null when nothing
// changes.
export function rankedAfterDrop(lane: Lane, moved: string, anchor: DropAnchor): string[] | null {
  const ranked = lane.cards.slice(0, lane.rankedCount).map((c) => c.filename);
  const without = ranked.filter((f) => f !== moved);
  let insertAt = without.length;
  if (anchor && anchor.filename !== moved) {
    const idx = without.indexOf(anchor.filename);
    if (idx >= 0) {
      insertAt = idx + (anchor.after ? 1 : 0);
    }
  }
  without.splice(insertAt, 0, moved);
  if (without.length === ranked.length && without.every((f, i) => f === ranked[i])) {
    return null;
  }
  return without;
}

// positionAfterDrop is the cross-lane sibling: the index into the
// destination lane's on-disk ranked list for a transition-and-place beside
// anchor.
export function positionAfterDrop(lane: Lane, anchor: DropAnchor): number {
  const ranked = lane.cards.slice(0, lane.rankedCount).map((c) => c.filename);
  if (!anchor) {
    return ranked.length;
  }
  const idx = ranked.indexOf(anchor.filename);
  if (idx < 0) {
    return ranked.length;
  }
  return idx + (anchor.after ? 1 : 0);
}
