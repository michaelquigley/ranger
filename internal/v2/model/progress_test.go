package model

import "testing"

func TestBuildingProgress(t *testing.T) {
	oneStage := func(verdict string) *Plan {
		return &Plan{Mercurius: verdict, Stages: []string{"one"}}
	}
	for _, tt := range []struct {
		name string
		item Item
		want Progress
	}{
		{"no design", Item{}, Progress{Position: NoDesign}},
		{"designing", Item{HasDesign: true}, Progress{Position: Designing}},
		{"empty plan", Item{Plan: &Plan{Mercurius: "ready_to_build s"}}, Progress{Position: NoStages, Verdict: Verdict{Mercurius, "ready_to_build s"}}},
		{"review", Item{Plan: oneStage("needs_changes s")}, Progress{Position: PlanInReview, Verdict: Verdict{Mercurius, "needs_changes s"}}},
		{"ready with missing file", Item{Plan: oneStage("ready_to_build s")}, Progress{Position: Ready, Verdict: Verdict{Mercurius, "ready_to_build s"}}},
		{"complete without plan verdict", Item{Plan: oneStage(""), Stages: map[string]Stage{"one": {Accepted: "2026-11-24"}}}, Progress{Position: Complete}},
		{"reordered stages", Item{
			Plan:   &Plan{Mercurius: "ready_to_build s", Stages: []string{"one", "two"}},
			Stages: map[string]Stage{"one": {Error: "bad accepted date"}, "two": {Accepted: "2026-11-24"}},
		}, Progress{Position: CurrentStage, Stage: "one", Number: 1, Total: 2}},
		{"only an unlisted stage has started", Item{Plan: oneStage(""), Stages: map[string]Stage{"extra": {Status: Built}}}, Progress{Position: PlanInReview}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.item.State = Building
			if got := BuildingProgress(tt.item); got != tt.want {
				t.Fatalf("progress = %+v, want %+v", got, tt.want)
			}
			project := ComputeProject(map[string]Item{"item": tt.item}, nil,
				Membership{Readable: true, Entries: []ViewEntry{{Slug: "item", Line: 1}}}, Membership{Readable: true})
			if flagged := len(project.Items["item"].Flags) > 0 && project.Items["item"].Flags[0].Kind == StillBuilding; flagged != (tt.want.Position == Complete) {
				t.Errorf("still building flag = %v", flagged)
			}
		})
	}
}

func TestQueueVerdictFollowsCurrentStage(t *testing.T) {
	item := Item{
		State:  Building,
		Plan:   &Plan{Mercurius: "ready_to_build session", Stages: []string{"one", "two"}},
		Stages: map[string]Stage{"one": {}, "two": {}},
	}
	queue := Membership{Readable: true, Entries: []ViewEntry{{Slug: "item", Line: 1}}}
	check := func(want Progress) {
		t.Helper()
		project := ComputeProject(map[string]Item{"item": item}, nil, Membership{Readable: true}, queue)
		if got := project.Items["item"].Progress; got != want {
			t.Errorf("queue progress = %+v, want %+v", got, want)
		}
	}
	check(Progress{Position: Ready, Verdict: Verdict{Mercurius, "ready_to_build session"}})
	item.Stages["one"] = Stage{Status: Started, Terminus: "needs_changes run"}
	check(Progress{Position: CurrentStage, Stage: "one", Number: 1, Total: 2,
		Status: Started, Verdict: Verdict{Terminus, "needs_changes run"}})
	item.Stages["one"] = Stage{Status: Built, Terminus: "clean run", Accepted: "2026-11-24"}
	check(Progress{Position: CurrentStage, Stage: "two", Number: 2, Total: 2})
}
