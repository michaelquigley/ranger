import { describe, expect, it } from "vitest";
import { projectFromPath } from "./project";

describe("projectFromPath", () => {
  it("reads the project segment from /p/{name}", () => {
    expect(projectFromPath("/p/ranger")).toBe("ranger");
    expect(projectFromPath("/p/my-repo/")).toBe("my-repo");
  });

  it("returns null when no project is in the URL", () => {
    expect(projectFromPath("/")).toBe(null);
    expect(projectFromPath("/p/")).toBe(null);
    expect(projectFromPath("/other")).toBe(null);
  });
});
