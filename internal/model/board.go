package model

import "sort"

// EntryDisposition classifies one order.yaml entry after board evaluation.
type EntryDisposition int

const (
	// EntryActive ranks its card in the lane.
	EntryActive EntryDisposition = iota
	// EntryInert is retained on disk but participates in nothing: its card
	// exists with an unreadable state, and this lane is not the card's
	// effective lane.
	EntryInert
	// EntryPrunable is discarded and may be removed on the next order
	// write; the model never decides when.
	EntryPrunable
)

// Lane is one board lane: the first RankedCount cards are the ranked prefix
// in order.yaml order, the rest the computed unranked tail.
type Lane struct {
	State       State
	Cards       []CardInput
	RankedCount int
}

// Board is the computed board: lanes in lifecycle order, plus the
// disposition of every order entry, parallel to the input lane lists.
type Board struct {
	Lanes        []Lane
	Dispositions map[State][]EntryDisposition
}

// ComputeBoard evaluates order against cards in the spec's fixed order:
// invalid entries are discarded first (nonexistent file unconditionally;
// lane mismatch only with positive evidence — the card exists and has a
// valid, different readable state), then first-occurrence-wins runs among
// the survivors, then each lane's unranked tail sorts created ascending,
// filename ascending, undated cards after every dated card. entries
// retained for an unreadable-state card are active in the card's effective
// lane (inbox) and inert everywhere else.
func ComputeBoard(cards []CardInput, order map[State][]string) Board {
	byName := make(map[string]CardInput, len(cards))
	for _, c := range cards {
		byName[c.Filename] = c
	}

	board := Board{Dispositions: make(map[State][]EntryDisposition, len(order))}
	ranked := make(map[State][]string, len(order))
	isRanked := make(map[string]bool)

	for _, lane := range LaneOrder {
		entries, ok := order[lane]
		if !ok {
			continue
		}
		dispositions := make([]EntryDisposition, len(entries))
		for i, filename := range entries {
			card, exists := byName[filename]
			switch {
			case !exists:
				dispositions[i] = EntryPrunable
			case card.State != "" && card.State != lane:
				dispositions[i] = EntryPrunable
			case card.EffectiveLane() != lane:
				dispositions[i] = EntryInert
			case isRanked[filename]:
				dispositions[i] = EntryPrunable
			default:
				dispositions[i] = EntryActive
				isRanked[filename] = true
				ranked[lane] = append(ranked[lane], filename)
			}
		}
		board.Dispositions[lane] = dispositions
	}

	for _, lane := range LaneOrder {
		l := Lane{State: lane}
		for _, filename := range ranked[lane] {
			l.Cards = append(l.Cards, byName[filename])
		}
		l.RankedCount = len(l.Cards)

		var tail []CardInput
		for _, c := range cards {
			if c.EffectiveLane() == lane && !isRanked[c.Filename] {
				tail = append(tail, c)
			}
		}
		sort.Slice(tail, func(i, j int) bool {
			a, b := tail[i], tail[j]
			if (a.Created == "") != (b.Created == "") {
				return b.Created == ""
			}
			if a.Created != b.Created {
				return a.Created < b.Created
			}
			return a.Filename < b.Filename
		})
		l.Cards = append(l.Cards, tail...)
		board.Lanes = append(board.Lanes, l)
	}
	return board
}
