package spec

import (
	"fmt"
	"strings"
)

// Platform identifies a target AI coding assistant or agent environment.
type Platform string

const (
	PlatformCodex          Platform = "codex"
	PlatformClaudeCode     Platform = "claude-code"
	PlatformGitLabDuo      Platform = "gitlab-duo"
	PlatformGitHubCopilot  Platform = "github-copilot"
	PlatformCursor         Platform = "cursor"
	PlatformWindsurf       Platform = "windsurf"
	PlatformOpenHands      Platform = "openhands"
	PlatformOpenCode       Platform = "opencode"
	PlatformOllama         Platform = "ollama"
	PlatformJunie          Platform = "junie"
	PlatformGeminiCLI      Platform = "gemini-cli"
	PlatformAmp            Platform = "amp"
	PlatformGoose          Platform = "goose"
	PlatformRooCode        Platform = "roo-code"
	PlatformKiro           Platform = "kiro"
	PlatformLetta          Platform = "letta"
	PlatformFirebender     Platform = "firebender"
	PlatformTRAE           Platform = "trae"
	PlatformMux            Platform = "mux"
	PlatformAutohand       Platform = "autohand"
	PlatformFactory        Platform = "factory"
	PlatformTabnine        Platform = "tabnine"
	PlatformQodo           Platform = "qodo"
	PlatformVibe           Platform = "vibe"
	PlatformCommandCode    Platform = "command-code"
	PlatformOna            Platform = "ona"
	PlatformFastAgent      Platform = "fast-agent"
	PlatformNanobot        Platform = "nanobot"
	PlatformBub            Platform = "bub"
	PlatformVita           Platform = "vita"
	PlatformSuperconductor Platform = "superconductor"
	PlatformEmdash         Platform = "emdash"
	PlatformWorkshop       Platform = "workshop"
	PlatformVTCode         Platform = "vt-code"
	PlatformPiebald        Platform = "piebald"
	PlatformSpringAI       Platform = "spring-ai"
	PlatformDatabricks     Platform = "databricks"
	PlatformSnowflake      Platform = "snowflake"
	PlatformLaravelBoost   Platform = "laravel-boost"
	PlatformGoogleAIEdge   Platform = "google-ai-edge"
	PlatformAgentman       Platform = "agentman"
	PlatformGeneric        Platform = "generic"
	PlatformAll            Platform = "all"
)

// knownPlatforms is the ordered set of all concrete canonical platforms (excludes "all").
var knownPlatforms = []Platform{
	PlatformCodex,
	PlatformClaudeCode,
	PlatformGitLabDuo,
	PlatformGitHubCopilot,
	PlatformCursor,
	PlatformWindsurf,
	PlatformOpenHands,
	PlatformOpenCode,
	PlatformOllama,
	PlatformJunie,
	PlatformGeminiCLI,
	PlatformAmp,
	PlatformGoose,
	PlatformRooCode,
	PlatformKiro,
	PlatformLetta,
	PlatformFirebender,
	PlatformTRAE,
	PlatformMux,
	PlatformAutohand,
	PlatformFactory,
	PlatformTabnine,
	PlatformQodo,
	PlatformVibe,
	PlatformCommandCode,
	PlatformOna,
	PlatformFastAgent,
	PlatformNanobot,
	PlatformBub,
	PlatformVita,
	PlatformSuperconductor,
	PlatformEmdash,
	PlatformWorkshop,
	PlatformVTCode,
	PlatformPiebald,
	PlatformSpringAI,
	PlatformDatabricks,
	PlatformSnowflake,
	PlatformLaravelBoost,
	PlatformGoogleAIEdge,
	PlatformAgentman,
	PlatformGeneric,
}

// platformAliases maps non-canonical strings to their canonical Platform value.
var platformAliases = map[string]Platform{
	"gitlab":            PlatformGitLabDuo,
	"duo":               PlatformGitLabDuo,
	"github":            PlatformGitHubCopilot,
	"copilot":           PlatformGitHubCopilot,
	"claude":            PlatformClaudeCode,
	"open-code":         PlatformOpenCode,
	"open-hands":        PlatformOpenHands,
	"gemini":            PlatformGeminiCLI,
	"roo":               PlatformRooCode,
	"laravel":           PlatformLaravelBoost,
	"databricks-genie":  PlatformDatabricks,
	"snowflake-cortex":  PlatformSnowflake,
	"mistral-vibe":      PlatformVibe,
	"jetbrains":         PlatformJunie,
}

// NormalizePlatform resolves a string (including aliases) to a canonical Platform.
func NormalizePlatform(value string) (Platform, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return "", fmt.Errorf("platform value must not be empty")
	}

	if alias, ok := platformAliases[v]; ok {
		return alias, nil
	}

	p := Platform(v)
	if p == PlatformAll {
		return PlatformAll, nil
	}
	for _, known := range knownPlatforms {
		if p == known {
			return known, nil
		}
	}
	return "", fmt.Errorf("unknown platform %q", value)
}

// NormalizePlatforms resolves and deduplicates a slice of platform strings.
// Order is preserved; duplicates are dropped.
func NormalizePlatforms(values []string) ([]Platform, error) {
	seen := make(map[Platform]bool)
	out := make([]Platform, 0, len(values))
	for _, v := range values {
		p, err := NormalizePlatform(v)
		if err != nil {
			return nil, err
		}
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out, nil
}

// IsKnownPlatform returns true if value resolves to a known platform or "all".
func IsKnownPlatform(value string) bool {
	_, err := NormalizePlatform(value)
	return err == nil
}

// ExpandAllPlatforms replaces any occurrence of PlatformAll with all concrete
// canonical platforms. Other values are passed through. The result is deduplicated.
func ExpandAllPlatforms(values []Platform) []Platform {
	seen := make(map[Platform]bool)
	out := make([]Platform, 0, len(values))

	add := func(p Platform) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	for _, v := range values {
		if v == PlatformAll {
			for _, p := range knownPlatforms {
				add(p)
			}
		} else {
			add(v)
		}
	}
	return out
}

// SupportsPlatform returns true if target is in compatible, or if compatible contains PlatformAll.
func SupportsPlatform(compatible []Platform, target Platform) bool {
	for _, p := range compatible {
		if p == PlatformAll || p == target {
			return true
		}
	}
	return false
}

// PlatformStrings converts a slice of Platform values to plain strings.
func PlatformStrings(values []Platform) []string {
	out := make([]string, len(values))
	for i, p := range values {
		out[i] = string(p)
	}
	return out
}
