// Package workspace composes document operations into the spec's gestures
// against a discovered repository root. stateless: every Load is a fresh
// read, because files are the truth and other writers never stop.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.hq.quigley.com/products/ranger/internal/document"
	"git.hq.quigley.com/products/ranger/internal/model"
)

// RoadmapRel is the roadmap directory, relative to the root.
const RoadmapRel = "docs/future/roadmap"

// capturePrefix marks in-flight capture temp files; enumeration skips them.
const capturePrefix = ".capture-"

// Workspace is a discovered root and nothing more — no snapshot lives here.
type Workspace struct {
	root string
}

// New returns a workspace over root.
func New(root string) *Workspace {
	return &Workspace{root: root}
}

// Root returns the workspace's repository root.
func (w *Workspace) Root() string { return w.root }

func (w *Workspace) roadmapDir() string {
	return filepath.Join(w.root, filepath.FromSlash(RoadmapRel))
}

func (w *Workspace) itemPath(filename string) string {
	return filepath.Join(w.roadmapDir(), filename)
}

func (w *Workspace) orderPath() string {
	return filepath.Join(w.roadmapDir(), "order.yaml")
}

// DiscoverRoot finds the repository root by a single upward walk from
// startDir: at each ancestor, a docs/future/roadmap/ directory claims the
// root; failing that, any entry named .git — file or directory, never
// opened — claims it and walls the walk, because a repo boundary is a wall,
// not a waypoint. exhaustion falls back to startDir.
func DiscoverRoot(startDir string) string {
	dir := startDir
	for {
		if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(RoadmapRel))); err == nil && info.IsDir() {
			return dir
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return startDir
		}
		dir = parent
	}
}

// Item is one enumerated roadmap file: raw bytes, guard hash, parsed doc.
type Item struct {
	Filename string
	Raw      []byte
	Hash     string
	Doc      *document.ItemDoc
}

// Snapshot is one fresh read of the workspace.
type Snapshot struct {
	Items []Item
	// Order is nil when no order.yaml exists; OrderVersion is its guard
	// token, the absent sentinel included.
	Order        *document.OrderDoc
	OrderRaw     []byte
	OrderVersion string

	byName map[string]*Item
}

// Load enumerates roadmap/*.md (flat, skipping capture temps and
// directories) and reads order.yaml. a missing or unreadable roadmap
// directory and an unreadable order.yaml are repository-level errors; any
// single bad item degrades to a flagged card instead.
func (w *Workspace) Load() (*Snapshot, error) {
	entries, err := os.ReadDir(w.roadmapDir())
	if err != nil {
		return nil, fmt.Errorf("roadmap directory: %w", err)
	}

	s := &Snapshot{byName: map[string]*Item{}, OrderVersion: document.VersionAbsent}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, capturePrefix) {
			continue
		}
		raw, err := os.ReadFile(w.itemPath(name))
		if err != nil {
			return nil, fmt.Errorf("item %s: %w", name, err)
		}
		s.Items = append(s.Items, Item{
			Filename: name,
			Raw:      raw,
			Hash:     document.Hash(raw),
			Doc:      document.ParseItem(raw),
		})
	}
	sort.Slice(s.Items, func(i, j int) bool { return s.Items[i].Filename < s.Items[j].Filename })
	for i := range s.Items {
		s.byName[s.Items[i].Filename] = &s.Items[i]
	}

	orderRaw, err := os.ReadFile(w.orderPath())
	switch {
	case err == nil:
		doc, err := document.ParseOrder(orderRaw)
		if err != nil {
			return nil, fmt.Errorf("order.yaml: %w", err)
		}
		s.Order = doc
		s.OrderRaw = orderRaw
		s.OrderVersion = document.Hash(orderRaw)
	case os.IsNotExist(err):
		// absence is a version
	default:
		return nil, fmt.Errorf("order.yaml: %w", err)
	}
	return s, nil
}

// Item returns the named item from the snapshot.
func (s *Snapshot) Item(filename string) (*Item, bool) {
	it, ok := s.byName[filename]
	return it, ok
}

// Cards classifies every item for the ordering computation, flags included.
func (s *Snapshot) Cards() []model.CardInput {
	cards := make([]model.CardInput, 0, len(s.Items))
	for _, it := range s.Items {
		var flags []model.Flag
		if it.Doc.Malformed {
			flags = append(flags, model.Flag{
				Kind:       model.FlagMalformed,
				Diagnostic: strings.Join(it.Doc.Diagnostics, "; "),
			})
		}
		if it.Doc.TitleOK && model.MismatchesSlug(it.Filename, it.Doc.Title) {
			flags = append(flags, model.Flag{
				Kind:       model.FlagFilenameMismatch,
				Diagnostic: fmt.Sprintf("filename is not slug(%q)", it.Doc.Title),
			})
		}
		cards = append(cards, model.CardInput{
			Filename: it.Filename,
			Title:    it.Doc.Title,
			State:    it.Doc.State,
			Created:  it.Doc.Created,
			Flags:    flags,
		})
	}
	return cards
}

// Lanes returns the order document's lanes, nil-safe.
func (s *Snapshot) Lanes() map[model.State][]string {
	if s.Order == nil {
		return nil
	}
	return s.Order.Lanes
}

// Board computes the board from this snapshot.
func (s *Snapshot) Board() model.Board {
	return model.ComputeBoard(s.Cards(), s.Lanes())
}

// verifyItem returns the named item after checking the caller's guard hash
// against this fresh read.
func (s *Snapshot) verifyItem(w *Workspace, filename, expectedHash string) (*Item, error) {
	it, ok := s.byName[filename]
	if !ok {
		return nil, &document.ConflictError{Path: w.itemPath(filename), Expected: expectedHash, Actual: document.VersionAbsent}
	}
	if it.Hash != expectedHash {
		return nil, &document.ConflictError{Path: w.itemPath(filename), Expected: expectedHash, Actual: it.Hash}
	}
	return it, nil
}

// verifyOrder checks the caller's order version against this fresh read.
func (s *Snapshot) verifyOrder(w *Workspace, expectedVersion string) error {
	if s.OrderVersion != expectedVersion {
		return &document.ConflictError{Path: w.orderPath(), Expected: expectedVersion, Actual: s.OrderVersion}
	}
	return nil
}
