package skillmeta

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/sklib/spec"
	"gopkg.in/yaml.v3"
)

const (
	DescriptorSchemaVersion  = "2"
	DescriptorFilename       = "descriptor.yaml"
	LegacyDescriptorFilename = "skill.yaml"
)

// DescriptorPath returns the canonical source descriptor when present and
// falls back to the pre-compiler skill.yaml name for backwards compatibility.
func DescriptorPath(skillDir string) string {
	canonical := filepath.Join(skillDir, DescriptorFilename)
	if _, err := os.Stat(canonical); err == nil {
		return canonical
	}
	return filepath.Join(skillDir, LegacyDescriptorFilename)
}

// Descriptor identifies a skill, its goal, and its authoritative contract and
// eval locations. The security boundary itself lives in contract.yaml.
type Descriptor struct {
	SchemaVersion  string          `yaml:"schema_version,omitempty" json:"schema_version,omitempty"`
	Name           string          `yaml:"name" json:"name"`
	Namespace      string          `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Version        string          `yaml:"version" json:"version"`
	Description    string          `yaml:"description" json:"description"`
	Owners         []string        `yaml:"owners,omitempty" json:"owners,omitempty"`
	Ownership      *Ownership      `yaml:"ownership,omitempty" json:"ownership,omitempty"`
	License        string          `yaml:"license,omitempty" json:"license,omitempty"`
	Entrypoint     string          `yaml:"entrypoint" json:"entrypoint"`
	CompatibleWith []spec.Platform `yaml:"compatible_with" json:"compatible_with"`
	Tags           []string        `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Security is retained only to read legacy descriptor hints without data
	// loss. It is deprecated and MUST NOT be used to authorize behavior;
	// contract.yaml is the sole behavioral/security boundary.
	Security     *spec.SkillSecurity    `yaml:"security,omitempty" json:"security,omitempty"`
	Metadata     map[string]string      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Goal         *Goal                  `yaml:"goal,omitempty" json:"goal,omitempty"`
	Contract     *ContractReference     `yaml:"contract,omitempty" json:"contract,omitempty"`
	Evals        *EvalsReference        `yaml:"evals,omitempty" json:"evals,omitempty"`
	Integrations *IntegrationReferences `yaml:"integrations,omitempty" json:"integrations,omitempty"`
	Dependencies *ArtifactReference     `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Assurance    *ArtifactReference     `yaml:"assurance,omitempty" json:"assurance,omitempty"`

	// LegacyEmbeddedContract is populated only when parsing the short-lived
	// descriptor-v2 format that embedded the whole contract in skill.yaml.
	// It is never serialized and retains non-destructive compatibility.
	LegacyEmbeddedContract *LegacyContract `yaml:"-" json:"-"`
}

type Ownership struct {
	Primary     string   `yaml:"primary" json:"primary"`
	Maintainers []string `yaml:"maintainers" json:"maintainers"`
	Publisher   string   `yaml:"publisher,omitempty" json:"publisher,omitempty"`
}

type Goal struct {
	Objective         string          `yaml:"objective" json:"objective"`
	SuccessCriteria   []GoalCriterion `yaml:"success_criteria" json:"success_criteria"`
	FailureConditions []GoalCriterion `yaml:"failure_conditions" json:"failure_conditions"`
}

type GoalCriterion struct {
	ID          string `yaml:"id" json:"id"`
	Description string `yaml:"description" json:"description"`
}

// UnmarshalYAML accepts the earlier v2 string criteria for compatibility.
func (c *GoalCriterion) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		c.Description = node.Value
		return nil
	}
	type plain GoalCriterion
	var decoded plain
	if err := node.Decode(&decoded); err != nil {
		return err
	}
	*c = GoalCriterion(decoded)
	return nil
}

type ContractReference struct {
	File string `yaml:"file" json:"file"`
}

type EvalsReference struct {
	Directory string `yaml:"directory" json:"directory"`
}

type ArtifactReference struct {
	File string `yaml:"file" json:"file"`
}

type IntegrationReferences struct {
	MCP ArtifactReference `yaml:"mcp" json:"mcp"`
	A2A ArtifactReference `yaml:"a2a" json:"a2a"`
}

func NewDescriptor(name, version, description, license string, owners []string, platforms []spec.Platform) Descriptor {
	objective := strings.TrimSpace(description)
	if objective == "" || objective == "Describe what this skill helps an agent do." {
		objective = "TODO: Define the outcome this skill is expected to achieve."
	}
	return Descriptor{
		SchemaVersion:  DescriptorSchemaVersion,
		Name:           name,
		Version:        version,
		Description:    description,
		Ownership:      newOwnership(owners),
		License:        license,
		Entrypoint:     spec.DefaultEntrypointValue,
		CompatibleWith: platforms,
		Goal: &Goal{
			Objective:         objective,
			SuccessCriteria:   []GoalCriterion{},
			FailureConditions: []GoalCriterion{},
		},
		Contract:     &ContractReference{File: "contract.yaml"},
		Evals:        &EvalsReference{Directory: "evals"},
		Integrations: &IntegrationReferences{MCP: ArtifactReference{File: "integrations/mcp.yaml"}, A2A: ArtifactReference{File: "integrations/a2a.yaml"}},
		Dependencies: &ArtifactReference{File: "dependencies.yaml"},
		Assurance:    &ArtifactReference{File: "assurance.yaml"},
	}
}

func newOwnership(owners []string) *Ownership {
	if len(owners) == 0 {
		return &Ownership{Maintainers: []string{}}
	}
	return &Ownership{Primary: owners[0], Maintainers: append([]string{}, owners[1:]...)}
}

func ParseDescriptor(data []byte) (Descriptor, error) {
	var header struct {
		SchemaVersion string `yaml:"schema_version"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return Descriptor{}, err
	}
	if header.SchemaVersion == "" {
		var legacy Descriptor
		if err := yaml.Unmarshal(data, &legacy); err != nil {
			return Descriptor{}, err
		}
		return legacy, nil
	}
	if header.SchemaVersion != DescriptorSchemaVersion {
		var unsupported Descriptor
		if err := yaml.Unmarshal(data, &unsupported); err != nil {
			return Descriptor{}, err
		}
		return unsupported, nil
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Descriptor{}, err
	}
	contractNode := mappingValue(&root, "contract")
	if contractNode != nil && mappingValue(contractNode, "file") == nil {
		// Compatibility with the previously generated embedded v2 contract.
		var wire struct {
			SchemaVersion  string              `yaml:"schema_version"`
			Name           string              `yaml:"name"`
			Namespace      string              `yaml:"namespace,omitempty"`
			Version        string              `yaml:"version"`
			Description    string              `yaml:"description"`
			Owners         []string            `yaml:"owners,omitempty"`
			License        string              `yaml:"license,omitempty"`
			Entrypoint     string              `yaml:"entrypoint"`
			CompatibleWith []spec.Platform     `yaml:"compatible_with"`
			Tags           []string            `yaml:"tags,omitempty"`
			Security       *spec.SkillSecurity `yaml:"security,omitempty"`
			Metadata       map[string]string   `yaml:"metadata,omitempty"`
			Goal           *Goal               `yaml:"goal,omitempty"`
			Contract       *LegacyContract     `yaml:"contract"`
			Evals          *struct {
				Directory string `yaml:"directory"`
				Baseline  string `yaml:"baseline,omitempty"`
			} `yaml:"evals,omitempty"`
		}
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(&wire); err != nil {
			return Descriptor{}, err
		}
		descriptor := Descriptor{
			SchemaVersion: wire.SchemaVersion, Name: wire.Name, Namespace: wire.Namespace,
			Version: wire.Version, Description: wire.Description, Owners: wire.Owners,
			License: wire.License, Entrypoint: wire.Entrypoint, CompatibleWith: wire.CompatibleWith,
			Tags: wire.Tags, Security: wire.Security, Metadata: wire.Metadata, Goal: wire.Goal,
			LegacyEmbeddedContract: wire.Contract,
		}
		if wire.Evals != nil {
			descriptor.Evals = &EvalsReference{Directory: wire.Evals.Directory}
		}
		return descriptor, nil
	}

	var descriptor Descriptor
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&descriptor); err != nil {
		return Descriptor{}, err
	}
	return descriptor, nil
}

func LoadDescriptor(path string) (Descriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Descriptor{}, err
	}
	descriptor, err := ParseDescriptor(data)
	if err != nil {
		return Descriptor{}, fmt.Errorf("%s: %w", path, err)
	}
	return descriptor, nil
}

// Parse is retained for callers of the initial structured-artifact
// implementation.
func Parse(data []byte) (Descriptor, error) { return ParseDescriptor(data) }

// Load is retained for callers of the initial structured-artifact
// implementation.
func Load(path string) (Descriptor, error) { return LoadDescriptor(path) }

func ValidateDirectory(skillDir string) []string {
	descriptorPath := DescriptorPath(skillDir)
	descriptor, err := LoadDescriptor(descriptorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"missing descriptor.yaml (legacy skill.yaml is also accepted)"}
		}
		return []string{fmt.Sprintf("malformed %s: %v", filepath.Base(descriptorPath), err)}
	}
	errs := ValidateDescriptor(descriptor, skillDir)
	var currentContract *Contract
	if descriptor.SchemaVersion == "" {
		return errs
	}
	if descriptor.LegacyEmbeddedContract != nil {
		errs = append(errs, ValidateLegacyContract(*descriptor.LegacyEmbeddedContract)...)
	} else if descriptor.Contract != nil {
		contractPath, safe := safeArtifactPath(skillDir, descriptor.Contract.File)
		if safe {
			contract, loadErr := LoadContract(contractPath)
			if loadErr != nil {
				if os.IsNotExist(loadErr) {
					errs = append(errs, fmt.Sprintf("referenced contract %q does not exist", descriptor.Contract.File))
				} else {
					errs = append(errs, fmt.Sprintf("malformed contract %q: %v", descriptor.Contract.File, loadErr))
				}
			} else {
				errs = append(errs, ValidateContract(contract)...)
				currentContract = &contract
			}
		}
	}
	if descriptor.Evals != nil && descriptor.LegacyEmbeddedContract == nil {
		errs = append(errs, ValidateEvalDirectory(skillDir, descriptor.Evals.Directory, descriptor, currentContract)...)
	}
	if descriptor.LegacyEmbeddedContract == nil {
		errs = append(errs, ValidatePhase3Artifacts(skillDir, descriptor, currentContract)...)
		if descriptor.Assurance != nil {
			if assurancePath, safe := safeArtifactPath(skillDir, descriptor.Assurance.File); safe {
				assurance, loadErr := LoadAssurance(assurancePath)
				if loadErr != nil {
					errs = append(errs, "malformed assurance requirements: "+loadErr.Error())
				} else if evalDocuments, evalErr := LoadEvalDocuments(skillDir, descriptor); evalErr != nil {
					errs = append(errs, "load evals for assurance requirements: "+evalErr.Error())
				} else {
					errs = append(errs, ValidateAssurance(assurance, evalDocuments)...)
				}
			}
		}
	}
	return errs
}

func ValidateDescriptor(d Descriptor, skillDir string) []string {
	if d.SchemaVersion == "" {
		return nil
	}
	if d.SchemaVersion != DescriptorSchemaVersion {
		return []string{fmt.Sprintf("unsupported descriptor schema_version %q", d.SchemaVersion)}
	}
	var errs []string
	if err := spec.ValidateSkillName(d.Name); err != nil {
		errs = append(errs, fmt.Sprintf("name %q is invalid", d.Name))
	} else if filepath.Base(skillDir) != d.Name {
		errs = append(errs, fmt.Sprintf("name %q does not match directory %q", d.Name, filepath.Base(skillDir)))
	}
	if _, err := spec.NormalizeVersion(d.Version); err != nil {
		errs = append(errs, fmt.Sprintf("version %q is not valid semver", d.Version))
	}
	if err := spec.ValidateNamespace(d.Namespace); err != nil {
		errs = append(errs, fmt.Sprintf("namespace %q is invalid", d.Namespace))
	}
	if strings.TrimSpace(d.Description) == "" {
		errs = append(errs, "description must not be blank")
	}
	if d.Ownership != nil {
		if strings.TrimSpace(d.Ownership.Primary) == "" {
			errs = append(errs, "ownership.primary must not be blank")
		}
		errs = append(errs, validateStringList("ownership.maintainers", d.Ownership.Maintainers, false)...)
	} else if len(d.Owners) == 0 {
		errs = append(errs, "ownership.primary is required")
	}
	if strings.TrimSpace(d.Entrypoint) == "" {
		errs = append(errs, "entrypoint must not be blank")
	}
	if d.CompatibleWith == nil {
		errs = append(errs, "compatible_with is required and must be a list")
	} else {
		for _, platform := range d.CompatibleWith {
			if _, err := models.NormalizePlatform(string(platform)); err != nil {
				errs = append(errs, fmt.Sprintf("compatible_with contains unsupported platform %q", platform))
			}
		}
	}
	if d.Goal == nil {
		errs = append(errs, "goal is required")
	} else {
		if strings.TrimSpace(d.Goal.Objective) == "" {
			errs = append(errs, "goal.objective must not be blank")
		}
		requireIDs := d.LegacyEmbeddedContract == nil
		errs = append(errs, validateCriteria("goal.success_criteria", d.Goal.SuccessCriteria, requireIDs)...)
		errs = append(errs, validateCriteria("goal.failure_conditions", d.Goal.FailureConditions, requireIDs)...)
	}
	if d.LegacyEmbeddedContract == nil {
		if d.Contract == nil {
			errs = append(errs, "contract reference is required")
		} else if _, safe := safeArtifactPath(skillDir, d.Contract.File); !safe {
			errs = append(errs, "contract.file must reference a file within the skill root")
		}
	}
	if d.Evals == nil {
		errs = append(errs, "evals reference is required")
	} else if _, safe := safeArtifactPath(skillDir, d.Evals.Directory); !safe {
		errs = append(errs, "evals.directory must be within the skill root")
	}
	if d.LegacyEmbeddedContract == nil {
		if d.Integrations == nil {
			errs = append(errs, "integrations references are required")
		} else {
			for name, reference := range map[string]ArtifactReference{"mcp": d.Integrations.MCP, "a2a": d.Integrations.A2A} {
				if _, safe := safeArtifactPath(skillDir, reference.File); strings.TrimSpace(reference.File) == "" || !safe {
					errs = append(errs, "integrations."+name+".file must reference a file within the skill root")
				}
			}
		}
		if d.Dependencies == nil {
			errs = append(errs, "dependencies reference is required")
		} else {
			if _, safe := safeArtifactPath(skillDir, d.Dependencies.File); strings.TrimSpace(d.Dependencies.File) == "" || !safe {
				errs = append(errs, "dependencies.file must reference a file within the skill root")
			}
		}
		if d.Assurance == nil {
			errs = append(errs, "assurance reference is required")
		} else if _, safe := safeArtifactPath(skillDir, d.Assurance.File); strings.TrimSpace(d.Assurance.File) == "" || !safe {
			errs = append(errs, "assurance.file must reference a file within the skill root")
		}
	}
	return errs
}

func validateCriteria(path string, criteria []GoalCriterion, requireID bool) []string {
	if criteria == nil {
		return []string{path + " is required and must be a list"}
	}
	var errs []string
	seen := map[string]struct{}{}
	for i, criterion := range criteria {
		if strings.TrimSpace(criterion.Description) == "" {
			errs = append(errs, fmt.Sprintf("%s[%d].description must not be blank", path, i))
		}
		if criterion.ID == "" {
			if requireID {
				errs = append(errs, fmt.Sprintf("%s[%d].id is required", path, i))
			}
			continue
		}
		if !validIdentifier(criterion.ID) {
			errs = append(errs, fmt.Sprintf("%s[%d].id %q is invalid", path, i, criterion.ID))
		} else if _, duplicate := seen[criterion.ID]; duplicate {
			errs = append(errs, fmt.Sprintf("%s contains duplicate id %q", path, criterion.ID))
		}
		seen[criterion.ID] = struct{}{}
	}
	return errs
}

func mappingValue(root *yaml.Node, key string) *yaml.Node {
	if root == nil {
		return nil
	}
	node := root
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func safeArtifactPath(root, reference string) (string, bool) {
	if strings.TrimSpace(reference) == "" || filepath.IsAbs(reference) {
		return "", false
	}
	clean := filepath.Clean(reference)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	path := filepath.Join(root, clean)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if _, err := os.Lstat(path); err == nil {
		resolvedRoot, rootErr := filepath.EvalSymlinks(absRoot)
		resolvedPath, pathErr := filepath.EvalSymlinks(absPath)
		if rootErr != nil || pathErr != nil {
			return "", false
		}
		resolvedRel, relErr := filepath.Rel(resolvedRoot, resolvedPath)
		if relErr != nil || resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(filepath.Separator)) {
			return "", false
		}
	}
	return path, true
}
