import type { Board, Card, SavedFilter } from "./api";

// the board filter carries polarity per dimension: a card must carry
// every included tag and subsystem, none of the excluded ones, and match
// the milestone selection. tags and subsystems are multi-select; the
// milestone is a single slot holding one value of either polarity,
// matching its single-select include behavior. a value never sits in
// both polarities — toggling it into one displaces it from the other, so
// a flip is one gesture.
export type BoardFilters = {
  tags: string[];
  notTags: string[];
  subsystems: string[];
  notSubsystems: string[];
  milestone: string | null;
  notMilestone: string | null;
};

export const emptyFilters: BoardFilters = {
  tags: [],
  notTags: [],
  subsystems: [],
  notSubsystems: [],
  milestone: null,
  notMilestone: null,
};

export function isFiltering(f: BoardFilters): boolean {
  return (
    f.tags.length > 0 ||
    f.notTags.length > 0 ||
    f.subsystems.length > 0 ||
    f.notSubsystems.length > 0 ||
    f.milestone !== null ||
    f.notMilestone !== null
  );
}

function toggled(list: string[], value: string): string[] {
  return list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
}

function without(list: string[], value: string): string[] {
  return list.filter((v) => v !== value);
}

export function toggleTag(f: BoardFilters, tag: string, exclude: boolean): BoardFilters {
  return exclude
    ? { ...f, notTags: toggled(f.notTags, tag), tags: without(f.tags, tag) }
    : { ...f, tags: toggled(f.tags, tag), notTags: without(f.notTags, tag) };
}

export function toggleSubsystem(f: BoardFilters, subsystem: string, exclude: boolean): BoardFilters {
  return exclude
    ? { ...f, notSubsystems: toggled(f.notSubsystems, subsystem), subsystems: without(f.subsystems, subsystem) }
    : { ...f, subsystems: toggled(f.subsystems, subsystem), notSubsystems: without(f.notSubsystems, subsystem) };
}

// the milestone keeps its single-select shape across both polarities:
// toggling the active value clears the slot, toggling anything else
// replaces it whole.
export function toggleMilestone(f: BoardFilters, milestone: string, exclude: boolean): BoardFilters {
  const active = exclude ? f.notMilestone : f.milestone;
  const next = active === milestone ? null : milestone;
  return exclude ? { ...f, milestone: null, notMilestone: next } : { ...f, milestone: next, notMilestone: null };
}

// toSaved shapes the active filter as a named wire filter, empty
// dimensions omitted — the rendered filters.yaml carries only what the
// filter actually says. the search query is deliberately not part of a
// saved filter; it is a different gesture with its own lifecycle.
export function toSaved(name: string, f: BoardFilters): SavedFilter {
  return {
    name,
    ...(f.tags.length > 0 ? { tags: f.tags } : {}),
    ...(f.notTags.length > 0 ? { notTags: f.notTags } : {}),
    ...(f.subsystems.length > 0 ? { subsystems: f.subsystems } : {}),
    ...(f.notSubsystems.length > 0 ? { notSubsystems: f.notSubsystems } : {}),
    ...(f.milestone !== null ? { milestone: f.milestone } : {}),
    ...(f.notMilestone !== null ? { notMilestone: f.notMilestone } : {}),
  };
}

// toFilters applies a saved filter as the whole active filter state.
export function toFilters(s: SavedFilter): BoardFilters {
  return {
    tags: s.tags ?? [],
    notTags: s.notTags ?? [],
    subsystems: s.subsystems ?? [],
    notSubsystems: s.notSubsystems ?? [],
    milestone: s.milestone ?? null,
    notMilestone: s.notMilestone ?? null,
  };
}

// isApplied reports whether a saved filter is what the board is showing
// right now — same members per dimension, order aside. derived rather
// than remembered, so the highlight can't lie: edit the filter after
// applying and the match drops; rebuild the same lens by hand and its
// pill lights. the search query stays outside the comparison, as it
// stays outside the saved filter.
export function isApplied(f: BoardFilters, s: SavedFilter): boolean {
  const same = (a: string[], b: string[]) => a.length === b.length && a.every((v) => b.includes(v));
  const saved = toFilters(s);
  return (
    same(f.tags, saved.tags) &&
    same(f.notTags, saved.notTags) &&
    same(f.subsystems, saved.subsystems) &&
    same(f.notSubsystems, saved.notSubsystems) &&
    f.milestone === saved.milestone &&
    f.notMilestone === saved.notMilestone
  );
}

// upsert replaces the same-named filter in place — saving under an
// existing name updates it without changing its position — and appends a
// new name at the end.
export function upsert(list: SavedFilter[], saved: SavedFilter): SavedFilter[] {
  if (list.some((s) => s.name === saved.name)) {
    return list.map((s) => (s.name === saved.name ? saved : s));
  }
  return [...list, saved];
}

export function withoutName(list: SavedFilter[], name: string): SavedFilter[] {
  return list.filter((s) => s.name !== name);
}

// filterBoard narrows every lane to cards satisfying the filter and, when
// a search is active, present in the match set. excluding a milestone
// keeps unmilestoned cards — the exclusion names a train, not the absence
// of one. rankedCount shrinks to the surviving members of the ranked
// prefix so the boundary rule still lands between ranked and unranked
// cards.
export function filterBoard(board: Board, f: BoardFilters, searchMatches: Set<string> | null): Board {
  const matches = (c: Card) =>
    f.tags.every((t) => (c.tags ?? []).includes(t)) &&
    !f.notTags.some((t) => (c.tags ?? []).includes(t)) &&
    f.subsystems.every((s) => (c.subsystems ?? []).includes(s)) &&
    !f.notSubsystems.some((s) => (c.subsystems ?? []).includes(s)) &&
    (f.milestone === null || c.milestone === f.milestone) &&
    (f.notMilestone === null || c.milestone !== f.notMilestone) &&
    (searchMatches === null || searchMatches.has(c.filename));
  return {
    ...board,
    lanes: board.lanes.map((lane) => ({
      ...lane,
      cards: lane.cards.filter(matches),
      rankedCount: lane.cards.slice(0, lane.rankedCount).filter(matches).length,
    })),
  };
}
