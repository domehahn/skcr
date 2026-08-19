package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/bake"
	"github.com/domehahn/skcr/v2/internal/catalog"
	"github.com/domehahn/skcr/v2/internal/models"
	"github.com/domehahn/skcr/v2/internal/scaffold"
	"github.com/domehahn/sklib/spec"
	"github.com/spf13/cobra"
)

// sdlcSkillSet is a O(1) lookup set built once from the canonical SDLC skill list.
var sdlcSkillSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(scaffold.SDLCSkillNames))
	for _, n := range scaffold.SDLCSkillNames {
		m[n] = struct{}{}
	}
	return m
}()

func isSDLCSkill(name string) bool {
	_, ok := sdlcSkillSet[name]
	return ok
}

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add resources to the bakefile",
	}
	cmd.AddCommand(newAddSkillCommand())
	cmd.AddCommand(newAddTargetCommand())
	return cmd
}

func newAddTargetCommand() *cobra.Command {
	var repoPath string
	var description string
	var platforms []string
	var inherits []string

	cmd := &cobra.Command{
		Use:   "target <name>",
		Short: "Add a new target to the bakefile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			if _, exists := cfg.Targets[name]; exists {
				return fmt.Errorf("target %q already exists in bakefile", name)
			}
			if cfg.Targets == nil {
				cfg.Targets = map[string]*models.TargetConfig{}
			}
			t := &models.TargetConfig{
				Description: description,
				Platforms:   platforms,
				Inherits:    inherits,
			}
			cfg.Targets[name] = t
			if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
				return err
			}
			fmt.Printf("Added target %q to agentic.bake.yaml\n", name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&description, "description", "", "Target description")
	cmd.Flags().StringArrayVar(&platforms, "platform", nil, "Platform(s) for this target; may be repeated")
	cmd.Flags().StringArrayVar(&inherits, "inherits", nil, "Target(s) this target inherits from; may be repeated")
	return cmd
}

func newAddSkillCommand() *cobra.Command {
	var target string
	var inTargets []string
	var noScaffold bool
	var description string
	var owner string
	var initialVersion string
	var platforms []string
	var fromPath string

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Add a skill to bakefile targets and scaffold its directory structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := spec.ValidateSkillName(name); err != nil {
				return fmt.Errorf("invalid skill name %q: use lowercase letters, digits, and hyphens", name)
			}

			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			// --from: copy an existing skill directory before registration.
			if fromPath != "" {
				absSrc, err := filepath.Abs(fromPath)
				if err != nil {
					return err
				}
				sourceBase := skillSourceOutputDir(cfg.SkillSources)
				destDir := filepath.Join(absTarget, sourceBase, name)
				if _, statErr := os.Stat(destDir); statErr == nil {
					return fmt.Errorf("destination %s already exists; remove it first or choose a different name", destDir)
				}
				if err := copyDir(absSrc, destDir); err != nil {
					return fmt.Errorf("copy from %s: %w", fromPath, err)
				}
				fmt.Printf("Copied %s → %s\n", fromPath, destDir)
				noScaffold = true // directory already populated; skip scaffolding
			}

			targetNames := append([]string(nil), inTargets...)
			if len(targetNames) == 0 {
				for n := range cfg.Targets {
					targetNames = append(targetNames, n)
				}
				sort.Strings(targetNames)
			}

			var added, alreadyIn []string
			for _, tn := range targetNames {
				t, ok := cfg.Targets[tn]
				if !ok {
					return fmt.Errorf("target %q not found in bakefile", tn)
				}
				found := false
				for _, s := range t.Skills {
					if s == name {
						found = true
						break
					}
				}
				if found {
					alreadyIn = append(alreadyIn, tn)
				} else {
					t.Skills = append(t.Skills, name)
					added = append(added, tn)
				}
			}

			if len(added) == 0 {
				fmt.Printf("Skill %q is already present in all targets.\n", name)
				return nil
			}

			if isSDLCSkill(name) {
				desc := catalog.SkillDescription(name)
				fmt.Printf("Hint: %q is a built-in SDLC skill — %s\n       Use `skcr scaffold skills` to generate all standard skills at once.\n", name, desc)
			}

			if cfg.SkillSources != nil {
				found := false
				for i, sd := range cfg.SkillSources.Skills {
					if sd.Name == name {
						// Update existing entry with any explicitly provided fields.
						if description != "" {
							cfg.SkillSources.Skills[i].Description = description
						}
						if owner != "" {
							cfg.SkillSources.Skills[i].Owner = owner
						}
						if initialVersion != "" {
							cfg.SkillSources.Skills[i].InitialVersion = initialVersion
						}
						if len(platforms) > 0 {
							cfg.SkillSources.Skills[i].CompatibleWith = platforms
						}
						found = true
						break
					}
				}
				if !found {
					def := models.SkillSourceDefinition{
						Name:        name,
						Description: description,
						Owner:       owner,
					}
					if initialVersion != "" {
						def.InitialVersion = initialVersion
					}
					if len(platforms) > 0 {
						def.CompatibleWith = platforms
					}
					cfg.SkillSources.Skills = append(cfg.SkillSources.Skills, def)
				}
			}

			if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
				return err
			}
			fmt.Printf("Added %q to targets: %s\n", name, strings.Join(added, ", "))
			if len(alreadyIn) > 0 {
				fmt.Printf("Already present in:  %s\n", strings.Join(alreadyIn, ", "))
			}

			if noScaffold {
				return nil
			}

			// Collect unique platforms across all affected targets.
			platSeen := map[string]struct{}{}
			var resolvedPlatforms []string
			for _, tn := range added {
				resolved, err := bake.ResolveTarget(cfg, tn)
				if err != nil {
					continue
				}
				for _, p := range resolved.Platforms {
					if _, dup := platSeen[p]; !dup {
						platSeen[p] = struct{}{}
						resolvedPlatforms = append(resolvedPlatforms, p)
					}
				}
			}

			created, skipped, err := scaffoldTargetSkills(absTarget, []string{name}, cfg.SkillSources, resolvedPlatforms, false)
			if err != nil {
				return err
			}
			fmt.Printf("Scaffold: %d files created, %d already existed.\n", created, skipped)
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().StringArrayVar(&inTargets, "in-target", nil, "Bake target(s) to add skill to (default: all)")
	cmd.Flags().BoolVar(&noScaffold, "no-scaffold", false, "Update bakefile only, skip directory scaffolding")
	cmd.Flags().StringVar(&description, "description", "", "Skill description")
	cmd.Flags().StringVar(&owner, "owner", "", "Skill owner")
	cmd.Flags().StringVar(&initialVersion, "initial-version", "", "Initial semver version for new scaffolds")
	cmd.Flags().StringArrayVar(&platforms, "platform", nil, "Compatible platform(s); may be repeated")
	cmd.Flags().StringVar(&fromPath, "from", "", "Copy an existing skill directory instead of scaffolding from scratch")
	_ = cmd.RegisterFlagCompletionFunc("in-target", completeBakeTargets)
	return cmd
}

// copyDir recursively copies src into dst, creating dst and all subdirectories.
// Symlinks are skipped to prevent escaping the source directory.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
