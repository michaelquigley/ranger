package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesktopDataHomeFromEnv(t *testing.T) {
	if got, err := desktopDataHomeFromEnv("/tmp/ranger-data", "/home/example"); err != nil || got != "/tmp/ranger-data" {
		t.Errorf("absolute XDG_DATA_HOME must win: got %q, err %v", got, err)
	}
	want := filepath.Join("/home/example", ".local", "share")
	if got, err := desktopDataHomeFromEnv("", "/home/example"); err != nil || got != want {
		t.Errorf("empty XDG_DATA_HOME must fall back to home: got %q, err %v", got, err)
	}
	if got, err := desktopDataHomeFromEnv("relative-data", "/home/example"); err != nil || got != want {
		t.Errorf("relative XDG_DATA_HOME must fall back to home: got %q, err %v", got, err)
	}
	if _, err := desktopDataHomeFromEnv("", ""); err == nil {
		t.Error("the fallback requires HOME")
	}
}

func TestNewDesktopInstallPaths(t *testing.T) {
	paths := newDesktopInstallPaths("/tmp/ranger-data")

	if want := filepath.Join("/tmp/ranger-data", "applications", desktopAppId+".desktop"); paths.desktopFile != want {
		t.Errorf("desktop file = %q, want %q", paths.desktopFile, want)
	}
	if len(paths.icons) != len(desktopIconSizes()) {
		t.Fatalf("icons = %d, want %d", len(paths.icons), len(desktopIconSizes()))
	}
	for _, icon := range paths.icons {
		dir := fmt.Sprintf("%dx%d", icon.size, icon.size)
		want := filepath.Join("/tmp/ranger-data", "icons", "hicolor", dir, "apps", desktopAppId+".png")
		if icon.path != want {
			t.Errorf("icon path = %q, want %q", icon.path, want)
		}
	}
}

func TestDesktopEntry(t *testing.T) {
	entry := desktopEntry(`/opt/ranger tools/ranger`)

	for _, want := range []string{
		"[Desktop Entry]\n",
		"Type=Application\n",
		"Name=ranger\n",
		`Exec="/opt/ranger tools/ranger" daemon` + "\n",
		"Icon=" + desktopAppId + "\n",
		"Terminal=false\n",
		"Categories=Utility;Development;\n",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("entry missing %q:\n%s", want, entry)
		}
	}
	// the browser owns the window, so there is no ranger window class to match.
	if strings.Contains(entry, "StartupWMClass") {
		t.Errorf("entry must not carry StartupWMClass:\n%s", entry)
	}
}

func TestDesktopExecPathEscapesDesktopEntrySpecials(t *testing.T) {
	if got := desktopExecPath(`/opt/ranger path/ranger`); got != `"/opt/ranger path/ranger"` {
		t.Errorf("got %s", got)
	}
	if got := desktopExecPath(`C:\Program Files\ranger "app"\ranger%`); got != `"C:\\Program Files\\ranger \"app\"\\ranger%%"` {
		t.Errorf("got %s", got)
	}
	if got := desktopExecPath("/tmp/$HOME/`cmd`"); got != "\"/tmp/\\$HOME/\\`cmd\\`\"" {
		t.Errorf("got %s", got)
	}
}

func TestInstallAndRemoveDesktopFiles(t *testing.T) {
	dataHome := t.TempDir()
	paths := newDesktopInstallPaths(dataHome)

	if err := installDesktopFiles(paths, "/opt/ranger"); err != nil {
		t.Fatal(err)
	}

	entry, err := os.ReadFile(paths.desktopFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(entry), `Exec="/opt/ranger" daemon`) {
		t.Errorf("entry = %s", entry)
	}

	// each installed icon is a real PNG at its directory's advertised size.
	for _, icon := range paths.icons {
		data, err := os.ReadFile(icon.path)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("%dpx icon: %v", icon.size, err)
		}
		if cfg.Width != icon.size || cfg.Height != icon.size {
			t.Errorf("%dpx icon decodes as %dx%d", icon.size, cfg.Width, cfg.Height)
		}
	}

	// the remove command deletes every installed file, and a second run
	// succeeds silently — already-gone files are skipped, not errors.
	t.Setenv("XDG_DATA_HOME", dataHome)
	for range 2 {
		remove := newDesktopRemoveCmd()
		remove.SetOut(io.Discard)
		if err := remove.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range paths.allFiles() {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("%s survives removal", file)
		}
	}
}
