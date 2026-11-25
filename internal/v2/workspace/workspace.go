// Package workspace detects the file convention and loads a fresh v2 project
// from disk.
package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelquigley/df/dd"
	shared "github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/v2/document"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

const (
	FutureRel  = "docs/future"
	RangerRel  = FutureRel + "/.ranger"
	versionRel = RangerRel + "/version.yaml"
)

type Convention int

const (
	V1 Convention = 1
	V2 Convention = 2
)

// Detect returns the convention version for a root. the v2 marker takes
// precedence over any directory named roadmap.
func Detect(root string) (Convention, error) {
	marker := filepath.Join(root, versionRel)
	info, err := lstatPath(root, RangerRel)
	if os.IsNotExist(err) || err == nil && !info.IsDir() {
		return V1, nil
	}
	if err != nil {
		return 0, err
	}
	raw, err := readFile(root, versionRel)
	if err != nil {
		return 0, err
	}
	fields, err := shared.ParseFields(raw, []string{"version"})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", marker, err)
	}
	var schema struct{ Version string }
	if len(fields) != 1 || fields[0].Err != nil || fields[0].Value == nil {
		return 0, fmt.Errorf("%s: expected version: 2", marker)
	}
	if err := dd.Bind(&schema, map[string]any{"version": fields[0].Value}); err != nil {
		return 0, fmt.Errorf("%s: %w", marker, err)
	}
	if schema.Version != "2" {
		return 0, fmt.Errorf("%s: expected version: 2, got %q", marker, schema.Version)
	}
	return V2, nil
}

type Item struct {
	model.Item
	Raw   []byte
	Hash  string
	Body  string
	Files []string
}

type ViewFile struct {
	document.View
	Raw     []byte
	Version string
	Err     error
}

func (v ViewFile) membership() model.Membership {
	return model.Membership{Entries: v.Entries, Readable: v.Err == nil}
}

type Snapshot struct {
	Items   map[string]Item
	Roadmap ViewFile
	Queue   ViewFile
	Project model.Project
}

// Load reads the project at root after Detect has returned V2. a view
// error lives on its view, and an item document error lives on its item.
func Load(root string) (*Snapshot, error) {
	if _, err := lstatPath(root, versionRel); err != nil {
		return nil, err
	}
	entries, err := readDir(root, FutureRel)
	if err != nil {
		return nil, fmt.Errorf("future directory: %w", err)
	}

	snapshot := &Snapshot{Items: map[string]Item{}}
	items := map[string]model.Item{}
	directories := map[string]bool{}
	for _, entry := range entries {
		slug := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(slug, ".") {
			continue
		}
		directories[slug] = true
		item, exists := loadItem(root, slug)
		if exists {
			snapshot.Items[slug] = item
			items[slug] = item.Item
		}
	}
	snapshot.Roadmap = loadView(root, model.Roadmap)
	snapshot.Queue = loadView(root, model.Queue)
	snapshot.Project = model.ComputeProject(items, directories, snapshot.Roadmap.membership(), snapshot.Queue.membership())
	return snapshot, nil
}

func loadItem(root, slug string) (Item, bool) {
	dir := filepath.Join(FutureRel, slug)
	raw, err := readFile(root, filepath.Join(dir, "index.md"))
	if os.IsNotExist(err) || errors.Is(err, errSymlink) {
		return Item{}, false
	}
	var item Item
	if err != nil {
		item.Diagnostics = []string{fmt.Sprintf("index.md: %v", err)}
	} else {
		item.Raw, item.Hash = raw, shared.Hash(raw)
		item.Item, item.Body = document.ParseItem(raw, slug)
	}

	design, err := lstatPath(root, filepath.Join(dir, "design.md"))
	item.HasDesign = err == nil && !design.IsDir()
	raw, err = readFile(root, filepath.Join(dir, "plan.md"))
	switch {
	case err == nil:
		plan, parseErr := document.ParsePlan(raw)
		if parseErr == nil {
			item.Plan = &plan
		} else {
			item.PlanError = parseErr.Error()
		}
	case !os.IsNotExist(err) && !errors.Is(err, errSymlink):
		item.PlanError = err.Error()
	}
	item.Stages = loadStages(root, dir)
	item.Files = itemFiles(root, dir)
	return item, true
}

func loadStages(root, dir string) map[string]model.Stage {
	dir = filepath.Join(dir, "stages")
	entries, _ := readDir(root, dir)
	stages := map[string]model.Stage{}
	for _, entry := range entries {
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".md") {
			continue
		}
		raw, err := readFile(root, filepath.Join(dir, filename))
		if errors.Is(err, errSymlink) {
			continue
		}
		var stage model.Stage
		if err == nil {
			stage, err = document.ParseStage(raw)
		}
		if err != nil {
			stage.Error = err.Error()
		}
		stages[strings.TrimSuffix(filename, ".md")] = stage
	}
	return stages
}

func itemFiles(root, dir string) []string {
	var files []string
	var visit func(string)
	visit = func(subdir string) {
		entries, _ := readDir(root, filepath.Join(dir, subdir))
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			name := entry.Name()
			relative := filepath.Join(subdir, name)
			if entry.IsDir() {
				if !strings.HasPrefix(name, ".") {
					visit(relative)
				}
			} else if relative != "index.md" {
				files = append(files, filepath.ToSlash(relative))
			}
		}
	}
	visit("")
	return files
}

func loadView(root string, name model.ViewName) ViewFile {
	relative := filepath.Join(RangerRel, string(name)+".yaml")
	raw, err := readFile(root, relative)
	switch {
	case os.IsNotExist(err):
		return ViewFile{Version: shared.VersionAbsent}
	case err != nil:
		return ViewFile{Err: err}
	}
	view, err := document.ParseView(raw)
	if err != nil {
		return ViewFile{Raw: raw, Err: fmt.Errorf("%s: %w", filepath.Join(root, relative), err)}
	}
	return ViewFile{View: view, Raw: raw, Version: shared.Hash(raw)}
}
