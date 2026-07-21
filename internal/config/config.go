// Package config reads the daemon's project set: a hand-edited YAML file
// naming every root ranger serves. read-only — ranger never writes its own
// config; the operator's editor is the config surface.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/ranger/internal/model"
)

// DefaultPort is the listen port when the config names none.
const DefaultPort = 4114

// ProjectRef names one served repository root. Name addresses the project
// everywhere on the wire; Root stays config-private and routes nothing.
type ProjectRef struct {
	Root string
	Name string
}

// Config is the project set after normalization: roots absolute, names
// slug-shaped and unique, Default naming a configured project, Port filled.
type Config struct {
	Projects []ProjectRef
	Default  string
	Port     int
}

// Load reads and normalizes the config at path. every violation is a load
// error — bootstrap tier, fail-loud — because a daemon quietly serving the
// wrong tree is worse than one that refuses.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if err := dd.BindYAMLFile(cfg, path); err != nil {
		return nil, err
	}
	if err := cfg.normalize(path); err != nil {
		return nil, err
	}
	return cfg, nil
}

// New builds serve's synthesized one-project config from a discovered
// root, through the same normalization the file-backed loader runs — serve
// stays a true single-project instance of the same server for every root
// the loader would accept.
func New(root string, port int) (*Config, error) {
	cfg := &Config{Projects: []ProjectRef{{Root: root}}, Port: port}
	if err := cfg.normalize("discovered root"); err != nil {
		return nil, err
	}
	return cfg, nil
}

// normalize applies the one grammar both loaders share; origin names the
// config's provenance in every error.
func (cfg *Config) normalize(origin string) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("%s: projects must name at least one root", origin)
	}
	seen := map[string]string{}
	for i := range cfg.Projects {
		p := &cfg.Projects[i]
		root, err := expandRoot(p.Root)
		if err != nil {
			return fmt.Errorf("%s: root %q: %w", origin, p.Root, err)
		}
		p.Root = root
		if p.Name == "" {
			p.Name = model.Slug(filepath.Base(p.Root))
			if p.Name == "" {
				return fmt.Errorf("%s: root %q: basename slugifies to nothing; add a name:", origin, p.Root)
			}
		} else if p.Name != model.Slug(p.Name) {
			return fmt.Errorf("%s: name %q is not slug-shaped (want %q)", origin, p.Name, model.Slug(p.Name))
		}
		if other, ok := seen[p.Name]; ok {
			return fmt.Errorf("%s: %s and %s share the name %q; add a name: to one", origin, other, p.Root, p.Name)
		}
		seen[p.Name] = p.Root
	}
	if cfg.Default == "" {
		cfg.Default = cfg.Projects[0].Name
	} else if _, ok := seen[cfg.Default]; !ok {
		return fmt.Errorf("%s: default %q names no configured project", origin, cfg.Default)
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultPort
	}
	return nil
}

// expandRoot expands a leading ~ and requires the result absolute after
// cleaning: a tray daemon's working directory is whatever the desktop
// session felt like, so a relative root is the wrong tree waiting to be
// selected silently.
func expandRoot(root string) (string, error) {
	if root == "~" || strings.HasPrefix(root, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, root[1:])
	}
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("not absolute; the daemon has no working directory to resolve against")
	}
	return root, nil
}
