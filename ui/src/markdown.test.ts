import { describe, expect, it } from "vitest";
import { makeRoadmapUrl } from "./markdown";

describe("makeRoadmapUrl", () => {
  const transform = makeRoadmapUrl("a");

  it("resolves relative paths against the project's roadmap prefix", () => {
    expect(transform("images/pic.png")).toBe("/roadmap/a/images/pic.png");
  });

  it("strips a leading ./", () => {
    expect(transform("./images/pic.png")).toBe("/roadmap/a/images/pic.png");
  });

  it("keeps a traversal that stays inside the project", () => {
    expect(transform("images/../pic.png")).toBe("/roadmap/a/pic.png");
  });

  it("leaves absolute urls alone", () => {
    expect(transform("https://example.com/pic.png")).toBe("https://example.com/pic.png");
  });

  it("leaves root-relative paths and fragments alone", () => {
    expect(transform("/already/rooted.png")).toBe("/already/rooted.png");
    expect(transform("#section")).toBe("#section");
  });

  it("keeps the default dangerous-protocol sanitization", () => {
    expect(transform("javascript:alert(1)")).toBe("");
  });

  it("renders any spelling of a cross-project traversal inert", () => {
    // browsers normalize all three before the request ever leaves; the
    // resolved-path check must catch every spelling the same way.
    expect(transform("../b/images/x.png")).toBe("");
    expect(transform("..\\b\\images\\x.png")).toBe("");
    expect(transform("%2e%2e/b/images/x.png")).toBe("");
    expect(transform(".%2e/b/images/x.png")).toBe("");
  });

  it("renders a host-changing reference inert", () => {
    expect(transform("\\\\evil.example\\x.png")).toBe("");
  });
});
