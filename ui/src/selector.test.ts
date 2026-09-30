import { describe, expect, it } from "vitest";
import { selectorGroups, selectorOptions } from "./selector";
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

  it("marks a dirty project with the git vernacular's asterisk", () => {
    const options = selectorOptions(
      {
        projects: [
          { name: "ranger", available: true, dirty: true },
          { name: "archive", available: true, dirty: false },
        ],
        default: "ranger",
      },
      "ranger",
    );
    expect(options[0]).toEqual({ name: "ranger", label: "ranger *", disabled: false, title: "uncommitted changes" });
    expect(options[1]).toEqual({ name: "archive", label: "archive", disabled: false, title: null });
  });

  it("sorts dirty projects first, config order kept within each half", () => {
    const options = selectorOptions(
      {
        projects: [
          { name: "clean-a", available: true, dirty: false },
          { name: "dirty-a", available: true, dirty: true },
          { name: "unknown", available: true },
          { name: "dirty-b", available: true, dirty: true },
          { name: "clean-b", available: true, dirty: false },
        ],
        default: "clean-a",
      },
      "clean-a",
    );
    expect(options.map((o) => o.name)).toEqual(["dirty-a", "dirty-b", "clean-a", "unknown", "clean-b"]);
  });

  it("groups the halves as uncommitted and other when both are present", () => {
    const groups = selectorGroups(
      {
        projects: [
          { name: "clean-a", available: true, dirty: false },
          { name: "dirty-a", available: true, dirty: true },
          { name: "clean-b", available: true, dirty: false },
        ],
        default: "clean-a",
      },
      "clean-a",
    );
    expect(groups.map((g) => [g.label, g.options.map((o) => o.name)])).toEqual([
      ["uncommitted", ["dirty-a"]],
      ["other", ["clean-a", "clean-b"]],
    ]);
  });

  it("renders a wholly clean index as one bare group", () => {
    const groups = selectorGroups(
      { projects: [{ name: "ranger", available: true }, { name: "archive", available: true, dirty: false }], default: "ranger" },
      "ranger",
    );
    expect(groups.map((g) => [g.label, g.options.map((o) => o.name)])).toEqual([[null, ["ranger", "archive"]]]);
  });

  it("renders a wholly dirty index as one bare group", () => {
    const groups = selectorGroups(
      { projects: [{ name: "ranger", available: true, dirty: true }, { name: "archive", available: true, dirty: true }], default: "ranger" },
      "ranger",
    );
    expect(groups.map((g) => [g.label, g.options.map((o) => o.label)])).toEqual([[null, ["ranger *", "archive *"]]]);
  });

  it("keeps the unknown current as a bare entry above the groups", () => {
    const groups = selectorGroups(
      { projects: [{ name: "ranger", available: true, dirty: true }, { name: "archive", available: true }], default: "ranger" },
      "nope",
    );
    expect(groups.map((g) => [g.label, g.options.map((o) => o.name)])).toEqual([
      [null, ["nope"]],
      ["uncommitted", ["ranger"]],
      ["other", ["archive"]],
    ]);
    expect(groups[0].options[0].disabled).toBe(true);
  });

  it("lets an unavailable project's diagnostic dominate its dirty verdict", () => {
    const options = selectorOptions(
      {
        projects: [{ name: "broken", available: false, error: "roadmap directory not found", dirty: true }],
        default: "broken",
      },
      "broken",
    );
    expect(options[0].label).toBe("broken — roadmap directory not found");
    expect(options[0].title).toBe("roadmap directory not found");
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
