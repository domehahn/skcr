package skillmeta

import (
	"fmt"
	"strings"
)

var evalOutputProperties = map[string]bool{"non_empty": true, "json": true, "no_secrets": true}
var evalSecurityAssertions = map[string]bool{"no_external_side_effects": true, "no_errors": true, "no_forbidden_capabilities": true, "containment_compliant": true}

func validateEvalV2(document EvalDocument, descriptor Descriptor, contract *Contract, path string) []string {
	if document.Scenarios == nil {
		return []string{path + ": scenarios is required and must be a list"}
	}
	goalIDs := map[string]bool{}
	if descriptor.Goal != nil {
		for _, criterion := range descriptor.Goal.SuccessCriteria {
			goalIDs[criterion.ID] = true
		}
	}
	invariantIDs := map[string]bool{}
	if contract != nil {
		for _, invariant := range contract.Invariants {
			invariantIDs[invariant.ID] = true
		}
	}
	seen := map[string]bool{}
	var errs []string
	for i, scenario := range document.Scenarios {
		prefix := fmt.Sprintf("%s: scenarios[%d]", path, i)
		if !validIdentifier(scenario.ID) {
			errs = append(errs, fmt.Sprintf("%s.id %q is invalid", prefix, scenario.ID))
		} else if seen[scenario.ID] {
			errs = append(errs, fmt.Sprintf("%s duplicate id %q", path, scenario.ID))
		}
		seen[scenario.ID] = true
		if strings.TrimSpace(scenario.Description) == "" {
			errs = append(errs, prefix+".description must not be blank")
		}
		if scenario.Type != "behavioral" && scenario.Type != "adversarial" {
			errs = append(errs, fmt.Sprintf("%s.type %q is unsupported", prefix, scenario.Type))
		}
		if strings.TrimSpace(scenario.Input.Message) == "" {
			errs = append(errs, prefix+".input.message must not be blank")
		}
		if scenario.Context == nil {
			errs = append(errs, prefix+".context is required and must be a mapping")
		}
		if scenario.Environment == nil {
			errs = append(errs, prefix+".environment is required and must be a mapping")
		}
		errs = append(errs, validateStringList(prefix+".tools.available", scenario.ToolsV2.Available, false)...)
		for field, values := range map[string][]string{"required": scenario.Expect.Required, "allowed": scenario.Expect.Allowed, "forbidden": scenario.Expect.Forbidden, "forbidden_capabilities": scenario.Expect.ForbiddenCapabilities, "output_properties": scenario.Expect.OutputProperties, "assertions": scenario.Expect.Assertions} {
			errs = append(errs, validateStringList(prefix+".expect."+field, values, false)...)
		}
		if scenario.Expect.Arguments == nil {
			errs = append(errs, prefix+".expect.arguments is required and must be a mapping")
		}
		available := stringSet(scenario.ToolsV2.Available)
		forbidden := stringSet(scenario.Expect.Forbidden)
		for _, group := range []struct {
			name   string
			values []string
		}{{"required", scenario.Expect.Required}, {"allowed", scenario.Expect.Allowed}, {"forbidden", scenario.Expect.Forbidden}} {
			for _, tool := range group.values {
				if _, ok := available[tool]; !ok {
					errs = append(errs, fmt.Sprintf("%s.expect.%s tool %q is not available", prefix, group.name, tool))
				}
			}
		}
		for _, tool := range append(append([]string{}, scenario.Expect.Required...), scenario.Expect.Allowed...) {
			if _, denied := forbidden[tool]; denied {
				errs = append(errs, fmt.Sprintf("%s tool %q is both permitted and forbidden", prefix, tool))
			}
		}
		for _, capability := range scenario.Expect.ForbiddenCapabilities {
			if !KnownCapability(capability) {
				errs = append(errs, fmt.Sprintf("%s references unknown capability %q", prefix, capability))
			}
		}
		for _, property := range scenario.Expect.OutputProperties {
			if !evalOutputProperties[property] {
				errs = append(errs, fmt.Sprintf("%s.expect.output_properties contains unsupported value %q", prefix, property))
			}
		}
		for _, assertion := range scenario.Expect.Assertions {
			if !evalSecurityAssertions[assertion] {
				errs = append(errs, fmt.Sprintf("%s.expect.assertions contains unsupported value %q", prefix, assertion))
			}
		}
		errs = append(errs, validateStringList(prefix+".goal_refs.must_satisfy", scenario.GoalRefs.MustSatisfy, false)...)
		for _, id := range scenario.GoalRefs.MustSatisfy {
			if !goalIDs[id] {
				errs = append(errs, fmt.Sprintf("%s references unknown goal criterion %q", prefix, id))
			}
		}
		errs = append(errs, validateStringList(prefix+".invariant_refs.must_hold", scenario.InvariantRefs.MustHold, false)...)
		for _, id := range scenario.InvariantRefs.MustHold {
			if contract != nil && !invariantIDs[id] {
				errs = append(errs, fmt.Sprintf("%s references unknown invariant %q", prefix, id))
			}
		}
		if scenario.Type == "adversarial" && (scenario.Attack == nil || strings.TrimSpace(scenario.Attack.Category) == "") {
			errs = append(errs, prefix+".attack.category is required for adversarial scenarios")
		}
		if scenario.Attack != nil && strings.TrimSpace(scenario.Attack.Category) == "" {
			errs = append(errs, prefix+".attack.category must not be blank")
		}
		errs = append(errs, validateContainment(prefix, scenario, contract)...)
	}
	return dedupeSorted(errs)
}

func validateContainment(prefix string, scenario EvalScenario, contract *Contract) []string {
	if scenario.Containment == nil {
		return []string{prefix + ".containment is required"}
	}
	c := scenario.Containment
	var errs []string
	if c.Required == nil {
		errs = append(errs, prefix+".containment.required is required")
	}
	if c.RequireEnforcement == nil {
		errs = append(errs, prefix+".containment.require_enforcement is required")
	}
	if c.RequireNativeIsolation == nil {
		errs = append(errs, prefix+".containment.require_native_isolation is required")
	}
	if c.AllowedTargets == nil {
		errs = append(errs, prefix+".containment.allowed_targets is required and must be a mapping")
	}
	if enabled(c.RequireNativeIsolation) && !enabled(c.RequireEnforcement) {
		errs = append(errs, prefix+".containment native isolation requires enforcement")
	}
	assertions := stringSet(scenario.Expect.Assertions)
	if enabled(c.Required) {
		if _, ok := assertions["containment_compliant"]; !ok {
			errs = append(errs, prefix+".containment.required requires containment_compliant assertion")
		}
	}
	for capability, targets := range c.AllowedTargets {
		if strings.TrimSpace(capability) == "" {
			errs = append(errs, prefix+".containment.allowed_targets key must not be blank")
		}
		errs = append(errs, validateStringList(prefix+".containment.allowed_targets."+capability, targets, false)...)
		if contract != nil {
			errs = append(errs, containmentSubset(prefix, capability, targets, *contract)...)
		}
	}
	return errs
}

func containmentSubset(prefix, capability string, targets []string, contract Contract) []string {
	if contract.SchemaVersion != ContractSchemaVersion {
		return nil
	}
	a := contract.Capabilities.Runtime.Allowed
	var allowed []string
	switch capability {
	case "filesystem.read":
		allowed = a.Filesystem.Read
	case "filesystem.write":
		allowed = a.Filesystem.Write
	case "filesystem.delete":
		allowed = a.Filesystem.Delete
	case "network.outbound":
		allowed = a.Network.Hosts
	case "tool.invoke":
		allowed = a.Tools.Allow
	case "mcp.invoke":
		allowed = a.MCP.Tools
	case "external.action":
		allowed = a.Agent.ExternalTargets
	default:
		return nil
	}
	set := stringSet(allowed)
	var errs []string
	for _, target := range targets {
		if _, ok := set[target]; !ok {
			errs = append(errs, fmt.Sprintf("%s.containment target %q for %s exceeds contract authority", prefix, target, capability))
		}
	}
	return errs
}
