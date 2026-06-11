package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/validator"
	"github.com/domehahn/sklib/spec"
	"github.com/spf13/cobra"
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
				printDoctorFindings(findings)
				return doctorExitCode(findings)
			}
			add("ok", "bakefile", "agentic.bake.yaml is valid")

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

			const agentsBase = ".agents/skills"
			for _, name := range skillNames {
				skillDir := filepath.Join(absTarget, agentsBase, name)
				if _, statErr := cliStatBake(skillDir); os.IsNotExist(statErr) {
					add("warn", "skills", fmt.Sprintf("%s/%s/ not scaffolded — run: skcr bake --write", agentsBase, name))
					continue
				}

				skillMD := filepath.Join(skillDir, "SKILL.md")
				if data, err := os.ReadFile(skillMD); err != nil {
					add("error", "skills", fmt.Sprintf("%s/%s/SKILL.md missing", agentsBase, name))
				} else {
					if err := checkSkillMDFrontmatter(string(data)); err != "" {
						add("warn", "skills", fmt.Sprintf("%s/%s/SKILL.md: %s", agentsBase, name, err))
					} else {
						add("ok", "skills", fmt.Sprintf("%s/%s/SKILL.md valid", agentsBase, name))
					}
					for _, warning := range validator.ValidateSkillWarnings(string(data)) {
						add("warn", "compat", fmt.Sprintf("%s/%s/SKILL.md: %s", agentsBase, name, warning))
					}
				}

				if _, err := os.ReadFile(filepath.Join(skillDir, "skill.yaml")); err != nil {
					add("error", "skills", fmt.Sprintf("%s/%s/skill.yaml missing", agentsBase, name))
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

			// ── Platform dir sync ─────────────────────────────────────────────────
			dirs := allPlatformBaseDirs(cfg)
			outOfSync := 0
			for _, name := range skillNames {
				canonicalPath := filepath.Join(absTarget, agentsBase, name, "SKILL.md")
				canonicalData, err := os.ReadFile(canonicalPath)
				if err != nil {
					continue // already reported above
				}
				for _, baseDir := range dirs {
					if baseDir == agentsBase {
						continue
					}
					destPath := filepath.Join(absTarget, baseDir, name, "SKILL.md")
					data, err := os.ReadFile(destPath)
					if os.IsNotExist(err) {
						continue // not scaffolded; not an error here
					}
					if string(data) != string(canonicalData) {
						add("warn", "sync", fmt.Sprintf("%s/%s/SKILL.md differs from canonical — run: skcr sync", baseDir, name))
						outOfSync++
					}
				}
			}
			if outOfSync == 0 && len(skillNames) > 0 {
				add("ok", "sync", "all platform SKILL.md files match canonical source")
			}

			// ── Lockfile ─────────────────────────────────────────────────────────
			lockPath := filepath.Join(absTarget, ".agentic-template.lock")
			if _, err := cliStatBake(lockPath); os.IsNotExist(err) {
				add("warn", "lockfile", ".agentic-template.lock missing — run: skcr bake --write")
			} else {
				add("ok", "lockfile", ".agentic-template.lock present")
			}

			printDoctorFindings(findings)
			return doctorExitCode(findings)
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	return cmd
}

func printDoctorFindings(findings []doctorFinding) {
	icons := map[string]string{"ok": "✓", "warn": "!", "error": "✗"}
	for _, f := range findings {
		fmt.Printf("  %s  [%-9s]  %s\n", icons[f.level], f.check, f.msg)
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
	fmt.Println()
	if errors == 0 && warns == 0 {
		fmt.Println("Everything looks healthy.")
	} else {
		parts := []string{}
		if errors > 0 {
			parts = append(parts, fmt.Sprintf("%d error(s)", errors))
		}
		if warns > 0 {
			parts = append(parts, fmt.Sprintf("%d warning(s)", warns))
		}
		fmt.Printf("%s found.\n", strings.Join(parts, ", "))
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
