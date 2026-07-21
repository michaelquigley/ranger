import { describe, expect, it } from "vitest";
import { selectorOptions } from "./selector";
import { projectPath } from "./project";

describe("selectorOptions", () => {
  it("lists available projects by name, enabled", () => {
    const options = selectorOptions(
      { projects: [{ name: "ranger", available: true }, { name: "archive", available: true }], default: "ranger" },
      "ranger",
    );
    expect(options).toEqual([
      { name: "ranger", label: "ranger", disabled: false, title: null },
      { name: "archive", label: "archive", disabled: false, title: null },
    ]);
  });

  it("keeps an unavailable project present, flagged with its diagnostic", () => {
    const options = selectorOptions(
      {
        projects: [
          { name: "ranger", available: true },
          { name: "anpheq", available: false, error: "roadmap directory not found" },
        ],
        default: "ranger",
      },
      "ranger",
    );
    expect(options[1]).toEqual({
      name: "anpheq",
      label: "anpheq — roadmap directory not found",
      disabled: true,
      title: "roadmap directory not found",
    });
  });

  it("survives the trap case: an unavailable default with a healthy sibling", () => {
    // the URL landed on the broken default; the selector must still show
    // it truthfully while leaving the healthy sibling one click away.
    const options = selectorOptions(
      {
        projects: [
          { name: "broken", available: false, error: "roadmap directory: no such file" },
          { name: "healthy", available: true },
        ],
        default: "broken",
      },
      "broken",
    );
    const current = options.find((o) => o.name === "broken");
    const sibling = options.find((o) => o.name === "healthy");
    expect(current).toBeDefined();
    expect(current?.disabled).toBe(true);
    expect(current?.label).toContain("roadmap directory");
    expect(sibling?.disabled).toBe(false);
    expect(projectPath(sibling!.name)).toBe("/p/healthy");
  });

  it("keeps an unknown current project as a disabled entry", () => {
    const options = selectorOptions(
      { projects: [{ name: "ranger", available: true }], default: "ranger" },
      "nope",
    );
    expect(options[0]).toEqual({
      name: "nope",
      label: "nope",
      disabled: true,
      title: "not a configured project",
    });
    expect(options[1].name).toBe("ranger");
  });
});

describe("projectPath", () => {
  it("is projectFromPath's dual", () => {
    expect(projectPath("my-repo")).toBe("/p/my-repo");
  });
});
