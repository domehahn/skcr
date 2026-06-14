package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newGraphCommand() *cobra.Command {
	var repoPath string
	var dot bool

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Show the target inheritance graph",
		Long: `Renders the bake target dependency tree as an ASCII diagram or DOT source.
Useful for understanding complex multi-level inheritance chains.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			names := make([]string, 0, len(cfg.Targets))
			for n := range cfg.Targets {
				names = append(names, n)
			}
			sort.Strings(names)

			if dot {
				fmt.Fprintln(cmd.OutOrStdout(), "digraph targets {")
				fmt.Fprintln(cmd.OutOrStdout(), "  rankdir=LR;")
				fmt.Fprintln(cmd.OutOrStdout(), "  node [shape=box];")
				for _, n := range names {
					t := cfg.Targets[n]
					for _, inh := range t.Inherits {
						fmt.Fprintf(cmd.OutOrStdout(), "  %q -> %q;\n", inh, n)
					}
				}
				fmt.Fprintln(cmd.OutOrStdout(), "}")
				return nil
			}

			// Build child map: parent → []children
			children := map[string][]string{}
			hasParent := map[string]bool{}
			for _, n := range names {
				t := cfg.Targets[n]
				for _, inh := range t.Inherits {
					children[inh] = append(children[inh], n)
					hasParent[n] = true
				}
			}

			// Roots: targets with no parent.
			var roots []string
			for _, n := range names {
				if !hasParent[n] {
					roots = append(roots, n)
				}
			}

			var printTree func(name, prefix string, last bool)
			printTree = func(name, prefix string, last bool) {
				connector := "├── "
				childPrefix := prefix + "│   "
				if last {
					connector = "└── "
					childPrefix = prefix + "    "
				}
				t := cfg.Targets[name]
				desc := ""
				if t != nil && t.Description != "" {
					desc = "  " + t.Description
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s\n", prefix, connector, name, desc)
				kids := append([]string(nil), children[name]...)
				sort.Strings(kids)
				for i, kid := range kids {
					printTree(kid, childPrefix, i == len(kids)-1)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Targets")
			for i, root := range roots {
				t := cfg.Targets[root]
				desc := ""
				if t != nil && t.Description != "" {
					desc = "  " + t.Description
				}
				isLast := i == len(roots)-1
				connector := "├── "
				childPrefix := "│   "
				if isLast {
					connector = "└── "
					childPrefix = "    "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n", connector, root, desc)
				kids := append([]string(nil), children[root]...)
				sort.Strings(kids)
				for j, kid := range kids {
					printTree(kid, childPrefix, j == len(kids)-1)
				}
			}

			// Warn about cycles (targets that inherit non-existent targets).
			for _, n := range names {
				t := cfg.Targets[n]
				for _, inh := range t.Inherits {
					if _, ok := cfg.Targets[inh]; !ok {
						fmt.Fprintf(cmd.OutOrStdout(), "\nWARN: %q inherits from %q which does not exist\n", n, inh)
					}
				}
			}

			_ = strings.TrimSpace // keep import used
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dot, "dot", false, "Output Graphviz DOT format instead of ASCII tree")
	return cmd
}
