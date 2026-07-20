package workspace

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/model"
)

// FinalizeOutcome is one of the four explicit ends of a capture.
type FinalizeOutcome int

const (
	// Finalized landed the item under its slug.
	Finalized FinalizeOutcome = iota
	// EmptyTitle cancels the capture; the temp file is kept.
	EmptyTitle
	// EmptySlug means a non-empty title reduced to nothing; the temp file
	// is kept for a rename by hand.
	EmptySlug
	// Collision means the slug's filename already exists; the temp file is
	// kept and both paths are reported.
	Collision
)

// Finalization reports how a capture ended. TempPath is always set; the
// temp file survives every non-finalized outcome.
type Finalization struct {
	Outcome  FinalizeOutcome
	Filename string
	TempPath string
	DestPath string
}

// CreateDraft writes a capture temp file into roadmap/ — creating the
// directory on demand, because the first idea in a fresh repo is exactly
// the moment entry must cost nothing — and returns its path. the temp lives
// inside the working tree, reviewable and resumable, never outside the
// judgment gate.
func (w *Workspace) CreateDraft(title, body string) (string, error) {
	if err := os.MkdirAll(w.roadmapDir(), 0o755); err != nil {
		return "", fmt.Errorf("create roadmap directory: %w", err)
	}
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	path := filepath.Join(w.roadmapDir(), capturePrefix+hex.EncodeToString(suffix)+".md")
	content := document.NewItem(title, time.Now().Format("2006-01-02"), body)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		return "", err
	}
	return path, nil
}

// FinalizeDraft rereads the saved temp bytes — the editor sat between
// create and here and may have rewritten anything, the title included —
// recovers the title from those bytes, derives the slug, and no-clobber
// links the unchanged bytes into place.
func (w *Workspace) FinalizeDraft(tempPath string) (*Finalization, error) {
	raw, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("capture draft: %w", err)
	}
	doc := document.ParseItem(raw)
	if !doc.TitleOK || doc.Title == "" {
		return &Finalization{Outcome: EmptyTitle, TempPath: tempPath}, nil
	}
	slug := model.Slug(doc.Title)
	if slug == "" {
		return &Finalization{Outcome: EmptySlug, TempPath: tempPath}, nil
	}
	filename := slug + ".md"
	dst := w.itemPath(filename)
	if err := document.FinalizeLink(tempPath, dst); err != nil {
		var collision *document.CollisionError
		if errors.As(err, &collision) {
			return &Finalization{Outcome: Collision, TempPath: tempPath, DestPath: dst}, nil
		}
		return nil, err
	}
	return &Finalization{Outcome: Finalized, Filename: filename, TempPath: tempPath}, nil
}
