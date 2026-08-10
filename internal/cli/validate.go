package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/domehahn/skcr/internal/skillmeta"
	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/domehahn/skcr/internal/validator"
	"github.com/spf13/cobra"
)

// cliLoadBakeFileValidate reuses the package-level var from bake.go (same package).
// This alias exists only for documentation clarity.
var _ = cliLoadBakeFile // ensure the bake.go var is used here too

var cliAbsPathValidate = filepath.Abs

func newValidateCommand() *cobra.Command {
	var target string
	var againstLock string
	var skills bool
	var platform string
	var ci bool
	var fix bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated project or standalone skill state",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := cliAbsPathValidate(target)
			if err != nil {
				return err
			}
			if skillDir, standalone, detectErr := standaloneSkillTarget(absTarget); detectErr != nil {
				return detectErr
			} else if standalone {
				if againstLock != "" || skills || platform != "" {
					return fmt.Errorf("--against-lock, --skills, and --platform require a project target with agentic.bake.yaml")
				}
				return validateStandaloneSkill(skillDir, fix, cmd)
			}
			errors, err := validator.ValidateProjectWithOptions(absTarget, validator.Options{
				AgainstLock: againstLock,
				Skills:      skills,
				Platform:    platform,
				CI:          ci,
			})
			if err != nil {
				return err
			}
			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Println("ERROR", e)
				}
				if !fix {
					return fmt.Errorf("validation failed")
				}
				fmt.Println("\nAttempting auto-fix…")
				if fixErr := runValidateFix(absTarget, cmd); fixErr != nil {
					return fixErr
				}
				return nil
			}
			fmt.Println("Validation passed")
			cfg, cfgErr := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if cfgErr == nil && cfg.SkillSources != nil && len(cfg.SkillSources.Skills) > 0 {
				fmt.Printf("\nFor package/lifecycle validation, run:\n  skpm validate %s/<name>\n", cfg.SkillSources.OutputDir)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository, skill directory, or SKILL.md path")
	cmd.Flags().StringVar(&againstLock, "against-lock", "", "Validate against skpm agent-skills.lock")
	cmd.Flags().BoolVar(&skills, "skills", false, "Validate configured skpm skill lock state")
	cmd.Flags().StringVar(&platform, "platform", "", "Validate only the selected canonical platform")
	cmd.Flags().BoolVar(&ci, "ci", false, "Enable CI-oriented validation output")
	cmd.Flags().BoolVar(&fix, "fix", false, "Auto-fix known issues: sync version drift between SKILL.md, VERSION, descriptor.yaml, and CHANGELOG.md")
	return cmd
}

func standaloneSkillTarget(absTarget string) (string, bool, error) {
	stat, err := os.Stat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	if !stat.IsDir() {
		if filepath.Base(absTarget) == "SKILL.md" {
			return filepath.Dir(absTarget), true, nil
		}
		return "", false, fmt.Errorf("target is not a directory or SKILL.md file: %s", absTarget)
	}
	if _, err := os.Stat(filepath.Join(absTarget, "agentic.bake.yaml")); err == nil {
		return "", false, nil
	} else if !os.IsNotExist(err) {
		return "", false, err
	}
	if _, err := os.Stat(filepath.Join(absTarget, "SKILL.md")); err == nil {
		return absTarget, true, nil
	} else if !os.IsNotExist(err) {
		return "", false, err
	}
	return "", false, nil
}

func validateStandaloneSkill(skillDir string, fix bool, cmd *cobra.Command) error {
	validate := func() ([]string, []string, error) {
		infos, err := skillversion.Check(skillDir)
		if err != nil {
			return nil, nil, err
		}
		errors := []string{}
		warnings := []string{}
		for _, info := range infos {
			for _, validationErr := range info.Errors {
				errors = append(errors, fmt.Sprintf("%s: %s", filepath.Base(skillDir), validationErr))
			}
			warnings = append(warnings, info.Warnings...)
		}
		for _, validationErr := range skillmeta.ValidateDirectory(skillDir) {
			errors = append(errors, fmt.Sprintf("%s: %s", filepath.Base(skillDir), validationErr))
		}
		sort.Strings(errors)
		sort.Strings(warnings)
		return errors, warnings, nil
	}

	errors, warnings, err := validate()
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		fmt.Fprintln(cmd.OutOrStdout(), "WARN", warning)
	}
	if len(errors) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Validation passed")
		fmt.Fprintf(cmd.OutOrStdout(), "\nFor package/lifecycle validation, run:\n  skpm validate %s\n", skillDir)
		return nil
	}
	for _, validationErr := range errors {
		fmt.Fprintln(cmd.OutOrStdout(), "ERROR", validationErr)
	}
	if !fix {
		return fmt.Errorf("validation failed")
	}

	fmt.Fprintln(cmd.OutOrStdout(), "\nAttempting auto-fix…")
	if _, err := skillversion.SyncArtifacts(skillDir); err != nil {
		return err
	}
	errors, _, err = validate()
	if err != nil {
		return err
	}
	if len(errors) > 0 {
		for _, validationErr := range errors {
			fmt.Fprintln(cmd.OutOrStdout(), "ERROR", validationErr)
		}
		return fmt.Errorf("validation failed after auto-fix")
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Validation passed after auto-fix")
	return nil
}

// runValidateFix applies automatic repairs to all scaffolded skills in the project.
// Currently fixes: VERSION / descriptor.yaml / CHANGELOG.md out of sync with SKILL.md frontmatter version.
func runValidateFix(absTarget string, cmd *cobra.Command) error {
	cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
	if err != nil {
		return err
	}
	sourceBase := skillSourceOutputDir(cfg.SkillSources)

	seen := map[string]struct{}{}
	var names []string
	for _, t := range cfg.Targets {
		for _, s := range t.Skills {
			if _, dup := seen[s]; !dup {
				seen[s] = struct{}{}
				names = append(names, s)
			}
		}
	}
	sort.Strings(names)

	fixed, skipped, errs := 0, 0, 0
	for _, name := range names {
		skillDir := filepath.Join(absTarget, sourceBase, name)
		if _, statErr := os.Stat(filepath.Join(skillDir, "SKILL.md")); os.IsNotExist(statErr) {
			skipped++
			continue
		}
		info, syncErr := skillversion.SyncArtifacts(skillDir)
		if syncErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  error  %-40s  %v\n", name, syncErr)
			errs++
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  fixed  %-40s  %s\n", name, info.Version)
		fixed++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nFix complete: %d fixed, %d skipped (not scaffolded), %d error(s).\n", fixed, skipped, errs)
	if errs > 0 {
		return fmt.Errorf("%d skill(s) could not be fixed", errs)
	}
	return nil
}
