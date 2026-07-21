// the project lives in the URL — /p/{name} — so windows are independently
// addressable and bookmarkable, and multiple windows on different projects
// need no coordination. no router library: read the segment at load; the
// selector navigates. names are slug-shaped by the config grammar, so the
// segment is taken verbatim — no decoding rule exists anywhere.
export function projectFromPath(pathname: string): string | null {
  const m = pathname.match(/^\/p\/([^/]+)/);
  return m ? m[1] : null;
}
