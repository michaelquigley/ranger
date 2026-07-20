package document

import (
	"fmt"
	"strings"

	"github.com/michaelquigley/df/dd"
	"gopkg.in/yaml.v3"
	"github.com/michaelquigley/ranger/internal/model"
)

// laneSchema is the dd-bind target for one recognized lane's shape: a list
// of filename strings.
type laneSchema struct {
	Entries []string
}

// OrderDoc is one parsed order.yaml. parsing failures here are
// repository-level — an unreadable order.yaml is a fail-fast, never a
// flagged card — so ParseOrder returns an error instead of a verdict.
type OrderDoc struct {
	// Lanes holds each recognized lane's entries in file order.
	Lanes map[model.State][]string

	lines   []string
	lane    map[model.State]*laneInfo
	unknown []unknownBlock
}

type laneInfo struct {
	span          fieldSpan
	flow          bool
	entryLines    []int
	entryComments []string
}

// unknownBlock is a top-level key that isn't a state name: ignored by
// ordering, carried in the line map, and prunable on the next order write —
// unless a surviving recognized node aliases an anchor defined inside it,
// where pruning the definition would leave our own write unreadable.
type unknownBlock struct {
	span     fieldSpan
	prunable bool
}

// ParseOrder parses raw as an order document. a syntax failure or a
// duplicate recognized lane key makes the document unreadable — YAML loaders
// disagree about which mapping survives, so guessing is coin-flipping.
func ParseOrder(raw []byte) (*OrderDoc, error) {
	d := &OrderDoc{
		Lanes: map[model.State][]string{},
		lines: strings.Split(string(raw), "\n"),
		lane:  map[model.State]*laneInfo{},
	}

	entries, err := parseMapping(string(raw))
	if err != nil {
		return nil, fmt.Errorf("order document does not parse: %w", err)
	}

	seen := map[model.State]bool{}
	spans := scanSpans(d.lines, entries, -1, len(d.lines)-1)

	// anchors defined anywhere in unknown territory, for the prune guard
	unknownAnchors := map[*yaml.Node]bool{}

	for i, e := range entries {
		lane, recognized := model.ParseState(e.name)
		if !recognized {
			collectAnchors(e.key, unknownAnchors)
			collectAnchors(e.val, unknownAnchors)
			d.unknown = append(d.unknown, unknownBlock{span: spans[i], prunable: true})
			continue
		}
		if seen[lane] {
			return nil, fmt.Errorf("duplicate lane key: %s", lane)
		}
		seen[lane] = true

		info := &laneInfo{span: spans[i], flow: e.val.Style == yaml.FlowStyle}
		val := e.val
		if val.Kind == yaml.AliasNode {
			// a lane whose whole value is an alias reads normally; any op
			// on it materializes the lane as its own block.
			val = val.Alias
			info.flow = true
		}
		switch {
		case val.Kind == yaml.ScalarNode && val.Tag == "!!null":
			// an empty lane
		case val.Kind == yaml.SequenceNode:
			plain, err := nodeToPlain(val, 0)
			if err != nil {
				return nil, fmt.Errorf("lane %s: %w", lane, err)
			}
			var schema laneSchema
			if err := dd.Bind(&schema, map[string]any{"entries": plain}); err != nil {
				return nil, fmt.Errorf("lane %s: %w", lane, err)
			}
			d.Lanes[lane] = schema.Entries
			for j, item := range val.Content {
				line := item.Line - 1
				if !info.flow {
					// the convention's entries are single lines; an entry
					// spanning more would strand bytes under every
					// line-targeted op, so it fails the repository tier.
					end := info.span.end
					if j+1 < len(val.Content) {
						end = val.Content[j+1].Line - 2
					}
					if trimSpanEnd(d.lines, line, end, -1) != line {
						return nil, fmt.Errorf("lane %s: entry %q spans multiple lines", lane, schema.Entries[j])
					}
				}
				info.entryLines = append(info.entryLines, line)
				info.entryComments = append(info.entryComments, item.LineComment)
			}
		default:
			return nil, fmt.Errorf("lane %s is not a list", lane)
		}
		d.lane[lane] = info
	}

	// one precondition guards the unknown-block prune: never orphan an
	// anchor a surviving recognized node aliases. the guard is deliberately
	// coarse — any recognized alias into unknown territory keeps every
	// unknown block this write — because unknown blocks may alias each
	// other: either all of them go (their cross-references die together) or
	// none do, so a prune can never leave a dangling alias.
	for _, e := range entries {
		if _, recognized := model.ParseState(e.name); !recognized {
			continue
		}
		if aliasesInto(e.val, unknownAnchors) {
			for i := range d.unknown {
				d.unknown[i].prunable = false
			}
			break
		}
	}
	return d, nil
}

// Version returns the document's guard token for the given raw bytes.
func Version(raw []byte) string {
	return Hash(raw)
}

// edit is a batch of line-level changes rendered in one pass.
type edit struct {
	deletes map[int]bool
	inserts map[int][]string
	tail    []string
}

func newEdit() *edit {
	return &edit{deletes: map[int]bool{}, inserts: map[int][]string{}}
}

func (e *edit) replaceSpan(span fieldSpan, with []string) {
	for i := span.start; i <= span.end; i++ {
		e.deletes[i] = true
	}
	e.inserts[span.start] = with
}

func (d *OrderDoc) render(e *edit) []byte {
	out := make([]string, 0, len(d.lines))
	for i, line := range d.lines {
		if ins, ok := e.inserts[i]; ok {
			out = append(out, ins...)
		}
		if !e.deletes[i] {
			out = append(out, line)
		}
	}
	if len(e.tail) > 0 {
		if len(out) > 0 && out[len(out)-1] == "" {
			out = append(out[:len(out)-1], e.tail...)
			out = append(out, "")
		} else {
			out = append(out, e.tail...)
			out = append(out, "")
		}
	}
	return []byte(strings.Join(out, "\n"))
}

// laneBlockLines emits a lane as block form: key line plus one entry line
// per filename.
func laneBlockLines(lane model.State, filenames []string, keyComment string) []string {
	key := string(lane) + ":"
	if keyComment != "" {
		key += " " + keyComment
	}
	out := []string{key}
	for _, f := range filenames {
		out = append(out, "  - "+emitScalar(f))
	}
	return out
}

func (d *OrderDoc) entryIndent(info *laneInfo) string {
	if len(info.entryLines) > 0 && !info.flow {
		return leadingSpace(d.lines[info.entryLines[0]])
	}
	return "  "
}

// dispAt treats entries beyond the disposition slice as active.
func dispAt(disp []model.EntryDisposition, i int) model.EntryDisposition {
	if i < len(disp) {
		return disp[i]
	}
	return model.EntryActive
}

// Prune returns new bytes with every entry classified prunable removed,
// along with prunable unknown blocks. it is the one sanctioned multi-line
// side effect, applied only when the file is already being written.
func (d *OrderDoc) Prune(dispositions map[model.State][]model.EntryDisposition) []byte {
	e := newEdit()
	for lane, info := range d.lane {
		disp := dispositions[lane]
		if info.flow {
			kept := make([]string, 0, len(d.Lanes[lane]))
			pruned := false
			for i, f := range d.Lanes[lane] {
				if dispAt(disp, i) == model.EntryPrunable {
					pruned = true
					continue
				}
				kept = append(kept, f)
			}
			if pruned {
				e.replaceSpan(info.span, laneBlockLines(lane, kept, info.span.comment))
			}
			continue
		}
		for i, line := range info.entryLines {
			if dispAt(disp, i) == model.EntryPrunable {
				e.deletes[line] = true
			}
		}
	}
	for _, ub := range d.unknown {
		if ub.prunable {
			e.replaceSpan(ub.span, nil)
		}
	}
	return d.render(e)
}

// RewriteLane returns new bytes with one lane's active entry lines replaced
// by filenames, in order. entries classified inert stay in place — they are
// retained references, not part of the ranked list. a missing lane block is
// created at the end of the file.
func (d *OrderDoc) RewriteLane(lane model.State, filenames []string, dispositions []model.EntryDisposition) []byte {
	info, ok := d.lane[lane]
	if !ok {
		e := newEdit()
		e.tail = laneBlockLines(lane, filenames, "")
		return d.render(e)
	}

	if info.flow {
		final := append([]string{}, filenames...)
		for i, f := range d.Lanes[lane] {
			if dispAt(dispositions, i) == model.EntryInert && !contains(final, f) {
				final = append(final, f)
			}
		}
		e := newEdit()
		e.replaceSpan(info.span, laneBlockLines(lane, final, info.span.comment))
		return d.render(e)
	}

	e := newEdit()
	anchor := -1
	original := map[string]int{}
	for i, line := range info.entryLines {
		if dispAt(dispositions, i) == model.EntryInert {
			continue
		}
		if _, ok := original[d.Lanes[lane][i]]; !ok {
			original[d.Lanes[lane][i]] = line
		}
		e.deletes[line] = true
		if anchor == -1 {
			anchor = line
		}
	}
	if anchor == -1 {
		if n := len(info.entryLines); n > 0 {
			anchor = info.entryLines[n-1] + 1
		} else {
			anchor = info.span.start + 1
		}
	}
	// a reorder moves lines: an entry the lane already holds keeps its
	// original bytes — inline comment, spacing, quoting — and only a
	// genuinely new entry is emitted fresh.
	indent := d.entryIndent(info)
	lines := make([]string, 0, len(filenames))
	for _, f := range filenames {
		if li, ok := original[f]; ok {
			lines = append(lines, d.lines[li])
			continue
		}
		lines = append(lines, indent+"- "+emitScalar(f))
	}
	e.inserts[anchor] = lines
	return d.render(e)
}

// RemoveEntry returns new bytes with the first entry for filename in lane
// removed — the entry that holds the card's rank. a missing entry is a
// no-op.
func (d *OrderDoc) RemoveEntry(lane model.State, filename string) []byte {
	info, ok := d.lane[lane]
	if !ok {
		return d.render(newEdit())
	}
	for i, f := range d.Lanes[lane] {
		if f != filename {
			continue
		}
		if info.flow {
			kept := append(append([]string{}, d.Lanes[lane][:i]...), d.Lanes[lane][i+1:]...)
			e := newEdit()
			e.replaceSpan(info.span, laneBlockLines(lane, kept, info.span.comment))
			return d.render(e)
		}
		e := newEdit()
		e.deletes[info.entryLines[i]] = true
		return d.render(e)
	}
	return d.render(newEdit())
}

// InsertEntry returns new bytes with filename inserted at position among the
// lane's active entries — position indexes the ranked list only, never
// inert lines. a missing lane block is created at the end of the file.
func (d *OrderDoc) InsertEntry(lane model.State, filename string, position int, dispositions []model.EntryDisposition) []byte {
	info, ok := d.lane[lane]
	if !ok {
		e := newEdit()
		e.tail = laneBlockLines(lane, []string{filename}, "")
		return d.render(e)
	}

	if info.flow {
		final := make([]string, 0, len(d.Lanes[lane])+1)
		active := 0
		inserted := false
		for i, f := range d.Lanes[lane] {
			if dispAt(dispositions, i) == model.EntryActive {
				if active == position {
					final = append(final, filename)
					inserted = true
				}
				active++
			}
			final = append(final, f)
		}
		if !inserted {
			final = append(final, filename)
		}
		e := newEdit()
		e.replaceSpan(info.span, laneBlockLines(lane, final, info.span.comment))
		return d.render(e)
	}

	anchor := -1
	active := 0
	for i, line := range info.entryLines {
		if dispAt(dispositions, i) != model.EntryActive {
			continue
		}
		if active == position {
			anchor = line
			break
		}
		active++
	}
	if anchor == -1 {
		if n := len(info.entryLines); n > 0 {
			anchor = info.entryLines[n-1] + 1
		} else {
			anchor = info.span.start + 1
		}
	}
	e := newEdit()
	e.inserts[anchor] = []string{d.entryIndent(info) + "- " + emitScalar(filename)}
	return d.render(e)
}

// ReplaceFilename returns new bytes with old replaced by new in every
// retained occurrence across all recognized lanes, active and inert alike,
// positions preserved — a rename that missed the inert ones would leave
// them pointing at a nonexistent file.
func (d *OrderDoc) ReplaceFilename(old, new string) []byte {
	e := newEdit()
	for lane, info := range d.lane {
		if info.flow {
			if !contains(d.Lanes[lane], old) {
				continue
			}
			final := make([]string, len(d.Lanes[lane]))
			for i, f := range d.Lanes[lane] {
				if f == old {
					f = new
				}
				final[i] = f
			}
			e.replaceSpan(info.span, laneBlockLines(lane, final, info.span.comment))
			continue
		}
		for i, f := range d.Lanes[lane] {
			if f != old {
				continue
			}
			line := info.entryLines[i]
			rebuilt := leadingSpace(d.lines[line]) + "- " + emitScalar(new)
			if c := info.entryComments[i]; c != "" {
				rebuilt += " " + c
			}
			e.replaceSpan(fieldSpan{start: line, end: line}, []string{rebuilt})
		}
	}
	return d.render(e)
}

// NewOrder emits a fresh order document ranking filenames in one lane — the
// first-ever ranking, when no order.yaml exists.
func NewOrder(lane model.State, filenames []string) []byte {
	return []byte(strings.Join(laneBlockLines(lane, filenames, ""), "\n") + "\n")
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
