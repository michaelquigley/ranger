package workspace

import (
	"fmt"
	"os"

	"github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/model"
)

// every gesture here follows one discipline: load fresh, verify every
// affected guard token before the first write, compute the minimal bytes
// that express the gesture, then write — item first, order second, with any
// partial failure reported plainly. opportunistic pruning rides along only
// when order.yaml is already being written.

// Refusal is a gesture precondition failure — the gesture cannot mean
// anything against this item. not a conflict, not a server fault; the
// caller's request is the thing to change.
type Refusal struct {
	Msg string
}

func (r *Refusal) Error() string { return r.Msg }

func effectiveLane(doc *document.ItemDoc) model.State {
	if doc.State == "" {
		return model.Inbox
	}
	return doc.State
}

// rankedActive reports whether filename holds an active entry in lane.
func rankedActive(board model.Board, lanes map[model.State][]string, lane model.State, filename string) bool {
	for i, f := range lanes[lane] {
		if f == filename && i < len(board.Dispositions[lane]) && board.Dispositions[lane][i] == model.EntryActive {
			return true
		}
	}
	return false
}

// cardsWithState returns cards with one card's state overridden — the truth
// the order file is about to live against.
func cardsWithState(cards []model.CardInput, filename string, state model.State) []model.CardInput {
	out := make([]model.CardInput, len(cards))
	for i, c := range cards {
		if c.Filename == filename {
			c.State = state
		}
		out[i] = c
	}
	return out
}

// cardsWithFilename returns cards with one card renamed.
func cardsWithFilename(cards []model.CardInput, old, new string) []model.CardInput {
	out := make([]model.CardInput, len(cards))
	for i, c := range cards {
		if c.Filename == old {
			c.Filename = new
		}
		out[i] = c
	}
	return out
}

// pruneOrder reparses freshly patched order bytes, recomputes dispositions
// against the cards the write will leave behind, and applies the
// opportunistic prune.
func pruneOrder(raw []byte, cards []model.CardInput) ([]byte, error) {
	doc, err := document.ParseOrder(raw)
	if err != nil {
		return nil, fmt.Errorf("order patch produced an unreadable document: %w", err)
	}
	board := model.ComputeBoard(cards, doc.Lanes)
	return doc.Prune(board.Dispositions), nil
}

func (w *Workspace) writeOrder(orderBytes []byte, expectedVersion, context string) error {
	if err := document.CompareAndWrite(w.orderPath(), expectedVersion, orderBytes); err != nil {
		return fmt.Errorf("%s, but order.yaml was not updated: %w", context, err)
	}
	return nil
}

// Transition flips an item's state. a ranked item's old-lane entry is
// removed in the same gesture — leaving a lane always costs your place in
// it — and a position moves the entry to the destination instead:
// transition-and-place. position indexes the destination's ranked list
// only; nil lands the card unranked.
func (w *Workspace) Transition(filename string, state model.State, expectedHash, expectedOrderVersion string, position *int) error {
	snap, err := w.Load()
	if err != nil {
		return err
	}
	item, err := snap.verifyItem(w, filename, expectedHash)
	if err != nil {
		return err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return err
	}

	newItem, err := item.Doc.SetState(state)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	cards := snap.Cards()
	cardsAfter := cardsWithState(cards, filename, state)
	oldLane := effectiveLane(item.Doc)
	board := model.ComputeBoard(cards, snap.Lanes())
	ranked := snap.Order != nil && rankedActive(board, snap.Lanes(), oldLane, filename)

	var orderBytes []byte
	switch {
	case state == oldLane && position == nil:
		// a same-state transition leaves no lane: the "leaving a lane costs
		// your place" cleanup has nothing to clean, and rank must survive.
		// a placement to the same lane is a legitimate move and falls
		// through below.
	case snap.Order == nil:
		if position != nil {
			orderBytes = document.NewOrder(state, []string{filename})
		}
	case ranked || position != nil:
		raw := snap.OrderRaw
		if ranked {
			raw = snap.Order.RemoveEntry(oldLane, filename)
		}
		if position != nil {
			doc, err := document.ParseOrder(raw)
			if err != nil {
				return fmt.Errorf("order patch produced an unreadable document: %w", err)
			}
			after := model.ComputeBoard(cardsAfter, doc.Lanes)
			raw = doc.InsertEntry(state, filename, *position, after.Dispositions[state])
		}
		if orderBytes, err = pruneOrder(raw, cardsAfter); err != nil {
			return err
		}
	}

	if err := document.CompareAndWrite(w.itemPath(filename), expectedHash, newItem); err != nil {
		return err
	}
	if orderBytes != nil {
		return w.writeOrder(orderBytes, expectedOrderVersion, fmt.Sprintf("%s moved to %s", filename, state))
	}
	return nil
}

// Reorder rewrites one lane's ranked prefix. filenames is only the
// resulting ranked list — cards absent from it stay unranked.
func (w *Workspace) Reorder(lane model.State, filenames []string, expectedOrderVersion string) error {
	snap, err := w.Load()
	if err != nil {
		return err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return err
	}

	if snap.Order == nil {
		if len(filenames) == 0 {
			return nil
		}
		return document.CompareAndWrite(w.orderPath(), expectedOrderVersion, document.NewOrder(lane, filenames))
	}

	board := snap.Board()
	raw := snap.Order.RewriteLane(lane, filenames, board.Dispositions[lane])
	orderBytes, err := pruneOrder(raw, snap.Cards())
	if err != nil {
		return err
	}
	return document.CompareAndWrite(w.orderPath(), expectedOrderVersion, orderBytes)
}

// Retitle patches the title and renames to the new slug, replacing the old
// filename in every retained order occurrence, active and inert alike,
// positions preserved. a title whose slug is empty patches in place — the
// existing filename becomes the hand-picked name such titles carry, rank
// intact.
func (w *Workspace) Retitle(filename, newTitle, expectedHash, expectedOrderVersion string) (string, error) {
	snap, err := w.Load()
	if err != nil {
		return "", err
	}
	item, err := snap.verifyItem(w, filename, expectedHash)
	if err != nil {
		return "", err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return "", err
	}

	patched, err := item.Doc.SetTitle(newTitle)
	if err != nil {
		return "", fmt.Errorf("%s: %w", filename, err)
	}

	slug := model.Slug(newTitle)
	newName := slug + ".md"
	if slug == "" || newName == filename {
		if err := document.CompareAndWrite(w.itemPath(filename), expectedHash, patched); err != nil {
			return "", err
		}
		return filename, nil
	}
	return newName, w.renameItem(snap, filename, newName, patched, expectedHash, expectedOrderVersion)
}

// RenameToSlug is the one-gesture repair for the filename-mismatch flag. it
// refuses when the title is unreadable or its slug is empty — there is no
// destination to repair toward, and no flag to repair.
func (w *Workspace) RenameToSlug(filename, expectedHash, expectedOrderVersion string) (string, error) {
	snap, err := w.Load()
	if err != nil {
		return "", err
	}
	item, err := snap.verifyItem(w, filename, expectedHash)
	if err != nil {
		return "", err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return "", err
	}

	if !item.Doc.TitleOK {
		return "", &Refusal{Msg: fmt.Sprintf("%s: title is unreadable; nothing to derive a filename from", filename)}
	}
	slug := model.Slug(item.Doc.Title)
	if slug == "" {
		return "", &Refusal{Msg: fmt.Sprintf("%s: title reduces to an empty slug; the filename is hand-picked and carries no flag", filename)}
	}
	newName := slug + ".md"
	if newName == filename {
		return filename, nil
	}
	return newName, w.renameItem(snap, filename, newName, nil, expectedHash, expectedOrderVersion)
}

// renameItem is the shared rename tail: optional patched item bytes land
// first, then the no-clobber rename, then the in-place order.yaml filename
// replacement across every retained occurrence.
func (w *Workspace) renameItem(snap *Snapshot, filename, newName string, patched []byte, expectedHash, expectedOrderVersion string) error {
	oldPath, newPath := w.itemPath(filename), w.itemPath(newName)
	if _, err := os.Lstat(newPath); err == nil {
		return &document.CollisionError{Src: oldPath, Dst: newPath}
	}

	if patched != nil {
		if err := document.CompareAndWrite(oldPath, expectedHash, patched); err != nil {
			return err
		}
	}
	if err := document.FinalizeLink(oldPath, newPath); err != nil {
		if patched != nil {
			return fmt.Errorf("title updated, but the rename did not land: %w", err)
		}
		return err
	}

	occurs := false
	for _, entries := range snap.Lanes() {
		for _, f := range entries {
			if f == filename {
				occurs = true
			}
		}
	}
	if !occurs {
		return nil
	}
	raw := snap.Order.ReplaceFilename(filename, newName)
	orderBytes, err := pruneOrder(raw, cardsWithFilename(snap.Cards(), filename, newName))
	if err != nil {
		return err
	}
	return w.writeOrder(orderBytes, expectedOrderVersion, fmt.Sprintf("%s renamed to %s", filename, newName))
}

// Delete removes an item file — the operator's explicit curation gesture,
// hash-guarded like every other write (design change 2026-07-16; v1's
// tool-never-deletes rule gave the curation act to the operator, and this
// is that act with a button). the item's order.yaml entries go in the same
// gesture — a deleted file's entries are prunable unconditionally. under
// the write model a tracked file's deletion is a reviewable diff; a
// never-committed file has no net, so callers confirm before invoking.
func (w *Workspace) Delete(filename, expectedHash, expectedOrderVersion string) error {
	snap, err := w.Load()
	if err != nil {
		return err
	}
	if _, err := snap.verifyItem(w, filename, expectedHash); err != nil {
		return err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return err
	}

	var orderBytes []byte
	occurs := false
	for _, entries := range snap.Lanes() {
		for _, f := range entries {
			if f == filename {
				occurs = true
			}
		}
	}
	if occurs {
		cardsAfter := make([]model.CardInput, 0, len(snap.Cards()))
		for _, c := range snap.Cards() {
			if c.Filename != filename {
				cardsAfter = append(cardsAfter, c)
			}
		}
		if orderBytes, err = pruneOrder(snap.OrderRaw, cardsAfter); err != nil {
			return err
		}
	}

	if err := os.Remove(w.itemPath(filename)); err != nil {
		return err
	}
	if orderBytes != nil {
		return w.writeOrder(orderBytes, expectedOrderVersion, fmt.Sprintf("%s deleted", filename))
	}
	return nil
}

// SaveContent lands raw bytes verbatim — the operator's own bytes, no
// normalization. a save that changes state is a transition made through
// ranger and gets the transition's discipline, compared in effective lanes:
// departing a lane the card was actively ranked in costs that rank. entries
// are retained only when the new state is unreadable — still not positive
// evidence. renames are never a side effect of a text save; a changed title
// surfaces as the mismatch flag instead.
func (w *Workspace) SaveContent(filename string, content []byte, expectedHash, expectedOrderVersion string) error {
	snap, err := w.Load()
	if err != nil {
		return err
	}
	item, err := snap.verifyItem(w, filename, expectedHash)
	if err != nil {
		return err
	}
	if err := snap.verifyOrder(w, expectedOrderVersion); err != nil {
		return err
	}

	newDoc := document.ParseItem(content)
	oldLane := effectiveLane(item.Doc)

	var orderBytes []byte
	if snap.Order != nil && newDoc.State != "" && newDoc.State != oldLane {
		board := snap.Board()
		if rankedActive(board, snap.Lanes(), oldLane, filename) {
			raw := snap.Order.RemoveEntry(oldLane, filename)
			if orderBytes, err = pruneOrder(raw, cardsWithState(snap.Cards(), filename, newDoc.State)); err != nil {
				return err
			}
		}
	}

	if err := document.CompareAndWrite(w.itemPath(filename), expectedHash, content); err != nil {
		return err
	}
	if orderBytes != nil {
		return w.writeOrder(orderBytes, expectedOrderVersion, fmt.Sprintf("%s saved", filename))
	}
	return nil
}
