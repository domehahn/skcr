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
	root.AddCommand(newUpdateCommand())
	root.AddCommand(newAddCommand())
	root.AddCommand(newRemoveCommand())
	root.AddCommand(newRenameCommand())
	root.AddCommand(newListCommand())
	root.AddCommand(newListTargetsCommand()) // kept for backward compatibility
	root.AddCommand(newBakeCommand())
	root.AddCommand(newSyncCommand())
	root.AddCommand(newStatusCommand())
	root.AddCommand(newDoctorCommand())
	root.AddCommand(newExportCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newCleanCommand())
	root.AddCommand(newScaffoldCommand())
	root.AddCommand(newInspectCommand())
	root.AddCommand(newDiffCommand())
	root.AddCommand(newGraphCommand())
	root.AddCommand(newMigrateCommand())
	root.AddCommand(newAuditCommand())
	root.AddCommand(newCloneCommand())
	root.AddCommand(newFmtCommand())
	root.AddCommand(newSearchCommand())
	root.AddCommand(newReportCommand())
	root.AddCommand(newWatchCommand())
	root.AddCommand(newStatsCommand())
	root.AddCommand(newLintCommand())
	root.AddCommand(newHistoryCommand())
	root.AddCommand(newReleaseCommand())
	root.AddCommand(newCheckCommand())
	root.AddCommand(newCICommand())
	root.AddCommand(newHooksCommand())
	root.AddCommand(newSnapshotCommand())
	root.AddCommand(newVersionCommand())
	root.AddCommand(newCompatibilityCommand())
	root.AddCommand(newCompletionCommand())
	root.AddCommand(newPublishCommand())
	root.AddCommand(newCompareCommand())
	root.AddCommand(newTestCommand())
	root.AddCommand(newArchiveCommand())
	root.AddCommand(newRevertCommand())
	root.AddCommand(newTagCommand())
	root.AddCommand(newChangelogCommand())
	root.AddCommand(newCoverageCommand())
	root.AddCommand(newShowCommand())
	root.AddCommand(newBundleCommand())
	root.AddCommand(newRenameTargetToplevelCommand())
	root.AddCommand(newImportCommand())
	root.AddCommand(newPinCommand())
	root.AddCommand(newFlowCommand())
	root.AddCommand(newDepsCommand())
	root.AddCommand(newUpgradeSkillCommand())
	root.AddCommand(newTouchCommand())
	root.AddCommand(newSkillVersionsCommand())
	root.AddCommand(newTargetSkillsCommand())
	root.AddCommand(newBatchTouchCommand())
	root.AddCommand(newSkillDiffCommand())
	root.AddCommand(newExportLockCommand())
	root.AddCommand(newCheckCompatCommand())
	root.AddCommand(newBulkRenameCommand())
	root.AddCommand(newSetMetadataCommand())
	root.AddCommand(newOrphansCommand())
	root.AddCommand(newSkillSizeCommand())
	root.AddCommand(newFilterByCommand())
	root.AddCommand(newCopySkillCommand())
	root.AddCommand(newContractCommand())
	root.AddCommand(newCompileCommand())
	root.AddCommand(newASPSCommand())

	return root
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
