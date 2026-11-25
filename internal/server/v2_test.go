package server

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/document"
)

func copyRehearsal(t *testing.T, name string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS(filepath.Join("..", "v2", "workspace", "testdata", name))); err != nil {
		t.Fatal(err)
	}
	return root
}

func v2FixtureClient(t *testing.T, refs ...config.ProjectRef) *api.Client {
	t.Helper()
	projects := NewProjects(constantSource(&config.Config{Projects: refs, Default: refs[0].Name}))
	h, err := api.NewServer(NewServer(projects))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	client, err := api.NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func readV2Project(t *testing.T, client *api.Client, name string) *api.V2Project {
	t.Helper()
	res, err := client.GetV2Project(context.Background(), api.GetV2ProjectParams{Project: name})
	if project, ok := res.(*api.V2Project); ok && err == nil {
		return project
	}
	t.Fatalf("GET v2 project %s = %#v, %v", name, res, err)
	return nil
}

func readV2Item(t *testing.T, client *api.Client, name, slug string) *api.V2Item {
	t.Helper()
	res, err := client.GetV2Item(context.Background(), api.GetV2ItemParams{Project: name, Slug: slug})
	if item, ok := res.(*api.V2Item); ok && err == nil {
		return item
	}
	t.Fatalf("GET v2 item %s/%s = %#v, %v", name, slug, res, err)
	return nil
}

func TestV2ReadsOverHTTP(t *testing.T) {
	baab, screen := copyRehearsal(t, "baab"), copyRehearsal(t, "screeningroom")
	client := v2FixtureClient(t, config.ProjectRef{Root: baab, Name: "baab"}, config.ProjectRef{Root: screen, Name: "screen"})
	const score = "agent-mcp-score-creation-from-a-local-file"
	const stage = "file-backed-score-calls"

	project := readV2Project(t, client, "baab")
	if project.Project != "baab" || !project.Queue.Version.Set || !project.Roadmap.Version.Set {
		t.Fatalf("project = %+v", project)
	}
	wantLanes := []api.V2State{api.V2StateInbox, api.V2StateHorizon, api.V2StatePending, api.V2StateBuilding, api.V2StateEvaluating}
	for _, view := range []api.V2View{project.Roadmap, project.Queue} {
		var got []api.V2State
		for _, lane := range view.Lanes {
			got = append(got, lane.State)
		}
		if !slices.Equal(got, wantLanes) || !slices.Equal(view.Lanes[3].Items, []string{score}) {
			t.Fatalf("lanes = %+v", view.Lanes)
		}
	}
	progress := project.Items[score].Progress.Value
	if progress.Stage.Value != stage || progress.Number.Value != 1 || progress.Total.Value != 1 ||
		progress.Status.Value != api.V2StageStatusBuilt ||
		progress.Verdict.Value != (api.V2Verdict{Broker: api.V2VerdictBrokerTerminus, Value: "clean rehearsal-run"}) {
		t.Errorf("progress = %+v", progress)
	}
	if len(project.Needs) != 2 || !project.Needs[1].Acceptance || project.Needs[1].Stage.Value != stage || len(project.Flags) != 0 {
		t.Errorf("needs = %+v, flags = %+v", project.Needs, project.Flags)
	}

	item := readV2Item(t, client, "baab", score)
	raw, err := os.ReadFile(filepath.Join(baab, "docs", "future", score, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if item.Content != string(raw) || item.Hash != document.Hash(raw) || !item.HasDesign {
		t.Errorf("item content, hash or design missing: %+v", item)
	}
	if !slices.Equal(item.Plan.Value.Stages, []string{stage}) ||
		item.Plan.Value.Mercurius.Value != "ready_to_build s_GtmwEX60upbB" ||
		item.Stages[stage].Status.Value != api.V2StageStatusBuilt ||
		item.Stages[stage].Terminus.Value != "clean rehearsal-run" ||
		len(item.Stages[stage].Needs) != 1 ||
		!slices.Equal(item.Files, []string{"design.md", "plan.md", "stages/" + stage + ".md"}) {
		t.Errorf("item plan, stages or files = %+v", item)
	}

	const punchlist = "look-and-feel-pass"
	const fullscreen = "fullscreen-in-the-client"
	punch := readV2Item(t, client, "screen", punchlist)
	if punch.Summary.Checklist != (api.V2Checklist{Open: 2, Total: 4}) || len(punch.LinksOut) != 3 {
		t.Errorf("punchlist = %+v", punch)
	}
	full := readV2Item(t, client, "screen", fullscreen)
	if len(full.LinksIn) != 2 || full.LinksIn[0].From != punchlist ||
		len(full.LinksOut) != 1 || full.LinksOut[0].Project.Value != "dfw" ||
		full.LinksOut[0].Slug != "window-fullscreen" {
		t.Errorf("fullscreen links in = %+v, out = %+v", full.LinksIn, full.LinksOut)
	}
	if project.Roadmap.Git.Set || project.Queue.Git.Set || project.Items[score].Git.Set || item.Summary.Git.Set {
		t.Error("git marks must be absent outside a repository")
	}
}

func TestV2ReadErrorTiers(t *testing.T) {
	ctx := context.Background()
	root := copyRehearsal(t, "screeningroom")
	client := v2FixtureClient(t, config.ProjectRef{Root: root, Name: "screen"})
	for _, file := range []string{".ranger/queue.yaml", "broken/index.md"} {
		writeFuture(t, root, file, "---\nstate: unknown\n---\n")
	}

	project := readV2Project(t, client, "screen")
	if !project.Queue.Error.Set || project.Queue.Version.Set || len(project.Queue.Lanes) != 0 ||
		project.Roadmap.Error.Set || !project.Roadmap.Version.Set || len(project.Roadmap.Lanes) != 5 ||
		len(project.Items) != 4 {
		t.Fatalf("degraded project = %+v", project)
	}
	broken := readV2Item(t, client, "screen", "broken")
	if broken.Summary.State != api.V2StateInbox || len(broken.Summary.Flags) == 0 ||
		broken.Summary.Flags[0].Kind != api.V2FlagKindMalformed {
		t.Errorf("broken summary = %+v", broken.Summary)
	}
	index, err := client.GetProjects(ctx)
	if err != nil || !index.Projects[0].Available || index.Projects[0].Convention.Value != api.ProjectStatusConvention2 {
		t.Fatalf("index = %+v, %v", index, err)
	}

	if err := os.Remove(filepath.Join(root, "docs", "future", ".ranger", "version.yaml")); err != nil {
		t.Fatal(err)
	}
	index, err = client.GetProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	status := index.Projects[0]
	if status.Available || status.Convention.Set || status.Dirty.Set || !strings.Contains(status.Error.Value, "version.yaml") {
		t.Errorf("index after removing marker = %+v", status)
	}
}

func TestConventionRefusalsOverHTTP(t *testing.T) {
	ctx := context.Background()
	v1Root, v2Root := healthyRoot(t), copyRehearsal(t, "baab")
	writeFuture(t, v2Root, "roadmap/order.yaml", "# v1 would migrate this file\n")
	client := v2FixtureClient(t, config.ProjectRef{Root: v1Root, Name: "one"}, config.ProjectRef{Root: v2Root, Name: "two"})

	board, err := client.GetBoard(ctx, api.GetBoardParams{Project: "two"})
	if refusal, ok := board.(*api.ConventionMismatch); err != nil || !ok || !strings.Contains(refusal.Message, "uses convention v2") {
		t.Fatalf("v1 board on v2 = %#v, %v", board, err)
	}
	capture, err := client.CreateItem(ctx, &api.CreateItemReq{Title: "Wrong project"}, api.CreateItemParams{Project: "two"})
	if refusal, ok := capture.(*api.CreateItemBadRequest); err != nil || !ok || !strings.Contains(refusal.Message, "uses convention v2") {
		t.Fatalf("v1 capture on v2 = %#v, %v", capture, err)
	}
	project, err := client.GetV2Project(ctx, api.GetV2ProjectParams{Project: "one"})
	if refusal, ok := project.(*api.ConventionMismatch); err != nil || !ok || !strings.Contains(refusal.Message, "uses convention v1") {
		t.Fatalf("v2 project on v1 = %#v, %v", project, err)
	}
	item, err := client.GetV2Item(ctx, api.GetV2ItemParams{Project: "one", Slug: "missing"})
	if _, ok := item.(*api.ConventionMismatch); err != nil || !ok {
		t.Fatalf("v2 item on v1 = %#v, %v", item, err)
	}
	item, err = client.GetV2Item(ctx, api.GetV2ItemParams{Project: "two", Slug: "missing"})
	if _, ok := item.(*api.ErrorResponse); err != nil || !ok {
		t.Fatalf("missing v2 item = %#v, %v", item, err)
	}

	index, err := client.GetProjects(ctx)
	if err != nil || !index.Projects[0].Available || !index.Projects[1].Available ||
		index.Projects[0].Convention.Value != api.ProjectStatusConvention1 ||
		index.Projects[1].Convention.Value != api.ProjectStatusConvention2 {
		t.Fatalf("mixed index = %+v, %v", index, err)
	}
	if _, err := os.Stat(filepath.Join(v2Root, "docs", "future", "roadmap", "order.yaml")); err != nil {
		t.Fatalf("v1 loader touched the v2 root: %v", err)
	}
}

func writeFuture(t *testing.T, root, path, content string) {
	t.Helper()
	dst := filepath.Join(root, "docs", "future", filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestV2GitMarksAndScope(t *testing.T) {
	ctx := context.Background()
	v1Root, v2Root := healthyRoot(t), t.TempDir()
	writeFuture(t, v2Root, ".ranger/version.yaml", "version: 2\n")
	writeFuture(t, v2Root, ".ranger/roadmap.yaml", "- unstaged\n")
	itemText := func(slug string) string {
		return "---\ntitle: " + slug + "\ncreated: 2026-11-25\nstate: inbox\n---\n"
	}
	for _, slug := range []string{"clean", "unstaged", "staged", "sidecar", "across-files", "restaged", "deleted"} {
		writeFuture(t, v2Root, slug+"/index.md", itemText(slug))
	}
	writeFuture(t, v2Root, "across-files/plan.md", "Plan\n")
	gitify(t, v2Root)
	writeFuture(t, v1Root, "roadmap/clean.md", itemText("clean"))
	gitify(t, v1Root)
	future := filepath.Join(v2Root, "docs", "future")
	assertStatusCalls := traceGitStatus(t)
	client := v2FixtureClient(t, config.ProjectRef{Root: v1Root, Name: "one"}, config.ProjectRef{Root: v2Root, Name: "two"})

	runGit(t, v2Root, "rm", "docs/future/deleted/index.md")
	index, err := client.GetProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !index.Projects[1].Dirty.Set || !index.Projects[1].Dirty.Value {
		t.Errorf("the deleted item alone should dirty the v2 project: %+v", index.Projects[1])
	}
	assertStatusCalls(v1Root+" docs/future/roadmap", v2Root+" docs/future")

	writeFuture(t, v1Root, "outside-v1.txt", "outside the v1 scope")
	for _, slug := range []string{"unstaged", "staged", "across-files", "restaged"} {
		writeFuture(t, v2Root, slug+"/index.md", itemText(slug)+"First edit\n")
	}
	runGit(t, future, "add", "staged/index.md", "across-files/index.md", "restaged/index.md")
	writeFuture(t, v2Root, "across-files/plan.md", "Edited plan\n")
	writeFuture(t, v2Root, "restaged/index.md", itemText("restaged")+"Second edit\n")
	writeFuture(t, v2Root, "sidecar/attachments/note.txt", "New sidecar\n")
	writeFuture(t, v2Root, "new/index.md", itemText("new"))
	writeFuture(t, v2Root, ".ranger/roadmap.yaml", "- staged\n")
	runGit(t, future, "add", ".ranger/roadmap.yaml")
	writeFuture(t, v2Root, ".ranger/roadmap.yaml", "- clean\n")
	writeFuture(t, v2Root, ".ranger/queue.yaml", "- new\n")

	project := readV2Project(t, client, "two")
	assertStatusCalls(v2Root + " docs/future")
	for slug, want := range map[string]api.V2GitMark{
		"clean": {}, "unstaged": {Unstaged: true}, "staged": {Staged: true},
		"sidecar": {Unstaged: true}, "new": {New: true},
		"across-files": {Staged: true, Unstaged: true}, "restaged": {Staged: true, Unstaged: true},
	} {
		if got := project.Items[slug].Git; !got.Set || got.Value != want {
			t.Errorf("git mark for %s = %+v, want %+v", slug, got, want)
		}
	}
	if got := project.Roadmap.Git; !got.Set || got.Value != (api.V2GitMark{Staged: true, Unstaged: true}) {
		t.Errorf("roadmap mark = %+v", got)
	}
	if got := project.Queue.Git; !got.Set || got.Value != (api.V2GitMark{New: true}) {
		t.Errorf("queue mark = %+v", got)
	}

	item := readV2Item(t, client, "two", "across-files")
	if item.Summary.Git != project.Items["across-files"].Git {
		t.Errorf("item and project git marks disagree: %+v", item.Summary.Git)
	}
	assertStatusCalls(v2Root + " docs/future")
	index, err = client.GetProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !index.Projects[0].Dirty.Set || index.Projects[0].Dirty.Value {
		t.Errorf("file outside v1 scope dirtied its project: %+v", index.Projects[0])
	}
	assertStatusCalls(v1Root+" docs/future/roadmap", v2Root+" docs/future")
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func traceGitStatus(t *testing.T) func(...string) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	logFile := filepath.Join(dir, "calls")
	script := `#!/bin/sh
if [ "$3" = status ]; then
    [ "$4 $5 $6 $7" = "--porcelain -z --untracked-files=all --" ] || exit 1
    printf '%s %s\n' "$2" "$8" >> "$RANGER_TEST_GIT_LOG"
fi
exec "$RANGER_TEST_GIT" "$@"
`
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RANGER_TEST_GIT", git)
	t.Setenv("RANGER_TEST_GIT_LOG", logFile)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return func(want ...string) {
		t.Helper()
		raw, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n"); !slices.Equal(got, want) {
			t.Errorf("git status calls = %q, want %q", got, want)
		}
		if err := os.Remove(logFile); err != nil {
			t.Fatal(err)
		}
	}
}
