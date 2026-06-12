package skillversion

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/domehahn/skcr/internal/validator"
	"github.com/domehahn/sklib/spec"
	"gopkg.in/yaml.v3"
)

type BumpKind string

const (
	BumpMajor BumpKind = "major"
	BumpMinor BumpKind = "minor"
	BumpPatch BumpKind = "patch"
)

type SkillInfo struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	LastModified string   `json:"last_modified"`
	LatestChange string   `json:"latest_change"`
	Warnings     []string `json:"warnings,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

type ReleaseEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Date    string `json:"date"`
	Change  string `json:"change"`
	Path    string `json:"path"`
}

type BumpOptions struct {
	Kind   BumpKind `json:"kind"`
	Date   string   `json:"date,omitempty"`
	Change string   `json:"change"`
	DryRun bool     `json:"dry_run,omitempty"`
}

type BumpResult struct {
	Info       SkillInfo `json:"info"`
	OldVersion string    `json:"old_version"`
	NewVersion string    `json:"new_version"`
	DryRun     bool      `json:"dry_run,omitempty"`
}

type ChangedSkill struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	OldVersion     string   `json:"old_version,omitempty"`
	CurrentVersion string   `json:"current_version,omitempty"`
	VersionChanged bool     `json:"version_changed"`
	Files          []string `json:"files"`
	Errors         []string `json:"errors,omitempty"`
}

type ReleaseBundle struct {
	Checks       []SkillInfo    `json:"checks"`
	Changed      []ChangedSkill `json:"changed,omitempty"`
	Changelog    []ReleaseEntry `json:"changelog"`
	ReleaseNotes string         `json:"release_notes"`
}

var bodyChangelogRE = regexp.MustCompile(`(?m)^## Changelog\s*$`)
var bodyEntryRE = regexp.MustCompile(`(?m)^###\s+([0-9A-Za-z.+-]+)\s+-\s+(\d{4}-\d{2}-\d{2})\s*$`)
var changelogHeadingRE = regexp.MustCompile(`(?m)^##\s+([0-9A-Za-z.+-]+)(?:\s+-\s+\d{4}-\d{2}-\d{2})?\s*$`)
var gitCommand = exec.Command

func Check(path string) ([]SkillInfo, error) {
	files, err := skillFiles(path)
	if err != nil {
		return nil, err
	}
	infos := make([]SkillInfo, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		info, err := parseInfo(file, string(content))
		if err != nil {
			info = SkillInfo{Path: file, Errors: []string{err.Error()}}
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func Bump(path string, kind BumpKind, date, change string) (SkillInfo, error) {
	result, err := BumpWithOptions(path, BumpOptions{Kind: kind, Date: date, Change: change})
	return result.Info, err
}

func BumpWithOptions(path string, opts BumpOptions) (BumpResult, error) {
	kind := opts.Kind
	if kind == "" {
		kind = BumpPatch
	}
	change := opts.Change
	if change == "" {
		return BumpResult{}, fmt.Errorf("change message is required")
	}
	date := opts.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date) {
		return BumpResult{}, fmt.Errorf("date must use YYYY-MM-DD format")
	}
	file := skillFile(path)
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		return BumpResult{}, err
	}
	content := string(contentBytes)
	info, err := parseInfo(file, content)
	if err != nil {
		return BumpResult{}, err
	}
	next, err := nextVersion(info.Version, kind)
	if err != nil {
		return BumpResult{}, err
	}
	updated, err := updateSkillMD(content, next, date, change)
	if err != nil {
		return BumpResult{}, err
	}
	result := BumpResult{OldVersion: info.Version, NewVersion: next, DryRun: opts.DryRun}
	if opts.DryRun {
		dryInfo := info
		dryInfo.Version = next
		dryInfo.LastModified = date
		dryInfo.LatestChange = change
		result.Info = dryInfo
		return result, nil
	}
	if err := os.WriteFile(file, []byte(updated), 0o644); err != nil {
		return BumpResult{}, err
	}
	dir := filepath.Dir(file)
	if err := updateTextFileIfExists(filepath.Join(dir, "VERSION"), next+"\n"); err != nil {
		return BumpResult{}, err
	}
	if err := updateSkillYAMLIfExists(filepath.Join(dir, "skill.yaml"), next); err != nil {
		return BumpResult{}, err
	}
	if err := prependChangelogIfExists(filepath.Join(dir, "CHANGELOG.md"), next, date, change); err != nil {
		return BumpResult{}, err
	}
	info, err = parseInfo(file, updated)
	result.Info = info
	return result, err
}

func SyncArtifacts(path string) (SkillInfo, error) {
	file := skillFile(path)
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		return SkillInfo{}, err
	}
	content := string(contentBytes)
	info, err := parseInfo(file, content)
	if err != nil {
		return info, err
	}
	frontmatter, _, ok := splitFrontmatter(content)
	if !ok {
		return info, fmt.Errorf("missing YAML frontmatter")
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return info, err
	}
	date := fm.LastModified
	change := "Version metadata synchronized"
	if len(fm.Changelog) > 0 {
		date = fm.Changelog[0].Date
		change = fm.Changelog[0].Change
	}
	dir := filepath.Dir(file)
	if err := updateTextFileIfExists(filepath.Join(dir, "VERSION"), fm.Version+"\n"); err != nil {
		return info, err
	}
	if err := updateSkillYAMLIfExists(filepath.Join(dir, "skill.yaml"), fm.Version); err != nil {
		return info, err
	}
	if err := syncChangelogIfExists(filepath.Join(dir, "CHANGELOG.md"), fm.Version, date, change); err != nil {
		return info, err
	}
	return parseInfo(file, content)
}

func Changelog(path string) ([]ReleaseEntry, error) {
	files, err := skillFiles(path)
	if err != nil {
		return nil, err
	}
	entries := []ReleaseEntry{}
	for _, file := range files {
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		frontmatter, _, ok := splitFrontmatter(string(contentBytes))
		if !ok {
			continue
		}
		fm := spec.SkillMDFrontmatter{}
		if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
			continue
		}
		for _, entry := range fm.Changelog {
			entries = append(entries, ReleaseEntry{
				Name:    fm.Name,
				Version: entry.Version,
				Date:    entry.Date,
				Change:  entry.Change,
				Path:    file,
			})
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Date == entries[j].Date {
			if entries[i].Name == entries[j].Name {
				return entries[i].Version > entries[j].Version
			}
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Date > entries[j].Date
	})
	return entries, nil
}

func ReleaseNotes(path string, since string) (string, error) {
	entries, err := Changelog(path)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("# Release Notes\n\n")
	for _, entry := range entries {
		if since != "" && entry.Date < since {
			continue
		}
		fmt.Fprintf(&b, "## %s %s - %s\n\n- %s\n\n", entry.Name, entry.Version, entry.Date, entry.Change)
	}
	return b.String(), nil
}

func ReleaseBundleFor(path, since string, includeChanged bool) (ReleaseBundle, error) {
	checks, err := Check(path)
	if err != nil {
		return ReleaseBundle{}, err
	}
	changelog, err := Changelog(path)
	if err != nil {
		return ReleaseBundle{}, err
	}
	notes, err := ReleaseNotes(path, since)
	if err != nil {
		return ReleaseBundle{}, err
	}
	bundle := ReleaseBundle{Checks: checks, Changelog: changelog, ReleaseNotes: notes}
	if includeChanged {
		changed, err := Changed(path)
		if err == nil {
			bundle.Changed = changed
		}
	}
	return bundle, nil
}

func Changed(path string) ([]ChangedSkill, error) {
	root, err := gitRoot(path)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	statusDir := abs
	if filepath.Base(path) == "SKILL.md" {
		statusDir = filepath.Dir(abs)
	}
	if resolvedRoot, err := filepath.EvalSymlinks(root); err == nil {
		root = resolvedRoot
	}
	if resolvedAbs, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolvedAbs
	}
	changedFiles, err := gitChangedFiles(statusDir, ".")
	if err != nil {
		return nil, err
	}
	bySkill := map[string]*ChangedSkill{}
	for _, changedFile := range changedFiles {
		changedFile = normalizeRepoRelPath(root, changedFile)
		skillMD := nearestSkillFile(root, changedFile)
		if skillMD == "" {
			continue
		}
		currentBytes, err := os.ReadFile(filepath.Join(root, skillMD))
		if err != nil {
			continue
		}
		current, err := parseInfo(filepath.Join(root, skillMD), string(currentBytes))
		if err != nil {
			current.Path = filepath.Join(root, skillMD)
			current.Errors = append(current.Errors, err.Error())
		}
		item, ok := bySkill[skillMD]
		if !ok {
			oldVersion := headVersion(root, skillMD)
			item = &ChangedSkill{
				Name:           current.Name,
				Path:           filepath.Join(root, skillMD),
				OldVersion:     oldVersion,
				CurrentVersion: current.Version,
				VersionChanged: oldVersion == "" || oldVersion != current.Version,
			}
			if item.Name == "" {
				item.Name = filepath.Base(filepath.Dir(skillMD))
			}
			if oldVersion != "" && oldVersion == current.Version {
				item.Errors = append(item.Errors, "material skill change without version bump")
			}
			bySkill[skillMD] = item
		}
		item.Files = append(item.Files, filepath.ToSlash(changedFile))
	}
	out := make([]ChangedSkill, 0, len(bySkill))
	for _, item := range bySkill {
		sort.Strings(item.Files)
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func BumpAllChanged(path string, opts BumpOptions) ([]BumpResult, error) {
	changed, err := Changed(path)
	if err != nil {
		return nil, err
	}
	results := []BumpResult{}
	for _, item := range changed {
		if item.VersionChanged {
			continue
		}
		result, err := BumpWithOptions(item.Path, opts)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func parseInfo(path, content string) (SkillInfo, error) {
	info := SkillInfo{Path: path}
	frontmatter, _, ok := splitFrontmatter(content)
	if !ok {
		return info, fmt.Errorf("missing YAML frontmatter")
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return info, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	info.Name = fm.Name
	info.Version = fm.Version
	info.LastModified = fm.LastModified
	if len(fm.Changelog) > 0 {
		info.LatestChange = fm.Changelog[0].Change
	}
	if err := validator.ValidateSkillMetadata(content); err != "" {
		info.Errors = append(info.Errors, strings.TrimPrefix(err, "Skill metadata invalid: "))
	}
	info.Errors = append(info.Errors, artifactConsistencyErrors(path, info.Version)...)
	info.Warnings = validator.ValidateSkillWarnings(content)
	return info, nil
}

func skillFile(path string) string {
	if filepath.Base(path) == "SKILL.md" {
		return path
	}
	return filepath.Join(path, "SKILL.md")
}

func skillFiles(path string) ([]string, error) {
	if filepath.Base(path) == "SKILL.md" {
		return []string{path}, nil
	}
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a SKILL.md file or directory", path)
	}
	if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil {
		return []string{filepath.Join(path, "SKILL.md")}, nil
	}
	files := []string{}
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(p) == "SKILL.md" {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func splitFrontmatter(content string) (frontmatter, body string, ok bool) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content, false
	}
	end := strings.Index(content[len("---\n"):], "\n---")
	if end < 0 {
		return "", content, false
	}
	end += len("---\n")
	frontmatter = content[len("---\n"):end]
	bodyStart := end + len("\n---")
	if strings.HasPrefix(content[bodyStart:], "\n") {
		bodyStart++
	}
	return frontmatter, content[bodyStart:], true
}

func updateSkillMD(content, version, date, change string) (string, error) {
	frontmatter, body, ok := splitFrontmatter(content)
	if !ok {
		return "", fmt.Errorf("missing YAML frontmatter")
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return "", err
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal([]byte(frontmatter), &raw); err != nil {
		return "", err
	}
	changelog := append([]spec.SkillMDChangelogEntry{{
		Version: version,
		Date:    date,
		Change:  change,
	}}, fm.Changelog...)
	raw["version"] = version
	raw["last_modified"] = date
	raw["changelog"] = changelog
	if _, ok := raw["deprecated_since"]; !ok {
		raw["deprecated_since"] = nil
	}
	if _, ok := raw["replaces"]; !ok {
		raw["replaces"] = nil
	}
	if _, ok := raw["supersedes"]; !ok {
		raw["supersedes"] = []string{}
	}
	fmBytes, err := yaml.Marshal(raw)
	if err != nil {
		return "", err
	}
	body = prependBodyChangelog(body, version, date, change)
	return "---\n" + string(fmBytes) + "---\n" + body, nil
}

func prependBodyChangelog(body, version, date, change string) string {
	match := bodyChangelogRE.FindStringIndex(body)
	if match == nil {
		return body + fmt.Sprintf("\n## Changelog\n\n### %s - %s\n\n- %s\n", version, date, change)
	}
	insertAt := match[1]
	rest := body[insertAt:]
	if strings.HasPrefix(rest, "\n") {
		insertAt++
	}
	entry := fmt.Sprintf("\n### %s - %s\n\n- %s\n", version, date, change)
	return body[:insertAt] + entry + body[insertAt:]
}

func nextVersion(current string, kind BumpKind) (string, error) {
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("current version is not x.y.z: %s", current)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", err
	}
	switch kind {
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	case BumpMinor:
		minor++
		patch = 0
	case BumpPatch, "":
		patch++
	default:
		return "", fmt.Errorf("unsupported bump kind: %s", kind)
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func updateTextFileIfExists(path, content string) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func updateSkillYAMLIfExists(path, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return err
	}
	raw["version"] = version
	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func prependChangelogIfExists(path, version, date, change string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	updated := ensureChangelogEntry(string(content), version, date, change)
	return os.WriteFile(path, []byte(updated), 0o644)
}

func syncChangelogIfExists(path, version, date, change string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	updated := ensureChangelogEntry(string(content), version, date, change)
	return os.WriteFile(path, []byte(updated), 0o644)
}

func ensureChangelogEntry(content, version, date, change string) string {
	entry := fmt.Sprintf("## %s - %s\n\n- %s\n\n", version, date, change)
	if strings.HasPrefix(content, "# Changelog\n\n") {
		body := content[len("# Changelog\n\n"):]
		if strings.HasPrefix(body, entry) {
			return content
		}
		body = removeFirstChangelogEntryForVersion(body, version)
		return "# Changelog\n\n" + entry + strings.TrimLeft(body, "\n")
	}
	body := removeFirstChangelogEntryForVersion(content, version)
	return "# Changelog\n\n" + entry + strings.TrimLeft(body, "\n")
}

func removeFirstChangelogEntryForVersion(body, version string) string {
	matches := changelogHeadingRE.FindAllStringSubmatchIndex(body, -1)
	if len(matches) == 0 {
		return body
	}
	if body[matches[0][2]:matches[0][3]] != version {
		return body
	}
	end := len(body)
	if len(matches) > 1 {
		end = matches[1][0]
	}
	return body[:matches[0][0]] + body[end:]
}

func artifactConsistencyErrors(skillPath, version string) []string {
	if version == "" {
		return nil
	}
	dir := filepath.Dir(skillPath)
	errors := []string{}
	if got, ok, err := versionFileVersion(filepath.Join(dir, "VERSION")); err != nil {
		errors = append(errors, fmt.Sprintf("VERSION unreadable: %v", err))
	} else if ok && got != version {
		errors = append(errors, fmt.Sprintf("VERSION %q does not match SKILL.md version %q", got, version))
	}
	if got, ok, err := skillYAMLVersion(filepath.Join(dir, "skill.yaml")); err != nil {
		errors = append(errors, fmt.Sprintf("skill.yaml version unreadable: %v", err))
	} else if ok && got != version {
		errors = append(errors, fmt.Sprintf("skill.yaml version %q does not match SKILL.md version %q", got, version))
	}
	if got, ok, err := changelogVersion(filepath.Join(dir, "CHANGELOG.md")); err != nil {
		errors = append(errors, fmt.Sprintf("CHANGELOG.md unreadable: %v", err))
	} else if ok && got != version {
		errors = append(errors, fmt.Sprintf("CHANGELOG.md latest version %q does not match SKILL.md version %q", got, version))
	}
	return errors
}

func versionFileVersion(path string) (string, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(content)), true, nil
}

func skillYAMLVersion(path string) (string, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return "", true, err
	}
	version, _ := raw["version"].(string)
	return strings.TrimSpace(version), true, nil
}

func changelogVersion(path string) (string, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	match := changelogHeadingRE.FindSubmatch(content)
	if match == nil {
		return "", false, nil
	}
	return string(match[1]), true, nil
}

func gitRoot(path string) (string, error) {
	dir := path
	if filepath.Base(path) == "SKILL.md" {
		dir = filepath.Dir(path)
	}
	cmd := gitCommand("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func gitChangedFiles(root, rel string) ([]string, error) {
	cmd := gitCommand("git", "-C", root, "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %s", strings.TrimSpace(string(out)))
	}
	files := []string{}
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[len(parts)-1]
		}
		path = filepath.ToSlash(path)
		if rel != "" && rel != "." && path != rel && !strings.HasPrefix(path, strings.TrimSuffix(rel, "/")+"/") {
			continue
		}
		files = append(files, path)
	}
	sort.Strings(files)
	return files, nil
}

func nearestSkillFile(root, rel string) string {
	dir := filepath.Dir(filepath.Join(root, rel))
	for {
		candidate := filepath.Join(dir, "SKILL.md")
		if _, err := os.Stat(candidate); err == nil {
			out, _ := filepath.Rel(root, candidate)
			return filepath.ToSlash(out)
		}
		if samePath(dir, root) || dir == filepath.Dir(dir) {
			return ""
		}
		dir = filepath.Dir(dir)
	}
}

func normalizeRepoRelPath(root, rel string) string {
	if _, err := os.Stat(filepath.Join(root, rel)); err != nil && !strings.HasPrefix(rel, ".") {
		if _, dotErr := os.Stat(filepath.Join(root, "."+rel)); dotErr == nil {
			return "." + rel
		}
	}
	return rel
}

func headVersion(root, relSkillMD string) string {
	cmd := gitCommand("git", "-C", root, "show", "HEAD:"+filepath.ToSlash(relSkillMD))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	frontmatter, _, ok := splitFrontmatter(string(out))
	if !ok {
		return ""
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return ""
	}
	return fm.Version
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	return errA == nil && errB == nil && aa == bb
}
