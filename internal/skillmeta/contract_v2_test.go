package skillmeta

import (
	"strings"
	"testing"
)

func TestContractV2SecureDefaults(t *testing.T) {
	contract := NewContractV2()
	if errs := ValidateContract(contract); len(errs) != 0 {
		t.Fatalf("default v2 contract invalid: %v", errs)
	}
	allowed := contract.Capabilities.Runtime.Allowed
	if enabled(allowed.Network.Outbound) || enabled(allowed.Commands.Execute) || enabled(allowed.Persistence) || len(allowed.Filesystem.Delete) != 0 || len(allowed.Secrets.Read) != 0 {
		t.Fatal("v2 defaults are not least privilege")
	}
}

func TestContractV2RuntimeValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Contract)
		want   string
	}{
		{"outbound requires hosts", func(c *Contract) { *c.Capabilities.Runtime.Allowed.Network.Outbound = true }, "requires constrained hosts"},
		{"commands require allowlist", func(c *Contract) { *c.Capabilities.Runtime.Allowed.Commands.Execute = true }, "requires a command allowlist"},
		{"secret exposure requires IDs", func(c *Contract) { *c.Capabilities.Runtime.Allowed.Secrets.Expose = true }, "requires explicit secret IDs"},
		{"required subset", func(c *Contract) { c.Capabilities.Runtime.Required.Environment.Read = []string{"CI_TOKEN"} }, "outside allowed capabilities"},
		{"destructive requires approval", func(c *Contract) { c.Effects.Tools = []ToolEffect{{Tool: "production.delete", Class: "destructive"}} }, "requires a human approval rule"},
		{"credential ttl", func(c *Contract) { zero := 0; c.Identity.Credentials.MaxTTLSeconds = &zero }, "max_ttl_seconds must be greater than zero"},
		{"delegation constraints", func(c *Contract) { *c.Delegation.Allowed = true }, "enabled delegation requires positive depth"},
		{"data purpose policy", func(c *Contract) { c.Data.Policies = c.Data.Policies[1:] }, "requires a purpose and retention policy"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			contract := NewContractV2()
			tc.mutate(&contract)
			if errs := strings.Join(ValidateContract(contract), "\n"); !strings.Contains(errs, tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, errs)
			}
		})
	}
}

func TestContractV2DiffDetectsRuntimeExpansion(t *testing.T) {
	oldContract, newContract := NewContractV2(), NewContractV2()
	*newContract.Capabilities.Runtime.Allowed.Network.Outbound = true
	newContract.Capabilities.Runtime.Allowed.Network.Hosts = []string{"api.example.com"}
	diff := DiffContracts(oldContract, newContract)
	if diff.Classification != ImpactExpansion {
		t.Fatalf("expected expansion, got %#v", diff)
	}
}

func TestContractV2DiffDetectsDelegationAndRetentionExpansion(t *testing.T) {
	oldContract, newContract := NewContractV2(), NewContractV2()
	depth := 1
	oldContract.Resources.MaxDelegationDepth = &depth
	newContract.Resources.MaxDelegationDepth = &depth
	*newContract.Delegation.Allowed = true
	*newContract.Delegation.MaxDepth = 1
	newContract.Data.Policies[0].Retention = "persistent"
	if diff := DiffContracts(oldContract, newContract); diff.Classification != ImpactExpansion {
		t.Fatalf("expected agentic security expansion, got %#v", diff)
	}
}
