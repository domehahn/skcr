package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/domehahn/skcr/internal/models"
	"github.com/spf13/cobra"
)

func newBundleCommand() *cobra.Command {
	var target string
	var out string
	var bakeTarget string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Pack all skills for a bakefile target into a distributable tarball",
		Long: `Reads the bakefile, collects all skill directories under skill_sources.output_dir,
and writes a compressed tarball suitable for offline or air-gapped deployment.

Each skill is archived as <skill-name>/SKILL.md (and any sibling files).
Use --bake-target to restrict to a single bakefile target's skill list.`,
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

			// Collect skills to bundle.
			skills, err := bundleCollectSkills(cfg, sourceBase, bakeTarget)
			if err != nil {
				return err
			}
			if len(skills) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No skills found to bundle.")
				return nil
			}

			outPath := out
			if outPath == "" {
				outName := "skills-bundle"
				if bakeTarget != "" {
					outName = bakeTarget + "-skills"
				}
				outPath = filepath.Join(abs, outName+".tar.gz")
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would write bundle to: %s\n", outPath)
				for _, s := range skills {
					fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", s)
				}
				return nil
			}

			created, err := bundleSkills(abs, sourceBase, skills, outPath)
			if err != nil {
				return fmt.Errorf("write bundle: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Bundled %d skill(s) → %s\n", created, outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().StringVar(&out, "out", "", "Output tarball path (default: <target>/<bake-target>-skills.tar.gz)")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Restrict to skills listed in a specific bakefile target")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be bundled without writing")
	return cmd
}

// bundleCollectSkills returns the skill names to include. If bakeTarget is set, only
// skills listed under that target are included; otherwise all skills from skill_sources are used.
// When skill_sources.skills is empty the filesystem is walked to discover skill directories.
func bundleCollectSkills(cfg *models.BakeConfig, sourceBase, bakeTarget string) ([]string, error) {
	if bakeTarget != "" {
		tc, ok := cfg.Targets[bakeTarget]
		if !ok {
			return nil, fmt.Errorf("target %q not found in bakefile", bakeTarget)
		}
		return tc.Skills, nil
	}
	if cfg.SkillSources != nil && len(cfg.SkillSources.Skills) > 0 {
		names := make([]string, 0, len(cfg.SkillSources.Skills))
		for _, sd := range cfg.SkillSources.Skills {
			names = append(names, sd.Name)
		}
		return names, nil
	}
	// Collect unique skill names from all targets as fallback.
	seen := map[string]struct{}{}
	var names []string
	for _, tc := range cfg.Targets {
		for _, s := range tc.Skills {
			if _, dup := seen[s]; !dup {
				seen[s] = struct{}{}
				names = append(names, s)
			}
		}
	}
	return names, nil
}

// bundleSkills creates a tar.gz archive of the given skill directories.
func bundleSkills(root, sourceBase string, skills []string, outPath string) (int, error) {
	f, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	added := 0
	for _, name := range skills {
		skillDir := filepath.Join(root, sourceBase, name)
		info, err := os.Stat(skillDir)
		if err != nil || !info.IsDir() {
			continue
		}
		if err := filepath.Walk(skillDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(filepath.Join(root, sourceBase), path)
			rel = filepath.ToSlash(rel)
			hdr := &tar.Header{
				Name:    rel,
				Mode:    int64(fi.Mode()),
				Size:    fi.Size(),
				ModTime: fi.ModTime(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()
			_, err = io.Copy(tw, src)
			return err
		}); err != nil {
			return added, fmt.Errorf("walk %s: %w", name, err)
		}
		added++
	}
	return added, nil
}
