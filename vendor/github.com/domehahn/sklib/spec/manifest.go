package spec

import (
	"fmt"
	"strings"
)

// AgentSkillsManifest is the Go model for agent-skills.yaml.
type AgentSkillsManifest struct {
	Version         int                       `json:"version" yaml:"version"`
	DefaultRegistry string                    `json:"default_registry,omitempty" yaml:"default_registry,omitempty"`
	Registries      map[string]RegistryConfig `json:"registries,omitempty" yaml:"registries,omitempty"`
	Skills          []ManifestSkill           `json:"skills" yaml:"skills"`
	Metadata        map[string]string         `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ManifestSkill is an entry in the agent-skills.yaml skills list.
type ManifestSkill struct {
	Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string            `json:"name" yaml:"name"`
	Version   string            `json:"version,omitempty" yaml:"version,omitempty"`
	Source    string            `json:"source,omitempty" yaml:"source,omitempty"`
	Target    string            `json:"target,omitempty" yaml:"target,omitempty"`
	Platforms []Platform        `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	Path      string            `json:"path,omitempty" yaml:"path,omitempty"`
	Ref       string            `json:"ref,omitempty" yaml:"ref,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// RegistryConfig describes a registry entry in agent-skills.yaml.
type RegistryConfig struct {
	Type      string            `json:"type" yaml:"type"`
	URL       string            `json:"url,omitempty" yaml:"url,omitempty"`
	Repo      string            `json:"repo,omitempty" yaml:"repo,omitempty"`
	Project   string            `json:"project,omitempty" yaml:"project,omitempty"`
	Path      string            `json:"path,omitempty" yaml:"path,omitempty"`
	Auth      AuthConfig        `json:"auth,omitempty" yaml:"auth,omitempty"`
	Headers   map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Endpoints EndpointConfig    `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// AuthConfig holds authentication configuration for a registry.
type AuthConfig struct {
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Token       string `json:"token,omitempty" yaml:"token,omitempty"`
	TokenEnv    string `json:"token_env,omitempty" yaml:"token_env,omitempty"`
	Username    string `json:"username,omitempty" yaml:"username,omitempty"`
	UsernameEnv string `json:"username_env,omitempty" yaml:"username_env,omitempty"`
	Password    string `json:"password,omitempty" yaml:"password,omitempty"`
	PasswordEnv string `json:"password_env,omitempty" yaml:"password_env,omitempty"`
	HeaderName  string `json:"header_name,omitempty" yaml:"header_name,omitempty"`
}

// EndpointConfig overrides individual API endpoint paths for a registry.
type EndpointConfig struct {
	Capabilities string `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Search       string `json:"search,omitempty" yaml:"search,omitempty"`
	Info         string `json:"info,omitempty" yaml:"info,omitempty"`
	Versions     string `json:"versions,omitempty" yaml:"versions,omitempty"`
	Resolve      string `json:"resolve,omitempty" yaml:"resolve,omitempty"`
	Download     string `json:"download,omitempty" yaml:"download,omitempty"`
	Publish      string `json:"publish,omitempty" yaml:"publish,omitempty"`
}

// KnownRegistryTypes lists all supported registry type identifiers.
var KnownRegistryTypes = []string{
	"skillforge",
	"artifactory",
	"gitlab",
	"github",
	"local",
	"generic-http",
}

// IsKnownRegistryType returns true if t is a recognized registry type.
func IsKnownRegistryType(t string) bool {
	lower := strings.ToLower(t)
	for _, k := range KnownRegistryTypes {
		if k == lower {
			return true
		}
	}
	return false
}

// ValidateManifestBasic performs structural validation of an AgentSkillsManifest.
// It checks for duplicate skills, valid registry types, and required skill names.
func ValidateManifestBasic(m AgentSkillsManifest) error {
	for alias, reg := range m.Registries {
		if reg.Type == "" {
			return fmt.Errorf("registry %q: type must not be empty", alias)
		}
		if !IsKnownRegistryType(reg.Type) {
			return fmt.Errorf("registry %q: unknown type %q", alias, reg.Type)
		}
	}

	type key struct{ ns, name string }
	seen := make(map[key]bool)
	for i, s := range m.Skills {
		if s.Name == "" {
			return fmt.Errorf("skills[%d]: name must not be empty", i)
		}
		k := key{ns: DefaultNamespace(s.Namespace), name: s.Name}
		if seen[k] {
			return fmt.Errorf("duplicate skill %q/%q in manifest", k.ns, k.name)
		}
		seen[k] = true
	}

	return nil
}

// NormalizeManifestDefaults fills in default values for fields that have canonical defaults.
func NormalizeManifestDefaults(m *AgentSkillsManifest) {
	if m.Version == 0 {
		m.Version = 1
	}
	for i := range m.Skills {
		if m.Skills[i].Namespace == "" {
			m.Skills[i].Namespace = DefaultNamespaceValue
		}
	}
}
