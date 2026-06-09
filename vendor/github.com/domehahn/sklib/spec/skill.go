package spec

import (
	"fmt"
	"regexp"
)

var (
	skillNameRe  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	namespaceRe  = regexp.MustCompile(`^[a-z0-9]+(?:[._\-][a-z0-9]+)*$`)
)

const (
	DefaultNamespaceValue  = "default"
	DefaultEntrypointValue = "SKILL.md"
)

// Skill represents the canonical model for a skill.yaml file.
type Skill struct {
	Name           string            `json:"name" yaml:"name"`
	Namespace      string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Version        string            `json:"version" yaml:"version"`
	Description    string            `json:"description" yaml:"description"`
	Owners         []string          `json:"owners,omitempty" yaml:"owners,omitempty"`
	License        string            `json:"license,omitempty" yaml:"license,omitempty"`
	Entrypoint     string            `json:"entrypoint" yaml:"entrypoint"`
	CompatibleWith []Platform        `json:"compatible_with" yaml:"compatible_with"`
	Tags           []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Security       SkillSecurity     `json:"security,omitempty" yaml:"security,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SkillSecurity captures which elevated capabilities a skill requires.
type SkillSecurity struct {
	RequiresNetwork bool `json:"requires_network" yaml:"requires_network"`
	RequiresSecrets bool `json:"requires_secrets" yaml:"requires_secrets"`
	WritesFiles     bool `json:"writes_files" yaml:"writes_files"`
	RunsCommands    bool `json:"runs_commands" yaml:"runs_commands"`
}

// ValidateSkillName returns an error if name does not satisfy the skill name rules.
func ValidateSkillName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if !skillNameRe.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return nil
}

// ValidateNamespace returns an error if namespace is non-empty and does not satisfy the namespace rules.
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return nil
	}
	if !namespaceRe.MatchString(namespace) {
		return fmt.Errorf("%w: %q", ErrInvalidNamespace, namespace)
	}
	return nil
}

// DefaultNamespace returns the given namespace, or "default" if empty.
func DefaultNamespace(namespace string) string {
	if namespace == "" {
		return DefaultNamespaceValue
	}
	return namespace
}

// DefaultEntrypoint returns the given entrypoint, or "SKILL.md" if empty.
func DefaultEntrypoint(entrypoint string) string {
	if entrypoint == "" {
		return DefaultEntrypointValue
	}
	return entrypoint
}

// Ref builds a SkillRef from the skill's namespace and name.
// The namespace defaults to "default" if unset.
func (s Skill) Ref() SkillRef {
	return SkillRef{
		Namespace:  DefaultNamespace(s.Namespace),
		Name:       s.Name,
		Constraint: s.Version,
	}
}

// SkillRef is a parsed reference to a skill, optionally with a version constraint.
type SkillRef struct {
	Namespace  string `json:"namespace" yaml:"namespace"`
	Name       string `json:"name" yaml:"name"`
	Constraint string `json:"constraint,omitempty" yaml:"constraint,omitempty"`
}

// SkillVersionRef is a reference to an exact resolved version of a skill.
type SkillVersionRef struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Version   string `json:"version" yaml:"version"`
}

// SupportsPlatform returns true if the skill lists the given platform in CompatibleWith,
// or if it lists PlatformAll.
func (s Skill) SupportsPlatform(platform Platform) bool {
	for _, p := range s.CompatibleWith {
		if p == PlatformAll || p == platform {
			return true
		}
	}
	return false
}
