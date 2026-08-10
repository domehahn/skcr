package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"github.com/domehahn/skcr/v2/internal/validator"
	"github.com/domehahn/sklib/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type doctorFinding struct {
	level string // ok | warn | error
	check string
	msg   string
}

func newDoctorCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check project health: bakefile, skills, platform dirs, and toolchain",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			var findings []doctorFinding
			add := func(level, check, msg string) {
				findings = append(findings, doctorFinding{level, check, msg})
			}

			// ── Toolchain ────────────────────────────────────────────────────────
			if _, err := exec.LookPath("skpm"); err != nil {
				add("warn", "toolchain", "skpm not found in PATH — skill lifecycle commands unavailable")
			} else {
				add("ok", "toolchain", "skpm found")
			}

			// ── Bakefile ─────────────────────────────────────────────────────────
			bakePath := filepath.Join(absTarget, "agentic.bake.yaml")
			cfg, err := cliLoadBakeFile(bakePath)
			if err != nil {
				add("error", "bakefile", fmt.Sprintf("cannot parse agentic.bake.yaml: %v", err))
				printDoctorFindings(cmd.OutOrStdout(), findings)
				return doctorExitCode(findings)
			}
			add("ok", "bakefile", "agentic.bake.yaml is valid")

			// ── skill_sources migration check ─────────────────────────────────────
			if rawBytes, readErr := os.ReadFile(bakePath); readErr == nil {
				raw := map[string]any{}
				if yaml.Unmarshal(rawBytes, &raw) == nil {
					if ss, ok := raw["skill_sources"].(map[string]any); ok {
						if defaults, ok := ss["defaults"].(map[string]any); ok {
							if _, hasOldKey := defaults["version"]; hasOldKey {
								add("warn", "bakefile", "skill_sources.defaults contains deprecated key 'version' — rename to 'initial_version'")
							}
						}
						if skills, ok := ss["skills"].([]any); ok {
							for i, s := range skills {
								if sm, ok := s.(map[string]any); ok {
									if _, hasOldKey := sm["version"]; hasOldKey {
										name, _ := sm["name"].(string)
										add("warn", "bakefile", fmt.Sprintf("skill_sources.skills[%d] (%s): deprecated key 'version' — rename to 'initial_version'", i, name))
									}
								}
							}
						}
					}
				}
			}

			// ── Targets ──────────────────────────────────────────────────────────
			if len(cfg.Targets) == 0 {
				add("error", "targets", "no targets defined in bakefile")
			} else {
				add("ok", "targets", fmt.Sprintf("%d target(s) defined", len(cfg.Targets)))
			}

			// Check for duplicate skill names within each target.
			for tn, t := range cfg.Targets {
				seen := map[string]struct{}{}
				for _, s := range t.Skills {
					if _, dup := seen[s]; dup {
						add("error", "targets", fmt.Sprintf("target %q has duplicate skill %q", tn, s))
					}
					seen[s] = struct{}{}
				}
			}

			// ── Skills ───────────────────────────────────────────────────────────
			skillTargets := map[string]struct{}{}
			for _, t := range cfg.Targets {
				for _, s := range t.Skills {
					skillTargets[s] = struct{}{}
				}
			}
			skillNames := make([]string, 0, len(skillTargets))
			for s := range skillTargets {
				skillNames = append(skillNames, s)
			}
			sort.Strings(skillNames)

			if len(skillNames) == 0 {
				add("warn", "skills", "no skills defined in any target")
			}

			agentsBase := skillSourceOutputDir(cfg.SkillSources)
			for _, name := range skillNames {
				skillDir := filepath.Join(absTarget, agentsBase, name)
				if _, statErr := cliStatBake(skillDir); os.IsNotExist(statErr) {
					add("warn", "skills", fmt.Sprintf("%s/%s/ not scaffolded — run: skcr bake --write", agentsBase, name))
					continue
				}

				skillMD := filepath.Join(skillDir, "SKILL.md")
				skillMDContent, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					add("error", "skills", fmt.Sprintf("%s/%s/SKILL.md missing", agentsBase, name))
				} else {
					if err := checkSkillMDFrontmatter(string(skillMDContent)); err != "" {
						add("warn", "skills", fmt.Sprintf("%s/%s/SKILL.md: %s", agentsBase, name, err))
					} else {
						add("ok", "skills", fmt.Sprintf("%s/%s/SKILL.md valid", agentsBase, name))
					}
					for _, warning := range validator.ValidateSkillWarnings(string(skillMDContent)) {
						add("warn", "compat", fmt.Sprintf("%s/%s/SKILL.md: %s", agentsBase, name, warning))
					}
					if fm, ok := parseFrontmatterDates(string(skillMDContent)); ok {
						if fm.Since > fm.LastModified {
							add("error", "dates", fmt.Sprintf("%s/%s/SKILL.md: since (%s) is later than last_modified (%s)", agentsBase, name, fm.Since, fm.LastModified))
						}
					}
				}

				descriptorPresent := false
				descriptorPath := skillmeta.DescriptorPath(skillDir)
				descriptorName := filepath.Base(descriptorPath)
				if yamlData, err := os.ReadFile(descriptorPath); err != nil {
					add("error", "skills", fmt.Sprintf("%s/%s/descriptor.yaml missing", agentsBase, name))
				} else {
					descriptorPresent = true
					if readErr == nil {
						if fm, ok := parseFrontmatterDates(string(skillMDContent)); ok && fm.Version != "" {
							if yamlVersion := extractYAMLStringKey(string(yamlData), "version"); yamlVersion != "" && yamlVersion != fm.Version {
								add("error", "skills", fmt.Sprintf("%s/%s/%s version %q does not match SKILL.md version %q", agentsBase, name, descriptorName, yamlVersion, fm.Version))
							}
						}
					}
				}
				if descriptorPresent {
					if descriptor, loadErr := skillmeta.LoadDescriptor(descriptorPath); loadErr == nil && descriptor.Security != nil {
						add("warn", "contract", fmt.Sprintf(
							"%s/%s/%s: security is deprecated compatibility metadata and is ignored for authorization; declare the boundary only in contract.yaml",
							agentsBase, name, descriptorName,
						))
					}
					for _, descriptorErr := range skillmeta.ValidateDirectory(skillDir) {
						add("error", "contract", fmt.Sprintf("%s/%s: %s", agentsBase, name, descriptorErr))
					}
				}

				versionData, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
				if err != nil {
					add("error", "skills", fmt.Sprintf("%s/%s/VERSION missing", agentsBase, name))
				} else {
					v := strings.TrimSpace(string(versionData))
					if !spec.IsSemVer(strings.TrimPrefix(v, "v")) {
						add("error", "skills", fmt.Sprintf("%s/%s/VERSION %q is not valid semver", agentsBase, name, v))
					}
				}
			}

			// ── skill_sources orphan check ────────────────────────────────────────
			if cfg.SkillSources != nil && len(cfg.SkillSources.Skills) > 0 {
				targetSkillSet := map[string]struct{}{}
				for _, t := range cfg.Targets {
					for _, s := range t.Skills {
						targetSkillSet[s] = struct{}{}
					}
				}
				sourceSkillSet := map[string]struct{}{}
				for _, sd := range cfg.SkillSources.Skills {
					sourceSkillSet[sd.Name] = struct{}{}
					if _, inTarget := targetSkillSet[sd.Name]; !inTarget {
						add("warn", "skill_sources", fmt.Sprintf("skill %q is in skill_sources.skills but not in any target — add via: skcr add skill %s", sd.Name, sd.Name))
					}
				}
				for _, s := range skillNames {
					if _, inSource := sourceSkillSet[s]; !inSource {
						add("warn", "skill_sources", fmt.Sprintf("skill %q is in a target but not in skill_sources.skills — add via: skcr add skill %s", s, s))
					}
				}
			}

			// ── Platform dir sync ─────────────────────────────────────────────────
			dirs := allPlatformBaseDirs(cfg)
			outOfSync := 0
			for _, name := range skillNames {
				canonicalPath := filepath.Join(absTarget, agentsBase, name, "SKILL.md")
				_, err := os.ReadFile(canonicalPath)
				if err != nil {
					continue // already reported above
				}
				for _, baseDir := range dirs {
					if baseDir == agentsBase {
						continue
					}
					destDir := filepath.Join(absTarget, baseDir, name)
					_, err := os.Stat(filepath.Join(destDir, "SKILL.md"))
					if os.IsNotExist(err) {
						continue // not scaffolded; not an error here
					}
					if !skillArtifactsEqual(filepath.Dir(canonicalPath), destDir) {
						add("warn", "sync", fmt.Sprintf("%s/%s canonical artifacts differ — run: skcr sync", baseDir, name))
						outOfSync++
					}
				}
			}
			if outOfSync == 0 && len(skillNames) > 0 {
				add("ok", "sync", "all platform skill artifacts match canonical source")
			}

			// ── Lockfile ─────────────────────────────────────────────────────────
			lockPath := filepath.Join(absTarget, ".agentic-template.lock")
			if _, err := cliStatBake(lockPath); os.IsNotExist(err) {
				add("warn", "lockfile", ".agentic-template.lock missing — run: skcr bake --write")
			} else {
				add("ok", "lockfile", ".agentic-template.lock present")
			}

			printDoctorFindings(cmd.OutOrStdout(), findings)
			return doctorExitCode(findings)
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	return cmd
}

func printDoctorFindings(w io.Writer, findings []doctorFinding) {
	icons := map[string]string{"ok": "✓", "warn": "!", "error": "✗"}
	for _, f := range findings {
		fmt.Fprintf(w, "  %s  [%-9s]  %s\n", icons[f.level], f.check, f.msg)
	}
	errors, warns := 0, 0
	for _, f := range findings {
		switch f.level {
		case "error":
			errors++
		case "warn":
			warns++
		}
	}
	fmt.Fprintln(w)
	if errors == 0 && warns == 0 {
		fmt.Fprintln(w, "Everything looks healthy.")
	} else {
		parts := []string{}
		if errors > 0 {
			parts = append(parts, fmt.Sprintf("%d error(s)", errors))
		}
		if warns > 0 {
			parts = append(parts, fmt.Sprintf("%d warning(s)", warns))
		}
		fmt.Fprintf(w, "%s found.\n", strings.Join(parts, ", "))
	}
}

func doctorExitCode(findings []doctorFinding) error {
	for _, f := range findings {
		if f.level == "error" {
			return fmt.Errorf("doctor found errors")
		}
	}
	return nil
}

func checkSkillMDFrontmatter(content string) string {
	return validator.ValidateSkillMetadata(content)
}

type frontmatterDates struct {
	Version      string `yaml:"version"`
	Since        string `yaml:"since"`
	LastModified string `yaml:"last_modified"`
}

// extractYAMLStringKey returns the string value of a top-level YAML key without
// a full unmarshal — sufficient for reading "version:" from skill.yaml.
func extractYAMLStringKey(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+":") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
			return strings.Trim(val, `"'`)
		}
	}
	return ""
}

func parseFrontmatterDates(content string) (frontmatterDates, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return frontmatterDates{}, false
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return frontmatterDates{}, false
	}
	raw := content[4 : 4+end]
	var fm frontmatterDates
	if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
		return frontmatterDates{}, false
	}
	return fm, true
}
