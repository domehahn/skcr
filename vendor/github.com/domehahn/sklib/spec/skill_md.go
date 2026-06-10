package spec

// SkillMDFrontmatter represents the YAML frontmatter block at the top of a SKILL.md file.
// All fields marked as required must be present; optional fields may be empty or nil.
//
// Canonical location: SKILL.md (the human- and agent-readable skill entrypoint).
// The machine-readable counterpart is skill.yaml (see Skill).
type SkillMDFrontmatter struct {
	// Required fields.
	Name           string                     `yaml:"name"`
	Description    string                     `yaml:"description"`
	Version        string                     `yaml:"version"`
	Since          string                     `yaml:"since"`
	LastModified   string                     `yaml:"last_modified"`
	Authors        []string                   `yaml:"authors"`
	Stability      SkillStability             `yaml:"stability"`
	MinPlatformVer map[string]string          `yaml:"min_platform_version"`
	Changelog      []SkillMDChangelogEntry    `yaml:"changelog"`

	// Optional fields.
	DeprecatedSince string   `yaml:"deprecated_since,omitempty"`
	Replaces        string   `yaml:"replaces,omitempty"`
	Supersedes      []string `yaml:"supersedes,omitempty"`
}

// SkillMDChangelogEntry is one entry in the SKILL.md frontmatter changelog list.
type SkillMDChangelogEntry struct {
	Version string `yaml:"version"`
	Date    string `yaml:"date"`
	Change  string `yaml:"change"`
}

// SkillStability describes the production-readiness of a skill.
type SkillStability string

const (
	StabilityExperimental SkillStability = "experimental"
	StabilityStable       SkillStability = "stable"
	StabilityDeprecated   SkillStability = "deprecated"
)

// ValidateSkillMDFrontmatter checks that a parsed SkillMDFrontmatter satisfies
// the mandatory versioning schema. Returns a slice of human-readable errors.
func ValidateSkillMDFrontmatter(fm SkillMDFrontmatter) []string {
	var errs []string

	if fm.Name == "" {
		errs = append(errs, "missing required field: name")
	}
	if fm.Version == "" {
		errs = append(errs, "missing required field: version")
	} else if !IsSemVer(fm.Version) {
		errs = append(errs, "version is not valid SemVer (use x.y.z without leading v): "+fm.Version)
	}
	if fm.Since == "" {
		errs = append(errs, "missing required field: since")
	}
	if fm.LastModified == "" {
		errs = append(errs, "missing required field: last_modified")
	}
	if len(fm.Authors) == 0 {
		errs = append(errs, "missing required field: authors (must have at least one entry)")
	}
	switch fm.Stability {
	case StabilityExperimental, StabilityStable, StabilityDeprecated:
		// valid
	case "":
		errs = append(errs, "missing required field: stability")
	default:
		errs = append(errs, "invalid stability value: "+string(fm.Stability)+` (must be "experimental", "stable", or "deprecated")`)
	}
	if fm.Stability == StabilityDeprecated && fm.DeprecatedSince == "" {
		errs = append(errs, "stability is deprecated but deprecated_since is empty")
	}
	if len(fm.MinPlatformVer) == 0 {
		errs = append(errs, "missing required field: min_platform_version (must have at least one entry)")
	}
	if len(fm.Changelog) == 0 {
		errs = append(errs, "missing required field: changelog (must have at least one entry)")
	} else if fm.Version != "" && fm.Changelog[0].Version != fm.Version {
		errs = append(errs, "version mismatch: frontmatter version "+fm.Version+" does not match newest changelog entry "+fm.Changelog[0].Version)
	}

	return errs
}
