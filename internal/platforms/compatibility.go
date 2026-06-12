package platforms

import "strings"

type CompatibilityEntry struct {
	Name       string
	MinVersion string
	Status     string
	Source     string
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
