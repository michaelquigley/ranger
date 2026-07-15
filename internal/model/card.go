package model

// CardInput is the classification the ordering computation consumes,
// produced upstream by the document and workspace layers.
type CardInput struct {
	Filename string
	Title    string
	// State is the card's readable state; the zero value means state:
	// couldn't be read.
	State State
	// Created is the card's valid YYYY-MM-DD date, or "" when absent or
	// unreadable. the format sorts lexically.
	Created string
	Flags   []Flag
}

// EffectiveLane is the lane the card belongs to for all ordering purposes:
// its readable state, or inbox when the state couldn't be read.
func (c CardInput) EffectiveLane() State {
	if c.State == "" {
		return Inbox
	}
	return c.State
}
