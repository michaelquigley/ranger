package document

import (
	"errors"
	"fmt"
	"os"
)

// ConflictError is a guard refusal: the disk no longer matches the version
// the caller's view was painted from. the fix is a reload, not a retry.
type ConflictError struct {
	Path     string
	Expected string
	Actual   string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s changed on disk (expected %s, found %s)", e.Path, short(e.Expected), short(e.Actual))
}

// CollisionError is a no-clobber refusal: the destination already exists.
// the in-flight content survives at Src; both paths are reported so the
// operator resolves the collision.
type CollisionError struct {
	Src string
	Dst string
}

func (e *CollisionError) Error() string {
	return fmt.Sprintf("%s already exists; content preserved at %s", e.Dst, e.Src)
}

func short(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// CompareAndWrite writes newBytes to path only if the file on disk still
// hashes to expectedHash — the hash carried by the caller from the read that
// painted the state acted on, never a fresh self-comparison. VersionAbsent
// expects no file and creates no-clobber, so a racing creator wins and we
// refuse. best-effort detection: no locks, no fsync ceremony — the git gate
// is the real net.
func CompareAndWrite(path, expectedHash string, newBytes []byte) error {
	if expectedHash == VersionAbsent {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return &ConflictError{Path: path, Expected: VersionAbsent, Actual: "present"}
			}
			return err
		}
		defer f.Close()
		_, err = f.Write(newBytes)
		return err
	}

	onDisk, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ConflictError{Path: path, Expected: expectedHash, Actual: VersionAbsent}
		}
		return err
	}
	if actual := Hash(onDisk); actual != expectedHash {
		return &ConflictError{Path: path, Expected: expectedHash, Actual: actual}
	}
	return os.WriteFile(path, newBytes, 0o644)
}

// FinalizeLink moves src to dst no-clobber: link then remove, an atomic
// refuse-if-exists on POSIX. on collision both paths are reported and src
// survives untouched.
func FinalizeLink(src, dst string) error {
	if err := os.Link(src, dst); err != nil {
		if errors.Is(err, os.ErrExist) {
			return &CollisionError{Src: src, Dst: dst}
		}
		return err
	}
	return os.Remove(src)
}
