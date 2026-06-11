package platforms

import "strings"

type CompatibilityEntry struct {
	Name       string
	MinVersion string
	Status     string
	Source     string
}

var CompatibilityMatrix = []CompatibilityEntry{
	{Name: "codex", MinVersion: "0.51.0", Status: "verified", Source: "minimum-tested"},
	{Name: "claude-code", MinVersion: "1.0.44", Status: "verified", Source: "minimum-tested"},
	{Name: "github-copilot", MinVersion: "1.300.0", Status: "verified", Source: "minimum-tested"},
	{Name: "gitlab-duo", MinVersion: "18.0", Status: "verified", Source: "minimum-tested"},
	{Name: "opencode", MinVersion: "0.6.0", Status: "verified", Source: "minimum-tested"},
	{Name: "openhands", MinVersion: "0.39.0", Status: "verified", Source: "minimum-tested"},
	{Name: "cursor", MinVersion: "1.0.0", Status: "verified", Source: "minimum-tested"},
	{Name: "roo-code", MinVersion: "3.20.0", Status: "verified", Source: "minimum-tested"},
	{Name: "kiro", MinVersion: "0.2.0", Status: "verified", Source: "minimum-tested"},
	{Name: "junie", MinVersion: "2025.2", Status: "verified", Source: "minimum-tested"},
	{Name: "gemini-cli", MinVersion: "0.1.12", Status: "verified", Source: "minimum-tested"},
	{Name: "windsurf", MinVersion: "1.9.0", Status: "verified", Source: "minimum-tested"},
	{Name: "ollama", MinVersion: "0.7.0", Status: "verified", Source: "minimum-tested"},
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
