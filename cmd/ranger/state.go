package main

import (
	"fmt"
	"strings"

	"github.com/michaelquigley/ranger/internal/model"
	v2 "github.com/michaelquigley/ranger/internal/v2/workspace"
	"github.com/spf13/cobra"
)

// newStateCmd is the transition gesture from the terminal: no placement,
// the card lands unranked in its new lane like any transition-without-place.
func newStateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "state <filename|slug> <state>",
		Short: "transition an item to a new state",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			w, err := discovered()
			if err != nil {
				return err
			}
			convention, err := v2.Detect(w.Root())
			if err != nil {
				return err
			}
			if convention == v2.V2 {
				return fmt.Errorf("state is not built yet for ranger v2")
			}
			state, ok := model.ParseState(args[1])
			if !ok {
				return fmt.Errorf("unknown state %q; states are %s", args[1], laneNames())
			}
			snap, err := w.Load()
			if err != nil {
				return err
			}
			filename := args[0]
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}
			item, ok := snap.Item(filename)
			if !ok {
				return fmt.Errorf("no item named %s", filename)
			}
			if err := w.Transition(filename, state, item.Hash, snap.OrderVersion, nil); err != nil {
				return err
			}
			fmt.Printf("%s -> %s\n", filename, state)
			return nil
		},
	}
}

func laneNames() string {
	names := make([]string, len(model.LaneOrder))
	for i, lane := range model.LaneOrder {
		names[i] = string(lane)
	}
	return strings.Join(names, ", ")
}
