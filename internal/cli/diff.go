package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var repoPath string
	var bakeTarget string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show content diff between rendered output and files on disk",
		Long: `Renders the bake target in memory and compares each file against what is
currently on disk. Exits with code 1 if any file differs — useful in CI to
detect uncommitted drift after a template change.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			targetName := bakeTarget
			if targetName == "" {
				targetName = "default"
				if _, ok := cfg.Targets[targetName]; !ok {
					names := make([]string, 0, len(cfg.Targets))
					for n := range cfg.Targets {
						names = append(names, n)
					}
					if len(names) == 1 {
						targetName = names[0]
					} else {
						sort.Strings(names)
						return fmt.Errorf("no %q target found; specify one with --bake-target: %s", "default", strings.Join(names, ", "))
					}
				}
			}

			resolved, err := cliResolveTarget(cfg, targetName)
			if err != nil {
				return err
			}

			files, err := cliRenderFiles(cfg, resolved)
			if err != nil {
				return err
			}

			var changed []string
			for _, rendered := range files {
				diskPath := filepath.Join(absTarget, rendered.Destination)
				current, err := os.ReadFile(diskPath)
				if err != nil {
					if os.IsNotExist(err) {
						changed = append(changed, rendered.Destination)
						fmt.Fprintf(cmd.OutOrStdout(), "missing  %s\n", rendered.Destination)
					}
					continue
				}
				if string(current) != rendered.Content {
					changed = append(changed, rendered.Destination)
					diffText := unifiedDiff(string(current), rendered.Content, rendered.Destination, false)
					if strings.TrimSpace(diffText) != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", diffText)
					}
				}
			}

			if len(changed) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No drift detected — all rendered files match disk.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n%d file(s) differ. Run `skcr bake --write` to apply.\n", len(changed))
			return fmt.Errorf("drift detected")
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Bake target to render (default: default)")
	_ = cmd.RegisterFlagCompletionFunc("bake-target", completeBakeTargets)
	return cmd
}
