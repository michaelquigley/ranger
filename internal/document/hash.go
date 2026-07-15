// Package document owns everything byte-shaped: parsing and validating item
// and order.yaml documents (dd/yaml.v3 on the read side only), content
// hashes, and surgical byte-level patches. no write ever passes through a
// YAML encoder.
package document

import (
	"crypto/sha256"
	"encoding/hex"
)

// VersionAbsent is the distinguished version for a file that doesn't exist.
// absence is a version: a mutation whose expected state is absent succeeds
// only by creating the file no-clobber.
const VersionAbsent = "absent"

// Hash returns the SHA-256 of raw as lowercase hex — the guard token carried
// by board payloads, mutation requests, and CLI gestures.
func Hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
