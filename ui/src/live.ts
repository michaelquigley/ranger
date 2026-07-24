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

// the open item modal rides the same clock: the fresh board's hash for
// the item is the change signal.
export type ItemReloadInput = {
  // in-flight operator input is never stomped — raw edit, an active
  // retitle, a delete confirm, or a shown collision notice each suspend
  // the reload; the hash guard owns any conflict at save, as ever.
  editing: boolean;
  retitling: boolean;
  confirming: boolean;
  noticed: boolean;
  loadedHash: string;
  // the board's current hash for the item; undefined when the board no
  // longer carries the card (deleted or renamed elsewhere), which reloads
  // into the plain 404 rather than presenting stale bytes as current.
  boardHash: string | undefined;
};

export function shouldReloadItem(i: ItemReloadInput): boolean {
  if (i.editing || i.retitling || i.confirming || i.noticed) return false;
  return i.boardHash !== i.loadedHash;
}
