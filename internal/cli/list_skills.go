package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newListSkillsCommand() *cobra.Command {
	var target string
	var inTarget string
	var withTargets bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List all skills defined across bakefile targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			// skill name → sorted list of target names that contain it
			skillTargets := map[string][]string{}

			for tn, t := range cfg.Targets {
				if inTarget != "" && tn != inTarget {
					continue
				}
				for _, s := range t.Skills {
					skillTargets[s] = append(skillTargets[s], tn)
				}
			}

			if len(skillTargets) == 0 {
				if inTarget != "" {
					fmt.Printf("No skills defined in target %q.\n", inTarget)
				} else {
					fmt.Println("No skills defined in any target.")
				}
				return nil
			}

			names := make([]string, 0, len(skillTargets))
			for name := range skillTargets {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				if withTargets {
					targets := skillTargets[name]
					sort.Strings(targets)
					fmt.Printf("%-40s  %s\n", name, strings.Join(targets, ", "))
				} else {
					fmt.Println(name)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&inTarget, "in-target", "", "List skills only from this bake target")
	cmd.Flags().BoolVar(&withTargets, "with-targets", false, "Show which bake targets each skill belongs to")
	return cmd
}
