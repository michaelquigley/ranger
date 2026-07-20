// ranger — your roadmap lives in your repo. the root command is capture;
// subcommands read and gesture over the same convention any hand can.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/push/build"
	"github.com/spf13/cobra"
	"git.hq.quigley.com/products/ranger/internal/workspace"
)

func main() {
	dl.Init(dl.DefaultOptions().SetTrimPrefix("git.hq.quigley.com/products/"))
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:           "ranger [title words...]",
		Short:         "ranger - your roadmap lives in your repo",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose {
				dl.Init(dl.DefaultOptions().SetTrimPrefix("git.hq.quigley.com/products/").SetLevel(slog.LevelDebug))
			}
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runCapture(args)
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	cmd.AddCommand(newServeCmd(), newListCmd(), newStateCmd(), build.NewVersionCmd("ranger"))
	return cmd
}

// discovered builds a workspace over the root discovered from the working
// directory.
func discovered() (*workspace.Workspace, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root := workspace.DiscoverRoot(cwd)
	dl.Debugf("workspace root: %s", root)
	return workspace.New(root), nil
}

func runCapture(args []string) error {
	editor := os.Getenv("RANGER_EDITOR")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return fmt.Errorf("no editor configured: set RANGER_EDITOR or EDITOR")
	}

	w, err := discovered()
	if err != nil {
		return err
	}
	temp, err := w.CreateDraft(strings.Join(args, " "), "")
	if err != nil {
		return err
	}

	parts := strings.Fields(editor)
	edit := exec.Command(parts[0], append(parts[1:], temp)...)
	edit.Stdin, edit.Stdout, edit.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := edit.Run(); err != nil {
		return fmt.Errorf("editor failed (%w); draft kept at %s", err, temp)
	}

	fin, err := w.FinalizeDraft(temp)
	if err != nil {
		return err
	}
	switch fin.Outcome {
	case workspace.Finalized:
		fmt.Printf("captured %s/%s\n", workspace.RoadmapRel, fin.Filename)
	case workspace.EmptyTitle:
		fmt.Printf("capture canceled (empty title); draft kept at %s\n", fin.TempPath)
	case workspace.EmptySlug:
		fmt.Printf("title reduces to an empty slug; pick a filename and rename by hand: %s\n", fin.TempPath)
	case workspace.Collision:
		fmt.Printf("%s already exists; draft kept at %s\n", fin.DestPath, fin.TempPath)
	}
	return nil
}
