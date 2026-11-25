package document

import (
	"errors"
	"fmt"
	"strings"

	shared "github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

func ParsePlan(raw []byte) (model.Plan, error) {
	front, _, err := splitFrontmatter(raw, false)
	if err != nil {
		return model.Plan{}, err
	}
	var plan model.Plan
	_, diagnostics := bindFields(front, &plan, []string{"mercurius", "stages"}, nil)
	if len(diagnostics) > 0 {
		return model.Plan{}, errors.New(strings.Join(diagnostics, ", "))
	}
	return plan, nil
}

func ParseStage(raw []byte) (model.Stage, error) {
	front, _, err := splitFrontmatter(raw, true)
	if err != nil {
		return model.Stage{}, err
	}
	var schema struct {
		Title    string
		Status   string
		Terminus string
		Needs    []string
		Accepted string
	}
	readable, diagnostics := bindFields(front, &schema, []string{"title", "status", "terminus", "needs", "accepted"}, nil)
	status := model.StageStatus(schema.Status)
	if readable["status"] && status != model.Started && status != model.Built {
		diagnostics = append(diagnostics, fmt.Sprintf("invalid status: %q", status))
	}
	if readable["accepted"] && !shared.ValidDate(schema.Accepted) {
		diagnostics = append(diagnostics, fmt.Sprintf("invalid accepted date: %q", schema.Accepted))
	}
	if len(diagnostics) > 0 {
		return model.Stage{}, errors.New(strings.Join(diagnostics, ", "))
	}
	return model.Stage{
		Title: schema.Title, Status: status, Terminus: schema.Terminus,
		Needs: schema.Needs, Accepted: schema.Accepted,
	}, nil
}
