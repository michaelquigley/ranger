package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var errSymlink = errors.New("symlink refused")

// lstatPath checks each component below the project root, stopping before
// any symlink's target. errors name the component that could not be used.
func lstatPath(root, relative string) (os.FileInfo, error) {
	path := root
	var info os.FileInfo
	for name := range strings.SplitSeq(filepath.ToSlash(relative), "/") {
		path = filepath.Join(path, name)
		var err error
		info, err = os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, &os.PathError{Op: "lstat", Path: path, Err: errSymlink}
		}
	}
	return info, nil
}

func readFile(root, relative string) ([]byte, error) {
	if _, err := lstatPath(root, relative); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, relative))
}

func readDir(root, relative string) ([]os.DirEntry, error) {
	if _, err := lstatPath(root, relative); err != nil {
		return nil, err
	}
	return os.ReadDir(filepath.Join(root, relative))
}
