package skillmeta

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const IntegrationSchemaVersion = "1"
const DependenciesSchemaVersion = "1"

var sha256DigestRE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type MCPDocument struct {
	SchemaVersion string      `yaml:"schema_version" json:"schema_version"`
	Servers       []MCPServer `yaml:"servers" json:"servers"`
}
type MCPServer struct {
	ID             string            `yaml:"id" json:"id"`
	Transport      string            `yaml:"transport" json:"transport"`
	Endpoint       string            `yaml:"endpoint" json:"endpoint"`
	Authentication MCPAuthentication `yaml:"authentication" json:"authentication"`
	Tools          Tools             `yaml:"tools" json:"tools"`
}
type MCPAuthentication struct {
	Type             string `yaml:"type" json:"type"`
	AudienceBinding  string `yaml:"audience_binding" json:"audience_binding"`
	TokenPassthrough string `yaml:"token_passthrough" json:"token_passthrough"`
}

type A2ADocument struct {
	SchemaVersion string        `yaml:"schema_version" json:"schema_version"`
	Agent         A2AAgent      `yaml:"agent" json:"agent"`
	Delegation    A2ADelegation `yaml:"delegation" json:"delegation"`
	Trust         A2ATrust      `yaml:"trust" json:"trust"`
}
type A2AAgent struct {
	ID string `yaml:"id" json:"id"`
}
type A2ADelegation struct {
	AcceptFrom                 []string `yaml:"accept_from" json:"accept_from"`
	DelegateTo                 []string `yaml:"delegate_to" json:"delegate_to"`
	MaxDepth                   int      `yaml:"max_depth" json:"max_depth"`
	RequireAuthenticatedOrigin bool     `yaml:"require_authenticated_origin" json:"require_authenticated_origin"`
	RequireCapabilitySubset    bool     `yaml:"require_capability_subset" json:"require_capability_subset"`
}
type A2ATrust struct {
	IncomingOutputs   string   `yaml:"incoming_outputs" json:"incoming_outputs"`
	PromotionRequires []string `yaml:"promotion_requires" json:"promotion_requires"`
}

type DependenciesDocument struct {
	SchemaVersion   string                `yaml:"schema_version" json:"schema_version"`
	Packages        []PackageDependency   `yaml:"packages" json:"packages"`
	RemoteResources []RemoteResource      `yaml:"remote_resources" json:"remote_resources"`
	Containers      []ContainerDependency `yaml:"containers" json:"containers"`
	MCPServers      []MCPDependency       `yaml:"mcp_servers" json:"mcp_servers"`
}
type PackageDependency struct {
	Ecosystem string `yaml:"ecosystem" json:"ecosystem"`
	Name      string `yaml:"name" json:"name"`
	Version   string `yaml:"version" json:"version"`
}

type RemoteResource struct {
	ID       string `yaml:"id" json:"id"`
	URL      string `yaml:"url" json:"url"`
	Digest   string `yaml:"digest" json:"digest"`
	Required bool   `yaml:"required" json:"required"`
}
type ContainerDependency struct {
	Image string `yaml:"image" json:"image"`
}
type MCPDependency struct {
	ID             string `yaml:"id" json:"id"`
	IdentityDigest string `yaml:"identity_digest" json:"identity_digest"`
}

func NewMCPDocument() MCPDocument {
	return MCPDocument{SchemaVersion: IntegrationSchemaVersion, Servers: []MCPServer{}}
}
func NewA2ADocument(skillName string) A2ADocument {
	return A2ADocument{SchemaVersion: IntegrationSchemaVersion, Agent: A2AAgent{ID: skillName}, Delegation: A2ADelegation{AcceptFrom: []string{}, DelegateTo: []string{}, RequireAuthenticatedOrigin: true, RequireCapabilitySubset: true}, Trust: A2ATrust{IncomingOutputs: "untrusted", PromotionRequires: []string{"verification"}}}
}
func NewDependenciesDocument() DependenciesDocument {
	return DependenciesDocument{SchemaVersion: DependenciesSchemaVersion, Packages: []PackageDependency{}, RemoteResources: []RemoteResource{}, Containers: []ContainerDependency{}, MCPServers: []MCPDependency{}}
}

func parseStrict[T any](data []byte, out *T) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(out)
}
func ParseMCP(data []byte) (MCPDocument, error) {
	var value MCPDocument
	err := parseStrict(data, &value)
	return value, err
}
func ParseA2A(data []byte) (A2ADocument, error) {
	var value A2ADocument
	err := parseStrict(data, &value)
	return value, err
}
func ParseDependencies(data []byte) (DependenciesDocument, error) {
	var value DependenciesDocument
	err := parseStrict(data, &value)
	return value, err
}
func loadStrict[T any](path string, out *T) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := parseStrict(data, out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func ValidatePhase3Artifacts(skillDir string, d Descriptor, contract *Contract) []string {
	var errs []string
	var mcp *MCPDocument
	if d.Integrations != nil {
		if mcpPath, safe := safeArtifactPath(skillDir, d.Integrations.MCP.File); safe {
			var document MCPDocument
			if err := loadStrict(mcpPath, &document); err != nil {
				errs = append(errs, "malformed MCP integration: "+err.Error())
			} else {
				errs = append(errs, ValidateMCP(document)...)
				mcp = &document
				if contract != nil && contract.SchemaVersion == ContractSchemaVersion {
					errs = append(errs, validateMCPAuthority(document, *contract)...)
				}
			}
		}
		if a2aPath, safe := safeArtifactPath(skillDir, d.Integrations.A2A.File); safe {
			var a2a A2ADocument
			if err := loadStrict(a2aPath, &a2a); err != nil {
				errs = append(errs, "malformed A2A integration: "+err.Error())
			} else {
				errs = append(errs, ValidateA2A(a2a, d, contract)...)
			}
		}
	}
	if d.Dependencies != nil {
		if path, safe := safeArtifactPath(skillDir, d.Dependencies.File); safe {
			var document DependenciesDocument
			if err := loadStrict(path, &document); err != nil {
				errs = append(errs, "malformed dependencies: "+err.Error())
			} else {
				errs = append(errs, ValidateDependencies(document, mcp)...)
			}
		}
	}
	if contract != nil && contract.SchemaVersion == ContractSchemaVersion && (len(contract.Capabilities.Runtime.Allowed.MCP.Servers) > 0 || len(contract.Capabilities.Runtime.Allowed.MCP.Tools) > 0) && mcp == nil {
		errs = append(errs, "runtime MCP authority requires an integrations.mcp descriptor")
	}
	if contract != nil && contract.SchemaVersion == ContractSchemaVersion && enabled(contract.Delegation.Allowed) && d.Integrations == nil {
		errs = append(errs, "enabled delegation requires an integrations.a2a descriptor")
	}
	return dedupeSorted(errs)
}

func ValidateMCP(d MCPDocument) []string {
	if d.SchemaVersion != IntegrationSchemaVersion {
		return []string{fmt.Sprintf("unsupported MCP schema_version %q", d.SchemaVersion)}
	}
	var errs []string
	seenServers := map[string]bool{}
	for i, server := range d.Servers {
		prefix := fmt.Sprintf("mcp.servers[%d]", i)
		if !validIdentifier(server.ID) || seenServers[server.ID] {
			errs = append(errs, prefix+".id is invalid or duplicate")
		}
		seenServers[server.ID] = true
		if server.Transport != "https" {
			errs = append(errs, prefix+`.transport must be "https"`)
		}
		parsed, err := url.Parse(server.Endpoint)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			errs = append(errs, prefix+".endpoint must be an HTTPS URL without userinfo")
		}
		if !map[string]bool{"oauth": true, "token": true, "mtls": true, "none": true}[server.Authentication.Type] {
			errs = append(errs, prefix+".authentication.type is unsupported")
		}
		if !map[string]bool{"required": true, "optional": true}[server.Authentication.AudienceBinding] {
			errs = append(errs, prefix+".authentication.audience_binding is unsupported")
		}
		if !map[string]bool{"forbidden": true, "allowed": true}[server.Authentication.TokenPassthrough] {
			errs = append(errs, prefix+".authentication.token_passthrough is unsupported")
		}
		if server.Authentication.TokenPassthrough == "allowed" && server.Authentication.AudienceBinding != "required" {
			errs = append(errs, prefix+" token passthrough requires audience binding")
		}
		errs = append(errs, validateStringList(prefix+".tools.allow", server.Tools.Allow, false)...)
		errs = append(errs, validateStringList(prefix+".tools.deny", server.Tools.Deny, false)...)
		denied := stringSet(server.Tools.Deny)
		for _, tool := range server.Tools.Allow {
			if _, deniedTool := denied[tool]; deniedTool {
				errs = append(errs, prefix+" tool "+tool+" is both allowed and denied")
			}
		}
	}
	return errs
}

func validateMCPAuthority(document MCPDocument, contract Contract) []string {
	servers, tools := map[string]bool{}, map[string]bool{}
	for _, server := range document.Servers {
		servers[server.ID] = true
		for _, tool := range server.Tools.Allow {
			tools[server.ID+"."+tool] = true
		}
	}
	var errs []string
	for _, server := range contract.Capabilities.Runtime.Allowed.MCP.Servers {
		if !servers[server] {
			errs = append(errs, "contract runtime MCP server "+server+" is not declared by integrations.mcp")
		}
	}
	for _, tool := range contract.Capabilities.Runtime.Allowed.MCP.Tools {
		if !tools[tool] {
			errs = append(errs, "contract runtime MCP tool "+tool+" is not allowed by integrations.mcp")
		}
	}
	return errs
}

func ValidateA2A(a A2ADocument, d Descriptor, contract *Contract) []string {
	if a.SchemaVersion != IntegrationSchemaVersion {
		return []string{fmt.Sprintf("unsupported A2A schema_version %q", a.SchemaVersion)}
	}
	var errs []string
	if !validIdentifier(a.Agent.ID) {
		errs = append(errs, "a2a.agent.id is invalid")
	} else if a.Agent.ID != d.Name {
		errs = append(errs, "a2a.agent.id must match descriptor name")
	}
	errs = append(errs, validateStringList("a2a.delegation.accept_from", a.Delegation.AcceptFrom, false)...)
	errs = append(errs, validateStringList("a2a.delegation.delegate_to", a.Delegation.DelegateTo, false)...)
	if a.Delegation.MaxDepth < 0 {
		errs = append(errs, "a2a.delegation.max_depth must be zero or greater")
	}
	if len(a.Delegation.DelegateTo) > 0 && (!a.Delegation.RequireAuthenticatedOrigin || !a.Delegation.RequireCapabilitySubset) {
		errs = append(errs, "A2A delegation requires authenticated origin and capability subset")
	}
	if a.Trust.IncomingOutputs != "untrusted" {
		errs = append(errs, `a2a.trust.incoming_outputs must be "untrusted"`)
	}
	errs = append(errs, validateStringList("a2a.trust.promotion_requires", a.Trust.PromotionRequires, false)...)
	if contract != nil && contract.SchemaVersion == ContractSchemaVersion && contract.Delegation.MaxDepth != nil {
		if a.Delegation.MaxDepth > *contract.Delegation.MaxDepth {
			errs = append(errs, "A2A max_depth exceeds contract delegation authority")
		}
		if !enabled(contract.Delegation.Allowed) && (len(a.Delegation.DelegateTo) > 0 || len(a.Delegation.AcceptFrom) > 0) {
			errs = append(errs, "A2A relationships require contract delegation.allowed")
		}
	}
	return errs
}

func ValidateDependencies(d DependenciesDocument, mcp *MCPDocument) []string {
	if d.SchemaVersion != DependenciesSchemaVersion {
		return []string{fmt.Sprintf("unsupported dependencies schema_version %q", d.SchemaVersion)}
	}
	var errs []string
	seen := map[string]bool{}
	for i, p := range d.Packages {
		prefix := fmt.Sprintf("dependencies.packages[%d]", i)
		if strings.TrimSpace(p.Ecosystem) == "" || strings.TrimSpace(p.Name) == "" {
			errs = append(errs, prefix+" ecosystem and name are required")
		}
		if strings.TrimSpace(p.Version) == "" || strings.EqualFold(p.Version, "latest") || regexp.MustCompile(`(?i)(^|[.\-])x($|[.\-])|[*^~<>=]|\s`).MatchString(p.Version) {
			errs = append(errs, prefix+" requires an exact version")
		}
		key := p.Ecosystem + "\x00" + p.Name
		if seen[key] {
			errs = append(errs, prefix+" is duplicate")
		}
		seen[key] = true
	}
	for i, resource := range d.RemoteResources {
		prefix := fmt.Sprintf("dependencies.remote_resources[%d]", i)
		parsed, err := url.Parse(resource.URL)
		if !validIdentifier(resource.ID) || err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			errs = append(errs, prefix+" requires a valid id and HTTPS URL")
		}
		if !sha256DigestRE.MatchString(resource.Digest) {
			errs = append(errs, prefix+" requires a sha256 digest")
		}
	}
	for i, container := range d.Containers {
		if !regexp.MustCompile(`@sha256:[0-9a-f]{64}$`).MatchString(container.Image) {
			errs = append(errs, fmt.Sprintf("dependencies.containers[%d].image must be digest-pinned", i))
		}
	}
	knownServers := map[string]bool{}
	if mcp != nil {
		for _, server := range mcp.Servers {
			knownServers[server.ID] = true
		}
	}
	for i, server := range d.MCPServers {
		prefix := fmt.Sprintf("dependencies.mcp_servers[%d]", i)
		if !knownServers[server.ID] {
			errs = append(errs, prefix+" references unknown MCP server "+server.ID)
		}
		if !sha256DigestRE.MatchString(server.IdentityDigest) {
			errs = append(errs, prefix+" requires an identity sha256 digest")
		}
	}
	return errs
}
