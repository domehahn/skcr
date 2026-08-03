package skillmeta

import (
	"fmt"
	"strings"
)

func requiredContractV2Presence() map[string]bool {
	paths := []string{
		"schema_version", "capabilities", "capabilities.semantic", "capabilities.semantic.required",
		"capabilities.runtime", "capabilities.runtime.required", "capabilities.runtime.allowed",
		"identity", "identity.accepted_principals", "identity.credentials", "identity.credentials.max_ttl_seconds",
		"identity.credentials.require_audience_binding", "identity.credentials.allow_token_passthrough", "identity.credentials.required_scopes",
		"delegation", "delegation.allowed", "delegation.max_depth", "delegation.require_child_capability_subset", "delegation.require_authenticated_origin",
		"data", "data.classifications", "data.policies", "data.egress", "data.egress.allow", "data.flows",
		"preconditions", "postconditions", "invariants", "human_approval", "human_approval.rules",
		"effects", "effects.tools", "resources", "resources.scope", "output", "output.format",
		"output.required_fields", "output.forbidden_content",
	}
	for _, scope := range []string{"required", "allowed"} {
		prefix := "capabilities.runtime." + scope
		for _, path := range []string{
			"filesystem", "filesystem.read", "filesystem.write", "filesystem.delete",
			"network", "network.inbound", "network.outbound", "network.hosts",
			"commands", "commands.execute", "commands.allow", "secrets", "secrets.read", "secrets.expose",
			"environment", "environment.read", "tools", "tools.allow", "tools.deny",
			"mcp", "mcp.servers", "mcp.tools", "persistence", "agent", "agent.autonomous_actions",
			"agent.external_side_effects", "agent.confirm_destructive", "agent.confirm_external", "agent.external_targets",
		} {
			paths = append(paths, prefix+"."+path)
		}
	}
	out := make(map[string]bool, len(paths))
	for _, path := range paths {
		out[path] = true
	}
	return out
}

func validateContractV2(c Contract) []string {
	var errs []string
	for path := range requiredContractV2Presence() {
		if !c.presence[path] {
			errs = append(errs, path+" is required")
		}
	}
	if c.Capabilities.Semantic.Required == nil {
		errs = append(errs, "capabilities.semantic.required is required and must be a mapping")
	}
	for id, purposes := range c.Capabilities.Semantic.Required {
		if !KnownCapability(id) {
			errs = append(errs, fmt.Sprintf("capabilities.semantic.required contains unknown capability %q", id))
		}
		errs = append(errs, validateStringList("capabilities.semantic.required."+id, purposes, false)...)
	}
	errs = append(errs, validateRuntimeSet("capabilities.runtime.required", c.Capabilities.Runtime.Required)...)
	errs = append(errs, validateRuntimeSet("capabilities.runtime.allowed", c.Capabilities.Runtime.Allowed)...)
	errs = append(errs, validateRuntimeSubset(c.Capabilities.Runtime.Required, c.Capabilities.Runtime.Allowed)...)
	errs = append(errs, validateIdentityAndDelegation(c)...)
	errs = append(errs, validateSharedContract(c)...)
	if c.Resources.Scope != "invocation" {
		errs = append(errs, `resources.scope must be "invocation"`)
	}
	for path, value := range map[string]*int{
		"max_runtime_seconds": c.Resources.MaxRuntimeSeconds, "max_memory_mb": c.Resources.MaxMemoryMB,
		"max_network_bytes": c.Resources.MaxNetworkBytes, "max_network_requests": c.Resources.MaxNetworkRequests,
		"max_tool_calls": c.Resources.MaxToolCalls, "max_steps": c.Resources.MaxSteps, "max_replans": c.Resources.MaxReplans,
		"max_delegation_depth": c.Resources.MaxDelegationDepth, "max_model_tokens": c.Resources.MaxModelTokens,
		"max_external_mutations": c.Resources.MaxExternalMutations,
	} {
		if value != nil && *value < 0 {
			errs = append(errs, "resources."+path+" must be zero or greater")
		}
	}
	errs = append(errs, validateEffects(c)...)
	return dedupeSorted(errs)
}

func validateRuntimeSet(path string, set RuntimeCapabilitySet) []string {
	var errs []string
	for name, values := range map[string][]string{
		"filesystem.read": set.Filesystem.Read, "filesystem.write": set.Filesystem.Write, "filesystem.delete": set.Filesystem.Delete,
		"network.hosts": set.Network.Hosts, "secrets.read": set.Secrets.Read, "environment.read": set.Environment.Read,
		"tools.allow": set.Tools.Allow, "tools.deny": set.Tools.Deny, "mcp.servers": set.MCP.Servers, "mcp.tools": set.MCP.Tools,
		"agent.external_targets": set.Agent.ExternalTargets,
	} {
		errs = append(errs, validateStringList(path+"."+name, values, false)...)
	}
	for name, value := range map[string]*bool{
		"network.inbound": set.Network.Inbound, "network.outbound": set.Network.Outbound, "commands.execute": set.Commands.Execute,
		"secrets.expose": set.Secrets.Expose, "persistence": set.Persistence, "agent.autonomous_actions": set.Agent.AutonomousActions,
		"agent.external_side_effects": set.Agent.ExternalSideEffects, "agent.confirm_destructive": set.Agent.ConfirmDestructive,
		"agent.confirm_external": set.Agent.ConfirmExternal,
	} {
		if value == nil {
			errs = append(errs, path+"."+name+" is required and must be boolean")
		}
	}
	seenCommands := map[string]bool{}
	for i, rule := range set.Commands.Allow {
		if strings.TrimSpace(rule.Executable) == "" {
			errs = append(errs, fmt.Sprintf("%s.commands.allow[%d].executable must not be blank", path, i))
		}
		if seenCommands[rule.Executable] {
			errs = append(errs, path+".commands.allow contains duplicate executable "+rule.Executable)
		}
		seenCommands[rule.Executable] = true
		errs = append(errs, validateStringList(fmt.Sprintf("%s.commands.allow[%d].argv_prefix", path, i), nonnilStrings(rule.ArgvPrefix), false)...)
	}
	if set.Network.Outbound != nil && *set.Network.Outbound && len(set.Network.Hosts) == 0 {
		errs = append(errs, path+".network.outbound requires constrained hosts")
	}
	if set.Commands.Execute != nil && *set.Commands.Execute && len(set.Commands.Allow) == 0 {
		errs = append(errs, path+".commands.execute requires a command allowlist")
	}
	if set.Secrets.Expose != nil && *set.Secrets.Expose && len(set.Secrets.Read) == 0 {
		errs = append(errs, path+".secrets.expose requires explicit secret IDs")
	}
	denied := stringSet(set.Tools.Deny)
	for _, tool := range set.Tools.Allow {
		if _, exists := denied[tool]; exists {
			errs = append(errs, fmt.Sprintf("%s.tools contains %q in both allow and deny", path, tool))
		}
	}
	if enabled(set.Agent.ExternalSideEffects) && len(set.Agent.ExternalTargets) == 0 {
		errs = append(errs, path+".agent.external_side_effects requires external_targets")
	}
	if (len(set.Filesystem.Delete) > 0 || enabled(set.Agent.ExternalSideEffects)) && !enabled(set.Agent.ConfirmDestructive) {
		errs = append(errs, path+" requires confirm_destructive for destructive authority")
	}
	if enabled(set.Agent.ExternalSideEffects) && !enabled(set.Agent.ConfirmExternal) {
		errs = append(errs, path+" requires confirm_external for external side effects")
	}
	return errs
}

func validateRuntimeSubset(required, allowed RuntimeCapabilitySet) []string {
	var errs []string
	checkLists := func(name string, req, allow []string) {
		allowedSet := stringSet(allow)
		for _, value := range req {
			if _, ok := allowedSet[value]; !ok {
				errs = append(errs, fmt.Sprintf("required runtime capability %s scope %q is outside allowed capabilities", name, value))
			}
		}
	}
	checkLists("filesystem.read", required.Filesystem.Read, allowed.Filesystem.Read)
	checkLists("filesystem.write", required.Filesystem.Write, allowed.Filesystem.Write)
	checkLists("filesystem.delete", required.Filesystem.Delete, allowed.Filesystem.Delete)
	checkLists("network.hosts", required.Network.Hosts, allowed.Network.Hosts)
	checkLists("secrets.read", required.Secrets.Read, allowed.Secrets.Read)
	checkLists("environment.read", required.Environment.Read, allowed.Environment.Read)
	checkLists("tools.allow", required.Tools.Allow, allowed.Tools.Allow)
	checkLists("mcp.servers", required.MCP.Servers, allowed.MCP.Servers)
	checkLists("mcp.tools", required.MCP.Tools, allowed.MCP.Tools)
	checkLists("agent.external_targets", required.Agent.ExternalTargets, allowed.Agent.ExternalTargets)
	for name, pair := range map[string][2]*bool{"network.inbound": {required.Network.Inbound, allowed.Network.Inbound}, "network.outbound": {required.Network.Outbound, allowed.Network.Outbound}, "commands.execute": {required.Commands.Execute, allowed.Commands.Execute}, "secrets.expose": {required.Secrets.Expose, allowed.Secrets.Expose}, "persistence": {required.Persistence, allowed.Persistence}, "agent.autonomous_actions": {required.Agent.AutonomousActions, allowed.Agent.AutonomousActions}, "agent.external_side_effects": {required.Agent.ExternalSideEffects, allowed.Agent.ExternalSideEffects}} {
		if enabled(pair[0]) && !enabled(pair[1]) {
			errs = append(errs, "required runtime capability "+name+" is outside allowed capabilities")
		}
	}
	allowedCommands := map[string]bool{}
	for _, r := range allowed.Commands.Allow {
		allowedCommands[commandKey(r)] = true
	}
	for _, r := range required.Commands.Allow {
		if !allowedCommands[commandKey(r)] {
			errs = append(errs, "required runtime command "+r.Executable+" is outside allowed capabilities")
		}
	}
	return errs
}

func validateSharedContract(c Contract) []string {
	var errs []string
	errs = append(errs, validateStringList("data.classifications", c.Data.Classifications, false)...)
	errs = append(errs, validateStringList("data.egress.allow", c.Data.Egress.Allow, false)...)
	classifications := stringSet(c.Data.Classifications)
	seenPolicies := map[string]bool{}
	for i, policy := range c.Data.Policies {
		prefix := fmt.Sprintf("data.policies[%d]", i)
		if _, ok := classifications[policy.ID]; !ok {
			errs = append(errs, fmt.Sprintf("%s.id %q is not a declared classification", prefix, policy.ID))
		} else if seenPolicies[policy.ID] {
			errs = append(errs, "data.policies contains duplicate classification "+policy.ID)
		}
		seenPolicies[policy.ID] = true
		if !map[string]bool{"public": true, "internal": true, "confidential": true, "restricted": true}[policy.Sensitivity] {
			errs = append(errs, fmt.Sprintf("%s.sensitivity %q is unsupported", prefix, policy.Sensitivity))
		}
		errs = append(errs, validateStringList(prefix+".purposes", policy.Purposes, false)...)
		if !map[string]bool{"invocation": true, "session": true, "ephemeral": true, "persistent": true}[policy.Retention] {
			errs = append(errs, fmt.Sprintf("%s.retention %q is unsupported", prefix, policy.Retention))
		}
	}
	if c.SchemaVersion == ContractSchemaVersion {
		for classification := range classifications {
			if !seenPolicies[classification] {
				errs = append(errs, "data classification "+classification+" requires a purpose and retention policy")
			}
		}
	}
	egress := stringSet(normalizedHosts(c.Data.Egress.Allow))
	for i, flow := range c.Data.Flows {
		if strings.TrimSpace(flow.Source.Classification) == "" {
			errs = append(errs, fmt.Sprintf("data.flows[%d].source.classification is required", i))
		} else if _, ok := classifications[flow.Source.Classification]; !ok {
			errs = append(errs, fmt.Sprintf("data.flows[%d] uses undeclared classification %q", i, flow.Source.Classification))
		}
		if strings.TrimSpace(flow.Sink.Type) == "" {
			errs = append(errs, fmt.Sprintf("data.flows[%d].sink.type is required", i))
		}
		if flow.Allowed && flow.Sink.Type == "network" {
			if flow.Sink.Destination == "" {
				errs = append(errs, fmt.Sprintf("data.flows[%d] allowed network sink requires destination", i))
			} else if _, ok := egress[strings.ToLower(strings.TrimSpace(flow.Sink.Destination))]; !ok {
				errs = append(errs, fmt.Sprintf("data.flows[%d] network destination %q is not in data.egress.allow", i, flow.Sink.Destination))
			}
		}
	}
	errs = append(errs, validateConditions("preconditions", c.Preconditions)...)
	errs = append(errs, validateConditions("postconditions", c.Postconditions)...)
	errs = append(errs, validateConditions("invariants", c.Invariants)...)
	errs = append(errs, validateApproval(c.HumanApproval)...)
	if strings.TrimSpace(c.Output.Format) == "" {
		errs = append(errs, "output.format must not be blank")
	}
	errs = append(errs, validateStringList("output.required_fields", c.Output.RequiredFields, false)...)
	errs = append(errs, validateStringList("output.forbidden_content", c.Output.ForbiddenContent, false)...)
	return errs
}

func validateIdentityAndDelegation(c Contract) []string {
	var errs []string
	errs = append(errs, validateStringList("identity.accepted_principals", c.Identity.AcceptedPrincipals, false)...)
	if len(c.Identity.AcceptedPrincipals) == 0 {
		errs = append(errs, "identity.accepted_principals must not be empty")
	}
	credentials := c.Identity.Credentials
	if credentials.MaxTTLSeconds == nil || *credentials.MaxTTLSeconds <= 0 {
		errs = append(errs, "identity.credentials.max_ttl_seconds must be greater than zero")
	}
	if credentials.RequireAudienceBinding == nil || credentials.AllowTokenPassthrough == nil {
		errs = append(errs, "identity credential booleans are required")
	}
	errs = append(errs, validateStringList("identity.credentials.required_scopes", credentials.RequiredScopes, false)...)
	if enabled(credentials.AllowTokenPassthrough) && !enabled(credentials.RequireAudienceBinding) {
		errs = append(errs, "identity token passthrough requires audience binding")
	}
	d := c.Delegation
	if d.Allowed == nil || d.MaxDepth == nil || d.RequireChildCapabilitySubset == nil || d.RequireAuthenticatedOrigin == nil {
		errs = append(errs, "delegation fields are required")
		return errs
	}
	if *d.MaxDepth < 0 {
		errs = append(errs, "delegation.max_depth must be zero or greater")
	}
	if !*d.Allowed && *d.MaxDepth != 0 {
		errs = append(errs, "delegation.max_depth must be zero when delegation is disabled")
	}
	if *d.Allowed && (*d.MaxDepth == 0 || !*d.RequireChildCapabilitySubset || !*d.RequireAuthenticatedOrigin) {
		errs = append(errs, "enabled delegation requires positive depth, child capability subset, and authenticated origin")
	}
	if c.Resources.MaxDelegationDepth != nil && *d.MaxDepth > *c.Resources.MaxDelegationDepth {
		errs = append(errs, "delegation.max_depth exceeds resources.max_delegation_depth")
	}
	return errs
}

func validateEffects(c Contract) []string {
	var errs []string
	seen := map[string]bool{}
	allowed := map[string]bool{"pure": true, "read": true, "reversible_write": true, "irreversible_write": true, "external_side_effect": true, "destructive": true}
	approvalTools := map[string]bool{}
	for _, rule := range c.HumanApproval.Rules {
		if rule.Action.Tool != "" {
			approvalTools[rule.Action.Tool] = true
		}
	}
	for i, effect := range c.Effects.Tools {
		if strings.TrimSpace(effect.Tool) == "" {
			errs = append(errs, fmt.Sprintf("effects.tools[%d].tool must not be blank", i))
		} else if seen[effect.Tool] {
			errs = append(errs, "effects.tools contains duplicate tool "+effect.Tool)
		}
		seen[effect.Tool] = true
		if !allowed[effect.Class] {
			errs = append(errs, fmt.Sprintf("effects.tools[%d].class %q is unsupported", i, effect.Class))
		}
		if (effect.Class == "destructive" || effect.Class == "irreversible_write") && !approvalTools[effect.Tool] {
			errs = append(errs, "destructive or irreversible tool "+effect.Tool+" requires a human approval rule")
		}
	}
	return errs
}

func enabled(value *bool) bool { return value != nil && *value }
func commandKey(rule CommandRule) string {
	return rule.Executable + "\x00" + strings.Join(rule.ArgvPrefix, "\x00")
}
func nonnilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
