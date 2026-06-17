package cli

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCICommand() *cobra.Command {
	var repoPath string
	var failFast bool

	cmd := &cobra.Command{
		Use:   "ci",
		Short: "Strict CI gate: bake drift + lint + audit",
		Long: `Runs a strict set of checks intended for CI pipelines:

  1. diff    — bake drift check (exits non-zero if files on disk differ from rendered output)
  2. lint    -- content quality (fails on warnings as well as errors in CI mode)
  3. audit   — structural completeness

All steps run by default. Use --fail-fast to stop at the first failure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			type step struct {
				name string
				args []string
			}

			steps := []step{
				{"diff", []string{"diff", "--target", repoPath}},
				{"lint", []string{"lint", "--target", repoPath, "--fail-on-warn"}},
				{"audit", []string{"audit", "--target", repoPath}},
			}

			type result struct {
				name   string
				passed bool
				output string
			}

			var results []result
			anyFailed := false

			for _, s := range steps {
				var buf bytes.Buffer
				root := NewRootCommand()
				root.SetArgs(s.args)
				root.SetOut(&buf)
				root.SetErr(&buf)
				err := root.Execute()

				passed := err == nil
				if !passed {
					anyFailed = true
				}
				results = append(results, result{
					name:   s.name,
					passed: passed,
					output: strings.TrimSpace(buf.String()),
				})

				if !passed && failFast {
					break
				}
			}

			for _, r := range results {
				icon := "✓"
				label := "pass"
				if !r.passed {
					icon = "✗"
					label = "FAIL"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-8s  %s\n", icon, r.name, label)
				if !r.passed && r.output != "" {
					for _, line := range strings.Split(r.output, "\n") {
						if strings.TrimSpace(line) == "" {
							continue
						}
						fmt.Fprintf(cmd.OutOrStdout(), "         %s\n", line)
					}
				}
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if anyFailed {
				fmt.Fprintf(cmd.OutOrStdout(), "CI: FAILED\n")
				return fmt.Errorf("CI checks failed")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "CI: passed\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop at the first failing step")
	return cmd
}
