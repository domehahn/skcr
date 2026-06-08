package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version, Commit, Date are set via -ldflags.
// With `go install`, version is read from embedded build info when available.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func init() {
	applyBuildInfo(debug.ReadBuildInfo())
}

func applyBuildInfo(info *debug.BuildInfo, ok bool) {
	if Version != "dev" {
		return
	}
	if !ok {
		return
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				Commit = s.Value[:7]
			}
		case "vcs.time":
			Date = s.Value
		}
	}
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "skcr",
		Short: "Generate agentic DevSecOps SDLC templates for multiple agent platforms",
	}

	root.AddCommand(newInitCommand())
	root.AddCommand(newListTargetsCommand())
	root.AddCommand(newBakeCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newCleanCommand())
	root.AddCommand(newScaffoldCommand())
	root.AddCommand(newVersionCommand())

	return root
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
