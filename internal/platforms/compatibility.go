package platforms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type CompatibilityEntry struct {
	Name       string `json:"name" yaml:"name"`
	MinVersion string `json:"min_version" yaml:"min_version"`
	Status     string `json:"status" yaml:"status"`
	Source     string `json:"source" yaml:"source"`
	Evidence   string `json:"evidence,omitempty" yaml:"evidence,omitempty"`
	Validated  string `json:"validated,omitempty" yaml:"validated,omitempty"`
	Notes      string `json:"notes,omitempty" yaml:"notes,omitempty"`
}

var CompatibilityMatrix = []CompatibilityEntry{
	{Name: "codex", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "amazon-q", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "antigravity", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "auggie", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "bob", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "claude-code", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "cline", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "codebuddy", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "continue", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "costrict", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "crush", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "github-copilot", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "gitlab-duo", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "factory", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "forgecode", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "opencode", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "openhands", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "cursor", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "roo-code", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "kiro", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "junie", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "gemini-cli", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "iflow", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "kilocode", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "kimi", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "lingma", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "pi", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "qoder", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "qwen", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "windsurf", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
	{Name: "ollama", MinVersion: "unknown", Status: "unverified", Source: "pending-platform-validation"},
}

func MinVersion(platform string) (string, bool) {
	for _, entry := range CompatibilityMatrix {
		if entry.Name == platform {
			return entry.MinVersion, true
		}
	}
	return "", false
}

func MinVersionOrUnknown(platform string) string {
	if version, ok := MinVersion(platform); ok {
		return version
	}
	return "unknown"
}

func AllMinVersions() []CompatibilityEntry {
	out := make([]CompatibilityEntry, len(CompatibilityMatrix))
	copy(out, CompatibilityMatrix)
	return out
}

func LoadMatrix(root string) ([]CompatibilityEntry, error) {
	entries := AllMinVersions()
	if strings.TrimSpace(root) == "" {
		return entries, nil
	}
	path := filepath.Join(root, "agentic.compatibility.yaml")
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}
	override := struct {
		Platforms []CompatibilityEntry `yaml:"platforms"`
	}{}
	if err := yaml.Unmarshal(payload, &override); err != nil {
		return nil, fmt.Errorf("invalid agentic.compatibility.yaml: %w", err)
	}
	byName := map[string]int{}
	for i, entry := range entries {
		byName[entry.Name] = i
	}
	for _, entry := range override.Platforms {
		if err := ValidateEvidenceEntry(root, entry); err != nil {
			return nil, err
		}
		if idx, ok := byName[entry.Name]; ok {
			entries[idx] = entry
		} else {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func SaveEvidence(root string, entry CompatibilityEntry) error {
	if err := ValidateEvidenceEntry(root, entry); err != nil {
		return err
	}
	path := filepath.Join(root, "agentic.compatibility.yaml")
	matrix, err := LoadMatrix(root)
	if err != nil {
		return err
	}
	found := false
	for i, existing := range matrix {
		if existing.Name == entry.Name {
			matrix[i] = entry
			found = true
			break
		}
	}
	if !found {
		matrix = append(matrix, entry)
	}
	out := struct {
		Platforms []CompatibilityEntry `yaml:"platforms"`
	}{Platforms: verifiedOnly(matrix)}
	payload, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func ValidateEvidenceEntry(root string, entry CompatibilityEntry) error {
	if strings.TrimSpace(entry.Name) == "" {
		return fmt.Errorf("platform name is required")
	}
	if strings.TrimSpace(entry.MinVersion) == "" {
		return fmt.Errorf("min version is required for %s", entry.Name)
	}
	if entry.MinVersion == "unknown" {
		if entry.Status != "unverified" {
			return fmt.Errorf("platform %s with unknown min version must be unverified", entry.Name)
		}
		return nil
	}
	if entry.Status != "verified" {
		return fmt.Errorf("platform %s with concrete min version must be verified", entry.Name)
	}
	if strings.TrimSpace(entry.Evidence) == "" {
		return fmt.Errorf("platform %s with concrete min version requires evidence", entry.Name)
	}
	evidencePath, err := resolveEvidencePath(root, entry.Evidence)
	if err != nil {
		return fmt.Errorf("platform %s invalid evidence path: %w", entry.Name, err)
	}
	if _, err := os.Stat(evidencePath); err != nil {
		return fmt.Errorf("platform %s evidence file not found: %s", entry.Name, entry.Evidence)
	}
	if strings.TrimSpace(entry.Validated) == "" {
		return fmt.Errorf("platform %s with concrete min version requires validated date", entry.Name)
	}
	if _, err := time.Parse("2006-01-02", entry.Validated); err != nil {
		return fmt.Errorf("platform %s validated date must use YYYY-MM-DD: %s", entry.Name, entry.Validated)
	}
	return nil
}

func resolveEvidencePath(root, evidence string) (string, error) {
	if filepath.IsAbs(evidence) {
		return "", fmt.Errorf("must be relative to target")
	}
	clean := filepath.Clean(evidence)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("must stay inside target")
	}
	return filepath.Join(root, clean), nil
}

func verifiedOnly(entries []CompatibilityEntry) []CompatibilityEntry {
	out := []CompatibilityEntry{}
	for _, entry := range entries {
		if entry.Status == "verified" && entry.MinVersion != "unknown" {
			out = append(out, entry)
		}
	}
	return out
}

func MinVersionsFor(platforms []string) []CompatibilityEntry {
	if len(platforms) == 0 {
		return AllMinVersions()
	}
	wanted := map[string]struct{}{}
	for _, platform := range platforms {
		platform = strings.TrimSpace(platform)
		if platform != "" {
			wanted[platform] = struct{}{}
		}
	}
	out := []CompatibilityEntry{}
	for _, entry := range CompatibilityMatrix {
		if _, ok := wanted[entry.Name]; ok {
			out = append(out, entry)
		}
	}
	return out
}
