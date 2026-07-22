import { describe, expect, it } from "vitest";
import { freshest } from "./live";

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
