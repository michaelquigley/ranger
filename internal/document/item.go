package document

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/michaelquigley/df/dd"
	"git.hq.quigley.com/products/vane/internal/model"
)

var itemClaimed = map[string]bool{
	"title": true, "state": true, "created": true,
	"tags": true, "source": true, "log": true,
}

var dateShape = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// LogEntry is one dated event stamp from an item's log spine.
type LogEntry struct {
	Stamp string
	Note  string
}

// itemSchema is the dd-bind target for claimed fields; each field binds
// independently so one broken sibling never erases another.
type itemSchema struct {
	Title   string
	State   string
	Created string
	Tags    []string
	Source  string
	Log     []LogEntry
}

// ItemDoc is one parsed item file. Malformed is a verdict, not a blackout:
// each claimed field decodes and validates independently, and whichever of
// state, title, and created were individually readable survive alongside the
// flag.
type ItemDoc struct {
	// Title is the readable title; TitleOK distinguishes an empty title from
	// an unreadable one.
	Title   string
	TitleOK bool
	// State is the readable state; the zero value means state: couldn't be
	// read and the card falls to inbox.
	State model.State
	// Created is the valid YYYY-MM-DD date, or "" when absent or invalid.
	Created string
	Tags    []string
	Source  string
	Log     []LogEntry
	// Malformed reports whether the document fails the schema table;
	// Diagnostics carries what broke.
	Malformed   bool
	Diagnostics []string

	lines []string
	spans map[string]fieldSpan
}

// ParseItem parses raw as an item document. it always returns a document —
// a file that fails the schema table comes back with Malformed set and
// whatever fields were individually readable populated.
func ParseItem(raw []byte) *ItemDoc {
	d := &ItemDoc{
		lines: strings.Split(string(raw), "\n"),
		spans: map[string]fieldSpan{},
	}

	if len(d.lines) == 0 || strings.TrimRight(d.lines[0], " \t") != "---" {
		d.flaw("missing frontmatter fence")
		return d
	}
	fenceClose := -1
	for i := 1; i < len(d.lines); i++ {
		t := strings.TrimRight(d.lines[i], " \t")
		if t == "---" || t == "..." {
			fenceClose = i
			break
		}
	}
	if fenceClose == -1 {
		d.flaw("unterminated frontmatter fence")
		return d
	}

	yamlText := strings.Join(d.lines[1:fenceClose], "\n")
	entries, err := parseMapping(yamlText)
	if err != nil {
		d.flaw(fmt.Sprintf("frontmatter does not parse: %v", err))
		return d
	}

	// duplicate detection among claimed keys only: a duplicate inside
	// someone else's extension data must not flag the item.
	counts := map[string]int{}
	for _, e := range entries {
		if itemClaimed[e.name] {
			counts[e.name]++
		}
	}

	// yaml line 1 is file line index 1 (the opening fence shifts by one).
	spans := scanSpans(d.lines, entries, 0, fenceClose-1)

	schema := itemSchema{}
	readable := map[string]bool{}
	for i, e := range entries {
		if !itemClaimed[e.name] {
			continue
		}
		if counts[e.name] > 1 {
			continue
		}
		d.spans[e.name] = spans[i]
		plain, err := nodeToPlain(e.val, 0)
		if err != nil {
			d.flaw(fmt.Sprintf("%s: %v", e.name, err))
			continue
		}
		if plain == nil {
			// dd would treat a null as an absent merge value; a claimed key
			// with no value violates its declared shape instead.
			d.flaw(fmt.Sprintf("%s: null value", e.name))
			continue
		}
		if err := dd.Bind(&schema, map[string]any{e.name: plain}); err != nil {
			d.flaw(fmt.Sprintf("%s: %v", e.name, err))
			continue
		}
		readable[e.name] = true
	}
	for name, n := range counts {
		if n > 1 {
			d.flaw(fmt.Sprintf("duplicate key: %s", name))
		}
	}
	for _, name := range []string{"title", "state", "created"} {
		if counts[name] == 0 {
			d.flaw(fmt.Sprintf("missing required field: %s", name))
		}
	}

	if readable["title"] {
		d.Title = schema.Title
		d.TitleOK = true
	}
	if readable["state"] {
		if s, ok := model.ParseState(schema.State); ok {
			d.State = s
		} else {
			d.flaw(fmt.Sprintf("invalid state: %q", schema.State))
		}
	}
	if readable["created"] {
		if validDate(schema.Created) {
			d.Created = schema.Created
		} else {
			d.flaw(fmt.Sprintf("invalid created date: %q", schema.Created))
		}
	}
	d.Tags = schema.Tags
	d.Source = schema.Source
	d.Log = schema.Log
	if readable["log"] {
		for _, entry := range schema.Log {
			if !validDate(entry.Stamp) {
				d.flaw(fmt.Sprintf("invalid log stamp: %q", entry.Stamp))
			}
		}
	}
	return d
}

func (d *ItemDoc) flaw(diagnostic string) {
	d.Malformed = true
	d.Diagnostics = append(d.Diagnostics, diagnostic)
}

func validDate(s string) bool {
	if !dateShape.MatchString(s) {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// NewItem emits a fresh capture skeleton: title from the argument or bare,
// inbox state, the given created date, and the body after a blank line —
// what you edit is exactly what lands.
func NewItem(title, created, body string) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	if title == "" {
		b.WriteString("title:\n")
	} else {
		b.WriteString("title: " + emitScalar(title) + "\n")
	}
	b.WriteString("state: inbox\n")
	b.WriteString("created: " + created + "\n")
	b.WriteString("---\n\n")
	if body != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteString("\n")
		}
	}
	return []byte(b.String())
}

// SetState returns new file bytes with the state: field's line range
// replaced in place — same indentation, inline comment preserved, every
// other byte untouched.
func (d *ItemDoc) SetState(state model.State) ([]byte, error) {
	return d.patchScalar("state", string(state))
}

// SetTitle returns new file bytes with the title: field's complete line
// range replaced by a single line — a multiline scalar's continuation lines
// go with it, never stranded.
func (d *ItemDoc) SetTitle(title string) ([]byte, error) {
	return d.patchScalar("title", title)
}

func (d *ItemDoc) patchScalar(name, value string) ([]byte, error) {
	span, ok := d.spans[name]
	if !ok {
		return nil, fmt.Errorf("no patchable %s: field", name)
	}
	line := span.indent + name + ": " + emitScalar(value)
	if span.comment != "" {
		line += " " + span.comment
	}
	out := make([]string, 0, len(d.lines))
	out = append(out, d.lines[:span.start]...)
	out = append(out, line)
	out = append(out, d.lines[span.end+1:]...)
	return []byte(strings.Join(out, "\n")), nil
}
