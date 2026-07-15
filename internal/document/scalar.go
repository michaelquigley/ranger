package document

import (
	"strconv"
	"strings"
)

// emitScalar renders s as YAML scalar text: plain when unambiguously safe,
// double-quoted otherwise. every path that hand-emits a scalar — title
// patches, capture skeletons, order entry lines — goes through here, because
// writes never pass through a YAML encoder and each emission path is a
// chance to write malformed YAML ourselves.
func emitScalar(s string) string {
	if plainSafe(s) {
		return s
	}
	return strconv.Quote(s)
}

// plainSafe reports whether s stays a string under YAML's implicit typing
// and clear of its structural forms: it must start with an ASCII letter
// (which rules out every numeric, timestamp, and indicator-led form), carry
// no control characters, avoid the `: ` / trailing-colon / ` #` sequences
// that end a plain scalar, and not spell a YAML boolean or null. everything
// else quotes — quoting when unsure costs nothing.
func plainSafe(s string) bool {
	if s == "" || s != strings.TrimSpace(s) {
		return false
	}
	c := s[0]
	if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	if strings.Contains(s, ": ") || strings.HasSuffix(s, ":") || strings.Contains(s, " #") {
		return false
	}
	switch strings.ToLower(s) {
	case "null", "true", "false", "yes", "no", "on", "off", "y", "n":
		return false
	}
	return true
}
