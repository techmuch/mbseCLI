// Package cli defines the mbsecli command surface.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "0.1.1"

func Execute() {
	root := &cobra.Command{
		Use:   "mbsecli",
		Short: "Live visualization and review tool for SysML v2 textual models",
		Long: "mbsecli watches .sysml files, parses them, and serves an interactive\n" +
			"web UI (object tree, dockable diagram views, element inspector) that\n" +
			"live-updates as the model changes on disk.",
	}

	root.AddCommand(newServeCmd())
	root.AddCommand(newVersionCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the mbsecli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("mbsecli " + Version)
		},
	}
}
