package skillmeta

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type SecurityImpact string

const (
	ImpactNarrowing SecurityImpact = "NARROWING"
	ImpactNone      SecurityImpact = "NO_SECURITY_IMPACT"
	ImpactExpansion SecurityImpact = "EXPANSION"
	ImpactMixed     SecurityImpact = "MIXED"
)

type ContractChange struct {
	Type       string `json:"type" yaml:"type"`
	Capability string `json:"capability,omitempty" yaml:"capability,omitempty"`
	Scope      string `json:"scope,omitempty" yaml:"scope,omitempty"`
	Tool       string `json:"tool,omitempty" yaml:"tool,omitempty"`
	Value      string `json:"value,omitempty" yaml:"value,omitempty"`
	Impact     string `json:"impact" yaml:"impact"`
}

type ContractDiff struct {
	Classification SecurityImpact   `json:"classification" yaml:"classification"`
	Changes        []ContractChange `json:"changes" yaml:"changes"`
}

func NormalizeContract(contract Contract) Contract {
	normalized := contract
	if normalized.SchemaVersion == ContractSchemaVersion {
		normalizeRuntime := func(set *RuntimeCapabilitySet) {
			set.Filesystem.Read = normalizedStrings(set.Filesystem.Read, true)
			set.Filesystem.Write = normalizedStrings(set.Filesystem.Write, true)
			set.Filesystem.Delete = normalizedStrings(set.Filesystem.Delete, true)
			set.Network.Hosts = normalizedHosts(set.Network.Hosts)
			set.Secrets.Read = normalizedStrings(set.Secrets.Read, false)
			set.Environment.Read = normalizedStrings(set.Environment.Read, false)
			set.Tools.Allow = normalizedStrings(set.Tools.Allow, false)
			set.Tools.Deny = normalizedStrings(set.Tools.Deny, false)
			set.MCP.Servers = normalizedStrings(set.MCP.Servers, false)
			set.MCP.Tools = normalizedStrings(set.MCP.Tools, false)
			set.Agent.ExternalTargets = normalizedStrings(set.Agent.ExternalTargets, false)
			sort.Slice(set.Commands.Allow, func(i, j int) bool { return commandKey(set.Commands.Allow[i]) < commandKey(set.Commands.Allow[j]) })
		}
		normalizeRuntime(&normalized.Capabilities.Runtime.Required)
		normalizeRuntime(&normalized.Capabilities.Runtime.Allowed)
		for key, values := range normalized.Capabilities.Semantic.Required {
			normalized.Capabilities.Semantic.Required[key] = normalizedStrings(values, false)
		}
		normalized.Identity.AcceptedPrincipals = normalizedStrings(normalized.Identity.AcceptedPrincipals, false)
		normalized.Identity.Credentials.RequiredScopes = normalizedStrings(normalized.Identity.Credentials.RequiredScopes, false)
		normalized.Data.Policies = append([]DataPolicy(nil), normalized.Data.Policies...)
		for i := range normalized.Data.Policies {
			normalized.Data.Policies[i].Purposes = normalizedStrings(normalized.Data.Policies[i].Purposes, false)
		}
		sort.Slice(normalized.Data.Policies, func(i, j int) bool { return normalized.Data.Policies[i].ID < normalized.Data.Policies[j].ID })
	}
	normalizeSet := func(set *CapabilitySet) {
		set.Repository.Read = normalizedStrings(set.Repository.Read, false)
		set.Repository.Write = normalizedStrings(set.Repository.Write, false)
		set.Filesystem.Read = normalizedStrings(set.Filesystem.Read, true)
		set.Filesystem.Write = normalizedStrings(set.Filesystem.Write, true)
		set.Network.Allow = normalizedHosts(set.Network.Allow)
	}
	normalizeSet(&normalized.Capabilities.Required)
	normalizeSet(&normalized.Capabilities.Allowed)
	normalized.Tools.Allow = normalizedStrings(normalized.Tools.Allow, false)
	normalized.Tools.Deny = normalizedStrings(normalized.Tools.Deny, false)
	normalized.Data.Classifications = normalizedStrings(normalized.Data.Classifications, false)
	normalized.Data.Egress.Allow = normalizedHosts(normalized.Data.Egress.Allow)
	normalized.Output.RequiredFields = normalizedStrings(normalized.Output.RequiredFields, false)
	normalized.Output.ForbiddenContent = normalizedStrings(normalized.Output.ForbiddenContent, false)
	normalized.Preconditions = append([]Condition(nil), normalized.Preconditions...)
	normalized.Postconditions = append([]Condition(nil), normalized.Postconditions...)
	normalized.Invariants = append([]Condition(nil), normalized.Invariants...)
	normalized.HumanApproval.Rules = append([]ApprovalRule(nil), normalized.HumanApproval.Rules...)
	normalized.Data.Flows = append([]DataFlow(nil), normalized.Data.Flows...)
	for i := range normalized.Data.Flows {
		normalized.Data.Flows[i].Source.Classification = strings.TrimSpace(normalized.Data.Flows[i].Source.Classification)
		normalized.Data.Flows[i].Sink.Type = strings.ToLower(strings.TrimSpace(normalized.Data.Flows[i].Sink.Type))
		normalized.Data.Flows[i].Sink.Destination = strings.TrimSpace(normalized.Data.Flows[i].Sink.Destination)
		if normalized.Data.Flows[i].Sink.Type == "network" {
			normalized.Data.Flows[i].Sink.Destination = strings.ToLower(normalized.Data.Flows[i].Sink.Destination)
		}
	}
	for i := range normalized.Preconditions {
		normalizeCondition(&normalized.Preconditions[i])
	}
	for i := range normalized.Postconditions {
		normalizeCondition(&normalized.Postconditions[i])
	}
	for i := range normalized.Invariants {
		normalizeCondition(&normalized.Invariants[i])
	}
	sort.Slice(normalized.Preconditions, func(i, j int) bool { return normalized.Preconditions[i].ID < normalized.Preconditions[j].ID })
	sort.Slice(normalized.Postconditions, func(i, j int) bool { return normalized.Postconditions[i].ID < normalized.Postconditions[j].ID })
	sort.Slice(normalized.Invariants, func(i, j int) bool { return normalized.Invariants[i].ID < normalized.Invariants[j].ID })
	sort.Slice(normalized.HumanApproval.Rules, func(i, j int) bool {
		return normalized.HumanApproval.Rules[i].ID < normalized.HumanApproval.Rules[j].ID
	})
	sort.Slice(normalized.Data.Flows, func(i, j int) bool {
		left := fmt.Sprintf("%s|%s|%s|%t", normalized.Data.Flows[i].Source.Classification, normalized.Data.Flows[i].Sink.Type, normalized.Data.Flows[i].Sink.Destination, normalized.Data.Flows[i].Allowed)
		right := fmt.Sprintf("%s|%s|%s|%t", normalized.Data.Flows[j].Source.Classification, normalized.Data.Flows[j].Sink.Type, normalized.Data.Flows[j].Sink.Destination, normalized.Data.Flows[j].Allowed)
		return left < right
	})
	return normalized
}

func CanonicalContractBytes(contract Contract) ([]byte, error) {
	normalized := NormalizeContract(contract)
	return yaml.Marshal(normalized)
}

func ContractDigest(contract Contract) (string, error) {
	data, err := CanonicalContractBytes(contract)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func DiffContracts(oldContract, newContract Contract) ContractDiff {
	oldContract = NormalizeContract(oldContract)
	newContract = NormalizeContract(newContract)
	if oldContract.SchemaVersion == ContractSchemaVersion && newContract.SchemaVersion == ContractSchemaVersion {
		return diffRuntimeContracts(oldContract, newContract)
	}
	changes := []ContractChange{}
	compareSet := func(capability string, oldValues, newValues []string, addedImpact, removedImpact string) {
		oldSet, newSet := stringSet(oldValues), stringSet(newValues)
		for _, value := range newValues {
			if _, exists := oldSet[value]; !exists {
				changes = append(changes, ContractChange{Type: "capability_scope_added", Capability: capability, Scope: value, Impact: addedImpact})
			}
		}
		for _, value := range oldValues {
			if _, exists := newSet[value]; !exists {
				changes = append(changes, ContractChange{Type: "capability_scope_removed", Capability: capability, Scope: value, Impact: removedImpact})
			}
		}
	}
	compareCapabilitySet := func(prefix string, oldSet, newSet CapabilitySet) {
		addedImpact, removedImpact, boundary := "none", "none", false
		if prefix == "allowed." {
			addedImpact, removedImpact, boundary = "expansion", "narrowing", true
		}
		compareSet(prefix+"repository.read", oldSet.Repository.Read, newSet.Repository.Read, addedImpact, removedImpact)
		compareSet(prefix+"repository.write", oldSet.Repository.Write, newSet.Repository.Write, addedImpact, removedImpact)
		compareSet(prefix+"filesystem.read", oldSet.Filesystem.Read, newSet.Filesystem.Read, addedImpact, removedImpact)
		compareSet(prefix+"filesystem.write", oldSet.Filesystem.Write, newSet.Filesystem.Write, addedImpact, removedImpact)
		if prefix == "required." {
			compareSet(prefix+"network.connect", oldSet.Network.Allow, newSet.Network.Allow, "none", "none")
		}
		compareBool(prefix+"process.execute", oldSet.Process.Execute, newSet.Process.Execute, boundary, &changes)
		compareBool(prefix+"secrets.read", oldSet.Secrets.Read, newSet.Secrets.Read, boundary, &changes)
	}
	compareCapabilitySet("required.", oldContract.Capabilities.Required, newContract.Capabilities.Required)
	compareCapabilitySet("allowed.", oldContract.Capabilities.Allowed, newContract.Capabilities.Allowed)
	compareNamedSet("tool_allowed", oldContract.Tools.Allow, newContract.Tools.Allow, "tool", &changes, false)
	// Adding a deny is narrowing; removing one is expansion.
	compareNamedSet("tool_denied", oldContract.Tools.Deny, newContract.Tools.Deny, "tool", &changes, true)
	compareNamedSet("network_destination", oldContract.Capabilities.Allowed.Network.Allow, newContract.Capabilities.Allowed.Network.Allow, "value", &changes, false)
	compareNamedSet("invariant", conditionIDs(oldContract.Invariants), conditionIDs(newContract.Invariants), "value", &changes, true)
	compareNamedSet("approval_requirement", approvalIDs(oldContract.HumanApproval.Rules), approvalIDs(newContract.HumanApproval.Rules), "value", &changes, true)
	compareLimit("max_tool_calls", oldContract.Limits.MaxToolCalls, newContract.Limits.MaxToolCalls, &changes)
	compareLimit("max_runtime_seconds", oldContract.Limits.MaxRuntimeSeconds, newContract.Limits.MaxRuntimeSeconds, &changes)
	compareLimit("max_network_requests", oldContract.Limits.MaxNetworkRequests, newContract.Limits.MaxNetworkRequests, &changes)
	sort.Slice(changes, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%s|%s", changes[i].Type, changes[i].Capability, changes[i].Scope, changes[i].Value) <
			fmt.Sprintf("%s|%s|%s|%s", changes[j].Type, changes[j].Capability, changes[j].Scope, changes[j].Value)
	})
	return ContractDiff{Classification: classifyChanges(changes), Changes: changes}
}

func diffRuntimeContracts(oldContract, newContract Contract) ContractDiff {
	changes := []ContractChange{}
	oldSet, newSet := oldContract.Capabilities.Runtime.Allowed, newContract.Capabilities.Runtime.Allowed
	compare := func(name string, oldValues, newValues []string, reverse bool) {
		compareNamedSet("capability_"+strings.ReplaceAll(name, ".", "_"), oldValues, newValues, "value", &changes, reverse)
	}
	compare("filesystem.read", oldSet.Filesystem.Read, newSet.Filesystem.Read, false)
	compare("filesystem.write", oldSet.Filesystem.Write, newSet.Filesystem.Write, false)
	compare("filesystem.delete", oldSet.Filesystem.Delete, newSet.Filesystem.Delete, false)
	compare("network.hosts", oldSet.Network.Hosts, newSet.Network.Hosts, false)
	compare("secrets.read", oldSet.Secrets.Read, newSet.Secrets.Read, false)
	compare("environment.read", oldSet.Environment.Read, newSet.Environment.Read, false)
	compare("tools.allow", oldSet.Tools.Allow, newSet.Tools.Allow, false)
	compare("tools.deny", oldSet.Tools.Deny, newSet.Tools.Deny, true)
	compare("mcp.servers", oldSet.MCP.Servers, newSet.MCP.Servers, false)
	compare("mcp.tools", oldSet.MCP.Tools, newSet.MCP.Tools, false)
	compare("agent.external_targets", oldSet.Agent.ExternalTargets, newSet.Agent.ExternalTargets, false)
	compareBool("allowed.network.inbound", oldSet.Network.Inbound, newSet.Network.Inbound, true, &changes)
	compareBool("allowed.network.outbound", oldSet.Network.Outbound, newSet.Network.Outbound, true, &changes)
	compareBool("allowed.commands.execute", oldSet.Commands.Execute, newSet.Commands.Execute, true, &changes)
	compareBool("allowed.secrets.expose", oldSet.Secrets.Expose, newSet.Secrets.Expose, true, &changes)
	compareBool("allowed.persistence", oldSet.Persistence, newSet.Persistence, true, &changes)
	compareBool("allowed.agent.autonomous_actions", oldSet.Agent.AutonomousActions, newSet.Agent.AutonomousActions, true, &changes)
	compareBool("allowed.agent.external_side_effects", oldSet.Agent.ExternalSideEffects, newSet.Agent.ExternalSideEffects, true, &changes)
	oldCommands, newCommands := []string{}, []string{}
	for _, r := range oldSet.Commands.Allow {
		oldCommands = append(oldCommands, commandKey(r))
	}
	for _, r := range newSet.Commands.Allow {
		newCommands = append(newCommands, commandKey(r))
	}
	compare("commands.allow", oldCommands, newCommands, false)
	compareNamedSet("accepted_principal", oldContract.Identity.AcceptedPrincipals, newContract.Identity.AcceptedPrincipals, "value", &changes, false)
	compareNamedSet("credential_scope", oldContract.Identity.Credentials.RequiredScopes, newContract.Identity.Credentials.RequiredScopes, "value", &changes, true)
	compareLimit("credential_max_ttl_seconds", oldContract.Identity.Credentials.MaxTTLSeconds, newContract.Identity.Credentials.MaxTTLSeconds, &changes)
	compareBool("identity.allow_token_passthrough", oldContract.Identity.Credentials.AllowTokenPassthrough, newContract.Identity.Credentials.AllowTokenPassthrough, true, &changes)
	compareConstraintBool("identity.require_audience_binding", oldContract.Identity.Credentials.RequireAudienceBinding, newContract.Identity.Credentials.RequireAudienceBinding, &changes)
	compareBool("delegation.allowed", oldContract.Delegation.Allowed, newContract.Delegation.Allowed, true, &changes)
	compareLimit("delegation.max_depth", oldContract.Delegation.MaxDepth, newContract.Delegation.MaxDepth, &changes)
	compareConstraintBool("delegation.require_child_capability_subset", oldContract.Delegation.RequireChildCapabilitySubset, newContract.Delegation.RequireChildCapabilitySubset, &changes)
	compareConstraintBool("delegation.require_authenticated_origin", oldContract.Delegation.RequireAuthenticatedOrigin, newContract.Delegation.RequireAuthenticatedOrigin, &changes)
	compareDataPolicies(oldContract.Data.Policies, newContract.Data.Policies, &changes)
	compareNamedSet("invariant", conditionIDs(oldContract.Invariants), conditionIDs(newContract.Invariants), "value", &changes, true)
	compareNamedSet("approval_requirement", approvalIDs(oldContract.HumanApproval.Rules), approvalIDs(newContract.HumanApproval.Rules), "value", &changes, true)
	compareLimit("max_tool_calls", oldContract.Resources.MaxToolCalls, newContract.Resources.MaxToolCalls, &changes)
	compareLimit("max_runtime_seconds", oldContract.Resources.MaxRuntimeSeconds, newContract.Resources.MaxRuntimeSeconds, &changes)
	compareLimit("max_network_requests", oldContract.Resources.MaxNetworkRequests, newContract.Resources.MaxNetworkRequests, &changes)
	compareLimit("max_memory_mb", oldContract.Resources.MaxMemoryMB, newContract.Resources.MaxMemoryMB, &changes)
	compareLimit("max_network_bytes", oldContract.Resources.MaxNetworkBytes, newContract.Resources.MaxNetworkBytes, &changes)
	compareLimit("max_external_mutations", oldContract.Resources.MaxExternalMutations, newContract.Resources.MaxExternalMutations, &changes)
	compareLimit("max_delegation_depth", oldContract.Resources.MaxDelegationDepth, newContract.Resources.MaxDelegationDepth, &changes)
	compareLimit("max_model_tokens", oldContract.Resources.MaxModelTokens, newContract.Resources.MaxModelTokens, &changes)
	sort.Slice(changes, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%s", changes[i].Type, changes[i].Capability, changes[i].Value) < fmt.Sprintf("%s|%s|%s", changes[j].Type, changes[j].Capability, changes[j].Value)
	})
	return ContractDiff{Classification: classifyChanges(changes), Changes: changes}
}

func compareConstraintBool(name string, oldValue, newValue *bool, changes *[]ContractChange) {
	oldEnabled, newEnabled := enabled(oldValue), enabled(newValue)
	if oldEnabled == newEnabled {
		return
	}
	impact, changeType := "expansion", "constraint_removed"
	if newEnabled {
		impact, changeType = "narrowing", "constraint_added"
	}
	*changes = append(*changes, ContractChange{Type: changeType, Capability: name, Impact: impact})
}

func compareDataPolicies(oldPolicies, newPolicies []DataPolicy, changes *[]ContractChange) {
	oldByID, newByID := map[string]DataPolicy{}, map[string]DataPolicy{}
	for _, policy := range oldPolicies {
		oldByID[policy.ID] = policy
	}
	for _, policy := range newPolicies {
		newByID[policy.ID] = policy
	}
	retentionRank := map[string]int{"invocation": 0, "ephemeral": 1, "session": 2, "persistent": 3}
	for id, oldPolicy := range oldByID {
		newPolicy, ok := newByID[id]
		if !ok {
			*changes = append(*changes, ContractChange{Type: "data_policy_removed", Value: id, Impact: "expansion"})
			continue
		}
		if oldPolicy.Retention != newPolicy.Retention {
			impact := "narrowing"
			if retentionRank[newPolicy.Retention] > retentionRank[oldPolicy.Retention] {
				impact = "expansion"
			}
			*changes = append(*changes, ContractChange{Type: "data_retention_changed", Value: id + ":" + newPolicy.Retention, Impact: impact})
		}
	}
	for id := range newByID {
		if _, ok := oldByID[id]; !ok {
			*changes = append(*changes, ContractChange{Type: "data_policy_added", Value: id, Impact: "narrowing"})
		}
	}
}

func normalizedStrings(values []string, cleanPath bool) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if cleanPath {
			value = strings.ReplaceAll(value, "\\", "/")
			value = path.Clean(value)
		}
		set[value] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizedHosts(values []string) []string {
	normalized := make([]string, len(values))
	for i := range values {
		normalized[i] = strings.ToLower(strings.TrimSpace(values[i]))
	}
	return normalizedStrings(normalized, false)
}

func normalizeCondition(condition *Condition) {
	condition.Capabilities = normalizedStrings(condition.Capabilities, false)
	condition.Classifications = normalizedStrings(condition.Classifications, false)
}

func compareBool(name string, oldValue, newValue *bool, securityBoundary bool, changes *[]ContractChange) {
	oldEnabled := oldValue != nil && *oldValue
	newEnabled := newValue != nil && *newValue
	if oldEnabled == newEnabled {
		return
	}
	impact := "narrowing"
	changeType := "capability_removed"
	if !securityBoundary {
		impact = "none"
	}
	if newEnabled {
		if securityBoundary {
			impact = "expansion"
		}
		changeType = "capability_added"
	}
	*changes = append(*changes, ContractChange{Type: changeType, Capability: name, Impact: impact})
}

func compareNamedSet(changeType string, oldValues, newValues []string, field string, changes *[]ContractChange, reverse bool) {
	oldSet, newSet := stringSet(oldValues), stringSet(newValues)
	appendChange := func(value, impact, suffix string) {
		change := ContractChange{Type: changeType + "_" + suffix, Impact: impact}
		switch field {
		case "tool":
			change.Tool = value
		default:
			change.Value = value
		}
		*changes = append(*changes, change)
	}
	for _, value := range newValues {
		if _, exists := oldSet[value]; !exists {
			impact := "expansion"
			if reverse {
				impact = "narrowing"
			}
			appendChange(value, impact, "added")
		}
	}
	for _, value := range oldValues {
		if _, exists := newSet[value]; !exists {
			impact := "narrowing"
			if reverse {
				impact = "expansion"
			}
			appendChange(value, impact, "removed")
		}
	}
}

func compareLimit(name string, oldValue, newValue *int, changes *[]ContractChange) {
	if intPtrEqual(oldValue, newValue) {
		return
	}
	impact := "narrowing"
	changeType := "limit_tightened"
	if oldValue != nil && (newValue == nil || *newValue > *oldValue) {
		impact = "expansion"
		changeType = "limit_relaxed"
	}
	if oldValue == nil && newValue != nil {
		impact = "narrowing"
		changeType = "limit_tightened"
	}
	*changes = append(*changes, ContractChange{Type: changeType, Value: name + ":" + limitValue(newValue), Impact: impact})
}

func intPtrEqual(left, right *int) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
func limitValue(value *int) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *value)
}
func conditionIDs(values []Condition) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = value.ID
	}
	return out
}
func approvalIDs(values []ApprovalRule) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = value.ID
	}
	return out
}
func classifyChanges(changes []ContractChange) SecurityImpact {
	expansion, narrowing := false, false
	for _, change := range changes {
		expansion = expansion || change.Impact == "expansion"
		narrowing = narrowing || change.Impact == "narrowing"
	}
	switch {
	case expansion && narrowing:
		return ImpactMixed
	case expansion:
		return ImpactExpansion
	case narrowing:
		return ImpactNarrowing
	default:
		return ImpactNone
	}
}
