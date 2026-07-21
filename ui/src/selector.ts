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

// selectorOptions turns the index into the dropdown's option list. an
// unavailable project keeps its place with the diagnostic in the label
// ("anpheq — roadmap directory not found") and rides disabled — its error
// already fills the body region when the URL lands on it. a current name
// the index doesn't carry (an unknown project in the URL) is kept as a
// disabled entry so the header reflects the URL truthfully.
export function selectorOptions(index: ProjectIndex, current: string): SelectorOption[] {
  const options = index.projects.map((p) => ({
    name: p.name,
    label: p.available ? p.name : `${p.name} — ${p.error ?? "unavailable"}`,
    disabled: !p.available,
    title: p.error ?? null,
  }));
  if (!options.some((o) => o.name === current)) {
    options.unshift({ name: current, label: current, disabled: true, title: "not a configured project" });
  }
  return options;
}
