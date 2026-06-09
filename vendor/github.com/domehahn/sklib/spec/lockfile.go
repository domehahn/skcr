package spec

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// AgentSkillsLock is the Go model for agent-skills.lock.yaml.
type AgentSkillsLock struct {
	Version     int               `json:"version" yaml:"version"`
	GeneratedBy string            `json:"generated_by,omitempty" yaml:"generated_by,omitempty"`
	ResolvedAt  string            `json:"resolved_at,omitempty" yaml:"resolved_at,omitempty"`
	Skills      []LockedSkill     `json:"skills" yaml:"skills"`
}

// LockedSkill is a fully resolved skill entry in the lockfile.
type LockedSkill struct {
	Namespace      string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name           string            `json:"name" yaml:"name"`
	Version        string            `json:"version" yaml:"version"`
	Constraint     string            `json:"constraint,omitempty" yaml:"constraint,omitempty"`
	Source         string            `json:"source" yaml:"source"`
	RegistryType   string            `json:"registry_type,omitempty" yaml:"registry_type,omitempty"`
	RegistryURL    string            `json:"registry_url,omitempty" yaml:"registry_url,omitempty"`
	Artifact       string            `json:"artifact,omitempty" yaml:"artifact,omitempty"`
	DownloadURL    string            `json:"download_url,omitempty" yaml:"download_url,omitempty"`
	SHA256         string            `json:"sha256" yaml:"sha256"`
	PackageType    string            `json:"package_type,omitempty" yaml:"package_type,omitempty"`
	CompatibleWith []Platform        `json:"compatible_with,omitempty" yaml:"compatible_with,omitempty"`
	InstalledTo    []string          `json:"installed_to,omitempty" yaml:"installed_to,omitempty"`
	SourceCommit   string            `json:"source_commit,omitempty" yaml:"source_commit,omitempty"`
	Signature      string            `json:"signature,omitempty" yaml:"signature,omitempty"`
	Provenance     string            `json:"provenance,omitempty" yaml:"provenance,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ValidateSHA256 returns an error if value is not a valid lowercase hex SHA-256 digest (64 chars).
func ValidateSHA256(value string) error {
	if value == "" {
		return fmt.Errorf("sha256 must not be empty")
	}
	if len(value) != 64 {
		return fmt.Errorf("sha256 must be 64 hex characters, got %d", len(value))
	}
	if value != strings.ToLower(value) {
		return fmt.Errorf("sha256 must be lowercase hex")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("sha256 is not valid hex: %w", err)
	}
	return nil
}

// ValidateLockBasic performs structural validation of an AgentSkillsLock.
// Locked versions must be exact SemVer (stable or pre-release), not ranges or "latest".
func ValidateLockBasic(lock AgentSkillsLock) error {
	for i, s := range lock.Skills {
		if s.Name == "" {
			return fmt.Errorf("skills[%d]: name must not be empty", i)
		}
		if s.Version == "" {
			return fmt.Errorf("skills[%d] %q: version must not be empty", i, s.Name)
		}
		if !IsSemVer(s.Version) {
			return fmt.Errorf("skills[%d] %q: version %q must be exact SemVer (no ranges or 'latest')", i, s.Name, s.Version)
		}
		if s.Source == "" {
			return fmt.Errorf("skills[%d] %q: source must not be empty", i, s.Name)
		}
		if err := ValidateSHA256(s.SHA256); err != nil {
			return fmt.Errorf("skills[%d] %q: %w", i, s.Name, err)
		}
	}
	return nil
}
