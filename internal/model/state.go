// Package model holds vane's pure domain types and functions: lifecycle
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
	Done        State = "done"
	Dropped     State = "dropped"
)

// LaneOrder is the canonical lane order for board presentation.
var LaneOrder = []State{Inbox, Horizon, Researching, Building, Evaluating, Done, Dropped}

// ParseState returns the State named by s, or false for anything else.
func ParseState(s string) (State, bool) {
	switch State(s) {
	case Inbox, Horizon, Researching, Building, Evaluating, Done, Dropped:
		return State(s), true
	}
	return "", false
}
