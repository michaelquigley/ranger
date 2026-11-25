package model

import "github.com/michaelquigley/ranger/internal/slug"

// Slug derives an item's filename stem from its title.
func Slug(title string) string { return slug.FromTitle(title) }
