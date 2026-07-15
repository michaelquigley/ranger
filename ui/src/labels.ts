// label colors, matched to the house palette used across project boards:
// solid chips, white or dark text by fill. anything unlisted derives a
// deterministic solid hue from the tag text, so a label reads the same
// color on every board without configuration — the file convention stays
// free of presentation knowledge.

export type LabelColor = {
  background: string;
  borderColor: string;
  color: string;
};

const dark = "#2c3345";
const light = "#ffffff";

const house: Record<string, { background: string; color: string }> = {
  "defect": { background: "#fb0007", color: light },
  "documentation": { background: "#1f7ce8", color: light },
  "enhancement": { background: "#cbdcf5", color: dark },
  "epic": { background: "#fbca04", color: dark },
  "feature": { background: "#28a745", color: light },
  "spike": { background: "#a020f0", color: light },
  "story": { background: "#3e4b5b", color: light },
  "flo (creative workflow)": { background: "#12a99b", color: light },
  "reef (estate management)": { background: "#3d5069", color: light },
};

export function labelColor(tag: string): LabelColor {
  const known = house[tag.toLowerCase()];
  if (known) {
    return { ...known, borderColor: "rgba(0, 0, 0, 0.18)" };
  }
  const hue = hueOf(tag);
  return {
    background: `hsl(${hue}, 60%, 42%)`,
    borderColor: "rgba(0, 0, 0, 0.18)",
    color: light,
  };
}

// sortedTags orders tags for display — case-insensitive, alphabetical. a
// render-side sort only: the file's own tag order is the author's and never
// rewritten.
export function sortedTags(tags: string[] | undefined): string[] {
  return [...(tags ?? [])].sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));
}

function hueOf(tag: string): number {
  let h = 0;
  for (let i = 0; i < tag.length; i++) {
    h = (h * 31 + tag.charCodeAt(i)) >>> 0;
  }
  return h % 360;
}
