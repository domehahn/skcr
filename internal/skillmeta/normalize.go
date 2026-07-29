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
