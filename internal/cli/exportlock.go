package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type exportLockEntry struct {
	Name    string   `json:"name"`
	Version string   `json:"version,omitempty"`
	Skill   string   `json:"skill"`
	Targets []string `json:"targets"`
	Flows   []string `json:"flows,omitempty"`
}

type exportLockFile struct {
	GeneratedAt string            `json:"generated_at"`
	SourceDir   string            `json:"source_dir"`
	Skills      []exportLockEntry `json:"skills"`
}

func newExportLockCommand() *cobra.Command {
	var target string
	var out string

	cmd := &cobra.Command{
		Use:   "export-lock",
		Short: "Export a portable JSON lock summarising all skill versions across targets",
		Long: `Walks every target in agentic.bake.yaml (resolving inherits:), collects the
complete skill list, reads each skill's version from SKILL.md frontmatter, and
writes a machine-readable agentic.lock.json.

Useful for CI fingerprinting, supply-chain audits, and reproducibility checks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			skillTargets := map[string][]string{}
			skillFlows := map[string][]string{}

			for targetName := range cfg.Targets {
				tc, err := cliResolveTarget(cfg, targetName)
				if err != nil {
					return fmt.Errorf("resolve target %q: %w", targetName, err)
				}
				for _, s := range tc.Skills {
					skillTargets[s] = append(skillTargets[s], targetName)
				}
				for _, f := range tc.Flows {
					if validateSkillArg(f) != nil {
						continue
					}
					flowFile := filepath.Join(abs, sourceBase, f, "flow.yaml")
					if data, err2 := os.ReadFile(flowFile); err2 == nil {
						var fm lintFlowManifest
						if yaml.Unmarshal(data, &fm) == nil {
							for _, step := range fm.Steps {
								if step.Skill != "" && validateSkillArg(step.Skill) == nil {
									skillFlows[step.Skill] = append(skillFlows[step.Skill], f)
								}
							}
						}
					}
				}
			}

			skillSet := map[string]struct{}{}
			for s := range skillTargets {
				skillSet[s] = struct{}{}
			}
			for s := range skillFlows {
				skillSet[s] = struct{}{}
			}
			skillNames := make([]string, 0, len(skillSet))
			for s := range skillSet {
				skillNames = append(skillNames, s)
			}
			sort.Strings(skillNames)

			entries := make([]exportLockEntry, 0, len(skillNames))
			for _, s := range skillNames {
				var version string
				if data, err := cliReadFile(filepath.Join(abs, sourceBase, s, "SKILL.md")); err == nil {
					version = parseFrontmatterScalar(string(data), "version")
				}

				targets := skillTargets[s]
				if targets == nil {
					targets = []string{}
				}
				sort.Strings(targets)
				flows := deduplicateStrings(skillFlows[s])
				sort.Strings(flows)

				entry := exportLockEntry{
					Name:    s,
					Version: version,
					Skill:   filepath.Join(sourceBase, s),
					Targets: targets,
				}
				if len(flows) > 0 {
					entry.Flows = flows
				}
				entries = append(entries, entry)
			}

			lf := exportLockFile{
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				SourceDir:   sourceBase,
				Skills:      entries,
			}

			if out == "-" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(lf)
			}

			outPath := out
			if outPath == "" {
				outPath = filepath.Join(abs, "agentic.lock.json")
			}

			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create %s: %w", outPath, err)
			}

			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(lf); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("close %s: %w", outPath, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s (%d skill(s))\n", outPath, len(entries))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().StringVar(&out, "out", "", "Output path (default: agentic.lock.json); use '-' to print to stdout")
	return cmd
}
