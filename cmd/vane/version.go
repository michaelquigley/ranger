package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print build info",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			version := "dev"
			if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
			fmt.Printf("vane %s\n", version)
		},
	}
}
