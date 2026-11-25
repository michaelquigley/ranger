package document

import (
	shared "github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

type View struct {
	Entries []model.ViewEntry
	Flow    bool
}

func ParseView(raw []byte) (View, error) {
	entries, flow, err := shared.ParseSequence(raw)
	if err != nil {
		return View{}, err
	}
	view := View{Flow: flow}
	for _, entry := range entries {
		view.Entries = append(view.Entries, model.ViewEntry{Slug: entry.Value, Line: entry.Line})
	}
	return view, nil
}
