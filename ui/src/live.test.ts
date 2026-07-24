import { describe, expect, it } from "vitest";
import { freshest, shouldReloadItem } from "./live";

describe("freshest", () => {
  it("keeps the previous identity when content is unchanged", () => {
    const prev = { lanes: [{ state: "inbox", cards: [] }], orderVersion: "abc" };
    const next = { lanes: [{ state: "inbox", cards: [] }], orderVersion: "abc" };
    expect(freshest(prev, next)).toBe(prev);
  });

  it("replaces on any content change", () => {
    const prev = { orderVersion: "abc" };
    const next = { orderVersion: "def" };
    expect(freshest(prev, next)).toBe(next);
  });

  it("adopts the first read", () => {
    const next = { orderVersion: "abc" };
    expect(freshest(null, next)).toBe(next);
  });
});

describe("shouldReloadItem", () => {
  const quiet = { editing: false, retitling: false, confirming: false, noticed: false };

  it("reloads when the board carries a different hash", () => {
    expect(shouldReloadItem({ ...quiet, loadedHash: "a", boardHash: "b" })).toBe(true);
  });

  it("stays quiet on a matching hash", () => {
    expect(shouldReloadItem({ ...quiet, loadedHash: "a", boardHash: "a" })).toBe(false);
  });

  it("reloads when the board no longer carries the card", () => {
    // deleted or renamed elsewhere: the reload surfaces the 404 plainly
    // rather than presenting stale bytes as current.
    expect(shouldReloadItem({ ...quiet, loadedHash: "a", boardHash: undefined })).toBe(true);
  });

  it("never stomps in-flight operator input", () => {
    const changed = { loadedHash: "a", boardHash: "b" };
    expect(shouldReloadItem({ ...quiet, ...changed, editing: true })).toBe(false);
    expect(shouldReloadItem({ ...quiet, ...changed, retitling: true })).toBe(false);
    expect(shouldReloadItem({ ...quiet, ...changed, confirming: true })).toBe(false);
    expect(shouldReloadItem({ ...quiet, ...changed, noticed: true })).toBe(false);
  });
});
