package main

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/michaelquigley/ranger/internal/v2/model"
	"github.com/michaelquigley/ranger/internal/v2/workspace"
	"github.com/spf13/cobra"
)

func v2Commands() []*cobra.Command {
	return []*cobra.Command{
		newV2Command("roadmap", "show the v2 roadmap by lane", cobra.NoArgs,
			func(out io.Writer, snapshot *workspace.Snapshot, _ []string) error {
				return printView(out, snapshot, model.Roadmap)
			}),
		newV2Command("queue", "show the v2 work queue by lane", cobra.NoArgs,
			func(out io.Writer, snapshot *workspace.Snapshot, _ []string) error {
				return printView(out, snapshot, model.Queue)
			}),
		newV2Command("needs", "show requests and stages waiting for acceptance", cobra.NoArgs,
			func(out io.Writer, snapshot *workspace.Snapshot, _ []string) error {
				printNeeds(out, snapshot.Project.Needs)
				return nil
			}),
		newV2Command("show <item>", "show a v2 item by its directory name", cobra.ExactArgs(1),
			func(out io.Writer, snapshot *workspace.Snapshot, args []string) error {
				return printItem(out, snapshot, args[0])
			}),
		newV2Command("flags", "show v2 item and view flags", cobra.NoArgs,
			func(out io.Writer, snapshot *workspace.Snapshot, _ []string) error {
				printFlags(out, snapshot.Project.Flags)
				return nil
			}),
	}
}

func newV2Command(use, short string, args cobra.PositionalArgs,
	render func(io.Writer, *workspace.Snapshot, []string) error) *cobra.Command {
	return &cobra.Command{
		Use: use, Short: short, Args: args,
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := discovered()
			if err != nil {
				return err
			}
			convention, err := workspace.Detect(w.Root())
			if err != nil {
				return err
			}
			if convention != workspace.V2 {
				return fmt.Errorf("ranger %s requires a v2 project (this project uses v1)", cmd.Name())
			}
			snapshot, err := workspace.Load(w.Root())
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), snapshot, args)
		},
	}
}

func printView(out io.Writer, snapshot *workspace.Snapshot, name model.ViewName) error {
	view, lanes := snapshot.Roadmap, snapshot.Project.Roadmap
	if name == model.Queue {
		view, lanes = snapshot.Queue, snapshot.Project.Queue
	}
	if view.Err != nil {
		return view.Err
	}
	shown := false
	for _, lane := range lanes {
		if len(lane.Items) == 0 {
			continue
		}
		shown = true
		fmt.Fprintf(out, "%s:\n", lane.State)
		for _, slug := range lane.Items {
			item, summary := snapshot.Items[slug], snapshot.Project.Items[slug]
			fmt.Fprintf(out, "  - %s", v2ItemLabel(slug, item.Title))
			if item.Kind != "" {
				fmt.Fprintf(out, " [%s]", item.Kind)
			}
			if name == model.Roadmap && len(item.Tags) > 0 {
				fmt.Fprintf(out, " [%s]", strings.Join(item.Tags, ", "))
			}
			if len(summary.Needs) > 0 {
				fmt.Fprint(out, " [needs you]")
			}
			if len(summary.Flags) > 0 {
				fmt.Fprint(out, " [flagged]")
			}
			fmt.Fprintln(out)
			if name == model.Queue {
				printQueueDetails(out, item, summary)
			}
		}
	}
	if !shown {
		fmt.Fprintln(out, "no items")
	}
	return nil
}

func printQueueDetails(out io.Writer, item workspace.Item, summary *model.ItemSummary) {
	if summary.Progress.Position != "" {
		fmt.Fprintf(out, "    %s\n", progressLabel(summary.Progress))
	}
	if item.Checklist.Total > 0 {
		fmt.Fprintf(out, "    %d open of %d\n", item.Checklist.Open, item.Checklist.Total)
	}
	printTargets(out, "    requires", item.Requires)
	printTargets(out, "    blocked on", item.BlockedOn)
	for _, need := range summary.Needs {
		fmt.Fprintf(out, "    %s\n", needLabel(need))
	}
}

func printTargets(out io.Writer, label string, targets []model.Target) {
	if len(targets) == 0 {
		return
	}
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	fmt.Fprintf(out, "%s: %s\n", label, strings.Join(values, ", "))
}

func v2ItemLabel(slug, title string) string {
	if title == "" || title == slug {
		return slug
	}
	return fmt.Sprintf("%s (%s)", title, slug)
}

func progressLabel(progress model.Progress) string {
	label := string(progress.Position)
	switch progress.Position {
	case model.NoDesign:
		label = "no design yet"
	case model.PlanInReview:
		label = "plan in review"
	case model.Ready:
		label = "ready to hand off"
	case model.NoStages:
		label = "no stages listed"
	case model.Complete:
		label = "all stages accepted"
	case model.CurrentStage:
		label = fmt.Sprintf("stage %d of %d (%s)", progress.Number, progress.Total, progress.Stage)
		if progress.Status != "" {
			label += ": " + string(progress.Status)
		}
	}
	if progress.Verdict.Value != "" {
		label += fmt.Sprintf(" [%s: %s]", progress.Verdict.Broker, progress.Verdict.Value)
	}
	return label
}

func needLabel(need model.Need) string {
	request := need.Request
	if need.Acceptance {
		request = "awaiting acceptance"
	}
	if need.Stage != "" {
		return need.Stage + ": " + request
	}
	return request
}

func printNeeds(out io.Writer, needs []model.Need) {
	if len(needs) == 0 {
		fmt.Fprintln(out, "nothing needs you")
	}
	for _, need := range needs {
		fmt.Fprintf(out, "%s: %s\n", need.Item, needLabel(need))
	}
}

func printFlags(out io.Writer, flags []model.Flag) {
	if len(flags) == 0 {
		fmt.Fprintln(out, "no flags")
	}
	for _, flag := range flags {
		owner := flag.Item
		if flag.View != "" {
			owner = fmt.Sprintf(".ranger/%s.yaml:%d", flag.View, flag.Line)
		}
		fmt.Fprintf(out, "%s: %s\n", owner, flagLabel(flag))
	}
}

func flagLabel(flag model.Flag) string {
	switch flag.Kind {
	case model.StillBuilding:
		return "all stages accepted, still building"
	case model.MissingLink:
		return fmt.Sprintf("%s link to missing item %s", flag.Relation, flag.Target)
	case model.MissingViewItem:
		return "missing item " + flag.Target
	case model.InNeitherView:
		return "in neither view"
	case model.SlugMismatch:
		return "directory name differs from title slug " + flag.Target
	case model.MissingStage:
		return "missing stage " + flag.Target
	case model.UnlistedStage:
		return "stage not listed in plan: " + flag.Target
	case model.UnreadableStage:
		return fmt.Sprintf("unreadable stage %s: %s", flag.Target, flag.Diagnostic)
	case model.UnreadablePlan:
		return "unreadable plan: " + flag.Diagnostic
	default:
		return "malformed index.md: " + flag.Diagnostic
	}
}

func printItem(out io.Writer, snapshot *workspace.Snapshot, slug string) error {
	item, ok := snapshot.Items[slug]
	if !ok {
		return fmt.Errorf("no item named %s", slug)
	}
	summary := snapshot.Project.Items[slug]
	fmt.Fprintln(out, v2ItemLabel(slug, item.Title))
	fmt.Fprintf(out, "state: %s\n", item.Lane())
	for _, field := range []struct{ name, value string }{
		{"kind", string(item.Kind)}, {"created", item.Created}, {"since", item.Since},
		{"tags", strings.Join(item.Tags, ", ")}, {"subsystems", strings.Join(item.Subsystems, ", ")},
		{"source", item.Source}, {"milestone", item.Milestone},
	} {
		if field.value != "" {
			fmt.Fprintf(out, "%s: %s\n", field.name, field.value)
		}
	}
	for _, entry := range item.Log {
		fmt.Fprintf(out, "log: %s %s\n", entry.Stamp, entry.Note)
	}
	if item.Checklist.Total > 0 {
		fmt.Fprintf(out, "%d open of %d\n", item.Checklist.Open, item.Checklist.Total)
	}
	if body := strings.TrimSpace(item.Body); body != "" {
		fmt.Fprintf(out, "\n%s\n\n", body)
	}
	for _, link := range summary.LinksOut {
		fmt.Fprintf(out, "link out (%s): %s\n", link.Relation, link.Target)
	}
	for _, link := range summary.LinksIn {
		fmt.Fprintf(out, "link in (%s): %s\n", link.Relation, link.From)
	}
	if summary.Progress.Position != "" {
		fmt.Fprintf(out, "progress: %s\n", progressLabel(summary.Progress))
	}
	printPlan(out, item)
	for _, need := range summary.Needs {
		fmt.Fprintf(out, "needs: %s\n", needLabel(need))
	}
	for _, file := range item.Files {
		fmt.Fprintf(out, "file: %s\n", file)
	}
	if len(summary.Flags) > 0 {
		printFlags(out, summary.Flags)
	}
	return nil
}

func printPlan(out io.Writer, item workspace.Item) {
	var listed []string
	if item.Plan != nil {
		listed = item.Plan.Stages
		fmt.Fprintln(out, "plan:")
		if item.Plan.Mercurius != "" {
			fmt.Fprintf(out, "  mercurius: %s\n", item.Plan.Mercurius)
		}
		for i, name := range listed {
			printStage(out, fmt.Sprintf("  %d.", i+1), name, item.Stages)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(item.Stages)) {
		if !slices.Contains(listed, name) {
			printStage(out, "stage:", name, item.Stages)
		}
	}
}

func printStage(out io.Writer, label, name string, stages map[string]model.Stage) {
	stage, exists := stages[name]
	fmt.Fprintf(out, "%s %s\n", label, v2ItemLabel(name, stage.Title))
	switch {
	case !exists:
		fmt.Fprintln(out, "    missing stage")
	case stage.Error != "":
		fmt.Fprintf(out, "    unreadable stage: %s\n", stage.Error)
	default:
		for _, fact := range []struct{ name, value string }{
			{"status", string(stage.Status)}, {"terminus", stage.Terminus}, {"accepted", stage.Accepted},
		} {
			if fact.value != "" {
				fmt.Fprintf(out, "    %s: %s\n", fact.name, fact.value)
			}
		}
	}
}
