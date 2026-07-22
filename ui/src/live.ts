// live reload: the freshness contract is a quiet clock on the same fresh
// reads a manual refresh runs. two seconds is close enough to live for a
// localhost tool, and trivial against reads that were already
// fresh-per-request on the server side.
export const POLL_MS = 2000;

// freshest keeps the previous object's identity when the fresh read
// carries identical content, so a quiet poll never churns a repaint —
// replacement happens only when something actually changed.
export function freshest<T>(prev: T | null, next: T): T {
  if (prev !== null && JSON.stringify(prev) === JSON.stringify(next)) return prev;
  return next;
}
