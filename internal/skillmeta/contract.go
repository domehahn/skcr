package skillmeta

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ContractSchemaVersion       = "2"
	LegacyContractSchemaVersion = "1"
)

var identifierRE = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._-][a-z0-9]+)*$`)

type Contract struct {
	SchemaVersion  string                 `yaml:"schema_version" json:"schema_version"`
	Capabilities   ContractCapabilities   `yaml:"capabilities" json:"capabilities"`
	Identity       IdentityRequirements   `yaml:"identity,omitempty" json:"identity,omitempty"`
	Delegation     DelegationRequirements `yaml:"delegation,omitempty" json:"delegation,omitempty"`
	Tools          Tools                  `yaml:"tools,omitempty" json:"tools,omitempty"`
	Data           DataBoundary           `yaml:"data" json:"data"`
	Preconditions  []Condition            `yaml:"preconditions" json:"preconditions"`
	Postconditions []Condition            `yaml:"postconditions" json:"postconditions"`
	Invariants     []Condition            `yaml:"invariants" json:"invariants"`
	HumanApproval  HumanApproval          `yaml:"human_approval" json:"human_approval"`
	Limits         Limits                 `yaml:"limits,omitempty" json:"limits,omitempty"`
	Effects        Effects                `yaml:"effects,omitempty" json:"effects,omitempty"`
	Resources      Resources              `yaml:"resources,omitempty" json:"resources,omitempty"`
	Output         Output                 `yaml:"output" json:"output"`
	presence       map[string]bool
}

type ContractCapabilities struct {
	Required CapabilitySet               `yaml:"required,omitempty" json:"required,omitempty"`
	Allowed  CapabilitySet               `yaml:"allowed,omitempty" json:"allowed,omitempty"`
	Semantic SemanticCapabilities        `yaml:"semantic,omitempty" json:"semantic,omitempty"`
	Runtime  RuntimeContractCapabilities `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

type IdentityRequirements struct {
	AcceptedPrincipals []string               `yaml:"accepted_principals" json:"accepted_principals"`
	Credentials        CredentialRequirements `yaml:"credentials" json:"credentials"`
}

func (i IdentityRequirements) IsZero() bool {
	return i.AcceptedPrincipals == nil && i.Credentials.MaxTTLSeconds == nil
}

type CredentialRequirements struct {
	MaxTTLSeconds          *int     `yaml:"max_ttl_seconds" json:"max_ttl_seconds"`
	RequireAudienceBinding *bool    `yaml:"require_audience_binding" json:"require_audience_binding"`
	AllowTokenPassthrough  *bool    `yaml:"allow_token_passthrough" json:"allow_token_passthrough"`
	RequiredScopes         []string `yaml:"required_scopes" json:"required_scopes"`
}

type DelegationRequirements struct {
	Allowed                      *bool `yaml:"allowed" json:"allowed"`
	MaxDepth                     *int  `yaml:"max_depth" json:"max_depth"`
	RequireChildCapabilitySubset *bool `yaml:"require_child_capability_subset" json:"require_child_capability_subset"`
	RequireAuthenticatedOrigin   *bool `yaml:"require_authenticated_origin" json:"require_authenticated_origin"`
}

func (d DelegationRequirements) IsZero() bool { return d.Allowed == nil }

type SemanticCapabilities struct {
	Required map[string][]string `yaml:"required" json:"required"`
}

func (s SemanticCapabilities) IsZero() bool { return s.Required == nil }
func (r RuntimeContractCapabilities) IsZero() bool {
	return r.Required.Persistence == nil && r.Allowed.Persistence == nil
}

type RuntimeContractCapabilities struct {
	Required RuntimeCapabilitySet `yaml:"required" json:"required"`
	Allowed  RuntimeCapabilitySet `yaml:"allowed" json:"allowed"`
}

type RuntimeCapabilitySet struct {
	Filesystem  RuntimeFilesystem  `yaml:"filesystem" json:"filesystem"`
	Network     RuntimeNetwork     `yaml:"network" json:"network"`
	Commands    RuntimeCommands    `yaml:"commands" json:"commands"`
	Secrets     RuntimeSecrets     `yaml:"secrets" json:"secrets"`
	Environment RuntimeEnvironment `yaml:"environment" json:"environment"`
	Tools       Tools              `yaml:"tools" json:"tools"`
	MCP         RuntimeMCP         `yaml:"mcp" json:"mcp"`
	Persistence *bool              `yaml:"persistence" json:"persistence"`
	Agent       RuntimeAgent       `yaml:"agent" json:"agent"`
}

type RuntimeFilesystem struct {
	Read   []string `yaml:"read" json:"read"`
	Write  []string `yaml:"write" json:"write"`
	Delete []string `yaml:"delete" json:"delete"`
}
type RuntimeNetwork struct {
	Inbound  *bool    `yaml:"inbound" json:"inbound"`
	Outbound *bool    `yaml:"outbound" json:"outbound"`
	Hosts    []string `yaml:"hosts" json:"hosts"`
}
type RuntimeCommands struct {
	Execute *bool         `yaml:"execute" json:"execute"`
	Allow   []CommandRule `yaml:"allow" json:"allow"`
}
type CommandRule struct {
	Executable string   `yaml:"executable" json:"executable"`
	ArgvPrefix []string `yaml:"argv_prefix,omitempty" json:"argv_prefix,omitempty"`
}
type RuntimeSecrets struct {
	Read   []string `yaml:"read" json:"read"`
	Expose *bool    `yaml:"expose" json:"expose"`
}
type RuntimeEnvironment struct {
	Read []string `yaml:"read" json:"read"`
}
type RuntimeMCP struct {
	Servers []string `yaml:"servers" json:"servers"`
	Tools   []string `yaml:"tools" json:"tools"`
}
type RuntimeAgent struct {
	AutonomousActions   *bool    `yaml:"autonomous_actions" json:"autonomous_actions"`
	ExternalSideEffects *bool    `yaml:"external_side_effects" json:"external_side_effects"`
	ConfirmDestructive  *bool    `yaml:"confirm_destructive" json:"confirm_destructive"`
	ConfirmExternal     *bool    `yaml:"confirm_external" json:"confirm_external"`
	ExternalTargets     []string `yaml:"external_targets" json:"external_targets"`
}

type CapabilitySet struct {
	Repository RepositoryCapabilities `yaml:"repository" json:"repository"`
	Filesystem FilesystemCapabilities `yaml:"filesystem" json:"filesystem"`
	Network    NetworkCapabilities    `yaml:"network" json:"network"`
	Process    ProcessCapabilities    `yaml:"process" json:"process"`
	Secrets    SecretsCapabilities    `yaml:"secrets" json:"secrets"`
}

type RepositoryCapabilities struct {
	Read  []string `yaml:"read" json:"read"`
	Write []string `yaml:"write" json:"write"`
}
type FilesystemCapabilities struct {
	Read  []string `yaml:"read" json:"read"`
	Write []string `yaml:"write" json:"write"`
}
type NetworkCapabilities struct {
	Allow []string `yaml:"allow" json:"allow"`
}
type ProcessCapabilities struct {
	Execute *bool `yaml:"execute" json:"execute"`
}
type SecretsCapabilities struct {
	Read *bool `yaml:"read" json:"read"`
}
type Tools struct {
	Allow []string `yaml:"allow" json:"allow"`
	Deny  []string `yaml:"deny" json:"deny"`
}

func (t Tools) IsZero() bool { return t.Allow == nil && t.Deny == nil }

type DataBoundary struct {
	Classifications []string     `yaml:"classifications" json:"classifications"`
	Policies        []DataPolicy `yaml:"policies,omitempty" json:"policies,omitempty"`
	Egress          DataEgress   `yaml:"egress" json:"egress"`
	Flows           []DataFlow   `yaml:"flows" json:"flows"`
}
type DataPolicy struct {
	ID          string   `yaml:"id" json:"id"`
	Sensitivity string   `yaml:"sensitivity" json:"sensitivity"`
	Purposes    []string `yaml:"purposes" json:"purposes"`
	Retention   string   `yaml:"retention" json:"retention"`
}
type DataEgress struct {
	Allow []string `yaml:"allow" json:"allow"`
}
type DataFlow struct {
	Source  DataEndpoint `yaml:"source" json:"source"`
	Sink    DataEndpoint `yaml:"sink" json:"sink"`
	Allowed bool         `yaml:"allowed" json:"allowed"`
}
type DataEndpoint struct {
	Classification string `yaml:"classification,omitempty" json:"classification,omitempty"`
	Type           string `yaml:"type,omitempty" json:"type,omitempty"`
	Destination    string `yaml:"destination,omitempty" json:"destination,omitempty"`
}
type Condition struct {
	ID              string   `yaml:"id" json:"id"`
	Type            string   `yaml:"type" json:"type"`
	Description     string   `yaml:"description,omitempty" json:"description,omitempty"`
	Capabilities    []string `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Classifications []string `yaml:"classifications,omitempty" json:"classifications,omitempty"`
}
type HumanApproval struct {
	Rules []ApprovalRule `yaml:"rules" json:"rules"`
}
type ApprovalRule struct {
	ID     string         `yaml:"id" json:"id"`
	Action ApprovalAction `yaml:"action" json:"action"`
	Mode   string         `yaml:"mode" json:"mode"`
}
type ApprovalAction struct {
	Capability string `yaml:"capability" json:"capability"`
	Tool       string `yaml:"tool,omitempty" json:"tool,omitempty"`
}
type Limits struct {
	Scope              string `yaml:"scope" json:"scope"`
	MaxToolCalls       *int   `yaml:"max_tool_calls" json:"max_tool_calls"`
	MaxRuntimeSeconds  *int   `yaml:"max_runtime_seconds" json:"max_runtime_seconds"`
	MaxNetworkRequests *int   `yaml:"max_network_requests" json:"max_network_requests"`
}

func (l Limits) IsZero() bool {
	return l.Scope == "" && l.MaxToolCalls == nil && l.MaxRuntimeSeconds == nil && l.MaxNetworkRequests == nil
}

type Resources struct {
	Scope                string `yaml:"scope" json:"scope"`
	MaxRuntimeSeconds    *int   `yaml:"max_runtime_seconds" json:"max_runtime_seconds"`
	MaxMemoryMB          *int   `yaml:"max_memory_mb" json:"max_memory_mb"`
	MaxNetworkBytes      *int   `yaml:"max_network_bytes" json:"max_network_bytes"`
	MaxNetworkRequests   *int   `yaml:"max_network_requests" json:"max_network_requests"`
	MaxToolCalls         *int   `yaml:"max_tool_calls" json:"max_tool_calls"`
	MaxSteps             *int   `yaml:"max_steps" json:"max_steps"`
	MaxReplans           *int   `yaml:"max_replans" json:"max_replans"`
	MaxDelegationDepth   *int   `yaml:"max_delegation_depth" json:"max_delegation_depth"`
	MaxModelTokens       *int   `yaml:"max_model_tokens" json:"max_model_tokens"`
	MaxExternalMutations *int   `yaml:"max_external_mutations" json:"max_external_mutations"`
}
type Effects struct {
	Tools []ToolEffect `yaml:"tools" json:"tools"`
}

func (e Effects) IsZero() bool { return e.Tools == nil }

type ToolEffect struct {
	Tool       string `yaml:"tool" json:"tool"`
	Class      string `yaml:"class" json:"class"`
	Reversible bool   `yaml:"reversible" json:"reversible"`
	Idempotent bool   `yaml:"idempotent" json:"idempotent"`
}
type Output struct {
	Format           string   `yaml:"format" json:"format"`
	RequiredFields   []string `yaml:"required_fields" json:"required_fields"`
	ForbiddenContent []string `yaml:"forbidden_content" json:"forbidden_content"`
}

func boolPtr(value bool) *bool { return &value }

func emptyCapabilitySet() CapabilitySet {
	return CapabilitySet{
		Repository: RepositoryCapabilities{Read: []string{}, Write: []string{}},
		Filesystem: FilesystemCapabilities{Read: []string{}, Write: []string{}},
		Network:    NetworkCapabilities{Allow: []string{}},
		Process:    ProcessCapabilities{Execute: boolPtr(false)},
		Secrets:    SecretsCapabilities{Read: boolPtr(false)},
	}
}

func NewContract() Contract {
	contract := Contract{
		SchemaVersion: LegacyContractSchemaVersion,
		Capabilities: ContractCapabilities{
			Required: emptyCapabilitySet(),
			Allowed:  emptyCapabilitySet(),
		},
		Tools: Tools{Allow: []string{}, Deny: []string{}},
		Data: DataBoundary{
			Classifications: []string{"secret", "credential", "source-code", "personal-data"},
			Egress:          DataEgress{Allow: []string{}},
			Flows:           []DataFlow{},
		},
		Preconditions: []Condition{},
		Postconditions: []Condition{
			{ID: "repository-state-unchanged", Type: "repository.unchanged"},
			{ID: "no-secret-egress", Type: "data.no_egress", Classifications: []string{"secret", "credential"}},
		},
		Invariants: []Condition{
			{ID: "declared-capabilities-only", Type: "capability.declared_only"},
			{ID: "declared-tools-only", Type: "tool.declared_only"},
			{ID: "no-secret-egress", Type: "data.no_egress", Classifications: []string{"secret", "credential"}},
		},
		HumanApproval: HumanApproval{Rules: []ApprovalRule{}},
		Limits:        Limits{Scope: "invocation"},
		Output:        Output{Format: "structured", RequiredFields: []string{}, ForbiddenContent: []string{"secrets", "credentials"}},
	}
	contract.presence = requiredContractPresence()
	return contract
}

func emptyRuntimeCapabilitySet() RuntimeCapabilitySet {
	return RuntimeCapabilitySet{
		Filesystem:  RuntimeFilesystem{Read: []string{}, Write: []string{}, Delete: []string{}},
		Network:     RuntimeNetwork{Inbound: boolPtr(false), Outbound: boolPtr(false), Hosts: []string{}},
		Commands:    RuntimeCommands{Execute: boolPtr(false), Allow: []CommandRule{}},
		Secrets:     RuntimeSecrets{Read: []string{}, Expose: boolPtr(false)},
		Environment: RuntimeEnvironment{Read: []string{}}, Tools: Tools{Allow: []string{}, Deny: []string{}},
		MCP: RuntimeMCP{Servers: []string{}, Tools: []string{}}, Persistence: boolPtr(false),
		Agent: RuntimeAgent{AutonomousActions: boolPtr(false), ExternalSideEffects: boolPtr(false), ConfirmDestructive: boolPtr(true), ConfirmExternal: boolPtr(true), ExternalTargets: []string{}},
	}
}

func NewContractV2() Contract {
	contract := NewContract()
	contract.SchemaVersion = ContractSchemaVersion
	contract.Capabilities = ContractCapabilities{
		Semantic: SemanticCapabilities{Required: map[string][]string{}},
		Runtime:  RuntimeContractCapabilities{Required: emptyRuntimeCapabilitySet(), Allowed: emptyRuntimeCapabilitySet()},
	}
	contract.Tools = Tools{}
	contract.Limits = Limits{}
	contract.Identity = IdentityRequirements{AcceptedPrincipals: []string{"agent", "human"}, Credentials: CredentialRequirements{MaxTTLSeconds: intPtr(900), RequireAudienceBinding: boolPtr(true), AllowTokenPassthrough: boolPtr(false), RequiredScopes: []string{}}}
	contract.Delegation = DelegationRequirements{Allowed: boolPtr(false), MaxDepth: intPtr(0), RequireChildCapabilitySubset: boolPtr(true), RequireAuthenticatedOrigin: boolPtr(true)}
	contract.Data.Policies = []DataPolicy{
		{ID: "secret", Sensitivity: "restricted", Purposes: []string{"task-execution"}, Retention: "invocation"},
		{ID: "credential", Sensitivity: "restricted", Purposes: []string{"authentication"}, Retention: "invocation"},
		{ID: "source-code", Sensitivity: "internal", Purposes: []string{"task-execution"}, Retention: "invocation"},
		{ID: "personal-data", Sensitivity: "confidential", Purposes: []string{"task-execution"}, Retention: "invocation"},
	}
	contract.Effects = Effects{Tools: []ToolEffect{}}
	contract.Resources = Resources{Scope: "invocation"}
	contract.presence = requiredContractV2Presence()
	return contract
}

func intPtr(value int) *int { return &value }

func ParseContract(data []byte) (Contract, error) {
	var contract Contract
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return Contract{}, err
	}
	contract.presence = collectMappingPaths(&node)
	return contract, nil
}

func LoadContract(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	contract, err := ParseContract(data)
	if err != nil {
		return Contract{}, fmt.Errorf("%s: %w", path, err)
	}
	return contract, nil
}

func ValidateContract(c Contract) []string {
	if c.SchemaVersion == ContractSchemaVersion {
		return validateContractV2(c)
	}
	if c.SchemaVersion != LegacyContractSchemaVersion {
		return []string{fmt.Sprintf("unsupported contract schema_version %q", c.SchemaVersion)}
	}
	var errs []string
	for path := range requiredContractPresence() {
		if !c.presence[path] {
			errs = append(errs, path+" is required")
		}
	}
	errs = append(errs, validateCapabilitySet("capabilities.required", c.Capabilities.Required)...)
	errs = append(errs, validateCapabilitySet("capabilities.allowed", c.Capabilities.Allowed)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.required.filesystem.read", c.Capabilities.Required.Filesystem.Read, true, false)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.required.filesystem.write", c.Capabilities.Required.Filesystem.Write, true, false)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.allowed.filesystem.read", c.Capabilities.Allowed.Filesystem.Read, true, false)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.allowed.filesystem.write", c.Capabilities.Allowed.Filesystem.Write, true, false)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.required.network.allow", c.Capabilities.Required.Network.Allow, false, true)...)
	errs = append(errs, validateNormalizedDuplicates("capabilities.allowed.network.allow", c.Capabilities.Allowed.Network.Allow, false, true)...)
	errs = append(errs, validateRequiredSubset(c.Capabilities.Required, c.Capabilities.Allowed)...)
	errs = append(errs, validateStringList("tools.allow", c.Tools.Allow, true)...)
	errs = append(errs, validateStringList("tools.deny", c.Tools.Deny, true)...)
	denied := stringSet(c.Tools.Deny)
	for _, tool := range c.Tools.Allow {
		if _, exists := denied[tool]; exists {
			errs = append(errs, fmt.Sprintf("tools contains %q in both allow and deny", tool))
		}
	}
	errs = append(errs, validateStringList("data.classifications", c.Data.Classifications, false)...)
	errs = append(errs, validateStringList("data.egress.allow", c.Data.Egress.Allow, true)...)
	classifications := stringSet(c.Data.Classifications)
	egress := stringSet(normalizedHosts(c.Data.Egress.Allow))
	for i, flow := range c.Data.Flows {
		if strings.TrimSpace(flow.Source.Classification) == "" {
			errs = append(errs, fmt.Sprintf("data.flows[%d].source.classification is required", i))
		} else if _, declared := classifications[flow.Source.Classification]; !declared {
			errs = append(errs, fmt.Sprintf("data.flows[%d] uses undeclared classification %q", i, flow.Source.Classification))
		}
		if strings.TrimSpace(flow.Sink.Type) == "" {
			errs = append(errs, fmt.Sprintf("data.flows[%d].sink.type is required", i))
		}
		if flow.Allowed && flow.Sink.Type == "network" {
			if flow.Sink.Destination == "" {
				errs = append(errs, fmt.Sprintf("data.flows[%d] allowed network sink requires destination", i))
			} else if _, declared := egress[strings.ToLower(strings.TrimSpace(flow.Sink.Destination))]; !declared {
				errs = append(errs, fmt.Sprintf("data.flows[%d] network destination %q is not in data.egress.allow", i, flow.Sink.Destination))
			}
		}
	}
	errs = append(errs, validateConditions("preconditions", c.Preconditions)...)
	errs = append(errs, validateConditions("postconditions", c.Postconditions)...)
	errs = append(errs, validateConditions("invariants", c.Invariants)...)
	errs = append(errs, validateApproval(c.HumanApproval)...)
	if c.Limits.Scope != "invocation" {
		errs = append(errs, `limits.scope must be "invocation"`)
	}
	for path, value := range map[string]*int{
		"limits.max_tool_calls":       c.Limits.MaxToolCalls,
		"limits.max_runtime_seconds":  c.Limits.MaxRuntimeSeconds,
		"limits.max_network_requests": c.Limits.MaxNetworkRequests,
	} {
		if value != nil && *value < 0 {
			errs = append(errs, path+" must be zero or greater")
		}
	}
	if strings.TrimSpace(c.Output.Format) == "" {
		errs = append(errs, "output.format must not be blank")
	}
	errs = append(errs, validateStringList("output.required_fields", c.Output.RequiredFields, false)...)
	errs = append(errs, validateStringList("output.forbidden_content", c.Output.ForbiddenContent, false)...)
	return dedupeSorted(errs)
}

func requiredContractPresence() map[string]bool {
	paths := []string{
		"schema_version", "capabilities", "capabilities.required", "capabilities.allowed",
		"capabilities.required.repository", "capabilities.required.repository.read", "capabilities.required.repository.write",
		"capabilities.required.filesystem", "capabilities.required.filesystem.read", "capabilities.required.filesystem.write",
		"capabilities.required.network", "capabilities.required.network.allow", "capabilities.required.process",
		"capabilities.required.process.execute", "capabilities.required.secrets", "capabilities.required.secrets.read",
		"capabilities.allowed.repository", "capabilities.allowed.repository.read", "capabilities.allowed.repository.write",
		"capabilities.allowed.filesystem", "capabilities.allowed.filesystem.read", "capabilities.allowed.filesystem.write",
		"capabilities.allowed.network", "capabilities.allowed.network.allow", "capabilities.allowed.process",
		"capabilities.allowed.process.execute", "capabilities.allowed.secrets", "capabilities.allowed.secrets.read",
		"tools", "tools.allow", "tools.deny",
		"data", "data.classifications", "data.egress", "data.egress.allow", "data.flows",
		"preconditions", "postconditions", "invariants", "human_approval", "human_approval.rules",
		"limits", "limits.scope", "limits.max_tool_calls", "limits.max_runtime_seconds", "limits.max_network_requests",
		"output", "output.format", "output.required_fields", "output.forbidden_content",
	}
	out := make(map[string]bool, len(paths))
	for _, path := range paths {
		out[path] = true
	}
	return out
}

func collectMappingPaths(root *yaml.Node) map[string]bool {
	out := map[string]bool{}
	var walk func(*yaml.Node, string)
	walk = func(node *yaml.Node, prefix string) {
		if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
			walk(node.Content[0], prefix)
			return
		}
		if node.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			path := node.Content[i].Value
			if prefix != "" {
				path = prefix + "." + path
			}
			out[path] = true
			walk(node.Content[i+1], path)
		}
	}
	walk(root, "")
	return out
}

func validateCapabilitySet(path string, set CapabilitySet) []string {
	var errs []string
	for name, list := range map[string][]string{
		path + ".repository.read":  set.Repository.Read,
		path + ".repository.write": set.Repository.Write,
		path + ".filesystem.read":  set.Filesystem.Read,
		path + ".filesystem.write": set.Filesystem.Write,
		path + ".network.allow":    set.Network.Allow,
	} {
		errs = append(errs, validateStringList(name, list, true)...)
	}
	if set.Process.Execute == nil {
		errs = append(errs, path+".process.execute is required and must be boolean")
	}
	if set.Secrets.Read == nil {
		errs = append(errs, path+".secrets.read is required and must be boolean")
	}
	return errs
}

func validateRequiredSubset(required, allowed CapabilitySet) []string {
	var errs []string
	for _, pair := range []struct {
		name              string
		required, allowed []string
		normalize         func([]string) []string
	}{
		{"repository.read", required.Repository.Read, allowed.Repository.Read, func(v []string) []string { return normalizedStrings(v, false) }},
		{"repository.write", required.Repository.Write, allowed.Repository.Write, func(v []string) []string { return normalizedStrings(v, false) }},
		{"filesystem.read", required.Filesystem.Read, allowed.Filesystem.Read, func(v []string) []string { return normalizedStrings(v, true) }},
		{"filesystem.write", required.Filesystem.Write, allowed.Filesystem.Write, func(v []string) []string { return normalizedStrings(v, true) }},
		{"network.connect", required.Network.Allow, allowed.Network.Allow, normalizedHosts},
	} {
		allowedSet := stringSet(pair.normalize(pair.allowed))
		for _, scope := range pair.normalize(pair.required) {
			if _, exists := allowedSet[scope]; !exists {
				errs = append(errs, fmt.Sprintf("required capability %s scope %q is outside allowed capabilities", pair.name, scope))
			}
		}
	}
	if required.Process.Execute != nil && *required.Process.Execute && (allowed.Process.Execute == nil || !*allowed.Process.Execute) {
		errs = append(errs, "required capability process.execute is outside allowed capabilities")
	}
	if required.Secrets.Read != nil && *required.Secrets.Read && (allowed.Secrets.Read == nil || !*allowed.Secrets.Read) {
		errs = append(errs, "required capability secrets.read is outside allowed capabilities")
	}
	return errs
}

func validateConditions(path string, conditions []Condition) []string {
	if conditions == nil {
		return []string{path + " is required and must be a list"}
	}
	var errs []string
	seen := map[string]struct{}{}
	for i, condition := range conditions {
		if !validIdentifier(condition.ID) {
			errs = append(errs, fmt.Sprintf("%s[%d].id %q is invalid", path, i, condition.ID))
		} else if _, duplicate := seen[condition.ID]; duplicate {
			errs = append(errs, fmt.Sprintf("%s contains duplicate id %q", path, condition.ID))
		}
		seen[condition.ID] = struct{}{}
		if !validIdentifier(condition.Type) {
			errs = append(errs, fmt.Sprintf("%s[%d].type %q is invalid", path, i, condition.Type))
		}
		for _, capability := range condition.Capabilities {
			if !KnownCapability(capability) {
				errs = append(errs, fmt.Sprintf("%s[%d] contains unknown capability %q", path, i, capability))
			}
		}
	}
	return errs
}

func validateApproval(approval HumanApproval) []string {
	if approval.Rules == nil {
		return []string{"human_approval.rules is required and must be a list"}
	}
	var errs []string
	seen := map[string]struct{}{}
	for i, rule := range approval.Rules {
		if !validIdentifier(rule.ID) {
			errs = append(errs, fmt.Sprintf("human_approval.rules[%d].id %q is invalid", i, rule.ID))
		} else if _, duplicate := seen[rule.ID]; duplicate {
			errs = append(errs, fmt.Sprintf("human_approval.rules contains duplicate id %q", rule.ID))
		}
		seen[rule.ID] = struct{}{}
		if !KnownCapability(rule.Action.Capability) {
			errs = append(errs, fmt.Sprintf("human_approval.rules[%d] has unknown capability %q", i, rule.Action.Capability))
		}
		switch rule.Mode {
		case "per_action", "per_invocation", "per_session":
		default:
			errs = append(errs, fmt.Sprintf("human_approval.rules[%d].mode %q is unsupported", i, rule.Mode))
		}
	}
	return errs
}

func validateStringList(path string, values []string, rejectWildcard bool) []string {
	if values == nil {
		return []string{path + " is required and must be a list"}
	}
	var errs []string
	seen := map[string]struct{}{}
	for i, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			errs = append(errs, fmt.Sprintf("%s[%d] must not be blank", path, i))
		}
		if rejectWildcard && strings.Contains(value, "*") {
			errs = append(errs, path+` must not contain wildcard permissions`)
		}
		if _, duplicate := seen[value]; duplicate {
			errs = append(errs, fmt.Sprintf("%s contains duplicate value %q", path, value))
		}
		seen[value] = struct{}{}
	}
	return errs
}

func validateNormalizedDuplicates(path string, values []string, cleanPath, lower bool) []string {
	seen := map[string]string{}
	var errs []string
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if cleanPath {
			normalized = filepath.ToSlash(filepath.Clean(normalized))
		}
		if lower {
			normalized = strings.ToLower(normalized)
		}
		if previous, exists := seen[normalized]; exists && previous != value {
			errs = append(errs, fmt.Sprintf("%s values %q and %q normalize identically", path, previous, value))
		}
		seen[normalized] = value
	}
	return errs
}

func KnownCapability(value string) bool {
	switch value {
	case "repository.read", "repository.write", "filesystem.read", "filesystem.write",
		"filesystem.delete", "network.connect", "network.inbound", "network.outbound",
		"process.execute", "commands.execute", "secrets.read", "secrets.expose",
		"environment.read", "tool.invoke", "mcp.invoke", "persistence.write",
		"agent.autonomous", "agent.external_side_effect":
		return true
	default:
		return false
	}
}

func validIdentifier(value string) bool { return identifierRE.MatchString(value) }

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func dedupeSorted(values []string) []string {
	set := stringSet(values)
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

// LegacyContract mirrors the short-lived embedded descriptor-v2 contract.
type LegacyContract struct {
	Capabilities   LegacyCapabilities  `yaml:"capabilities"`
	Tools          Tools               `yaml:"tools"`
	Data           LegacyData          `yaml:"data"`
	Preconditions  []string            `yaml:"preconditions"`
	Postconditions []string            `yaml:"postconditions"`
	Invariants     []string            `yaml:"invariants"`
	HumanApproval  LegacyHumanApproval `yaml:"human_approval"`
	Limits         LegacyLimits        `yaml:"limits"`
	Output         Output              `yaml:"output"`
}
type LegacyCapabilities struct {
	Repository RepositoryCapabilities `yaml:"repository"`
	Filesystem FilesystemCapabilities `yaml:"filesystem"`
	Network    NetworkCapabilities    `yaml:"network"`
	Process    ProcessCapabilities    `yaml:"process"`
	Secrets    SecretsCapabilities    `yaml:"secrets"`
}
type LegacyData struct {
	Inputs    []string   `yaml:"inputs"`
	Sensitive []string   `yaml:"sensitive"`
	Egress    DataEgress `yaml:"egress"`
}
type LegacyHumanApproval struct {
	RequiredFor []string `yaml:"required_for"`
}
type LegacyLimits struct {
	MaxToolCalls       *int `yaml:"max_tool_calls"`
	MaxRuntimeSeconds  *int `yaml:"max_runtime_seconds"`
	MaxNetworkRequests *int `yaml:"max_network_requests"`
}

func ValidateLegacyContract(c LegacyContract) []string {
	var errs []string
	errs = append(errs, validateStringList("contract.tools.allow", c.Tools.Allow, true)...)
	errs = append(errs, validateStringList("contract.tools.deny", c.Tools.Deny, true)...)
	denied := stringSet(c.Tools.Deny)
	for _, tool := range c.Tools.Allow {
		if _, exists := denied[tool]; exists {
			errs = append(errs, fmt.Sprintf("contract.tools contains %q in both allow and deny", tool))
		}
	}
	for path, value := range map[string]*int{
		"contract.limits.max_tool_calls":       c.Limits.MaxToolCalls,
		"contract.limits.max_runtime_seconds":  c.Limits.MaxRuntimeSeconds,
		"contract.limits.max_network_requests": c.Limits.MaxNetworkRequests,
	} {
		if value != nil && *value < 0 {
			errs = append(errs, path+" must be zero or greater")
		}
	}
	return dedupeSorted(errs)
}
