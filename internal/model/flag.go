package model

// FlagKind names a card condition the board surfaces visibly.
type FlagKind string

const (
	// FlagMalformed marks an item that fails the schema table; the
	// diagnostic carries what broke.
	FlagMalformed FlagKind = "malformed"
	// FlagFilenameMismatch marks an item whose filename is not the slug of
	// its readable title.
	FlagFilenameMismatch FlagKind = "filename-mismatch"
)

// Flag is one visible card condition.
type Flag struct {
	Kind       FlagKind
	Diagnostic string
}

// MismatchesSlug reports whether filename should carry the
// filename-mismatch flag for a readable title. a title whose slug is empty
// legitimately carries a hand-picked filename and never flags.
func MismatchesSlug(filename, title string) bool {
	s := Slug(title)
	return s != "" && filename != s+".md"
}
