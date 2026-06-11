package scaffold

import (
	"bytes"
	"strings"
	"text/template"
)

// skillTemplateData holds all variables available inside skill-specific SKILL.md templates.
type skillTemplateData struct {
	Name         string   // kebab-case, e.g. "requirements-analyst"
	Title        string   // human-readable, e.g. "Requirements Analyst"
	Description  string   // one-line purpose
	Version      string   // semver
	Since        string   // YYYY-MM-DD first release
	LastModified string   // YYYY-MM-DD last change
	Owner        string   // responsible team or author
	Stability    string   // experimental | stable | deprecated
	License      string   // SPDX identifier
	Platforms    []string // target platforms
}

// skillFrontmatter is the common YAML frontmatter prepended to every registered skill template.
const skillFrontmatter = `---
name: {{.Name}}
description: {{.Description}}
version: "{{.Version}}"
since: "{{.Since}}"
last_modified: "{{.LastModified}}"
authors:
  - {{.Owner}}
stability: {{.Stability}}
min_platform_version:
{{- range .Platforms}}
  {{.}}: "unknown"
{{- end}}
deprecated_since:
replaces:
supersedes: []
changelog:
  - version: "{{.Version}}"
    date: "{{.LastModified}}"
    change: "Initial generated DevSecOps SDLC skill"
---
`

// skillBodies maps each registered skill name to its body content (appended after the frontmatter).
// Skills not listed here receive the generic fallback template from skillMarkdown.
var skillBodies = map[string]string{
	"requirements-analyst":             requirementsAnalystBody,
	"threat-modeler":                   threatModelerBody,
	"architecture-reviewer":            architectureReviewerBody,
	"secure-design-reviewer":           secureDesignReviewerBody,
	"safe-implementer":                 safeImplementerBody,
	"secure-code-reviewer":             secureCodeReviewerBody,
	"test-strategist":                  testStrategistBody,
	"dependency-risk-reviewer":         dependencyRiskReviewerBody,
	"ci-cd-security-reviewer":          ciCdSecurityReviewerBody,
	"container-security-reviewer":      containerSecurityReviewerBody,
	"iac-security-reviewer":            iacSecurityReviewerBody,
	"secrets-auditor":                  secretsAuditorBody,
	"privacy-reviewer":                 privacyReviewerBody,
	"release-readiness-reviewer":       releaseReadinessReviewerBody,
	"incident-response-helper":         incidentResponseHelperBody,
	"compliance-evidence-collector":    complianceEvidenceCollectorBody,
	"policy-as-code-reviewer":          policyAsCodeReviewerBody,
	"observability-readiness-reviewer": observabilityReadinessReviewerBody,
}

// SDLCSkillNames is the canonical ordered list of DevSecOps SDLC skills this package ships.
var SDLCSkillNames = []string{
	"requirements-analyst",
	"threat-modeler",
	"architecture-reviewer",
	"secure-design-reviewer",
	"safe-implementer",
	"secure-code-reviewer",
	"test-strategist",
	"dependency-risk-reviewer",
	"ci-cd-security-reviewer",
	"container-security-reviewer",
	"iac-security-reviewer",
	"secrets-auditor",
	"privacy-reviewer",
	"release-readiness-reviewer",
	"incident-response-helper",
	"compliance-evidence-collector",
	"policy-as-code-reviewer",
	"observability-readiness-reviewer",
}

// renderSkillTemplate renders the full SKILL.md for the named skill using its registered body.
// Returns ("", nil) when no specific template is registered so callers can fall back gracefully.
func renderSkillTemplate(name string, data skillTemplateData) (string, error) {
	body, ok := skillBodies[name]
	if !ok {
		return "", nil
	}
	full := skillFrontmatter + body
	tmpl, err := template.New(name).Parse(full)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// skillTitle converts a kebab-case skill name into a human-readable title.
func skillTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// ─── Skill body templates ─────────────────────────────────────────────────────

const requirementsAnalystBody = `
# {{.Title}}

## Purpose

Decompose raw requirements, epics, and change requests into unambiguous, independently testable acceptance criteria before any design or implementation begins. Identify functional requirements, non-functional requirements, security obligations, compliance controls, data-privacy constraints, and operational readiness requirements. Surface ambiguity, conflicts, and missing stakeholder input as explicit open questions rather than silent assumptions carried into design.

## When to Use This Skill

- A new feature, epic, user story, or change request is submitted for refinement.
- Requirements contain language that cannot be directly tested: "should", "as appropriate", "reasonable", "fast", "secure", or "user-friendly".
- Security, compliance, or data-privacy obligations are unstated, implied, or contradictory.
- Stakeholders disagree on scope, priority, or acceptance conditions.
- A change request has no documented acceptance criteria.
- You are routed here by the central agent via $requirements-analyst.

## Inputs

- Tickets, epics, user stories, PRDs, change requests, or RFC documents.
- Existing system documentation, ADRs, API specifications, and data dictionaries.
- Compliance frameworks and regulatory references cited by the project (GDPR, SOC 2, ISO 27001, PCI DSS, HIPAA).
- Stakeholder communications: email threads, meeting notes, interview transcripts.
- Constraints: timeline, budget, platform, technology choices, team capacity.
- Existing acceptance criteria or test suites that define the current behavioral boundary.

## Skill-Specific Operating Model

1. **Categorize every stated requirement.** Assign each item a type: FR (functional), NFR (non-functional), SEC (security), COMP (compliance), PRIV (privacy). Requirements spanning multiple types receive multiple tags.
2. **Flag non-testable language.** Mark every requirement that uses undefined thresholds, vague qualifiers, or aspirational language. Each flagged item becomes a blocking open question.
3. **Decompose epics into atomic stories.** Split each large requirement into the smallest unit that can be independently accepted or rejected by a tester who was not involved in elicitation.
4. **Write Given/When/Then acceptance criteria.** For every Must-have atomic requirement, write at least one criterion: Given [precondition], When [action], Then [observable, measurable outcome].
5. **Surface security requirements explicitly.** For each user interaction, API surface, data asset, and integration point: What authentication is required? What authorization controls apply? What input validation is needed? What must and must not be logged?
6. **Surface compliance requirements with control references.** Map each compliance obligation to a specific control clause (e.g., SOC 2 CC6.1, GDPR Art. 5(1)(b)). Obligations without a clause reference are marked "unverified compliance requirement".
7. **Identify personal data flows.** For every requirement touching user data, behavioral data, location data, or authentication data: document the data element, legal basis, retention period, and applicable data-subject rights.
8. **Document assumptions explicitly.** Record every fact assumed true but unconfirmed. Assign each an Owner and a Risk statement (what breaks if the assumption is wrong).
9. **Build the open-questions log.** For each ambiguity, conflict, or missing information item, create an entry with Owner, Blocking status (design / implementation / testing), and Due Date.
10. **Define and confirm scope boundaries.** State explicitly what is in scope and out of scope. Every scope boundary without a named stakeholder confirmation is marked TBD.
11. **Assign MoSCoW priority.** Must / Should / Could / Won't for each requirement. Mark TBD where stakeholder input is missing—never assign priority by assumption.

## Skill-Specific Checklist

- [ ] Every stated requirement is tagged with at least one type: FR, NFR, SEC, COMP, PRIV.
- [ ] Every requirement is written in testable, observable language without unmeasured adjectives.
- [ ] Given/When/Then acceptance criteria exist for every Must-have requirement.
- [ ] Security requirements explicitly cover: authentication, authorization, input validation, output encoding, encryption in transit, encryption at rest, and audit logging.
- [ ] Each compliance requirement is linked to a specific framework control clause, not just a framework name.
- [ ] All personal data flows are documented with data element, legal basis, retention period, and data-subject rights.
- [ ] Role and permission requirements are listed for every new or modified access path.
- [ ] Rate-limiting, throttling, and abuse-prevention requirements are addressed for every external-facing endpoint.
- [ ] Logging requirements specify both what to log and what must never appear in logs (e.g., passwords, tokens, PII).
- [ ] All assumptions are documented with Owner, Risk if wrong, and Review Date.
- [ ] All open questions are in the open-questions log with Owner, Blocking status, and Due Date.
- [ ] Scope boundaries are stated with at least one named stakeholder confirmation or flagged TBD.
- [ ] Conflicting requirements are identified and listed as blocking open questions, not silently resolved.
- [ ] MoSCoW priority is assigned or explicitly marked TBD for every requirement.
- [ ] Non-functional requirements include measurable thresholds for at least: response time, availability, scalability, and error rate.

## Decision Rules

- If a requirement uses non-testable language and no threshold can be derived from existing SLOs, mark it "not ready" and create a blocking open question requesting a measurable threshold before proceeding.
- If a requirement implies reading, storing, or transmitting personal data, create a PRIV-tagged requirement and flag it for $privacy-reviewer analysis before design begins.
- If a requirement changes authentication, authorization, or session-management behavior, flag it for $secure-design-reviewer review and block design until that review is complete.
- If two requirements conflict and the conflict cannot be resolved from available documentation, do not resolve it yourself—log it as a blocking open question with both stakeholders named as Owners.
- If a compliance obligation is cited without a specific framework clause, mark it "unverified" and create an open question for the compliance team.
- If a Must-have requirement cannot be expressed as a testable criterion, return it as "not ready for design" with a specific question to the requester.
- Do not assign or infer priority without stakeholder input; mark all unconfirmed priorities as TBD.

## DevSecOps Guardrails

- Do not advance a feature to design until all Must-have requirements have at least one testable acceptance criterion confirmed by a named stakeholder.
- Do not infer security or compliance requirements from the implementation context; surface them as explicit requirements before design begins.
- Do not include credentials, connection strings, API keys, or sensitive configuration values in requirements artifacts.
- Do not mark a compliance requirement as satisfied unless a specific control clause and an accountable owner are named.
- Do not finalize scope without at least one named stakeholder confirmation of the scope boundary.
- Do not skip privacy analysis for any feature that processes personal data, behavioral data, location data, or authentication credentials.
- Do not produce acceptance criteria that reference internal implementation details; all criteria must be observable from the outside.

## Output Requirements

- **Requirements register**: columns Name, Type (FR/NFR/SEC/COMP/PRIV), Description, Priority (MoSCoW), Acceptance Criteria, Owner, Status.
- **Assumptions log**: each assumption with Owner, Risk if wrong, and Review Date.
- **Open-questions log**: each question with Owner, Blocking status, and Due Date.
- **Scope statement**: in-scope and out-of-scope items, each with a named stakeholder confirmation or TBD.
- **Security and compliance cross-reference**: table mapping each SEC/COMP requirement to framework clause, control owner, and verification method.
- **Privacy data map**: data flows with element name, legal basis, retention period, data-subject rights, and DPIA indicator.
- **Ambiguity summary**: list of non-testable requirements with the specific clarifying question raised for each.

## Acceptance Criteria

- Every Must-have functional requirement has at least one Given/When/Then acceptance criterion.
- All security requirements are explicitly stated and linked to a design or test obligation.
- All compliance requirements are linked to specific framework control clauses.
- All personal data flows are documented with legal basis and retention period.
- Assumptions and open questions are in separate logs, not embedded in the requirements body.
- No conflicting requirements are left unresolved without a documented decision or escalation.
- The scope boundary is unambiguous and carries at least one named stakeholder confirmation.
- No requirement uses language that a tester cannot evaluate without additional stakeholder input.

## Anti-Patterns

- **Implicit security**: Writing "the feature must be secure" instead of decomposing into specific, testable security controls with measurable outcomes.
- **Assumption laundering**: Embedding an unconfirmed business rule into acceptance criteria without flagging it as an assumption with an owner.
- **Priority theater**: Assigning Must-have to every requirement to avoid difficult prioritization conversations with stakeholders.
- **Fake measurability**: Writing "response time must be acceptable" instead of "p95 latency must be under 300ms at 1,000 concurrent users under normal load".
- **Compliance name-dropping**: Citing GDPR or SOC 2 without mapping to a specific article or control clause.
- **Functional-only output**: Producing only functional acceptance criteria and omitting NFRs, security controls, and observability targets.
- **Skipping conflict resolution**: Moving to design without resolving a known conflict between two stakeholder requirements.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const threatModelerBody = `
# {{.Title}}

## Purpose

Identify and prioritize threats against features, services, APIs, and architectural changes using structured STRIDE analysis before implementation begins. Produce a threat register, attack-surface map, and control recommendations where every threat is linked to a concrete mitigation or an accepted residual risk with a named decision owner. The output must be actionable by an implementer, not just informational.

## When to Use This Skill

- A new feature, service, or API is in design and touches authentication, authorization, data storage, external integrations, or network communication.
- An architectural change modifies trust boundaries, data flows, or existing security controls.
- A security review or compliance audit has requested a threat model.
- A previous threat model is being refreshed after significant scope changes.
- You are routed here by the central agent via $threat-modeler.

## Inputs

- Architecture diagrams, sequence diagrams, or data flow descriptions.
- API specifications (OpenAPI, gRPC protobuf, GraphQL schema).
- Requirements or accepted-risks register from $requirements-analyst.
- Existing access-control model, roles, and permission boundaries.
- Infrastructure description: cloud provider, network topology, deployment model.
- Previous threat models or security assessment reports for context and delta analysis.

## Skill-Specific Operating Model

1. **Define scope and trust model.** State which system, service, feature, or change is being modeled. Define what is in scope, what is out of scope, and what is explicitly trusted.
2. **Enumerate assets.** List every asset: data assets (PII, credentials, session tokens, API keys, business-critical data), system assets (servers, containers, databases, message queues), and process assets (cryptographic operations, business logic).
3. **Map trust boundaries and entry points.** Identify every trust boundary the system crosses. List every entry point where untrusted input enters: API endpoints, UI inputs, file uploads, message queues, webhooks, OAuth callbacks, SSO assertions, admin interfaces.
4. **Describe data flows.** For each significant operation, trace the data flow from entry point through processing to storage and exit. Annotate which flows cross trust boundaries.
5. **Apply STRIDE per element.** For each process, data store, data flow, and external entity, evaluate each STRIDE category systematically.
6. **Write abuse cases.** For each STRIDE finding, write a realistic abuse case: who is the attacker (insider, external, privileged user), what is their goal, what path do they take, what assumption do they violate.
7. **Trace attack paths.** For Critical and High threats, trace the multi-step path from entry point to impact. Identify which control breaks the attack chain at each step.
8. **Assign risk ratings.** For each threat, assign Impact (High/Medium/Low) and Likelihood (High/Medium/Low) with justification. Derive Risk Level: Critical (High×High), High, Medium, Low.
9. **Specify controls.** For each threat, name the primary control, assign an owner, and link it to a security requirement or implementation task.
10. **Document residual risk.** For threats where a control is absent or incomplete, document the residual risk, the acceptance decision, and the decision owner. A threat without a mitigation and without an acceptance decision is a gap, not a finding.

## Skill-Specific Checklist

- [ ] Scope is defined: in-scope elements, trust boundary, and explicit exclusions.
- [ ] Every significant data asset is listed and classified (confidentiality, integrity, availability requirements).
- [ ] Every trust boundary crossing is identified.
- [ ] Every external entry point is listed including non-obvious ones: admin interfaces, debug endpoints, message queues, webhooks.
- [ ] Data flows are described for all significant operations with trust boundary crossings annotated.
- [ ] Spoofing threats address: token forgery, credential replay, identity confusion, OAuth/SAML bypass.
- [ ] Tampering threats address: unsigned payloads, injectable parameters, race conditions on shared state, database injection.
- [ ] Repudiation threats address: audit log completeness, log integrity, non-repudiation for critical operations.
- [ ] Information Disclosure threats address: data in transit, data at rest, error messages, debug output, logs, API response leakage.
- [ ] Denial of Service threats address: resource exhaustion, unbounded queries, amplification vectors, unauthenticated expensive operations.
- [ ] Elevation of Privilege threats address: IDOR, missing authorization checks, privilege escalation, RBAC/ABAC bypass, JWT algorithm confusion.
- [ ] Each threat has a realistic abuse case with named attacker type, goal, and attack path.
- [ ] Each threat has an Impact, Likelihood, and Risk Level with written justification.
- [ ] Each Critical and High threat has a named mitigation control with an owner and an implementation reference.
- [ ] Residual risks are documented with acceptance decision and decision owner.
- [ ] The threat register is structured such that each entry is independently actionable.

## Decision Rules

- If the system processes authentication tokens, session state, or OAuth/OIDC flows, include spoofing and session-hijacking threats by default, even when not explicitly requested.
- If a data flow crosses an external trust boundary without mutual TLS or equivalent encryption, classify it as a Critical Information Disclosure threat regardless of data sensitivity.
- If a Critical or High threat has no control specified, the threat model is incomplete; do not deliver it without a mitigation or a documented acceptance decision.
- If an abuse case requires capabilities disproportionate to the stated threat actor profile, document the assumption and reduce the Likelihood rating—but do not remove the threat.
- If a control is marked "planned" or "future work", the threat status remains Open in the register until the control is implemented and verified.
- If two threats share the same root cause, group them under a single root-cause entry rather than duplicating mitigations.
- Do not assess threats for elements outside the declared scope; reference out-of-scope systems with a pointer to the responsible owner.

## DevSecOps Guardrails

- Do not assume existing perimeter controls (firewall, WAF, VPN) fully mitigate a threat; model each threat as if the perimeter is already compromised.
- Do not mark a threat as mitigated without a specific, implemented, and verifiable control—not a planned or aspirational control.
- Do not skip AuthN/AuthZ threats for internal microservices; internal network location is not a substitute for explicit authentication and authorization.
- Do not finalize a threat model without residual-risk documentation for every unmitigated or partially mitigated threat.
- Do not use "low-value target" as justification to omit threat analysis; threat models assess technical risk, not attacker motivation.
- Do not include actual credentials, connection strings, or secret values in the threat model artifact.

## Output Requirements

- **Scope statement**: in-scope and out-of-scope elements, trust boundary definition, threat actor profile.
- **Asset register**: each asset with confidentiality, integrity, and availability classification.
- **Entry point and trust boundary map**: all entry points and trust boundary crossings with annotation.
- **Data flow descriptions**: per-operation flows with trust boundary crossings marked.
- **Threat register**: Threat ID, STRIDE category, Affected Component, Abuse Case, Impact, Likelihood, Risk Level, Control, Control Owner, Status (Open/Mitigated/Accepted).
- **Residual risk register**: accepted threats with decision owner and scheduled review date.
- **Control summary**: required security controls derived from this model, each mapped to an implementation task or existing requirement.

## Acceptance Criteria

- Every trust boundary crossing has at least one associated threat in the register.
- Every external entry point has at least one Information Disclosure or Tampering threat.
- All authentication mechanisms have at least one Spoofing threat with a corresponding control.
- All Critical and High threats have a named mitigation control with an owner.
- Residual risks are explicitly documented with acceptance decisions and review dates.
- No threat is marked "mitigated" without a reference to a specific, verifiable, implemented control.
- The threat register is traceable: each entry references the component or data flow that introduced it.

## Anti-Patterns

- **Checkbox STRIDE**: Applying STRIDE labels without writing concrete abuse cases or tracing attack paths.
- **Perimeter trust**: Assuming internal network traffic is inherently safe and skipping AuthN/AuthZ threats for service-to-service calls.
- **Generic mitigations**: Assigning "use encryption" without specifying the algorithm, mode, key management approach, and implementation location.
- **Risk inflation**: Marking every threat Critical to force prioritization without evidence-based Impact and Likelihood justification.
- **Scope narrowing**: Reducing scope to exclude the riskiest components in order to produce a cleaner model.
- **Missing residual risk**: Leaving threats without mitigations in the register without a documented acceptance decision.
- **Static model**: Treating the threat model as finished once written, rather than scheduling a review when the architecture changes materially.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const architectureReviewerBody = `
# {{.Title}}

## Purpose

Evaluate architecture decisions, module boundaries, coupling, cohesion, dependency direction, and security-by-design properties before and after implementation. Identify structural risks—circular dependencies, layering violations, unclear ownership, runtime coupling, and missing security controls at the design level—and produce prioritized, actionable recommendations. Distinguish critical structural risks from stylistic preferences.

## When to Use This Skill

- A new service, component, or significant module is being designed.
- A pull request introduces or changes module boundaries, API contracts, or cross-component dependencies.
- A refactoring changes the ownership, layering, or coupling of existing components.
- An ADR is being drafted or reviewed.
- You are routed here by the central agent via $architecture-reviewer.

## Inputs

- Architecture diagrams, C4 model descriptions, or service maps.
- Module structure, package layout, and import graphs from the repository.
- API specifications and interface contracts (OpenAPI, protobuf, internal interfaces).
- ADRs or design documents describing the intended architecture.
- Infrastructure description: service topology, deployment model, data stores.
- Pull request diff if reviewing a concrete change.

## Skill-Specific Operating Model

1. **Establish the intended architecture.** Read available ADRs, design documents, and diagrams. If none exist, reconstruct the intended layering from code structure and ask the team to confirm.
2. **Analyze module and package boundaries.** Identify each logical boundary and its stated responsibility. Check whether each module has a single clear owner.
3. **Check dependency direction.** Verify that dependencies flow in the intended direction (e.g., domain does not import infrastructure, presentation does not import persistence). List every violation.
4. **Detect circular dependencies.** Identify any import cycles or runtime circular dependencies. Classify each as blocking (prevents clean build or test) or architectural smell (solvable with interface inversion).
5. **Assess coupling and cohesion.** Identify modules with high afferent coupling (many dependents) or high efferent coupling (many dependencies). Assess whether cohesion within each module is logical, sequential, or coincidental.
6. **Review API contracts.** For each public interface, check whether the contract is explicit (typed, versioned), whether breaking changes are guarded by backward-compatibility mechanisms, and whether consumers are isolated from implementation details.
7. **Identify runtime coupling.** Check for shared mutable state, synchronous cross-service calls in critical paths, and cascading failure risks. Assess whether circuit breakers, retries, and timeouts are appropriate.
8. **Evaluate security-by-design properties.** Check whether authentication and authorization are enforced at the correct layer, input validation occurs at trust boundaries, and secrets are isolated from application logic.
9. **Recommend ADR when needed.** If a decision has broad impact or introduces a new pattern, recommend creating or updating an ADR. Propose the decision question, options considered, and decision rationale.
10. **Prioritize findings.** Classify findings as: Critical (blocks safe deployment), High (introduces structural debt or security risk), Medium (degrades maintainability), Low (stylistic or minor concern). Provide concrete refactoring recommendations for Critical and High findings.

## Skill-Specific Checklist

- [ ] Intended layering or architectural style is identified and explicitly stated (hexagonal, layered, modular monolith, microservices, event-driven).
- [ ] All circular dependencies are identified and classified.
- [ ] Dependency direction violations are listed with the specific offending import or call.
- [ ] Modules with unclear or overlapping responsibilities are flagged.
- [ ] High-coupling modules are identified with concrete reasons why coupling is problematic.
- [ ] Public API contracts are checked for explicitness, versioning, and consumer isolation.
- [ ] Runtime coupling risks (synchronous chains, shared mutable state) are identified.
- [ ] AuthN/AuthZ enforcement layer is verified: authentication happens before authorization, authorization happens before business logic.
- [ ] Input validation occurs at trust boundaries, not deep inside business logic.
- [ ] Secrets (keys, credentials, tokens) are not embedded in business logic, configuration files committed to VCS, or passed through environment variables without a secrets manager.
- [ ] Each finding is classified as Critical / High / Medium / Low with reasoning.
- [ ] An ADR is recommended for any decision that sets a precedent or has broad impact.
- [ ] Proposed changes to fix Critical and High findings are concrete and specific, not "consider refactoring".
- [ ] Findings distinguish architectural risks from style preferences—style preferences belong in a separate section or are omitted.
- [ ] If no issues are found in a category, it is explicitly confirmed as "no issues found" rather than omitted.

## Decision Rules

- If a circular dependency cannot be broken without an interface inversion or a new abstraction layer, recommend the specific interface and where it should be defined.
- If a module has more than one clearly distinct responsibility, flag it for decomposition only when the responsibilities are independently deployable or testable—otherwise mark it as a coupling risk, not a split recommendation.
- If an API contract has no version and is consumed by more than one caller, flag it as a breaking-change risk regardless of how stable it appears today.
- If authentication or authorization enforcement is located anywhere other than the outermost application boundary or a dedicated middleware layer, flag it as a High architectural risk.
- If an ADR already exists for a decision being reviewed, check whether the implementation conforms to the ADR. Deviations are Critical if they undermine the stated rationale.
- If a runtime coupling risk exists (synchronous chain, no circuit breaker) in a path that cannot be degraded gracefully, classify it as High and recommend an async alternative or fallback.

## DevSecOps Guardrails

- Do not recommend global refactors in the context of a single PR review; scope recommendations to what the PR introduced or changed, with broader concerns noted separately.
- Do not conflate security architecture findings with implementation-level code review findings; architectural findings are about design decisions, not specific lines of code.
- Do not propose an architectural change without acknowledging the migration cost and the risk of the migration itself.
- Do not mark a design as "secure by design" unless authentication, authorization, input validation, and secrets management are each explicitly addressed at the correct layer.
- Do not use "best practice" as justification without citing the specific risk that the practice mitigates.

## Output Requirements

- **Architecture summary**: stated or inferred architectural style, key components, and their roles.
- **Dependency analysis**: dependency direction violations, circular dependencies, and high-coupling modules, each with the specific import path or call.
- **API contract review**: each public interface assessed for versioning, breaking-change risk, and consumer isolation.
- **Runtime coupling findings**: synchronous chains, shared state risks, and failure propagation paths.
- **Security-by-design findings**: AuthN/AuthZ layer, input validation placement, secrets handling.
- **Prioritized findings table**: Severity, Component, Finding, Impact, Recommended Action.
- **ADR recommendations**: decision question, options, and recommended rationale for each significant decision.

## Acceptance Criteria

- All circular dependencies are identified or explicitly confirmed absent.
- Dependency direction violations are listed with specific source and target.
- AuthN/AuthZ enforcement layer is assessed and either confirmed correct or flagged.
- Each finding has a severity (Critical/High/Medium/Low) with written justification.
- Critical and High findings include a concrete, scoped recommendation.
- ADR recommendations are provided for any decision that sets a new architectural precedent.
- Style preferences are separated from structural risks in the output.

## Anti-Patterns

- **Style-as-risk**: Reporting naming conventions or formatting as architectural risks instead of structural problems.
- **Rewrite recommendation**: Recommending a full rewrite without scoping to the specific problem introduced by the change under review.
- **Missing prioritization**: Listing 20 findings at equal severity, making it impossible to triage.
- **Abstract mitigations**: Recommending "better abstractions" or "cleaner interfaces" without specifying what interface, where, and why.
- **Ignored ADRs**: Reviewing code against personal preferences while ignoring documented ADRs that explain the design rationale.
- **Security as afterthought**: Reviewing module boundaries and coupling without addressing whether security controls are architecturally sound.
- **Finding inflation**: Including low-severity style findings in the same section as Critical structural risks.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const secureDesignReviewerBody = `
# {{.Title}}

## Purpose

Review the security design of features, services, and APIs before implementation to identify missing or inadequate security controls at the design level. Evaluate authentication, authorization, session management, secrets handling, input/output boundaries, data classification, encryption, audit logging, and defense-in-depth before a single line of implementation code is written. Produce concrete design-change recommendations, not implementation patches.

## When to Use This Skill

- A new feature or service is entering the design phase and touches authentication, authorization, data storage, external integrations, or user-facing functionality.
- An existing service is adding a new API endpoint, role, or data access pattern.
- A design document or RFC has been written and needs a security review before sign-off.
- $threat-modeler identified threats that require design-level controls.
- You are routed here by the central agent via $secure-design-reviewer.

## Inputs

- Design document, RFC, or architecture description.
- Data flow diagrams, sequence diagrams, or API specifications.
- Threat model output from $threat-modeler, if available.
- Existing authentication and authorization model for the system.
- Data classification policy or data dictionary.
- Compliance or regulatory requirements that apply to this feature.

## Skill-Specific Operating Model

1. **Understand the design intent.** Read the design document, API specification, or change description. State the security-relevant properties the design claims to provide.
2. **Evaluate authentication design.** Check that every entry point has a defined authentication mechanism. Verify that authentication is enforced before any business logic executes.
3. **Evaluate authorization design.** Check that every operation has a defined authorization rule. Verify that the rule is expressed in terms of principal, resource, and action, not just role names.
4. **Evaluate session and token management.** Check token lifetime, rotation policy, revocation mechanism, binding (IP, device fingerprint, or none), and storage location.
5. **Evaluate secrets handling.** Check that secrets are never embedded in code, configuration files, or environment variables without a secrets manager. Verify key rotation and least-access policies.
6. **Evaluate input and output boundaries.** Check that validation occurs at every trust boundary entry. Check that output encoding is appropriate for every rendering context (HTML, JSON, SQL, shell, etc.).
7. **Evaluate data classification and protection.** Verify that every data element is classified and that the protection controls (encryption, access restriction, masking) match the classification.
8. **Evaluate encryption design.** Check algorithm selection, key management, certificate validation, and whether encryption is applied to data at rest, in transit, and in use where required.
9. **Evaluate audit logging design.** Check that security-relevant events (login, logout, access denied, privilege change, data export) are logged with sufficient detail for forensic reconstruction, without logging sensitive data.
10. **Evaluate least-privilege and defense-in-depth.** Check that each service, process, and user has only the permissions it needs. Check that the design has multiple independent control layers rather than relying on a single control.
11. **Identify secure-default gaps.** Check whether the design defaults to secure behavior (deny by default, opt-in access, explicit allowlists) or relies on administrators to apply hardening after deployment.
12. **Produce design-change recommendations.** For each gap, propose a specific design change: which component is responsible, what mechanism to use, and what the expected security property is.

## Skill-Specific Checklist

- [ ] Every entry point has an explicitly named authentication mechanism.
- [ ] Authentication enforcement precedes all business logic in the request flow.
- [ ] Every operation has an explicit authorization rule (principal, resource, action).
- [ ] Authorization rules are enforced at the service layer, not only at the API gateway or load balancer.
- [ ] Token lifetime and revocation mechanism are specified.
- [ ] Session tokens are not stored in localStorage or URLs where they are accessible to JavaScript or logged.
- [ ] Secrets are never embedded in code or config files committed to VCS.
- [ ] A secrets manager or vault is named and the access-control policy is described.
- [ ] Input validation is specified at every trust boundary, with the validation mechanism and failure behavior named.
- [ ] Output encoding is specified for every rendering context.
- [ ] Every data element is classified and protection controls match the classification.
- [ ] Encryption algorithms and modes are explicitly named (e.g., AES-256-GCM, TLS 1.2+, not just "encrypted").
- [ ] Audit log events are listed with the fields to be captured and what must not be captured.
- [ ] The design defaults to deny-by-default for access control decisions.
- [ ] Defense-in-depth: at least two independent control layers exist for any Critical asset or operation.

## Decision Rules

- If an entry point has no defined authentication mechanism, the design is incomplete; do not approve it until authentication is specified.
- If authorization is implemented only at the API gateway or perimeter, flag it as a High design risk; internal services can be accessed directly if the perimeter is bypassed.
- If token revocation is not specified, flag it as a High risk; stolen tokens must have a defined invalidation path.
- If a secret is referenced in a design document as an environment variable without a named secrets manager, flag it as a Medium risk and specify the required change.
- If encryption is mentioned without naming the algorithm and key management approach, mark the encryption control as "underspecified" and return it for revision.
- If an audit log design includes user-entered fields or PII without explicit redaction, flag it as a privacy and forensics risk.
- If the design relies on a single control layer for a Critical asset, recommend a second independent control with justification.

## DevSecOps Guardrails

- Do not approve a design that has no authentication mechanism for any non-public entry point.
- Do not accept "TLS" as a complete encryption specification; require algorithm, version minimum, and certificate validation behavior.
- Do not accept "admin can configure security later" as a design decision; security must be on by default.
- Do not conflate authentication and authorization; verify both independently for every operation.
- Do not accept "same as the existing service" as an authorization specification without confirming the existing service's model is documented and correct.
- Do not include actual credentials, keys, or sensitive tokens in design review artifacts.

## Output Requirements

- **Authentication assessment**: each entry point with its authentication mechanism, enforcement layer, and gaps.
- **Authorization assessment**: each operation with its authorization rule, enforcement layer, and gaps.
- **Session and token assessment**: lifetime, rotation, revocation, binding, and storage with gaps.
- **Secrets handling assessment**: storage mechanism, access policy, rotation plan, and gaps.
- **Input/output boundary assessment**: validation mechanism and encoding approach per trust boundary and rendering context.
- **Data classification and protection assessment**: data elements, classification, protection controls, and gaps.
- **Encryption assessment**: algorithm, key management, and coverage with gaps.
- **Audit logging assessment**: logged events, captured fields, redaction requirements, and gaps.
- **Prioritized design-change recommendations**: Severity, Component, Gap, Recommended Design Change.

## Acceptance Criteria

- Every entry point has a named, specified authentication mechanism.
- Every operation has an explicit, documented authorization rule.
- Token management (lifetime, revocation, storage) is fully specified.
- All secrets have a named storage mechanism with an access-control policy.
- All data elements are classified and protection controls are matched to classification.
- Encryption specifications name algorithm, mode, and key management—not just "encrypted".
- Audit log design lists required event types, fields, and explicit redaction rules.
- All Critical and High findings have concrete design-change recommendations.

## Anti-Patterns

- **Perimeter-only authorization**: Enforcing authorization only at the API gateway and assuming internal calls are trusted.
- **Algorithm vagueness**: Specifying "we use encryption" without naming the algorithm, mode, key length, or key management approach.
- **Security opt-in defaults**: Designing so that security controls must be explicitly enabled by operators rather than being active by default.
- **Shared secret sprawl**: Using the same secret across multiple services or environments without rotation or per-service isolation.
- **Audit log PII leakage**: Including user-entered data, passwords, tokens, or PII fields in log entries without explicit redaction.
- **Authentication-authorization confusion**: Treating a valid authentication token as proof of authorization for all operations.
- **Single-layer defense**: Relying on a single control (e.g., WAF) as the only protection for a Critical asset.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const safeImplementerBody = `
# {{.Title}}

## Purpose

Implement requested changes safely, minimally, and traceably. Make only what was asked for—no opportunistic refactoring, no unrequested API changes, no silent behavior changes. Every change must be testable, rollback-capable, and accompanied by updated tests. The goal is a diff that a reviewer can understand in full, a test suite that fails before the change and passes after, and a rollback path that is stated explicitly.

## When to Use This Skill

- A feature, bug fix, configuration change, or infrastructure update needs to be implemented.
- An accepted design from $architecture-reviewer or $secure-design-reviewer is ready for implementation.
- A failing test or regression needs to be fixed without introducing new risk.
- You are routed here by the central agent via $safe-implementer.

## Inputs

- Accepted requirements and acceptance criteria from $requirements-analyst.
- Design approval from $secure-design-reviewer, if applicable.
- The failing test, bug report, or feature specification that defines the change.
- Repository context: existing code structure, conventions, test framework, CI configuration.
- Rollback constraints: deployment model, migration reversibility, feature flag availability.

## Skill-Specific Operating Model

1. **Read the acceptance criteria.** State each acceptance criterion that the implementation must satisfy. Do not begin coding until every criterion is understood.
2. **Identify the minimal change surface.** List the files and interfaces that must change to satisfy the acceptance criteria. Prefer modifying existing code over introducing new abstractions unless the design requires them.
3. **Write or update failing tests first.** Before implementing, write or update tests that fail against the current code and will pass after the correct implementation. This defines the change boundary.
4. **Implement only what is tested.** Make the minimal code change that causes the failing tests to pass without breaking existing tests. Do not extend scope beyond what the tests define.
5. **Validate input at every trust boundary.** Verify that all inputs entering the changed code from outside the trust boundary are validated before use. Reject invalid inputs with appropriate error codes and without leaking internal state.
6. **Handle errors explicitly.** Wrap errors with context. Do not swallow errors silently. Do not expose internal error details in external-facing responses.
7. **Verify API compatibility.** If the change touches a public or internal API, confirm that existing callers are unaffected. If a breaking change is unavoidable and was explicitly approved, version the API.
8. **Check for side effects.** Verify that the implementation does not introduce global state mutation, shared mutable state between concurrent requests, or unexpected behavior on retries.
9. **State the rollback path.** For every change that alters data schema, configuration, or deployment behavior, state how to revert the change and whether reversion requires downtime.
10. **Produce a clean, minimal diff.** Separate functional changes from formatting or cleanup changes. Do not mix unrelated fixes in a single commit.

## Skill-Specific Checklist

- [ ] Every acceptance criterion is listed and each has a corresponding test that fails before the change.
- [ ] The change surface is explicitly minimized: no opportunistic refactoring, no unrequested API additions.
- [ ] All new code paths have corresponding tests (unit, integration, or both as appropriate).
- [ ] All inputs from external trust boundaries are validated before use, with explicit reject behavior for invalid input.
- [ ] Errors are wrapped with context, not swallowed or propagated raw.
- [ ] Error responses to external callers do not expose stack traces, internal paths, database errors, or secret values.
- [ ] Public and internal API contracts are unchanged unless a breaking change was explicitly requested and approved.
- [ ] No new global state, singleton state, or shared mutable state is introduced.
- [ ] No secrets, tokens, passwords, or environment-specific values are hardcoded in the implementation.
- [ ] The rollback path for data migrations, schema changes, and configuration changes is explicitly stated.
- [ ] The diff separates functional changes from formatting or whitespace changes.
- [ ] No test mocks bypass real validation logic in a way that would hide a regression in production.
- [ ] The implementation compiles, all tests pass, and linting is clean before delivery.
- [ ] If a feature flag was used to gate the change, the flag name and the expected enabled/disabled behavior are documented.
- [ ] Acceptance criteria are verified one by one after implementation: each criterion is either confirmed met or listed as a known gap.

## Decision Rules

- If an acceptance criterion cannot be satisfied without modifying an unrelated component, stop and flag it as a scope change requiring approval before proceeding.
- If implementing a requirement requires a breaking API change that was not approved, propose the change as a version bump and wait for explicit approval.
- If a data migration is required, implement it as a separate, independently reversible step from the application change.
- If a test cannot be written without mocking a real security control (e.g., actual token validation), use integration tests against a real implementation rather than a mock that bypasses the control.
- If a feature flag is used, define both the enabled and disabled code paths explicitly and test both.
- If the implementation introduces a new dependency, flag it to $dependency-risk-reviewer before merging.
- If the rollback path requires data deletion or schema downgrade, document the risk explicitly and require explicit stakeholder approval before deployment.

## DevSecOps Guardrails

- Do not implement features beyond the stated acceptance criteria; unrequested behavior is untested behavior.
- Do not hardcode secrets, API keys, credentials, or environment-specific values in any file committed to version control.
- Do not expose internal error details, stack traces, or database errors in API responses accessible to external callers.
- Do not introduce global mutable state or request-shared mutable state without explicit concurrency controls.
- Do not skip validation at external trust boundaries, even when the calling service is internal and assumed trusted.
- Do not deliver an implementation where the test suite cannot run in CI without manual environment setup.
- Do not treat "tests pass locally" as equivalent to "tests pass in CI"; environment-specific tests are a deployment risk.

## Output Requirements

- **Change summary**: list of files changed, their purpose in the change, and the acceptance criterion each addresses.
- **Test coverage evidence**: list of new or updated tests with the acceptance criterion each validates.
- **API compatibility statement**: confirmation that existing callers are unaffected, or a description of the version change made.
- **Rollback path**: how to revert each schema change, configuration change, or deployment change, and whether reversion requires downtime.
- **Known gaps**: acceptance criteria not fully met, with reason and proposed follow-up.
- **Security check confirmation**: validation behavior, error handling, secret handling, and global-state check results.

## Acceptance Criteria

- Every acceptance criterion has a corresponding test that fails before and passes after the change.
- The diff contains only changes necessary to satisfy the stated acceptance criteria.
- All external trust boundary inputs are validated with defined reject behavior.
- No secrets, tokens, or credentials appear in any committed file.
- The rollback path for every migration or deployment change is explicitly stated.
- All existing tests continue to pass after the implementation.
- The implementation is deliverable to CI without manual environment configuration.

## Anti-Patterns

- **Scope creep commits**: Fixing unrelated issues, renaming variables, or reformatting code in the same commit as a functional change.
- **Silent error swallowing**: Catching an error, logging nothing, and returning a default value without the caller knowing the operation failed.
- **Security-through-obscurity input handling**: Relying on callers to send well-formed input and omitting validation in the implementation.
- **Test-bypassing mocks**: Mocking authentication, authorization, or validation logic in tests in a way that the mock would pass even if the production implementation is wrong.
- **Rollback by prayer**: Deploying a schema migration without a documented, tested rollback path.
- **Feature-flag abandonment**: Implementing a feature behind a flag without documenting the enabled/disabled behavior or scheduling flag removal.
- **Hardcoded configuration**: Embedding environment-specific URLs, secrets, or credentials in code or configuration files.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const secureCodeReviewerBody = `
# {{.Title}}

## Purpose

Review code changes for security risks, maintainability problems, and policy violations before merge. Identify injection vulnerabilities, authentication and authorization flaws, secrets exposure, insecure error handling, unsafe logging, race conditions, and supply-chain risks with file and function-level precision. Produce findings at a severity level that allows a reviewer to triage immediately: Blocker, High, Medium, Low, or Informational.

## When to Use This Skill

- A pull request or merge request is ready for security review.
- A change touches authentication, authorization, input handling, cryptography, secrets, external calls, or persistence.
- A new dependency has been added or an existing dependency upgraded.
- A security finding from a prior review is being verified as resolved.
- You are routed here by the central agent via $secure-code-reviewer.

## Inputs

- The pull request diff or the set of changed files.
- The full file context for each changed function or class, not just the diff lines.
- Test files covering the changed code.
- Security requirements or threat model from earlier SDLC stages, if available.
- CI pipeline results: SAST output, linting results, test results.

## Skill-Specific Operating Model

1. **Read the full context, not only the diff.** For each changed function or class, read the surrounding code to understand the trust model, data flow, and caller assumptions.
2. **Identify injection risks.** Check every location where external input reaches a SQL query, shell command, file path, template, LDAP query, XML parser, or deserializer.
3. **Check authentication and authorization.** Verify that every protected operation checks authentication before authorization, and authorization before business logic. Verify that authorization is enforced on the server, not in the client.
4. **Check insecure deserialization.** Identify any location that deserializes untrusted data (JSON, XML, protobuf, binary formats) without schema validation or type restriction.
5. **Check path traversal and SSRF.** Identify locations where user-controlled input determines a file path or an outbound HTTP/DNS target.
6. **Check XSS and CSRF.** For web-facing code, verify output encoding in every rendering context and verify CSRF token presence for state-changing operations.
7. **Check secret exposure.** Scan for hardcoded secrets, API keys, tokens, passwords, and connection strings in code and configuration. Scan for secrets in log statements, error messages, and API responses.
8. **Check unsafe logging.** Verify that log statements do not include passwords, tokens, session IDs, PII, or full request/response bodies containing sensitive data.
9. **Check error handling.** Verify that errors are wrapped with context, not swallowed, and that internal details (stack traces, DB errors, file paths) do not reach external API responses.
10. **Check cryptography.** Identify use of deprecated algorithms (MD5, SHA-1, DES, RC4), weak key lengths, static IVs, or home-grown cryptographic constructs.
11. **Check race conditions and concurrency.** Identify shared mutable state, time-of-check to time-of-use (TOCTOU) races, and non-atomic read-modify-write operations.
12. **Check test coverage.** Verify that security-relevant paths (rejection of invalid input, authorization denial, error handling) have explicit test coverage.
13. **Produce a severity-classified finding report.** For each finding: file name, function name, line range, severity, description, and a specific remediation recommendation.

## Skill-Specific Checklist

- [ ] Every location where external input reaches a SQL query, shell command, or file path is checked for injection.
- [ ] Every protected endpoint checks authentication before authorization, and authorization before business logic.
- [ ] Authorization is enforced on the server; client-side checks are noted as supplemental only.
- [ ] Deserialization of untrusted data uses explicit schema validation or type allowlists.
- [ ] File paths derived from user input are canonicalized and checked against an allowlist of permitted directories.
- [ ] Outbound HTTP/DNS targets are validated against an allowlist or restricted to known safe domains.
- [ ] XSS: output encoding is present and correct for every rendering context (HTML body, HTML attribute, JSON, JavaScript context).
- [ ] CSRF: state-changing operations require a validated, unpredictable, per-session token.
- [ ] No secrets, API keys, tokens, or passwords appear in any committed file or log statement.
- [ ] Log statements do not include PII, session identifiers, or full request bodies containing sensitive data.
- [ ] Errors are wrapped with context; no bare error returns that lose diagnostic information.
- [ ] External API responses do not include stack traces, internal file paths, or database error details.
- [ ] No deprecated cryptographic algorithms (MD5, SHA-1, DES, RC4) are used for security purposes.
- [ ] No static, hardcoded, or predictable IVs or nonces are used with symmetric encryption.
- [ ] Shared mutable state accessed from multiple goroutines, threads, or requests is protected by explicit synchronization.
- [ ] Security-relevant paths (rejection, denial, error) have explicit test coverage.

## Decision Rules

- Classify a finding as Blocker if it would allow an unauthenticated or unauthorized attacker to read, modify, or delete data, or to execute code on the server.
- Classify a finding as High if it enables authenticated attackers to escalate privileges, access other users' data, or exfiltrate secrets.
- Classify a finding as Medium if it increases attack surface, degrades auditability, or violates a security policy without direct immediate exploitability.
- Classify a finding as Low if it represents a defense-in-depth gap or a code quality issue with indirect security implications.
- If a Blocker or High finding is present, the pull request must not be merged until the finding is resolved or accepted with explicit, documented risk acceptance by a named owner.
- If a finding involves a secret already exposed in a committed file, treat it as Blocker and recommend immediate rotation, even if the file is scheduled for deletion.
- Do not mark a finding as "not exploitable" based on upstream validation without verifying that the upstream validation is always enforced and cannot be bypassed.

## DevSecOps Guardrails

- Do not approve a pull request with a Blocker finding without explicit, documented risk acceptance by a named security owner.
- Do not accept "input is validated upstream" without verifying that the upstream validation is unconditional and covers all callers.
- Do not dismiss a secret exposure as "it's a test key" without confirming the key has no permissions in any production or staging environment.
- Do not review only the changed lines; always read the full function or class context to understand the security model.
- Do not accept "the test passes" as evidence that a security control is implemented correctly; verify the test actually exercises the rejection path.
- Do not report generic findings ("validate input") without citing the specific file, function, and line where the risk exists.

## Output Requirements

- **Findings table**: Severity (Blocker/High/Medium/Low/Info), File, Function, Line Range, Description, Remediation.
- **Injection risk summary**: list of all locations where external input reaches sensitive sinks, with finding status.
- **Authentication and authorization assessment**: each protected operation assessed for correct enforcement order and server-side enforcement.
- **Secrets scan result**: confirmation that no hardcoded secrets are present, or list of found secrets with remediation.
- **Logging and error handling assessment**: log statements and error paths assessed for sensitive data exposure.
- **Cryptography assessment**: algorithms and key management assessed with any deprecated or weak usage flagged.
- **Test coverage gap analysis**: security-relevant paths without test coverage.
- **Merge recommendation**: Approve / Request Changes / Block, with explicit reasoning.

## Acceptance Criteria

- All Blocker and High findings are either resolved with a code change or accepted with a documented, named risk owner.
- Every finding includes a specific file, function, and line reference—no generic findings.
- No hardcoded secrets, API keys, tokens, or passwords appear in any committed file.
- Every protected operation has server-side authentication and authorization enforcement verified.
- Deprecated cryptographic algorithms are not used for any security purpose.
- Security-relevant rejection paths have explicit test coverage.

## Anti-Patterns

- **Diff-only review**: Reading only changed lines without understanding the full function's trust model and caller assumptions.
- **Generic findings**: Reporting "SQL injection risk" without naming the specific file, function, and parameter.
- **Test-pass acceptance**: Treating passing tests as proof of security without verifying that the tests exercise rejection paths.
- **Upstream validation trust**: Accepting that input is safe because "it was validated somewhere else" without verifying the upstream control is unconditional.
- **Severity inflation**: Classifying every finding as Blocker to force attention, making the report impossible to triage.
- **False-negative test key dismissal**: Dismissing a hardcoded key as "test-only" without checking whether the key has production permissions.
- **Missing remediation**: Reporting a finding without a specific, actionable remediation recommendation.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const testStrategistBody = `
# {{.Title}}

## Purpose

Design and document a comprehensive test strategy for a feature, service, or change. Identify the required test types, coverage targets, test data requirements, CI gate configuration, and known coverage risks. The strategy must cover positive, negative, security, and edge-case paths. Output is a testable plan that a developer can execute and a CI pipeline can enforce.

## When to Use This Skill

- A new feature or service is entering implementation and a test plan is needed.
- A significant refactoring requires assessing which existing tests provide adequate coverage and which need updating.
- Acceptance criteria from $requirements-analyst need to be mapped to specific test types and locations.
- CI pipelines lack appropriate test gates and a strategy for adding them is needed.
- You are routed here by the central agent via $test-strategist.

## Inputs

- Acceptance criteria and requirements from $requirements-analyst.
- Architecture description and component boundaries.
- Existing test suite structure, coverage reports, and CI pipeline configuration.
- API contracts and interface specifications.
- Security requirements and threat model from $threat-modeler, if available.
- Constraints: test infrastructure, environments available, time budget.

## Skill-Specific Operating Model

1. **Map acceptance criteria to test types.** For each acceptance criterion, identify whether it requires a unit test, integration test, contract test, end-to-end test, or security test. State why.
2. **Define the test pyramid target.** State the target distribution: unit tests at the base (fast, isolated), integration tests in the middle (real dependencies, scoped), E2E and security tests at the top (slow, full-stack). Justify deviations.
3. **Identify unit test scope.** For each changed function or module, identify the pure logic paths, boundary conditions, error cases, and security-relevant inputs that require unit tests.
4. **Identify integration test scope.** For each integration point (database, external service, message queue, cache), identify the contracts to test, the failure modes to simulate, and the test data required.
5. **Identify contract tests.** For each API consumed or produced, identify whether a consumer-driven contract test exists and whether it needs to be created or updated.
6. **Identify security tests.** For each trust boundary, identify the security validation tests: authentication rejection, authorization denial, injection rejection, rate limiting, and boundary enforcement.
7. **Identify negative and edge-case tests.** For each requirement, write the negative: what must NOT happen. For each data input, identify boundary values, null/empty inputs, and malformed inputs.
8. **Define test data requirements.** State what test data is needed, whether it can be synthetic, whether PII must be masked, and whether it must match production schemas.
9. **Define CI gate configuration.** State the minimum coverage threshold, which test suites must pass before merge, and which suites must pass before deployment.
10. **Identify coverage risks.** List paths that cannot be adequately unit-tested due to external dependencies, flaky environment behavior, or infrastructure complexity. Recommend mitigations.

## Skill-Specific Checklist

- [ ] Every Must-have acceptance criterion has at least one test mapped to it by type (unit, integration, contract, E2E, security).
- [ ] Positive paths (happy paths) are covered by at least unit and integration tests.
- [ ] Negative paths (rejection, denial, error) are covered by explicit tests—not only inferred from passing positive tests.
- [ ] Security validation tests cover: authentication rejection (invalid/missing token), authorization denial (insufficient permission), and input validation rejection (malformed, oversized, injected input).
- [ ] Contract tests exist or are planned for every external API consumed or produced.
- [ ] Test data requirements are defined: data source, schema, PII masking, and refresh strategy.
- [ ] Flaky test risks are identified and mitigation strategy is stated (retry, quarantine, synthetic data).
- [ ] CI gate configuration is specified: coverage threshold, required test suites, and gate position (pre-merge vs. pre-deploy).
- [ ] Tests do not use shared mutable state between test cases; each test is independent and idempotent.
- [ ] Test mocks accurately reflect the real behavior of the mocked component, including error paths.
- [ ] Tests are deterministic: no dependency on execution order, wall-clock time, or random values without seeding.
- [ ] Security tests verify rejection paths, not only acceptance paths.
- [ ] Edge cases include: null/empty inputs, maximum length inputs, Unicode boundary characters, concurrent access scenarios, and retry behavior.
- [ ] Coverage gaps (paths that cannot be tested) are documented with the reason and a risk assessment.
- [ ] The strategy distinguishes Must-test (blocks merge) from Should-test (recommended) from Could-test (nice-to-have).

## Decision Rules

- If a security-relevant path has no explicit rejection test, flag it as a test coverage gap and block merge until the test is added.
- If a contract test for an external API does not exist and the API is consumed by multiple services, recommend creating a consumer-driven contract test before the change ships.
- If test data requires PII, recommend synthetic data generation and document the masking approach; do not copy production data without explicit approval and data classification review.
- If a test is marked flaky due to external environment dependency, recommend either fixing the root cause or quarantining the test from the blocking gate until it is fixed.
- If coverage drops below the project threshold, the change must not merge until the gap is addressed or the threshold reduction is explicitly approved.
- If the test requires mocking an authentication or authorization mechanism, use a real implementation in an integration test rather than a mock that would pass even if the production implementation is broken.

## DevSecOps Guardrails

- Do not use production data in test environments without an explicit data classification review and masking policy.
- Do not treat test coverage percentage as the only quality metric; coverage of rejection paths and security-relevant inputs matters more than line coverage.
- Do not allow tests that bypass real authentication or authorization implementations to serve as evidence that the controls work.
- Do not mark a test suite as green if flaky tests are being suppressed rather than fixed; document them as known gaps.
- Do not skip contract tests for external APIs under time pressure; contract drift is a production incident waiting to happen.
- Do not share test state between tests in ways that create order dependencies or mask failures in isolation.

## Output Requirements

- **Test type mapping**: table mapping each acceptance criterion to its required test type(s) and the location where the test belongs.
- **Test pyramid target**: target distribution with justification for any deviation.
- **Unit test scope**: functions and logic paths requiring unit tests, with boundary conditions and error cases listed.
- **Integration test scope**: integration points to test, failure modes to simulate, and test data required.
- **Contract test plan**: APIs to test, test framework, and creation/update schedule.
- **Security test plan**: trust boundary validation tests, with specific scenarios per boundary.
- **Negative and edge-case matrix**: for each requirement, the negative case and edge cases to cover.
- **CI gate specification**: coverage threshold, required suites, gate positions.
- **Coverage risk register**: paths that cannot be adequately tested with reason and risk assessment.

## Acceptance Criteria

- Every Must-have acceptance criterion has at least one test mapped to a specific type and location.
- Negative paths (rejection, denial, error) have explicit test coverage.
- Security validation tests cover authentication rejection, authorization denial, and input validation rejection for every trust boundary.
- Contract tests exist or are explicitly scheduled for every external API consumed or produced.
- CI gate configuration is specified with minimum coverage threshold and required test suites.
- Test data requirements are defined with PII handling stated.
- Coverage gaps are documented with risk assessment.

## Anti-Patterns

- **Happy-path-only testing**: Writing only tests that verify the system does the right thing when inputs are valid, ignoring rejection, failure, and abuse scenarios.
- **Coverage theater**: Achieving line coverage targets with trivial tests that do not verify behavior.
- **Mock-as-implementation**: Writing mocks that return expected values unconditionally, making the test pass even when the production implementation is wrong.
- **Shared test state**: Using class-level or module-level mutable state that causes tests to pass or fail depending on execution order.
- **PII in test data**: Using real or copied production data in tests without masking, creating a data breach risk in non-production environments.
- **Flaky test acceptance**: Merging with known flaky tests in the blocking gate rather than fixing or quarantining them.
- **Test-after-the-fact**: Writing tests only after a bug is found in production rather than as part of the feature development cycle.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`
