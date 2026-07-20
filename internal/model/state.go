// Package model holds ranger's pure domain types and functions: lifecycle
// states, the slug rule, card flags, and the board ordering computation. no
// I/O, no bytes, no rendering, no transport.
package model

// State is one of the seven lifecycle states.
type State string

const (
	Inbox       State = "inbox"
	Horizon     State = "horizon"
	Researching State = "researching"
	Building    State = "building"
	Evaluating  State = "evaluating"
)

// LaneOrder is the canonical lane order for board presentation. the
// lifecycle ends at evaluating: a realized (or declined) item is deleted,
// its information synthesized into the project — there are no terminal
// lanes (design change 2026-07-18; v1 shipped with done and dropped).
var LaneOrder = []State{Inbox, Horizon, Researching, Building, Evaluating}

// ParseState returns the State named by s, or false for anything else.
func ParseState(s string) (State, bool) {
	switch State(s) {
	case Inbox, Horizon, Researching, Building, Evaluating:
		return State(s), true
	}
	return "", false
}
