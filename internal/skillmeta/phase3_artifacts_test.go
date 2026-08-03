package skillmeta

import (
	"strings"
	"testing"
)

func TestMCPAuthorityCrossCheck(t *testing.T) {
	document := MCPDocument{SchemaVersion: "1", Servers: []MCPServer{{ID: "gitlab", Transport: "https", Endpoint: "https://gitlab.example.com/mcp", Authentication: MCPAuthentication{Type: "oauth", AudienceBinding: "required", TokenPassthrough: "forbidden"}, Tools: Tools{Allow: []string{"merge_requests.get"}, Deny: []string{}}}}}
	contract := NewContractV2()
	contract.Capabilities.Runtime.Allowed.MCP.Servers = []string{"gitlab"}
	contract.Capabilities.Runtime.Allowed.MCP.Tools = []string{"gitlab.merge_requests.get"}
	if errs := append(ValidateMCP(document), validateMCPAuthority(document, contract)...); len(errs) != 0 {
		t.Fatalf("valid MCP integration rejected: %v", errs)
	}
	contract.Capabilities.Runtime.Allowed.MCP.Tools = []string{"gitlab.merge_requests.merge"}
	if errs := strings.Join(validateMCPAuthority(document, contract), "\n"); !strings.Contains(errs, "is not allowed") {
		t.Fatalf("expected undeclared MCP tool error, got %q", errs)
	}
}

func TestA2ADelegationCannotExpandContract(t *testing.T) {
	descriptor := NewDescriptor("reviewer", "1.0.0", "Review.", "MIT", []string{"owner"}, nil)
	contract := NewContractV2()
	a2a := NewA2ADocument("reviewer")
	a2a.Delegation.DelegateTo = []string{"evidence-agent"}
	a2a.Delegation.MaxDepth = 1
	if errs := strings.Join(ValidateA2A(a2a, descriptor, &contract), "\n"); !strings.Contains(errs, "require contract delegation.allowed") || !strings.Contains(errs, "exceeds contract") {
		t.Fatalf("expected delegation authority errors, got %q", errs)
	}
}

func TestReviewedExecutionClosureRequiresDigestsAndExactVersions(t *testing.T) {
	document := NewDependenciesDocument()
	document.Packages = []PackageDependency{{Ecosystem: "npm", Name: "example", Version: "latest"}}
	document.RemoteResources = []RemoteResource{{ID: "policy-data", URL: "https://example.com/policy.json", Required: true}}
	document.Containers = []ContainerDependency{{Image: "registry.example.com/tool:latest"}}
	errs := strings.Join(ValidateDependencies(document, nil), "\n")
	for _, want := range []string{"exact version", "sha256 digest", "digest-pinned"} {
		if !strings.Contains(errs, want) {
			t.Fatalf("expected %q, got %q", want, errs)
		}
	}
}

func TestPhase3StrictParsersRejectUnknownFields(t *testing.T) {
	if _, err := ParseMCP([]byte("schema_version: \"1\"\nservers: []\nunknown: true\n")); err == nil {
		t.Fatal("expected strict MCP parse failure")
	}
}
