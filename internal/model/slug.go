package model

import "strings"

// Slug derives an item's filename stem from its title. the rule is
// ASCII-mechanical over code points, exactly as the spec fixes it: map
// A-Z to a-z; keep a-z, 0-9, space, and hyphen; discard every other code
// point; convert spaces to hyphens; collapse hyphen runs; trim hyphens from
// the ends. a title that reduces to nothing returns "".
func Slug(title string) string {
	var b strings.Builder
	for _, r := range title {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '-':
			b.WriteRune('-')
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}
