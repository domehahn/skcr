// Package asps exposes skcr's pinned authoring catalog for ASPS v1.0.
// It models requirements only; PASS/FAIL results remain owned by skil.
package asps

import (
	"fmt"
	"sort"
)

const Version = "1.0"
const Snapshot = "2026-07-31"

type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Property struct {
	ID       string `json:"id"`
	DomainID string `json:"domain_id"`
	Name     string `json:"name"`
}

var domainNames = []string{
	"Instruction & Goal Integrity", "Discovery, Metadata & Selection Integrity", "Data Confidentiality & Privacy",
	"Identity, Authorization & Consent", "Tool, Capability & Agency Safety", "Code Execution & Information-Flow Safety",
	"Memory, State & Persistence Integrity", "Inter-Agent & Delegation Security", "Supply Chain, Provenance & Artifact Integrity",
	"MCP & Integration Protocol Security", "Network, Filesystem & Runtime Boundary Security", "Resource, Availability & Failure Containment",
	"Human-Agent Trust & Safety", "Auditability, Observability & Accountability", "Dependency, Package & Container Supply-Chain Security",
}

var propertyNames = [][]string{
	{"Instruction Hierarchy Preservation", "Role & Context Integrity", "Refusal Preservation", "Warning & Safety-Context Preservation", "Guardrail Integrity", "Hidden Instruction Resistance", "Covert Behavioral Steering Resistance", "Goal & Scope Integrity"},
	{"Metadata Authenticity", "Description–Behavior Consistency", "Trigger Specificity", "Trigger Shadowing Resistance", "Retrieval Keyword-Stuffing Resistance", "Semantic Camouflage Resistance", "Sybil & Reputation Manipulation Resistance", "Planner Selection Integrity"},
	{"Secret & Environment Harvesting Protection", "Credential File Protection", "Conversation & Context Confidentiality", "Privileged Instruction Confidentiality", "External Data Exfiltration Prevention", "Cloud/Object-Storage Exfiltration Prevention", "Data Minimization", "Purpose & Retention Bound"},
	{"Least-Privilege Permission Bound", "Underdeclared Capability Detection", "Overdeclared Capability Minimization", "Approval & Confirmation Integrity", "Credential Scope Minimization", "Token Audience & Resource Binding", "Credential Non-Transferability", "Revocation & Stop Enforcement"},
	{"Restricted Tool Surface", "Tool Parameter Safety", "Tool Chaining Safety", "High-Impact Action Gating", "Safe Defaults", "Capability Scope Confinement", "Side-Effect Disclosure", "Autonomy Bound"},
	{"Dynamic Execution Control", "Process Execution Control", "Shell Execution Safety", "Unsafe Deserialization Control", "Generated Output Execution Control", "Cross-Context Output Validation", "Information-Flow Integrity", "Obfuscated & Malicious Payload Safety"},
	{"Memory Poisoning Resistance", "Context Stuffing Resistance", "State Manipulation Resistance", "Self-Modification Control", "Persistence Authorization", "State Ownership Isolation", "Memory Provenance & Trust Labels", "Memory Lifecycle & Expiry"},
	{"Agent Identity Authentication", "Inter-Agent Message Integrity", "Delegation Monotonicity", "Confused-Deputy Resistance", "Cross-Agent Trust Propagation Bound", "Inter-Agent Output Validation", "Collusion & Self-Replication Containment", "Delegation Chain Traceability"},
	{"Publisher Authenticity", "Artifact Signature Verification", "Payload Integrity", "Provenance & Build Attestation", "Version Lineage & Rollback Protection", "Update Revalidation", "Tool & Dependency Substitution Integrity", "Review-to-Execution Integrity"},
	{"MCP Wildcard Permission Prevention", "MCP Metadata Poisoning Resistance", "MCP Parameter Injection Resistance", "MCP Tool Identity Stability", "MCP Description–Behavior Consistency", "OAuth Authorization URL Safety", "MCP State-Handle Ownership Binding", "Local MCP/stdio Isolation"},
	{"Outbound Network Boundary", "Cloud Metadata SSRF Protection", "Internal & Loopback SSRF Protection", "Dynamic-Target SSRF Protection", "Agent & MCP Configuration Isolation", "Peer Skill Isolation", "Container Control-Plane Isolation", "Host Escape & Privilege-Escalation Prevention"},
	{"Resource Budget Bound", "Output Bound", "Rate & Quota Bound", "Loop & Recursion Bound", "Cascading Failure Containment", "Circuit Breaker & Kill-Switch Effectiveness", "Retry & Backoff Safety", "Idempotency & Duplicate-Action Safety"},
	{"Risk Disclosure Integrity", "Consent Specificity", "Deceptive Risk Framing Resistance", "Confirmation UI/Message Integrity", "Physical-Harm Operational Safety", "Malware & Phishing Objective Safety", "Destructive & Recovery-Inhibition Safety", "Security-Control Evasion Safety"},
	{"Action Attribution", "Authorization Decision Traceability", "Security Event Completeness", "Tamper-Evident Audit Records", "Artifact & Version Attribution", "Evidence-Chain Reproducibility", "Failure & Denial Auditability", "Coverage & Unknown-State Reporting"},
	{"Dependency Pinning", "Known Vulnerability Hygiene", "Package Identity & Typosquatting Resistance", "Dependency Maintenance & Reputation", "Dependency Namespace/Confusion Resistance", "Lockfile Integrity", "Remote Script & Runtime Dependency Integrity", "Container Image Trust"},
}

func Domains() []Domain {
	out := make([]Domain, len(domainNames))
	for i, name := range domainNames {
		out[i] = Domain{fmt.Sprintf("ASP-%02d", i+1), name}
	}
	return out
}
func Properties() []Property {
	out := make([]Property, 0, 120)
	for domain, names := range propertyNames {
		domainID := fmt.Sprintf("ASP-%02d", domain+1)
		for property, name := range names {
			out = append(out, Property{fmt.Sprintf("%s.%02d", domainID, property+1), domainID, name})
		}
	}
	return out
}
func FindProperty(id string) (Property, bool) {
	for _, property := range Properties() {
		if property.ID == id {
			return property, true
		}
	}
	return Property{}, false
}
func KnownProperty(id string) bool { _, ok := FindProperty(id); return ok }

var profiles = map[string][]string{
	"asps-core@1.0":         {"ASP-01.01", "ASP-01.08", "ASP-02.02", "ASP-03.01", "ASP-03.05", "ASP-03.08", "ASP-04.01", "ASP-04.04", "ASP-04.06", "ASP-05.01", "ASP-05.02", "ASP-05.04", "ASP-05.05", "ASP-05.06", "ASP-05.07", "ASP-05.08", "ASP-06.02", "ASP-06.07", "ASP-07.05", "ASP-08.01", "ASP-08.03", "ASP-08.04", "ASP-09.03", "ASP-09.04", "ASP-09.08", "ASP-11.01", "ASP-12.01", "ASP-12.04", "ASP-12.07", "ASP-12.08", "ASP-13.01", "ASP-14.01", "ASP-14.05", "ASP-14.06", "ASP-14.08", "ASP-15.01", "ASP-15.07", "ASP-15.08"},
	"asps-mcp@1.0":          {"ASP-04.06", "ASP-04.07", "ASP-09.07", "ASP-10.01", "ASP-10.02", "ASP-10.03", "ASP-10.04", "ASP-10.05", "ASP-10.06", "ASP-10.07", "ASP-10.08", "ASP-11.08"},
	"asps-a2a@1.0":          {"ASP-04.01", "ASP-04.06", "ASP-08.01", "ASP-08.02", "ASP-08.03", "ASP-08.04", "ASP-08.05", "ASP-08.06", "ASP-08.07", "ASP-08.08", "ASP-14.01", "ASP-14.02"},
	"asps-supply-chain@1.0": {"ASP-09.01", "ASP-09.02", "ASP-09.03", "ASP-09.04", "ASP-09.05", "ASP-09.06", "ASP-09.07", "ASP-09.08", "ASP-15.01", "ASP-15.02", "ASP-15.03", "ASP-15.04", "ASP-15.05", "ASP-15.06", "ASP-15.07", "ASP-15.08"},
}

func Profiles() []string {
	out := make([]string, 0, len(profiles))
	for profile := range profiles {
		out = append(out, profile)
	}
	sort.Strings(out)
	return out
}
func KnownProfile(profile string) bool          { _, ok := profiles[profile]; return ok }
func ProfileProperties(profile string) []string { return append([]string{}, profiles[profile]...) }

func EvidenceRoute(propertyID string) string {
	if len(propertyID) < 6 {
		return "unknown"
	}
	switch propertyID[:6] {
	case "ASP-01":
		return "descriptor.goal + evals"
	case "ASP-02":
		return "descriptor metadata"
	case "ASP-03":
		return "contract.data + evals"
	case "ASP-04":
		return "contract.identity + human_approval"
	case "ASP-05":
		return "contract capabilities/effects + evals"
	case "ASP-06":
		return "contract commands/data + evals"
	case "ASP-07":
		return "contract persistence/data"
	case "ASP-08":
		return "contract.delegation + integrations.a2a"
	case "ASP-09":
		return "descriptor/dependencies + build provenance"
	case "ASP-10":
		return "integrations.mcp"
	case "ASP-11":
		return "contract runtime boundary"
	case "ASP-12":
		return "contract resources/effects"
	case "ASP-13":
		return "human_approval + evals"
	case "ASP-14":
		return "evals + build provenance"
	case "ASP-15":
		return "dependencies reviewed closure"
	default:
		return "unknown"
	}
}
