package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/domehahn/skcr/v2/internal/catalog"
	"github.com/domehahn/skcr/v2/internal/lockfile"
	"github.com/domehahn/skcr/v2/internal/models"
	"github.com/domehahn/skcr/v2/internal/scaffold"
	"github.com/domehahn/skcr/v2/internal/skillversion"
	"github.com/domehahn/sklib/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const catalogStatePath = ".skcr/catalog.lock.yaml"

type catalogSkillState struct {
	AvailableDigest string `yaml:"available_digest"`
	InstalledDigest string `yaml:"installed_digest,omitempty"`
	Managed         bool   `yaml:"managed"`
}

type catalogState struct {
	Version       string                       `yaml:"version"`
	CatalogDigest string                       `yaml:"catalog_digest"`
	UpdatedAt     string                       `yaml:"updated_at"`
	Skills        map[string]catalogSkillState `yaml:"skills"`
}

type catalogSkillStatus struct {
	Name            string
	Status          string
	AvailableDigest string
	ActualDigest    string
	State           catalogSkillState
}

func newUpdateCommand() *cobra.Command {
	var target string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh the built-in skill catalog and report project updates",
		Long: `Compares the built-in skcr skill catalog with agentic.bake.yaml and the
canonical skill sources in the project. The command records catalog and installed
template digests in .skcr/catalog.lock.yaml but does not change skills. Run
'skcr upgrade' afterwards to add or refresh eligible skills.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			previous, err := loadCatalogState(absTarget)
			if err != nil {
				return err
			}
			state, statuses, err := scanSkillCatalog(absTarget, cfg, previous)
			if err != nil {
				return err
			}
			printCatalogStatus(cmd, state, statuses)
			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "\nDry run only. Catalog state was not written.")
				return nil
			}
			if err := writeCatalogState(absTarget, state); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nUpdated %s\n", catalogStatePath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without updating catalog state")
	return cmd
}

func runCatalogUpgrade(cmd *cobra.Command, target string, dryRun, force bool) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	bakePath := filepath.Join(absTarget, "agentic.bake.yaml")
	cfg, err := cliLoadBakeFile(bakePath)
	if err != nil {
		return err
	}
	previous, err := loadCatalogState(absTarget)
	if err != nil {
		return err
	}
	state, statuses, err := scanSkillCatalog(absTarget, cfg, previous)
	if err != nil {
		return err
	}
	printCatalogStatus(cmd, state, statuses)

	missingConfig := []string{}
	refresh := []string{}
	for _, item := range statuses {
		switch item.Status {
		case "new", "missing":
			missingConfig = append(missingConfig, item.Name)
		case "outdated":
			refresh = append(refresh, item.Name)
		case "modified":
			if force {
				refresh = append(refresh, item.Name)
			}
		}
	}
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "\nUpgrade plan: %d skill(s) to add/scaffold, %d skill(s) to refresh.\n", len(missingConfig), len(refresh))
		if !force {
			modified := countCatalogStatus(statuses, "modified")
			if modified > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%d locally modified or unmanaged skill(s) would remain unchanged; use --force to replace their instruction bodies.\n", modified)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Dry run only. Re-run without --dry-run to apply.")
		return nil
	}

	targetNames := catalogManagedTargets(cfg)
	if len(targetNames) == 0 {
		return fmt.Errorf("no bake target available for built-in skills")
	}
	if cfg.SkillSources == nil {
		cfg.SkillSources = &models.SkillSourceConfig{OutputDir: ".agents/skills"}
	}
	ensureCatalogRegistration(cfg, targetNames)
	if err := cliDumpBakeFile(cfg, bakePath); err != nil {
		return err
	}

	platforms, err := catalogTargetPlatforms(cfg, targetNames)
	if err != nil {
		return err
	}
	created, skipped, err := scaffoldTargetSkills(absTarget, catalog.CoreSkills, cfg.SkillSources, platforms, false)
	if err != nil {
		return err
	}

	refreshed := 0
	for _, name := range refresh {
		skillDir := filepath.Join(absTarget, skillSourceOutputDir(cfg.SkillSources), name)
		if err := refreshCatalogSkill(skillDir, name); err != nil {
			return fmt.Errorf("refresh %s: %w", name, err)
		}
		if _, err := syncCatalogSkill(absTarget, cfg, platforms, name); err != nil {
			return err
		}
		refreshed++
	}

	state, _, err = scanSkillCatalog(absTarget, cfg, state)
	if err != nil {
		return err
	}
	if err := writeCatalogState(absTarget, state); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nUpgrade complete: %d file(s) created, %d existing file(s) preserved, %d skill instruction body/bodies refreshed.\n", created, skipped, refreshed)
	if modified := countCatalogStatus(statuses, "modified"); modified > 0 && !force {
		fmt.Fprintf(cmd.OutOrStdout(), "%d locally modified or unmanaged skill(s) were preserved. Review them manually or use --force.\n", modified)
	}
	return nil
}

func scanSkillCatalog(root string, cfg *models.BakeConfig, previous catalogState) (catalogState, []catalogSkillStatus, error) {
	state := catalogState{Version: "1", UpdatedAt: time.Now().UTC().Format(time.RFC3339), Skills: map[string]catalogSkillState{}}
	statuses := make([]catalogSkillStatus, 0, len(catalog.CoreSkills))
	sourceBase := skillSourceOutputDir(cfg.SkillSources)
	configured := configuredSkillSet(cfg)
	digestParts := []string{}
	names := append([]string(nil), catalog.CoreSkills...)
	sort.Strings(names)
	for _, name := range names {
		available, err := catalogInstructionDigest(name)
		if err != nil {
			return catalogState{}, nil, err
		}
		digestParts = append(digestParts, name+"="+available)
		prior := previous.Skills[name]
		entry := catalogSkillState{AvailableDigest: available, InstalledDigest: prior.InstalledDigest, Managed: prior.Managed}
		item := catalogSkillStatus{Name: name, AvailableDigest: available, State: entry}
		if !configured[name] {
			item.Status = "new"
			state.Skills[name] = entry
			statuses = append(statuses, item)
			continue
		}
		payload, readErr := os.ReadFile(filepath.Join(root, sourceBase, name, "SKILL.md"))
		if readErr != nil {
			if os.IsNotExist(readErr) {
				item.Status = "missing"
				state.Skills[name] = entry
				statuses = append(statuses, item)
				continue
			}
			return catalogState{}, nil, readErr
		}
		actual := instructionDigest(string(payload))
		item.ActualDigest = actual
		switch {
		case actual == available:
			item.Status = "current"
			entry.InstalledDigest = available
			entry.Managed = true
		case prior.Managed && prior.InstalledDigest != "" && actual == prior.InstalledDigest:
			item.Status = "outdated"
		case prior.InstalledDigest != "" && actual != prior.InstalledDigest:
			item.Status = "modified"
		case prior.Managed:
			item.Status = "modified"
		default:
			item.Status = "modified"
			entry.InstalledDigest = actual
			entry.Managed = false
		}
		item.State = entry
		state.Skills[name] = entry
		statuses = append(statuses, item)
	}
	state.CatalogDigest = lockfile.Sha256Text(strings.Join(digestParts, "\n"))
	return state, statuses, nil
}

func catalogInstructionDigest(name string) (string, error) {
	rendered, ok, err := scaffold.RenderRegisteredSkillMarkdown(name, "", catalog.SkillDescription(name), "0.0.0", "2000-01-01", "2000-01-01", "skcr-catalog", "experimental", "MIT", []string{"codex"})
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("built-in catalog skill %q has no registered template", name)
	}
	return instructionDigest(rendered), nil
}

func instructionDigest(content string) string {
	_, body := splitFrontmatterBody(content)
	if changelog := strings.Index(body, "\n## Changelog\n"); changelog >= 0 {
		body = body[:changelog]
	}
	return lockfile.Sha256Text(strings.TrimSpace(body))
}

func configuredSkillSet(cfg *models.BakeConfig) map[string]bool {
	inTarget := map[string]bool{}
	inSources := map[string]bool{}
	for _, target := range cfg.Targets {
		if target == nil {
			continue
		}
		for _, name := range target.Skills {
			inTarget[name] = true
		}
	}
	if cfg.SkillSources != nil {
		for _, def := range cfg.SkillSources.Skills {
			inSources[def.Name] = true
		}
	}
	configured := map[string]bool{}
	for name := range inTarget {
		configured[name] = cfg.SkillSources == nil || inSources[name]
	}
	return configured
}

func catalogManagedTargets(cfg *models.BakeConfig) []string {
	builtIn := map[string]bool{}
	for _, name := range catalog.CoreSkills {
		builtIn[name] = true
	}
	maxCount := 0
	counts := map[string]int{}
	for name, target := range cfg.Targets {
		if target == nil {
			continue
		}
		for _, skill := range target.Skills {
			if builtIn[skill] {
				counts[name]++
			}
		}
		if counts[name] > maxCount {
			maxCount = counts[name]
		}
	}
	names := []string{}
	if maxCount > 0 {
		for name, count := range counts {
			if count == maxCount {
				names = append(names, name)
			}
		}
	} else if _, ok := cfg.Targets["default"]; ok {
		names = append(names, "default")
	} else {
		for name := range cfg.Targets {
			names = append(names, name)
			break
		}
	}
	sort.Strings(names)
	return names
}

func ensureCatalogRegistration(cfg *models.BakeConfig, targetNames []string) {
	for _, targetName := range targetNames {
		target := cfg.Targets[targetName]
		seen := map[string]bool{}
		for _, name := range target.Skills {
			seen[name] = true
		}
		for _, name := range catalog.CoreSkills {
			if !seen[name] {
				target.Skills = append(target.Skills, name)
			}
		}
	}
	seen := map[string]bool{}
	for _, def := range cfg.SkillSources.Skills {
		seen[def.Name] = true
	}
	for _, name := range catalog.CoreSkills {
		if !seen[name] {
			cfg.SkillSources.Skills = append(cfg.SkillSources.Skills, models.SkillSourceDefinition{Name: name, Description: catalog.SkillDescription(name)})
		}
	}
}

func catalogTargetPlatforms(cfg *models.BakeConfig, targets []string) ([]string, error) {
	seen := map[string]bool{}
	platforms := []string{}
	for _, name := range targets {
		resolved, err := cliResolveTarget(cfg, name)
		if err != nil {
			return nil, err
		}
		for _, platform := range resolved.Platforms {
			if !seen[platform] {
				seen[platform] = true
				platforms = append(platforms, platform)
			}
		}
	}
	return platforms, nil
}

func refreshCatalogSkill(skillDir, name string) error {
	skillPath := filepath.Join(skillDir, "SKILL.md")
	payload, err := os.ReadFile(skillPath)
	if err != nil {
		return err
	}
	fmRaw, _ := splitFrontmatterBody(string(payload))
	if fmRaw == "" {
		return fmt.Errorf("missing YAML frontmatter")
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return err
	}
	owner := "platform-engineering"
	if len(fm.Authors) > 0 {
		owner = fm.Authors[0]
	}
	rendered, ok, err := scaffold.RenderRegisteredSkillMarkdown(name, "", fm.Description, fm.Version, fm.Since, fm.LastModified, owner, string(fm.Stability), "MIT", nil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no registered template")
	}
	_, body := splitFrontmatterBody(rendered)
	updated := "---\n" + fmRaw + "\n---" + body
	if err := os.WriteFile(skillPath, []byte(updated), 0o644); err != nil {
		return err
	}
	_, err = skillversion.Bump(skillDir, skillversion.BumpPatch, time.Now().UTC().Format("2006-01-02"), "Refresh built-in skill template from skcr catalog")
	return err
}

func syncCatalogSkill(root string, cfg *models.BakeConfig, platforms []string, name string) (int, error) {
	sourceBase := skillSourceOutputDir(cfg.SkillSources)
	srcDir := filepath.Join(root, sourceBase, name)
	artifacts, err := canonicalSkillArtifacts(srcDir)
	if err != nil {
		return 0, err
	}
	seen := map[string]bool{sourceBase: true}
	updated := 0
	for _, platform := range platforms {
		base := canonicalPlatformSkillBaseDir(platform)
		if seen[base] {
			continue
		}
		seen[base] = true
		destDir := filepath.Join(root, base, name)
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			continue
		}
		for _, rel := range artifacts {
			payload, err := os.ReadFile(filepath.Join(srcDir, rel))
			if err != nil {
				return updated, err
			}
			dest := filepath.Join(destDir, rel)
			if err := rejectSymlinkArtifactPath(destDir, dest); err != nil {
				return updated, err
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return updated, err
			}
			if err := os.WriteFile(dest, payload, 0o644); err != nil {
				return updated, err
			}
			updated++
		}
	}
	return updated, nil
}

func loadCatalogState(root string) (catalogState, error) {
	payload, err := os.ReadFile(filepath.Join(root, catalogStatePath))
	if err != nil {
		if os.IsNotExist(err) {
			return catalogState{Version: "1", Skills: map[string]catalogSkillState{}}, nil
		}
		return catalogState{}, err
	}
	state := catalogState{}
	if err := yaml.Unmarshal(payload, &state); err != nil {
		return catalogState{}, fmt.Errorf("read %s: %w", catalogStatePath, err)
	}
	if state.Skills == nil {
		state.Skills = map[string]catalogSkillState{}
	}
	return state, nil
}

func writeCatalogState(root string, state catalogState) error {
	payload, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	path := filepath.Join(root, catalogStatePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func printCatalogStatus(cmd *cobra.Command, state catalogState, statuses []catalogSkillStatus) {
	fmt.Fprintf(cmd.OutOrStdout(), "Skill catalog: %s\n", state.CatalogDigest)
	fmt.Fprintln(cmd.OutOrStdout(), "Status\tSkill")
	for _, item := range statuses {
		if item.Status != "current" {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", item.Status, item.Name)
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n%d current, %d new, %d missing, %d outdated, %d modified/unmanaged.\n",
		countCatalogStatus(statuses, "current"), countCatalogStatus(statuses, "new"), countCatalogStatus(statuses, "missing"), countCatalogStatus(statuses, "outdated"), countCatalogStatus(statuses, "modified"))
}

func countCatalogStatus(statuses []catalogSkillStatus, status string) int {
	count := 0
	for _, item := range statuses {
		if item.Status == status {
			count++
		}
	}
	return count
}
