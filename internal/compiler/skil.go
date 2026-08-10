package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"gopkg.in/yaml.v3"
)

const Target = "skil"

type Options struct {
	OutputRoot      string
	RequireLossless bool
	CompilerVersion string
}

type Result struct {
	SkillName string
	OutputDir string
	Manifest  Manifest
}

type Manifest struct {
	SchemaVersion string          `json:"schema_version"`
	Source        SourceDigests   `json:"source"`
	Target        TargetMetadata  `json:"target"`
	Mapping       Mapping         `json:"mapping"`
	Provenance    BuildProvenance `json:"provenance"`
}

type SourceDigests struct {
	DescriptorDigest   string `json:"descriptor_digest"`
	ContractDigest     string `json:"contract_digest"`
	EvalDigest         string `json:"eval_digest"`
	InstructionsDigest string `json:"instructions_digest"`
	IntegrationsDigest string `json:"integrations_digest,omitempty"`
	DependenciesDigest string `json:"dependencies_digest,omitempty"`
	AssuranceDigest    string `json:"assurance_digest,omitempty"`
}

type BuildProvenance struct {
	Compiler             CompilerIdentity `json:"compiler"`
	SourceArtifactDigest string           `json:"source_artifact_digest"`
	MappingDigest        string           `json:"mapping_digest"`
}
type CompilerIdentity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type TargetMetadata struct {
	Type           string `json:"type"`
	ContractSchema string `json:"contract_schema"`
	EvalSchema     string `json:"eval_schema"`
	ArtifactDigest string `json:"artifact_digest"`
}

type Mapping struct {
	Mapped           []string `json:"mapped"`
	VerificationOnly []string `json:"verification_only"`
	Unsupported      []string `json:"unsupported"`
	Lossy            []string `json:"lossy"`
}

type skilContract struct {
	Version       int               `yaml:"version"`
	Owner         string            `yaml:"owner,omitempty"`
	Entrypoint    string            `yaml:"entrypoint,omitempty"`
	Compatibility skilCompatibility `yaml:"compatibility,omitempty"`
	Security      skilSecurity      `yaml:"security"`
	Skill         skilIdentity      `yaml:"skill"`
	Capabilities  skilCapabilities  `yaml:"capabilities"`
}

type skilCompatibility struct {
	Platforms []string `yaml:"platforms,omitempty"`
}
type skilSecurity struct {
	RequiresNetwork bool `yaml:"requires_network"`
	RequiresSecrets bool `yaml:"requires_secrets"`
	WritesFiles     bool `yaml:"writes_files"`
	RunsCommands    bool `yaml:"runs_commands"`
}
type skilIdentity struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description"`
}
type skilCapabilities struct {
	Filesystem  skilFilesystem  `yaml:"filesystem"`
	Network     skilNetwork     `yaml:"network"`
	Commands    skilCommands    `yaml:"commands"`
	Secrets     skilSecrets     `yaml:"secrets"`
	Environment skilEnvironment `yaml:"environment"`
	Tools       skilTools       `yaml:"tools"`
	MCP         skilMCP         `yaml:"mcp"`
	Persistence bool            `yaml:"persistence"`
	Agent       skilAgent       `yaml:"agent"`
	Resources   *skilResources  `yaml:"resources,omitempty"`
}
type skilFilesystem struct {
	Read   []string `yaml:"read"`
	Write  []string `yaml:"write"`
	Delete []string `yaml:"delete"`
}
type skilNetwork struct {
	Inbound  bool     `yaml:"inbound"`
	Outbound bool     `yaml:"outbound"`
	Hosts    []string `yaml:"hosts"`
}
type skilCommands struct {
	Execute bool     `yaml:"execute"`
	Allow   []string `yaml:"allow"`
}
type skilSecrets struct {
	Read   []string `yaml:"read"`
	Expose bool     `yaml:"expose"`
}
type skilEnvironment struct {
	Read []string `yaml:"read"`
}
type skilTools struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}
type skilMCP struct {
	Servers []string `yaml:"servers"`
	Tools   []string `yaml:"tools"`
}
type skilAgent struct {
	AutonomousActions   bool     `yaml:"autonomous_actions"`
	ExternalSideEffects bool     `yaml:"external_side_effects"`
	ConfirmDestructive  bool     `yaml:"confirm_destructive"`
	ConfirmExternal     bool     `yaml:"confirm_external"`
	ExternalTargets     []string `yaml:"external_targets,omitempty"`
}
type skilResources struct {
	MaxRuntimeSeconds    *int `yaml:"max_runtime_seconds,omitempty"`
	MaxMemoryMB          *int `yaml:"max_memory_mb,omitempty"`
	MaxNetworkBytes      *int `yaml:"max_network_bytes,omitempty"`
	MaxToolCalls         *int `yaml:"max_tool_calls,omitempty"`
	MaxModelTokens       *int `yaml:"max_model_tokens,omitempty"`
	MaxDelegationDepth   *int `yaml:"max_delegation_depth,omitempty"`
	MaxExternalMutations *int `yaml:"max_external_mutations,omitempty"`
}

type skilEval struct {
	Version     int                  `yaml:"version"`
	Name        string               `yaml:"name"`
	Type        string               `yaml:"type"`
	Input       skilEvalInput        `yaml:"input"`
	Context     map[string]any       `yaml:"context,omitempty"`
	Environment map[string]any       `yaml:"environment,omitempty"`
	Tools       skilEvalTools        `yaml:"tools"`
	Expect      skilEvalExpect       `yaml:"expect"`
	Attack      *skilEvalAttack      `yaml:"attack,omitempty"`
	Containment *skilEvalContainment `yaml:"containment,omitempty"`
}
type skilEvalInput struct {
	Message string `yaml:"message"`
}
type skilEvalTools struct {
	Available []string `yaml:"available"`
}
type skilEvalExpect struct {
	Required              []string          `yaml:"required,omitempty"`
	Allowed               []string          `yaml:"allowed,omitempty"`
	Forbidden             []string          `yaml:"forbidden,omitempty"`
	ForbiddenCapabilities []string          `yaml:"forbidden_capabilities,omitempty"`
	Arguments             map[string]string `yaml:"arguments,omitempty"`
	OutputProperties      []string          `yaml:"output_properties,omitempty"`
	Assertions            []string          `yaml:"assertions,omitempty"`
}
type skilEvalAttack struct {
	Category string `yaml:"category"`
}
type skilEvalContainment struct {
	Required               bool                `yaml:"required"`
	AllowedTargets         map[string][]string `yaml:"allowed_targets,omitempty"`
	RequireEnforcement     bool                `yaml:"require_enforcement,omitempty"`
	RequireNativeIsolation bool                `yaml:"require_native_isolation,omitempty"`
}

func CompileSkill(skillDir string, opts Options) (Result, error) {
	descriptorPath := skillmeta.DescriptorPath(skillDir)
	descriptor, err := skillmeta.LoadDescriptor(descriptorPath)
	if err != nil {
		return Result{}, err
	}
	if descriptor.SchemaVersion != skillmeta.DescriptorSchemaVersion || descriptor.LegacyEmbeddedContract != nil {
		return Result{}, fmt.Errorf("%s must be a split descriptor v%s source artifact", descriptorPath, skillmeta.DescriptorSchemaVersion)
	}
	if errs := skillmeta.ValidateDirectory(skillDir); len(errs) > 0 {
		return Result{}, fmt.Errorf("source validation failed: %s", strings.Join(errs, "; "))
	}
	contractPath := filepath.Join(skillDir, descriptor.Contract.File)
	contract, err := skillmeta.LoadContract(contractPath)
	if err != nil {
		return Result{}, err
	}

	mapping := Mapping{
		Mapped:           []string{"descriptor.identity", "descriptor.ownership.primary", "descriptor.entrypoint", "descriptor.compatible_with", "contract.capabilities.runtime.allowed", "contract.resources"},
		VerificationOnly: []string{"descriptor.goal", "contract.capabilities.semantic", "contract.capabilities.runtime.required", "contract.identity", "contract.delegation", "contract.data.policies", "contract.preconditions", "contract.postconditions", "contract.invariants", "contract.effects", "contract.output", "contract.resources.max_network_requests", "contract.resources.max_steps", "contract.resources.max_replans", "eval.goal_refs", "eval.invariant_refs", "integrations.mcp", "integrations.a2a", "dependencies.reviewed_execution_closure", "assurance.asps_requirements", "assurance.minimum_requested_level", "assurance.security_review"},
		Unsupported:      []string{},
		Lossy:            []string{},
	}
	if err := validateLegacySecuritySummary(descriptor, contract); err != nil {
		return Result{}, err
	}
	mapping.Unsupported = unsupportedSecurity(contract)
	if len(mapping.Unsupported) > 0 {
		return Result{}, fmt.Errorf("unsupported security field(s): %s", strings.Join(mapping.Unsupported, ", "))
	}
	if opts.RequireLossless && len(mapping.Lossy) > 0 {
		return Result{}, fmt.Errorf("lossless compilation unavailable; lossy field(s): %s", strings.Join(mapping.Lossy, ", "))
	}

	compiled := mapContract(descriptor, contract)
	contractBytes, err := yaml.Marshal(compiled)
	if err != nil {
		return Result{}, err
	}
	evals, evalDigest, err := compileEvals(skillDir, descriptor, contract)
	if err != nil {
		return Result{}, err
	}
	instructions, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return Result{}, err
	}
	descriptorBytes, _ := os.ReadFile(descriptorPath)
	contractSource, _ := os.ReadFile(contractPath)
	integrationsDigest, dependenciesDigest, assuranceDigest, err := extendedSourceDigests(skillDir, descriptor)
	if err != nil {
		return Result{}, err
	}

	outputRoot := opts.OutputRoot
	if outputRoot == "" {
		outputRoot = filepath.Join(".skcr", "build", Target)
	}
	outputDir := filepath.Join(outputRoot, descriptor.Name)
	if err := os.MkdirAll(filepath.Join(outputDir, "evals"), 0o755); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "SKILL.md"), instructions, 0o644); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "skill.yaml"), contractBytes, 0o644); err != nil {
		return Result{}, err
	}
	versionBytes := []byte(descriptor.Version + "\n")
	if sourceVersion, readErr := os.ReadFile(filepath.Join(skillDir, "VERSION")); readErr == nil {
		versionBytes = sourceVersion
	}
	changelogBytes := []byte("# Changelog\n\n## " + descriptor.Version + "\n\n- Compiled assurance artifact.\n")
	if sourceChangelog, readErr := os.ReadFile(filepath.Join(skillDir, "CHANGELOG.md")); readErr == nil {
		changelogBytes = sourceChangelog
	}
	if err := os.WriteFile(filepath.Join(outputDir, "VERSION"), versionBytes, 0o644); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "CHANGELOG.md"), changelogBytes, 0o644); err != nil {
		return Result{}, err
	}
	artifactParts := [][]byte{instructions, contractBytes, versionBytes, changelogBytes}
	for _, item := range evals {
		if err := os.WriteFile(filepath.Join(outputDir, "evals", item.name+".yaml"), item.data, 0o644); err != nil {
			return Result{}, err
		}
		artifactParts = append(artifactParts, []byte(item.name), item.data)
	}
	source := SourceDigests{DescriptorDigest: digest(descriptorBytes), ContractDigest: digest(contractSource), EvalDigest: evalDigest, InstructionsDigest: digest(instructions), IntegrationsDigest: integrationsDigest, DependenciesDigest: dependenciesDigest, AssuranceDigest: assuranceDigest}
	mappingBytes, _ := json.Marshal(mapping)
	compilerVersion := opts.CompilerVersion
	if compilerVersion == "" {
		compilerVersion = "dev"
	}
	manifest := Manifest{
		SchemaVersion: "1",
		Source:        source,
		Target:        TargetMetadata{Type: Target, ContractSchema: "1", EvalSchema: "1", ArtifactDigest: digest(join(artifactParts...))},
		Mapping:       mapping,
		Provenance:    BuildProvenance{Compiler: CompilerIdentity{Name: "skcr", Version: compilerVersion}, SourceArtifactDigest: digest(join([]byte(source.DescriptorDigest), []byte(source.ContractDigest), []byte(source.EvalDigest), []byte(source.InstructionsDigest), []byte(source.IntegrationsDigest), []byte(source.DependenciesDigest), []byte(source.AssuranceDigest))), MappingDigest: digest(mappingBytes)},
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Result{}, err
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "build-manifest.json"), manifestBytes, 0o644); err != nil {
		return Result{}, err
	}
	if err := writeChecksums(outputDir); err != nil {
		return Result{}, err
	}
	return Result{SkillName: descriptor.Name, OutputDir: outputDir, Manifest: manifest}, nil
}

func extendedSourceDigests(skillDir string, descriptor skillmeta.Descriptor) (string, string, string, error) {
	var integrationParts [][]byte
	if descriptor.Integrations != nil {
		for _, item := range []struct{ name, reference string }{{"a2a", descriptor.Integrations.A2A.File}, {"mcp", descriptor.Integrations.MCP.File}} {
			data, err := os.ReadFile(filepath.Join(skillDir, item.reference))
			if err != nil {
				return "", "", "", err
			}
			integrationParts = append(integrationParts, []byte(item.name), data)
		}
	}
	integrationDigest := ""
	if len(integrationParts) > 0 {
		integrationDigest = digest(join(integrationParts...))
	}
	dependenciesDigest := ""
	if descriptor.Dependencies != nil {
		data, err := os.ReadFile(filepath.Join(skillDir, descriptor.Dependencies.File))
		if err != nil {
			return "", "", "", err
		}
		dependenciesDigest = digest(data)
	}
	assuranceDigest := ""
	if descriptor.Assurance != nil {
		data, err := os.ReadFile(filepath.Join(skillDir, descriptor.Assurance.File))
		if err != nil {
			return "", "", "", err
		}
		assuranceDigest = digest(data)
	}
	return integrationDigest, dependenciesDigest, assuranceDigest, nil
}

func validateLegacySecuritySummary(d skillmeta.Descriptor, c skillmeta.Contract) error {
	if d.Security == nil {
		return nil
	}
	view := contractRuntimeView(c)
	derived := map[string]bool{
		"requires_network": view.networkOutbound,
		"requires_secrets": len(view.secrets) > 0,
		"writes_files":     len(view.write)+len(view.delete) > 0,
		"runs_commands":    view.commandsExecute,
	}
	declared := map[string]bool{
		"requires_network": d.Security.RequiresNetwork,
		"requires_secrets": d.Security.RequiresSecrets,
		"writes_files":     d.Security.WritesFiles,
		"runs_commands":    d.Security.RunsCommands,
	}
	var mismatches []string
	for field, value := range declared {
		if derived[field] != value {
			mismatches = append(mismatches, fmt.Sprintf("descriptor.security.%s=%t but contract derives %t", field, value, derived[field]))
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		return fmt.Errorf("inconsistent security posture: %s", strings.Join(mismatches, "; "))
	}
	return nil
}

func unsupportedSecurity(c skillmeta.Contract) []string {
	var out []string
	if c.SchemaVersion == skillmeta.ContractSchemaVersion {
		for _, rule := range c.Capabilities.Runtime.Allowed.Commands.Allow {
			if len(rule.ArgvPrefix) > 0 {
				out = append(out, "capabilities.runtime.allowed.commands.allow.argv_prefix")
			}
		}
		if len(c.HumanApproval.Rules) > 0 {
			out = append(out, "human_approval.rules")
		}
		if len(c.Data.Egress.Allow) > 0 || len(c.Data.Flows) > 0 {
			out = append(out, "data.egress/data.flows")
		}
		return uniqueSorted(out)
	}
	for _, set := range []struct {
		name string
		set  skillmeta.CapabilitySet
	}{{"required", c.Capabilities.Required}, {"allowed", c.Capabilities.Allowed}} {
		if len(set.set.Repository.Read)+len(set.set.Repository.Write) > 0 {
			out = append(out, "capabilities."+set.name+".repository")
		}
		if set.set.Process.Execute != nil && *set.set.Process.Execute {
			out = append(out, "capabilities."+set.name+".process.execute (command allowlist unavailable)")
		}
		if set.set.Secrets.Read != nil && *set.set.Secrets.Read {
			out = append(out, "capabilities."+set.name+".secrets.read (secret IDs unavailable)")
		}
	}
	if len(c.HumanApproval.Rules) > 0 {
		out = append(out, "human_approval.rules")
	}
	if len(c.Data.Egress.Allow) > 0 || len(c.Data.Flows) > 0 {
		out = append(out, "data.egress/data.flows")
	}
	return uniqueSorted(out)
}

func mapContract(d skillmeta.Descriptor, c skillmeta.Contract) skilContract {
	view := contractRuntimeView(c)
	platforms := make([]string, len(d.CompatibleWith))
	for i := range d.CompatibleWith {
		platforms[i] = string(d.CompatibleWith[i])
	}
	owner := ""
	if d.Ownership != nil {
		owner = d.Ownership.Primary
	} else if len(d.Owners) > 0 {
		owner = d.Owners[0]
	}
	result := skilContract{
		Version: 1, Owner: owner, Entrypoint: d.Entrypoint,
		Compatibility: skilCompatibility{Platforms: platforms},
		Security:      skilSecurity{RequiresNetwork: view.networkOutbound, RequiresSecrets: len(view.secrets) > 0, WritesFiles: len(view.write)+len(view.delete) > 0, RunsCommands: view.commandsExecute},
		Skill:         skilIdentity{Name: d.Name, Version: d.Version, Description: d.Description},
		Capabilities: skilCapabilities{
			Filesystem: skilFilesystem{Read: nonnil(view.read), Write: nonnil(view.write), Delete: nonnil(view.delete)},
			Network:    skilNetwork{Inbound: view.networkInbound, Outbound: view.networkOutbound, Hosts: nonnil(view.hosts)},
			Commands:   skilCommands{Execute: view.commandsExecute, Allow: nonnil(view.commands)}, Secrets: skilSecrets{Read: nonnil(view.secrets), Expose: view.secretExpose},
			Environment: skilEnvironment{Read: nonnil(view.environment)}, Tools: skilTools{Allow: nonnil(view.toolsAllow), Deny: nonnil(view.toolsDeny)},
			MCP: skilMCP{Servers: nonnil(view.mcpServers), Tools: nonnil(view.mcpTools)}, Persistence: view.persistence,
			Agent: skilAgent{AutonomousActions: view.autonomous, ExternalSideEffects: view.externalEffects, ConfirmDestructive: view.confirmDestructive, ConfirmExternal: view.confirmExternal, ExternalTargets: nonnil(view.externalTargets)},
		},
	}
	runtimeLimit, toolLimit := c.Limits.MaxRuntimeSeconds, c.Limits.MaxToolCalls
	if c.SchemaVersion == skillmeta.ContractSchemaVersion {
		runtimeLimit, toolLimit = c.Resources.MaxRuntimeSeconds, c.Resources.MaxToolCalls
	}
	if runtimeLimit != nil || toolLimit != nil {
		resources := &skilResources{}
		if runtimeLimit != nil {
			resources.MaxRuntimeSeconds = runtimeLimit
		}
		if toolLimit != nil {
			resources.MaxToolCalls = toolLimit
		}
		result.Capabilities.Resources = resources
	}
	if c.SchemaVersion == skillmeta.ContractSchemaVersion {
		if result.Capabilities.Resources == nil {
			result.Capabilities.Resources = &skilResources{}
		}
		resources := result.Capabilities.Resources
		if c.Resources.MaxMemoryMB != nil {
			resources.MaxMemoryMB = c.Resources.MaxMemoryMB
		}
		if c.Resources.MaxNetworkBytes != nil {
			resources.MaxNetworkBytes = c.Resources.MaxNetworkBytes
		}
		if c.Resources.MaxModelTokens != nil {
			resources.MaxModelTokens = c.Resources.MaxModelTokens
		}
		if c.Resources.MaxDelegationDepth != nil {
			resources.MaxDelegationDepth = c.Resources.MaxDelegationDepth
		}
		if c.Resources.MaxExternalMutations != nil {
			resources.MaxExternalMutations = c.Resources.MaxExternalMutations
		}
	}
	return result
}

type runtimeView struct {
	read, write, delete, hosts, commands, secrets, environment                    []string
	toolsAllow, toolsDeny, mcpServers, mcpTools, externalTargets                  []string
	networkInbound, networkOutbound, commandsExecute, secretExpose                bool
	persistence, autonomous, externalEffects, confirmDestructive, confirmExternal bool
}

func contractRuntimeView(c skillmeta.Contract) runtimeView {
	if c.SchemaVersion == skillmeta.ContractSchemaVersion {
		a := c.Capabilities.Runtime.Allowed
		commands := make([]string, 0, len(a.Commands.Allow))
		for _, rule := range a.Commands.Allow {
			commands = append(commands, rule.Executable)
		}
		return runtimeView{
			read: a.Filesystem.Read, write: a.Filesystem.Write, delete: a.Filesystem.Delete,
			hosts: a.Network.Hosts, commands: commands, secrets: a.Secrets.Read, environment: a.Environment.Read,
			toolsAllow: a.Tools.Allow, toolsDeny: a.Tools.Deny, mcpServers: a.MCP.Servers, mcpTools: a.MCP.Tools, externalTargets: a.Agent.ExternalTargets,
			networkInbound: enabled(a.Network.Inbound), networkOutbound: enabled(a.Network.Outbound), commandsExecute: enabled(a.Commands.Execute), secretExpose: enabled(a.Secrets.Expose),
			persistence: enabled(a.Persistence), autonomous: enabled(a.Agent.AutonomousActions), externalEffects: enabled(a.Agent.ExternalSideEffects),
			confirmDestructive: enabled(a.Agent.ConfirmDestructive), confirmExternal: enabled(a.Agent.ConfirmExternal),
		}
	}
	a := c.Capabilities.Allowed
	return runtimeView{read: a.Filesystem.Read, write: a.Filesystem.Write, hosts: a.Network.Allow, toolsAllow: c.Tools.Allow, toolsDeny: c.Tools.Deny,
		networkOutbound: len(a.Network.Allow) > 0, commandsExecute: enabled(a.Process.Execute), confirmDestructive: true, confirmExternal: true}
}

func enabled(value *bool) bool { return value != nil && *value }

type compiledEval struct {
	name string
	data []byte
}

func compileEvals(skillDir string, d skillmeta.Descriptor, c skillmeta.Contract) ([]compiledEval, string, error) {
	dir := filepath.Join(skillDir, d.Evals.Directory)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}
	var out []compiledEval
	seen := map[string]bool{}
	var sourceParts [][]byte
	view := contractRuntimeView(c)
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, "", err
		}
		sourceParts = append(sourceParts, []byte(entry.Name()), data)
		doc, err := skillmeta.ParseEval(data)
		if err != nil {
			return nil, "", err
		}
		for _, scenario := range doc.Scenarios {
			if seen[scenario.ID] {
				return nil, "", fmt.Errorf("duplicate eval scenario id %q", scenario.ID)
			}
			seen[scenario.ID] = true
			if doc.SchemaVersion == skillmeta.EvalSchemaVersion {
				e := skilEval{Version: 1, Name: scenario.ID, Type: scenario.Type, Input: skilEvalInput{scenario.Input.Message}, Context: scenario.Context, Environment: scenario.Environment,
					Tools: skilEvalTools{nonnil(scenario.ToolsV2.Available)}, Expect: skilEvalExpect{Required: nonnil(scenario.Expect.Required), Allowed: nonnil(scenario.Expect.Allowed), Forbidden: nonnil(scenario.Expect.Forbidden), ForbiddenCapabilities: nonnil(scenario.Expect.ForbiddenCapabilities), Arguments: scenario.Expect.Arguments, OutputProperties: nonnil(scenario.Expect.OutputProperties), Assertions: nonnil(scenario.Expect.Assertions)}}
				if scenario.Attack != nil {
					e.Attack = &skilEvalAttack{Category: scenario.Attack.Category}
				}
				if scenario.Containment != nil {
					e.Containment = &skilEvalContainment{Required: enabled(scenario.Containment.Required), AllowedTargets: scenario.Containment.AllowedTargets, RequireEnforcement: enabled(scenario.Containment.RequireEnforcement), RequireNativeIsolation: enabled(scenario.Containment.RequireNativeIsolation)}
				}
				encoded, err := yaml.Marshal(e)
				if err != nil {
					return nil, "", err
				}
				out = append(out, compiledEval{scenario.ID, encoded})
				continue
			}
			typ := "behavioral"
			if scenario.Category == "adversarial" {
				typ = "adversarial"
			}
			assertions := []string{}
			if len(scenario.Assertions.Capabilities.MustNotUse) > 0 {
				assertions = append(assertions, "no_forbidden_capabilities")
			}
			props := []string{"non_empty"}
			for _, f := range c.Output.ForbiddenContent {
				if f == "secret" || f == "secrets" || f == "credential" || f == "credentials" {
					props = append(props, "no_secrets")
					break
				}
			}
			available := append([]string{}, view.toolsAllow...)
			available = append(available, scenario.Assertions.Tools.MustNotUse...)
			e := skilEval{Version: 1, Name: scenario.ID, Type: typ, Input: skilEvalInput{scenario.Input.Prompt}, Tools: skilEvalTools{uniqueSorted(available)}, Expect: skilEvalExpect{Allowed: nonnil(view.toolsAllow), Forbidden: nonnil(scenario.Assertions.Tools.MustNotUse), ForbiddenCapabilities: nonnil(scenario.Assertions.Capabilities.MustNotUse), OutputProperties: uniqueSorted(props), Assertions: assertions}}
			encoded, err := yaml.Marshal(e)
			if err != nil {
				return nil, "", err
			}
			out = append(out, compiledEval{scenario.ID, encoded})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, digest(join(sourceParts...)), nil
}

func nonnil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
func uniqueSorted(values []string) []string {
	set := map[string]bool{}
	for _, v := range values {
		set[v] = true
	}
	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
func join(parts ...[]byte) []byte {
	var out []byte
	for _, part := range parts {
		out = append(out, part...)
		out = append(out, 0)
	}
	return out
}
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func writeChecksums(outputDir string) error {
	type entry struct{ path, checksum string }
	var entries []entry
	err := filepath.WalkDir(outputDir, func(path string, item os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if item.IsDir() || item.Name() == "checksums.txt" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{filepath.ToSlash(rel), hex.EncodeToString(sum[:])})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	var content strings.Builder
	for _, item := range entries {
		fmt.Fprintf(&content, "%s  %s\n", item.checksum, item.path)
	}
	return os.WriteFile(filepath.Join(outputDir, "checksums.txt"), []byte(content.String()), 0o644)
}
