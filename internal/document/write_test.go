package document

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareAndWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.md")
	original := []byte("original\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CompareAndWrite(path, Hash(original), []byte("updated\n")); err != nil {
		t.Fatalf("matching hash refused: %v", err)
	}
	onDisk, _ := os.ReadFile(path)
	if string(onDisk) != "updated\n" {
		t.Errorf("content = %q", onDisk)
	}

	var conflict *ConflictError
	err := CompareAndWrite(path, Hash(original), []byte("stale write\n"))
	if !errors.As(err, &conflict) {
		t.Fatalf("stale hash must conflict, got %v", err)
	}
	onDisk, _ = os.ReadFile(path)
	if string(onDisk) != "updated\n" {
		t.Error("refused write must leave the file untouched")
	}
}

func TestCompareAndWriteMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gone.md")
	var conflict *ConflictError
	err := CompareAndWrite(path, Hash([]byte("x")), []byte("y"))
	if !errors.As(err, &conflict) || conflict.Actual != VersionAbsent {
		t.Errorf("missing file must conflict with absent, got %v", err)
	}
}

func TestCompareAndWriteExpectedAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "order.yaml")
	if err := CompareAndWrite(path, VersionAbsent, []byte("researching:\n")); err != nil {
		t.Fatalf("create with expected-absent: %v", err)
	}
	var conflict *ConflictError
	err := CompareAndWrite(path, VersionAbsent, []byte("racer\n"))
	if !errors.As(err, &conflict) {
		t.Errorf("racing creation must conflict, got %v", err)
	}
	onDisk, _ := os.ReadFile(path)
	if string(onDisk) != "researching:\n" {
		t.Error("first creator must win")
	}
}

func TestFinalizeLink(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, ".capture-x.md")
	dst := filepath.Join(dir, "final.md")
	if err := os.WriteFile(src, []byte("typed words\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := FinalizeLink(src, dst); err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if _, err := os.Stat(src); !errors.Is(err, os.ErrNotExist) {
		t.Error("src must be removed after finalize")
	}
	onDisk, _ := os.ReadFile(dst)
	if string(onDisk) != "typed words\n" {
		t.Errorf("dst content = %q", onDisk)
	}

	src2 := filepath.Join(dir, ".capture-y.md")
	if err := os.WriteFile(src2, []byte("second attempt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var collision *CollisionError
	err := FinalizeLink(src2, dst)
	if !errors.As(err, &collision) {
		t.Fatalf("collision must refuse, got %v", err)
	}
	if collision.Src != src2 || collision.Dst != dst {
		t.Errorf("collision paths = %+v", collision)
	}
	if _, err := os.Stat(src2); err != nil {
		t.Error("temp must survive a collision")
	}
	onDisk, _ = os.ReadFile(dst)
	if string(onDisk) != "typed words\n" {
		t.Error("collision must not clobber the destination")
	}
}
