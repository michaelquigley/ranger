package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	shared "github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

func TestRehearsals(t *testing.T) {
	baab, err := Load("testdata/baab")
	if err != nil {
		t.Fatal(err)
	}
	const score = "agent-mcp-score-creation-from-a-local-file"
	item := baab.Items[score]
	summary := baab.Project.Items[score]
	if !item.HasDesign || item.Plan == nil || item.Hash != shared.Hash(item.Raw) {
		t.Fatalf("BAAB item = %+v", item)
	}
	if want := []string{"design.md", "plan.md", "stages/file-backed-score-calls.md"}; !slices.Equal(item.Files, want) {
		t.Errorf("files = %v, want %v", item.Files, want)
	}
	if progress := summary.Progress; progress.Position != model.CurrentStage || progress.Status != model.Built ||
		progress.Verdict != (model.Verdict{Broker: model.Terminus, Value: "clean rehearsal-run"}) {
		t.Errorf("progress = %+v", progress)
	}
	if needs := baab.Project.Needs; len(needs) != 2 || !strings.HasPrefix(needs[0].Request, "audition:") || !needs[1].Acceptance {
		t.Errorf("needs = %+v", needs)
	}
	if len(baab.Project.Flags) > 0 {
		t.Errorf("BAAB flags = %+v", baab.Project.Flags)
	}

	screen, err := Load("testdata/screeningroom")
	if err != nil {
		t.Fatal(err)
	}
	const punchlist = "look-and-feel-pass"
	const fullscreen = "fullscreen-in-the-client"
	if count := screen.Items[punchlist].Checklist; count != (model.Checklist{Open: 2, Total: 4}) {
		t.Errorf("checklist = %+v", count)
	}
	wantLinks := []model.Link{
		{From: punchlist, Relation: model.Includes, Target: model.Target{Slug: fullscreen}},
		{From: punchlist, Relation: model.BodyLink, Target: model.Target{Slug: "mark-unwatched-leaves-a-resume-card"}},
		{From: punchlist, Relation: model.BodyLink, Target: model.Target{Slug: fullscreen}},
	}
	if links := screen.Project.Items[punchlist].LinksOut; !slices.Equal(links, wantLinks) {
		t.Errorf("links = %+v, want %+v", links, wantLinks)
	}
	if links := screen.Project.Items[fullscreen].LinksIn; !slices.Equal(links, []model.Link{wantLinks[0], wantLinks[2]}) {
		t.Errorf("incoming fullscreen links = %+v", links)
	}
	if links := screen.Project.Items[fullscreen].LinksOut; len(links) != 1 || links[0].Target.String() != "dfw:window-fullscreen" {
		t.Errorf("outgoing fullscreen links = %+v", links)
	}
	if len(screen.Project.Flags) > 0 {
		t.Errorf("Screeningroom flags = %+v", screen.Project.Flags)
	}
}

func TestTitleEditAndMissingTargets(t *testing.T) {
	root := copyFixture(t, "screeningroom")
	const slug = "fullscreen-in-the-client"
	path := filepath.Join(root, FutureRel, slug, "index.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, path, strings.Replace(string(raw), "title: fullscreen in the client", "title: change the window mode", 1))
	snapshot, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	summary := snapshot.Project.Items[slug]
	if flags := summary.Flags; len(flags) != 1 || flags[0].Kind != model.SlugMismatch {
		t.Errorf("renamed title flags = %+v", flags)
	}
	if len(summary.LinksIn) != 2 || !slices.Equal(snapshot.Project.Queue[2].Items, []string{slug, "mark-unwatched-leaves-a-resume-card"}) {
		t.Errorf("title edit lost membership or links: %+v, %+v", snapshot.Project.Queue, summary)
	}

	if err := os.RemoveAll(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	snapshot, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []model.Flag{
		{Kind: model.MissingViewItem, View: model.Queue, Target: slug, Line: 2},
		{Kind: model.MissingLink, Item: "look-and-feel-pass", Target: slug, Relation: model.Includes},
		{Kind: model.MissingLink, Item: "look-and-feel-pass", Target: slug, Relation: model.BodyLink},
	}
	if !slices.Equal(snapshot.Project.Flags, want) {
		t.Errorf("flags = %+v, want %+v", snapshot.Project.Flags, want)
	}
}

func TestViewErrorsAreIndependent(t *testing.T) {
	for _, name := range []model.ViewName{model.Roadmap, model.Queue} {
		t.Run(string(name), func(t *testing.T) {
			root := copyFixture(t, "screeningroom")
			path := filepath.Join(root, RangerRel, string(name)+".yaml")
			writeFixtureFile(t, path, "malformed: [")
			snapshot, err := Load(root)
			if err != nil {
				t.Fatal(err)
			}
			broken, good := snapshot.Roadmap, snapshot.Queue
			brokenLanes, goodLanes := snapshot.Project.Roadmap, snapshot.Project.Queue
			if name == model.Queue {
				broken, good = good, broken
				brokenLanes, goodLanes = goodLanes, brokenLanes
			}
			if broken.Err == nil || !strings.Contains(broken.Err.Error(), filepath.Base(path)) ||
				broken.Version != "" || len(brokenLanes) != 0 || good.Err != nil || len(goodLanes) != 5 {
				t.Fatalf("broken view = %+v, intact view = %+v", broken, good)
			}
			if len(snapshot.Items) != 3 || len(snapshot.Project.Flags) != 0 {
				t.Errorf("items = %v, flags = %+v", len(snapshot.Items), snapshot.Project.Flags)
			}
		})
	}
}

func TestDetectionAndSparseProject(t *testing.T) {
	root := t.TempDir()
	if version, err := Detect(root); err != nil || version != V1 {
		t.Fatalf("unclaimed project: version = %v, error = %v", version, err)
	}
	marker := filepath.Join(root, versionRel)
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", "version: 1\n", "version: [\n", "version: 2\nversion: 2\n"} {
		if value != "" {
			writeFixtureFile(t, marker, value)
		}
		if _, err := Detect(root); err == nil || !strings.Contains(err.Error(), "version.yaml") {
			t.Errorf("invalid marker %q: error = %v", value, err)
		}
	}
	writeFixtureFile(t, marker, "version: 2\n")
	writeFixtureFile(t, filepath.Join(root, FutureRel, "roadmap", "index.md"),
		"---\ntitle: Roadmap\nstate: inbox\ncreated: 2026-11-24\n---\n")
	writeFixtureFile(t, filepath.Join(root, FutureRel, ".hidden", "index.md"), "ignored")
	if version, err := Detect(root); err != nil || version != V2 {
		t.Fatalf("v2 marker: version = %v, error = %v", version, err)
	}
	snapshot, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Items) != 1 || snapshot.Roadmap.Version != shared.VersionAbsent ||
		snapshot.Queue.Version != shared.VersionAbsent {
		t.Errorf("sparse snapshot = %+v", snapshot)
	}
	for _, lanes := range [][]model.Lane{snapshot.Project.Roadmap, snapshot.Project.Queue} {
		if !slices.Equal(lanes[0].Items, []string{"roadmap"}) {
			t.Errorf("lanes = %+v", lanes)
		}
	}
	writeFixtureFile(t, filepath.Join(root, FutureRel, "roadmap", "index.md"),
		"---\ntitle: Roadmap\nstate: pending\ncreated: 2026-11-24\n---\n")
	writeFixtureFile(t, filepath.Join(root, RangerRel, "roadmap.yaml"), "- roadmap\n")
	writeFixtureFile(t, filepath.Join(root, RangerRel, "queue.yaml"), "# no queue entries\n")
	snapshot, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(snapshot.Project.Roadmap[2].Items, []string{"roadmap"}) || len(snapshot.Project.Queue[2].Items) != 0 ||
		snapshot.Queue.Version != shared.Hash([]byte("# no queue entries\n")) {
		t.Errorf("views = %+v, %+v", snapshot.Project.Roadmap, snapshot.Project.Queue)
	}
}

func TestBrokenItemStillLoads(t *testing.T) {
	root := copyFixture(t, "baab")
	path := filepath.Join(root, FutureRel, "broken", "index.md")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	snapshot, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	summary := snapshot.Project.Items["broken"]
	if summary == nil || len(summary.Flags) != 1 || summary.Flags[0].Kind != model.Malformed {
		t.Fatalf("unreadable index.md did not get a malformed flag: %+v", summary)
	}
	if snapshot.Items["broken"].Lane() != model.Inbox {
		t.Errorf("unreadable index.md belongs in Inbox")
	}
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS(filepath.Join("testdata", name))); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeFixtureFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSymlinkedOptionalDocumentsCountAsMissing(t *testing.T) {
	const slug = "agent-mcp-score-creation-from-a-local-file"
	const stage = "file-backed-score-calls"
	stageFile := filepath.Join("stages", stage+".md")
	for _, relative := range []string{"design.md", "plan.md", "stages", stageFile} {
		t.Run(relative, func(t *testing.T) {
			root := copyFixture(t, "baab")
			link := filepath.Join(root, FutureRel, slug, relative)
			target := filepath.Join(t.TempDir(), filepath.Base(relative))
			if err := os.Rename(link, target); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			snapshot, err := Load(root)
			if err != nil {
				t.Fatal(err)
			}
			item := snapshot.Items[slug]
			summary := snapshot.Project.Items[slug]
			if slices.Contains(item.Files, filepath.ToSlash(relative)) {
				t.Errorf("symlinked file is listed: %v", item.Files)
			}
			switch relative {
			case "design.md":
				if item.HasDesign {
					t.Error("symlinked design counts as present")
				}
			case "plan.md":
				if item.Plan != nil || item.PlanError != "" {
					t.Errorf("symlinked plan did not count as missing: %+v, %q", item.Plan, item.PlanError)
				}
			default:
				want := model.Flag{Kind: model.MissingStage, Item: slug, Target: stage}
				if len(item.Stages) != 0 || len(summary.Needs) != 0 ||
					!slices.Equal(summary.Flags, []model.Flag{want}) || slices.Contains(item.Files, filepath.ToSlash(stageFile)) {
					t.Errorf("symlinked stage did not count as missing: stages %+v, summary %+v, files %v", item.Stages, summary, item.Files)
				}
			}
		})
	}
}

func TestLoadRefusesSymlinkedVersionMarker(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, versionRel)
	target := filepath.Join(t.TempDir(), "target.yaml")
	writeFixtureFile(t, target, "version: 2\n")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, marker); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), marker+": symlink refused") {
		t.Errorf("load error = %v, want symlink refusal at %s", err, marker)
	}
}
