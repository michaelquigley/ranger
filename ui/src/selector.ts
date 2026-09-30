import type { ProjectIndex } from "./api";

// the selector renders the project index — available projects by name,
// unavailable ones present but flagged with their diagnostic rather than
// hidden: the same explain-don't-hide posture the board takes toward
// malformed items.
export type SelectorOption = {
  name: string;
  label: string;
  disabled: boolean;
  title: string | null;
};

// SelectorGroup is one run of options in the dropdown. a labeled group
// renders as an <optgroup>; a null label renders its options bare.
export type SelectorGroup = {
  label: string | null;
  options: SelectorOption[];
};

// selectorGroups turns the index into the dropdown's grouped option list.
// an unavailable project keeps its place with the diagnostic in the label
// ("anpheq — roadmap directory not found") and rides disabled — its error
// already fills the body region when the URL lands on it. a current name
// the index doesn't carry (an unknown project in the URL) is kept as a
// disabled bare entry at the top so the header reflects the URL
// truthfully. a dirty project carries the git vernacular's own marker
// (" *") — and because the closed control renders the current option's
// label, the marker on the current project is the board-level glance for
// free. dirty projects come first, under an "uncommitted" group with the
// rest under "other" — never "committed", because an absent verdict means
// git couldn't answer, not cleanliness — config order kept within each; the groups only
// appear when both halves are non-empty, so a wholly clean (or wholly
// dirty) index renders as the plain list. the ordering is presentation
// only — the index on the wire stays in config order.
export function selectorGroups(index: ProjectIndex, current: string): SelectorGroup[] {
  const options = index.projects.map((p) => ({
    name: p.name,
    label: p.available ? (p.dirty ? `${p.name} *` : p.name) : `${p.name} — ${p.error ?? "unavailable"}`,
    disabled: !p.available,
    title: p.error ?? (p.dirty ? "uncommitted changes" : null),
    dirty: p.dirty ?? false,
  }));
  const strip = ({ dirty: _, ...o }: (typeof options)[number]): SelectorOption => o;
  const dirty = options.filter((o) => o.dirty).map(strip);
  const other = options.filter((o) => !o.dirty).map(strip);
  const groups: SelectorGroup[] =
    dirty.length > 0 && other.length > 0
      ? [
          { label: "uncommitted", options: dirty },
          { label: "other", options: other },
        ]
      : [{ label: null, options: options.map(strip) }];
  if (!options.some((o) => o.name === current)) {
    groups.unshift({
      label: null,
      options: [{ name: current, label: current, disabled: true, title: "not a configured project" }],
    });
  }
  return groups;
}

// selectorOptions is the grouped list flattened: the order the open
// control shows, top to bottom.
export function selectorOptions(index: ProjectIndex, current: string): SelectorOption[] {
  return selectorGroups(index, current).flatMap((g) => g.options);
}
