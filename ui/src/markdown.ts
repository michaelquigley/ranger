import { defaultUrlTransform } from "react-markdown";

// relative urls in item bodies resolve against the current project's
// roadmap directory — the reading obsidian and github give them — via the
// server's read-only /roadmap/{project}/ route. absolute urls,
// root-relative paths, and fragments pass through untouched;
// react-markdown's default transform keeps its dangerous-protocol
// sanitization either way.
//
// containment is a resolved-path check, never a spelling scan: the
// candidate resolves through real URL semantics, and the result must sit
// under /roadmap/{project}/ or it renders inert — browsers normalize ../,
// backslash separators, and percent-encoded dot segments alike before the
// request ever leaves, which no route-side check can see.
export function makeRoadmapUrl(project: string): (url: string) => string {
  const prefix = `/roadmap/${project}/`;
  const base = new URL(prefix, "http://roadmap.invalid");
  return (url: string) => {
    if (/^[a-z][a-z0-9+.-]*:/i.test(url) || url.startsWith("/") || url.startsWith("#")) {
      return defaultUrlTransform(url);
    }
    let resolved: URL;
    try {
      resolved = new URL(url, base);
    } catch {
      return "";
    }
    if (resolved.origin !== base.origin || !resolved.pathname.startsWith(prefix)) {
      return "";
    }
    return defaultUrlTransform(resolved.pathname + resolved.search + resolved.hash);
  };
}
