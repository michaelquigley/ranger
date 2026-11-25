package main

import (
	"fmt"

	"github.com/michaelquigley/ranger/internal/model"
	v2 "github.com/michaelquigley/ranger/internal/v2/workspace"
	"github.com/spf13/cobra"
)

// newListCmd renders the board as plain text: lanes in lifecycle order,
// ranked prefix numbered, unranked tail dashed, flags marked. a plain
// renderer over the same ComputeBoard output the UI consumes.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list roadmap items, lane-grouped in board order",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			w, err := discovered()
			if err != nil {
				return err
			}
			convention, err := v2.Detect(w.Root())
			if err != nil {
				return err
			}
			if convention == v2.V2 {
				return fmt.Errorf("use ranger roadmap or ranger queue in a v2 project")
			}
			snap, err := w.Load()
			if err != nil {
				return err
			}
			board := snap.Board()
			shown := 0
			for _, lane := range board.Lanes {
				if len(lane.Cards) == 0 {
					continue
				}
				shown += len(lane.Cards)
				fmt.Printf("%s:\n", lane.State)
				for i, card := range lane.Cards {
					marker := " -"
					if i < lane.RankedCount {
						marker = fmt.Sprintf("%2d", i+1)
					}
					fmt.Printf("  %s %s%s\n", marker, cardLabel(card), flagMarks(card))
				}
			}
			if shown == 0 {
				fmt.Println("no items")
			}
			return nil
		},
	}
}

func cardLabel(card model.CardInput) string {
	if card.Title == "" {
		return card.Filename
	}
	return fmt.Sprintf("%s (%s)", card.Title, card.Filename)
}

func flagMarks(card model.CardInput) string {
	out := ""
	for _, f := range card.Flags {
		out += fmt.Sprintf("  [%s: %s]", f.Kind, f.Diagnostic)
	}
	return out
}
