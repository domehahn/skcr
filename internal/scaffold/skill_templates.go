package scaffold

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/domehahn/skcr/v2/internal/cncf"
	platformcompat "github.com/domehahn/skcr/v2/internal/platforms"
)

type skillTemplateData struct {
	Name         string
	Title        string
	Description  string
	Version      string
	Since        string
	LastModified string
	Owner        string
	Stability    string
	License      string
	Platforms    []string
	MinPlatforms []platformcompat.CompatibilityEntry
}

// RenderRegisteredSkillMarkdown renders a built-in SDLC / DevSecOps skill.
// The second return value is false when name is not registered.
func RenderRegisteredSkillMarkdown(name, title, description, version, since, lastModified, owner, stability, license string, platforms []string) (string, bool, error) {
	return RenderRegisteredSkillMarkdownWithCompatibility(name, title, description, version, since, lastModified, owner, stability, license, platforms, platformcompat.AllMinVersions())
}

func RenderRegisteredSkillMarkdownWithCompatibility(name, title, description, version, since, lastModified, owner, stability, license string, platforms []string, minPlatforms []platformcompat.CompatibilityEntry) (string, bool, error) {
	if _, ok := skillBodies[name]; !ok {
		return "", false, nil
	}
	if title == "" {
		title = skillTitle(name)
	}
	data := skillTemplateData{
		Name:         name,
		Title:        title,
		Description:  description,
		Version:      version,
		Since:        since,
		LastModified: lastModified,
		Owner:        owner,
		Stability:    stability,
		License:      license,
		Platforms:    platforms,
		MinPlatforms: minPlatforms,
	}
	rendered, err := renderSkillTemplate(name, data)
	if err != nil {
		return "", true, err
	}
	return rendered, true, nil
}

const skillFrontmatter = `---
name: {{quote .Name}}
description: {{quote .Description}}
version: {{quote .Version}}
since: {{quote .Since}}
last_modified: {{quote .LastModified}}
authors:
  - {{quote .Owner}}
stability: {{quote .Stability}}
min_platform_version:
{{- range .MinPlatforms }}
  {{.Name}}: {{quote .MinVersion}}
{{- end }}
deprecated_since:
replaces:
supersedes: []
changelog:
  - version: "{{.Version}}"
    date: "{{.LastModified}}"
    change: "Initial generated production-ready SDLC / DevSecOps skill"
---
`

type skillContent struct {
	Purpose             string
	When                []string
	Operating           []string
	ReviewScope         []string
	Checklist           []string
	DecisionRules       []string
	FindingCategories   []string
	SeverityGuidance    []string
	DevSecOpsGuardrails []string
	OutputRequirements  []string
	AcceptanceCriteria  []string
	AntiPatterns        []string
}

var baseSDLCSkillNames = []string{
	"requirements-analyst",
	"cost-based-planner",
	"architecture-reviewer",
	"threat-modeler",
	"safe-implementer",
	"test-strategy-engineer",
	"verification-reviewer",
	"security-reviewer",
	"secrets-reviewer",
	"dependency-supply-chain-reviewer",
	"ci-cd-reviewer",
	"iac-gitops-reviewer",
	"compliance-governance-reviewer",
	"release-readiness-reviewer",
	"observability-reviewer",
	"incident-postmortem-assistant",
	"documentation-maintainer",
	"universal-skill-creator",
}

var AdditionalSDLCSkillNames = []string{
	"dora-readiness-reviewer",
	"ict-risk-management-reviewer",
	"ict-third-party-risk-reviewer",
	"ict-incident-reporting-reviewer",
	"operational-resilience-tester",
	"audit-evidence-reviewer",
	"control-mapping-reviewer",
	"outsourcing-exit-strategy-reviewer",
	"documentation-governance-reviewer",
	"runbook-playbook-maintainer",
	"architecture-decision-recorder",
	"audit-traceability-maintainer",
	"policy-documentation-maintainer",
	"evidence-package-creator",
	"devsecops-maturity-reviewer",
	"pipeline-security-architect",
	"software-supply-chain-architect",
	"policy-as-code-engineer",
	"secure-developer-platform-reviewer",
	"vulnerability-management-coordinator",
	"cloud-landing-zone-reviewer",
	"cloud-governance-reviewer",
	"finops-reviewer",
	"sre-reliability-reviewer",
	"kubernetes-platform-reviewer",
	"gitops-operations-reviewer",
	"aiops-signal-correlation-reviewer",
	"alert-quality-reviewer",
	"auto-remediation-reviewer",
	"mlops-governance-reviewer",
	"llmops-security-reviewer",
	"ai-change-risk-reviewer",
	"agent-containment-reviewer",
	"agent-runtime-enforcement-reviewer",
	"agent-behavior-eval-engineer",
	"backdoor-persistence-reviewer",
	"agentic-threat-modeler",
	"security-invariant-test-engineer",
	"privacy-data-protection-reviewer",
	"api-contract-reviewer",
	"secure-design-reviewer",
	"policy-as-code-reviewer",
	"container-security-reviewer",
	"identity-access-reviewer",
	"risk-acceptance-reviewer",
	"secure-code-reviewer",
	"performance-scalability-reviewer",
	"migration-change-reviewer",
	"sbom-vulnerability-management-reviewer",
	"developer-experience-reviewer",
	"resilience-reviewer",
	"backup-restore-reviewer",
}

var PaymentSkillNames = []string{
	"payment-integration-engineer",
	"payment-security-reviewer",
	"payment-webhook-reviewer",
	"payment-flow-tester",
	"refund-dispute-handler",
	"payment-reconciliation-reviewer",
	"subscription-billing-engineer",
	"sca-3ds-reviewer",
	"payment-fraud-risk-reviewer",
	"payment-observability-reviewer",
	"payment-provider-migration-reviewer",
	"payment-compliance-reviewer",
	"payment-operations-agent",
	"stripe-integration-engineer",
	"paypal-integration-engineer",
	"adyen-integration-engineer",
}

var LanguageSkillNames = []string{
	"java-reviewer", "golang-reviewer", "python-reviewer", "ruby-reviewer", "javascript-reviewer",
	"typescript-reviewer", "rust-reviewer", "csharp-reviewer", "kotlin-reviewer", "php-reviewer",
}

var FrameworkSkillNames = []string{
	"spring-boot-reviewer", "quarkus-reviewer", "angular-reviewer", "react-reviewer", "vuejs-reviewer",
	"nodejs-reviewer", "nextjs-reviewer", "django-reviewer", "fastapi-reviewer", "ruby-on-rails-reviewer",
}

var InfrastructureSkillNames = []string{
	"kubernetes-platform-reviewer", "aws-cloud-reviewer", "azure-cloud-reviewer", "gcp-cloud-reviewer",
	"terraform-reviewer", "opentofu-reviewer", "ansible-reviewer", "vagrant-reviewer",
	"virtualization-reviewer", "helm-reviewer",
}

var SDLCSkillNames = buildSDLCSkillNames()

func buildSDLCSkillNames() []string {
	names := []string{}
	for _, group := range [][]string{baseSDLCSkillNames, AdditionalSDLCSkillNames, PaymentSkillNames, LanguageSkillNames, FrameworkSkillNames, InfrastructureSkillNames, cncf.SkillNames()} {
		for _, name := range group {
			if !containsSkillName(names, name) {
				names = append(names, name)
			}
		}
	}
	return names
}

func containsSkillName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}

var sharedDevSecOpsGuardrails = []string{
	"Do not read secrets, `.env` files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.",
	"Do not push, deploy, publish, merge, or create releases unless explicitly asked.",
	"Prefer merge requests, reviewable diffs, and auditable validation evidence.",
	"Prefer least privilege, minimal changes, and explicit rollback notes.",
	"Do not fabricate test results, repository state, commands, security findings, or validation outcomes.",
	"Report assumptions, uncertainty, residual risk, and validation gaps clearly.",
}

var sharedSpecDrivenChangeContext = []string{
	"Treat repository specs, ADRs, runbooks, change proposals, design notes, and task files as durable context that outlives a chat session.",
	"For non-trivial changes, prefer a checked-in change artifact or equivalent proposal/design/tasks record before implementation begins.",
	"Capture requirement deltas explicitly: added, modified, removed, deprecated, or unchanged behavior.",
	"Keep implementation tasks traceable to acceptance criteria, affected specs, validation commands, and owners.",
	"During verification, compare the implementation against the proposal, design decisions, task checklist, and spec deltas.",
	"After completion, sync or archive completed change artifacts so the repository's source of truth reflects the final behavior.",
	"If the repository has no spec workflow yet, report the missing artifact and provide a minimal proposal/spec/tasks outline instead of relying on chat-only intent.",
}
var sdlcSkillContent = map[string]skillContent{
	"requirements-analyst": {
		Purpose: "Analyze, clarify, and make requirements testable before design or implementation begins. Separate stated requirements from assumptions, identify ambiguity and contradictions, and turn vague intent into acceptance criteria that can be verified.",
		When: []string{
			"A feature, epic, user story, or change request needs refinement.",
			"Acceptance criteria are missing, vague, contradictory, or not testable.",
			"Security, compliance, privacy, NFR, stakeholder, or dependency requirements are implied but unstated.",
			"Scope boundaries, priorities, ownership, or external-system dependencies are unclear.",
			"Implementation should not begin until open questions and acceptance criteria are explicit.",
		},
		Operating: []string{
			"Classify each requirement as functional, non-functional, security, compliance, privacy, operational, or out-of-scope.",
			"Extract assumptions, constraints, dependencies, owners, stakeholders, and open questions into separate lists.",
			"Rewrite ambiguous statements into measurable acceptance criteria without inventing stakeholder intent.",
			"Identify contradictions and missing ownership before recommending implementation.",
			"Prioritize requirements using must-have, should-have, could-have, and out-of-scope categories when evidence supports it.",
		},
		ReviewScope: []string{
			"Functional behavior and externally observable outcomes.",
			"Non-functional requirements such as performance, availability, scalability, usability, and reliability.",
			"Security, compliance, privacy, retention, audit, and data-processing requirements.",
			"Stakeholders, ownership, approvals, dependencies, constraints, and external systems.",
			"Ambiguity, contradictions, assumptions, open questions, prioritization, and scope boundaries.",
		},
		Checklist: []string{
			"Separate requirements from assumptions and open questions.",
			"Identify ambiguous terms and untestable statements.",
			"Convert vague requirements into testable acceptance criteria.",
			"Identify missing non-functional requirements.",
			"Identify security, compliance, and privacy requirements explicitly.",
			"Identify dependencies, constraints, and external systems.",
			"Distinguish must-have, should-have, could-have, and out-of-scope items.",
			"Identify contradictory requirements and unresolved decisions.",
			"Identify missing stakeholders, ownership, and approvers.",
			"Maintain traceability from each requirement to acceptance criteria, owner, and source.",
			"Produce actionable clarification questions with owners.",
			"Map each acceptance criterion to observable behavior.",
			"Call out requirements that are not ready for implementation.",
		},
		DecisionRules: []string{
			"If a requirement cannot be tested, mark it not ready and ask for a measurable criterion.",
			"If a requirement touches personal data, add explicit privacy and retention questions.",
			"If security or compliance is implied but not stated, record it as a missing requirement instead of assuming it away.",
			"If two requirements conflict, do not choose silently; document the contradiction and required decision owner.",
			"If scope is unclear, separate in-scope, out-of-scope, and unknown items before implementation planning.",
			"If priority is not evidenced, mark it unknown rather than assigning must-have status.",
		},
		FindingCategories: []string{
			"Ambiguous or untestable requirement.",
			"Missing non-functional requirement.",
			"Missing security, compliance, or privacy requirement.",
			"Contradictory stakeholder expectation.",
			"Missing owner, dependency, or external-system constraint.",
			"Incomplete or unverifiable acceptance criteria.",
		},
		SeverityGuidance: []string{
			"Critical: a must-have requirement is contradictory, legally unsafe, or impossible to verify.",
			"High: security, privacy, compliance, or external dependency requirements are missing.",
			"Medium: NFR thresholds, ownership, or prioritization are unclear but implementation can be scoped cautiously.",
			"Low: wording, examples, or documentation can be improved without changing scope.",
		},
		OutputRequirements: []string{
			"Requirements register with type, priority, owner, status, and acceptance criteria.",
			"Traceability map from requirement source to acceptance criteria and decision owner.",
			"Assumptions log separated from confirmed requirements.",
			"Open-questions list with owner, blocking status, and suggested wording.",
			"Scope summary with in-scope, out-of-scope, and unresolved items.",
			"Security, compliance, privacy, NFR, and dependency notes.",
			"Implementation-readiness recommendation: ready, ready with caveats, or not ready.",
		},
		AcceptanceCriteria: []string{
			"Every must-have requirement has at least one testable acceptance criterion.",
			"Assumptions and open questions are separated from requirements.",
			"Security, compliance, privacy, and NFR gaps are explicitly called out.",
			"Contradictions and ownership gaps are documented with decision owners.",
			"Scope boundaries and out-of-scope items are visible to implementers.",
			"The output gives a clear readiness recommendation.",
		},
		AntiPatterns: []string{
			"Treating assumptions as confirmed requirements.",
			"Using vague phrases such as fast, secure, intuitive, or reasonable without thresholds.",
			"Skipping privacy or compliance because the request sounds functional.",
			"Resolving stakeholder contradictions silently.",
			"Producing acceptance criteria that depend on implementation details instead of observable behavior.",
			"Marking everything must-have without evidence.",
		},
	},
	"cost-based-planner": {
		Purpose: "Plan work by balancing implementation cost, operational cost, maintenance cost, uncertainty, risk, and delivered value. Recommend the smallest useful implementation path and expose trade-offs before coding begins.",
		When: []string{
			"A request needs sizing, sequencing, or phased delivery.",
			"The implementation path has meaningful uncertainty, migration, infrastructure, or maintenance cost.",
			"Build-versus-buy, MVP scope, or cost-of-delay decisions are open.",
			"A broad change should be decomposed into lower-risk increments.",
			"The user needs a plan before implementation starts.",
		},
		Operating: []string{
			"Identify cost drivers across implementation, operation, maintenance, migration, licensing, infrastructure, and opportunity cost.",
			"Use repository evidence to estimate effort, blast radius, and validation cost.",
			"Separate MVP, incremental rollout, and full-scope options.",
			"Make uncertainty visible and reduce it with targeted file reads or experiments.",
			"Recommend the smallest useful path that preserves rollback and validation.",
		},
		ReviewScope: []string{
			"Implementation effort and complexity drivers.",
			"Operational, maintenance, migration, license, infrastructure, and support cost.",
			"Uncertainty, assumptions, dependencies, and cost-of-delay.",
			"MVP, phased delivery, rollout, rollback, and validation cost.",
			"Build-versus-buy and reuse-versus-new-code trade-offs.",
		},
		Checklist: []string{
			"Separate one-time implementation cost from recurring operational cost.",
			"Identify the main cost drivers and complexity drivers.",
			"Identify uncertainty factors and assumptions.",
			"Compare MVP, incremental rollout, and full-scope implementation.",
			"Identify build-versus-buy and reuse trade-offs.",
			"Check budget constraints, funding limits, and expected cost envelope.",
			"Identify hidden maintenance, migration, and support costs.",
			"Estimate effort using repository-specific evidence when possible.",
			"Identify high-cost dependencies, integrations, and platform changes.",
			"Recommend the smallest useful implementation path.",
			"Call out cost risks that affect prioritization.",
			"Include rollback and validation cost in the plan.",
			"Name decisions that require stakeholder input.",
		},
		DecisionRules: []string{
			"If a low-cost MVP can validate the goal, recommend it before full-scope work.",
			"If uncertainty dominates cost, plan a discovery step before implementation.",
			"If a dependency adds recurring operational burden, include it in prioritization.",
			"If migration cost is high, separate migration from feature delivery.",
			"If build-versus-buy is unresolved, compare license, integration, maintenance, and lock-in costs.",
			"If validation cost is high, include it as part of delivery cost rather than a footnote.",
		},
		FindingCategories: []string{
			"High implementation complexity.",
			"Hidden operational or maintenance cost.",
			"Migration or rollback cost risk.",
			"License, infrastructure, or vendor cost exposure.",
			"Unclear value, priority, or cost-of-delay.",
			"Uncertainty requiring discovery before build.",
		},
		SeverityGuidance: []string{
			"Critical: cost or migration risk makes the proposed path unsafe without a different plan.",
			"High: major recurring cost, irreversible migration, or expensive dependency is unaccounted for.",
			"Medium: cost estimates rely on assumptions but can be reduced with discovery.",
			"Low: minor sequencing or documentation issue affects planning clarity.",
		},
		OutputRequirements: []string{
			"Costed plan with MVP, incremental, and full-scope options.",
			"Budget impact summary covering expected spend, constraints, and unresolved approvals.",
			"List of cost drivers with evidence and assumptions.",
			"Risk and uncertainty register with reduction steps.",
			"Recommended implementation sequence and rollback points.",
			"Validation plan with expected command or review cost.",
			"Build-versus-buy or reuse rationale where applicable.",
		},
		AcceptanceCriteria: []string{
			"Costs are split into implementation, operation, maintenance, migration, license, and infrastructure where relevant.",
			"The recommended plan names the smallest useful implementation path.",
			"Uncertainty and assumptions are visible with next steps to reduce them.",
			"High-cost dependencies and rollback constraints are identified.",
			"Validation and release costs are included in the plan.",
			"The plan can be executed incrementally.",
		},
		AntiPatterns: []string{
			"Ignoring recurring operational cost.",
			"Treating the full solution as the only option.",
			"Estimating effort without reading relevant repository evidence.",
			"Hiding migration or rollback cost until implementation.",
			"Adding dependencies without considering maintenance and license cost.",
			"Optimizing for low upfront effort while increasing long-term support burden.",
		},
	},
	"architecture-reviewer": {
		Purpose: "Review architecture for module boundaries, service boundaries, coupling, cohesion, dependency direction, data ownership, resilience, security boundaries, and ADR-worthy decisions.",
		When: []string{
			"A design, PR, or refactor changes architecture, module boundaries, or service boundaries.",
			"API contracts, shared libraries, data ownership, or deployment topology change.",
			"Scalability, resilience, runtime coupling, or security boundaries need review.",
			"Circular dependencies, layering violations, or unclear ownership are suspected.",
			"An ADR should be created or updated.",
		},
		Operating: []string{
			"Map components, modules, services, APIs, data stores, queues, and deployment units.",
			"Verify dependency direction, layering, cohesion, and ownership boundaries.",
			"Trace data flows and ownership across tables, topics, buckets, and integrations.",
			"Review runtime coupling, fan-out, retry behavior, scalability, and cascading failure risk.",
			"Identify decisions that deserve ADRs and distinguish architecture risk from style preference.",
		},
		ReviewScope: []string{
			"Module boundaries, service boundaries, layering, coupling, cohesion, and circular dependencies.",
			"API contracts, interface stability, compatibility, and versioning.",
			"Data ownership, data flows, cross-boundary writes, and direct database access.",
			"Runtime and deployment coupling, scalability, resilience, retry storms, and fan-out.",
			"Security boundaries, trust boundaries, ownership, and ADR candidates.",
		},
		Checklist: []string{
			"Identify architectural layers and verify dependency direction.",
			"Detect circular dependencies between modules, packages, services, or libraries.",
			"Check whether module boundaries align with business capabilities or clear technical responsibilities.",
			"Review coupling between services, APIs, databases, queues, shared libraries, and deployment units.",
			"Identify shared mutable state, shared database writes, hidden dependencies, and temporal coupling.",
			"Check public interfaces for ownership, versioning, compatibility expectations, and tests.",
			"Verify explicit data ownership for domain objects, tables, topics, buckets, and integrations.",
			"Identify cross-boundary writes or direct database access across service boundaries.",
			"Review synchronous call chains for fan-out, latency amplification, retry storms, and cascading failures.",
			"Identify ADR-worthy decisions and missing architecture documentation.",
			"Check security and trust boundaries between components.",
			"Assess deployment coupling and independent rollback ability.",
		},
		DecisionRules: []string{
			"If a dependency violates the intended layer direction, classify it as architecture risk.",
			"If circular dependencies exist, recommend interface inversion or boundary redesign.",
			"If data ownership is unclear, block cross-boundary writes until an owner is named.",
			"If synchronous fan-out can cascade failures, recommend async, timeout, circuit breaker, or fallback design.",
			"If a public contract lacks versioning or compatibility tests, flag interface stability risk.",
			"If a decision changes long-term structure, recommend an ADR.",
		},
		FindingCategories: []string{
			"Circular dependencies and dependency direction violations.",
			"Excessive coupling or weak cohesion.",
			"Unclear module, service, or data ownership.",
			"Unstable API contracts and compatibility risk.",
			"Runtime coupling, fan-out, retry storm, and cascading failure risk.",
			"Missing ADR or architecture documentation.",
		},
		SeverityGuidance: []string{
			"Critical: architecture enables unsafe deployment, data corruption, or unavoidable cascading failure.",
			"High: coupling, ownership, or boundary issue blocks safe evolution or security isolation.",
			"Medium: maintainability, scalability, or compatibility risk is likely but controllable.",
			"Low: documentation or ADR gap that does not currently block delivery.",
		},
		OutputRequirements: []string{
			"Architecture summary with components, boundaries, and data flows.",
			"Findings table with severity, component, evidence, impact, and recommendation.",
			"Dependency and coupling analysis with circular dependencies called out.",
			"API contract and data ownership review.",
			"Runtime resilience, scalability, and deployment coupling notes.",
			"ADR recommendations with decision question and rationale.",
		},
		AcceptanceCriteria: []string{
			"Module and service boundaries are identified or marked unknown.",
			"Circular dependencies and layering violations are explicitly assessed.",
			"Data ownership and cross-boundary access are reviewed.",
			"Critical and High findings include concrete mitigation.",
			"ADR candidates are listed for broad decisions.",
			"Findings distinguish structural risk from style preference.",
		},
		AntiPatterns: []string{
			"Reporting style preferences as architecture risk.",
			"Recommending service extraction without ownership and operational cost analysis.",
			"Ignoring data ownership while reviewing module boundaries.",
			"Approving circular dependencies because they compile today.",
			"Treating ADRs as optional for precedent-setting decisions.",
			"Suggesting broad rewrites for localized coupling issues.",
		},
	},
	"threat-modeler": {
		Purpose: "Identify and prioritize threats for features, services, APIs, and architecture changes using assets, trust boundaries, entry points, data flows, STRIDE, abuse cases, mitigations, controls, and residual risk.",
		When: []string{
			"A feature or service crosses trust boundaries or handles sensitive assets.",
			"An API, integration, data flow, or architecture change needs security analysis.",
			"Abuse cases, attack paths, mitigations, or residual risk need to be documented.",
			"Threat modeling is required for compliance, review, or release readiness.",
			"Existing controls are unclear or unverified.",
		},
		Operating: []string{
			"Define assets, actors, entry points, trust boundaries, and data flows.",
			"Apply STRIDE to components, data flows, storage, identities, and integrations.",
			"Write realistic abuse cases and attack paths.",
			"Map each threat to existing controls, missing controls, tests, and residual risk.",
			"Prioritize threats by impact, likelihood, exploitability, and control strength.",
		},
		ReviewScope: []string{
			"Assets and sensitivity classification.",
			"Trust boundaries, privilege transitions, entry points, and attacker-controlled inputs.",
			"STRIDE threats, abuse cases, and attack paths.",
			"Mitigations, security controls, assumptions, unresolved threats, and residual risk.",
			"Security tests and validation for high-risk paths.",
		},
		Checklist: []string{
			"Identify assets and classify their sensitivity.",
			"Identify trust boundaries and privilege transitions.",
			"Identify entry points and attacker-controlled inputs.",
			"Map data flows across components and storage.",
			"Identify spoofing risks.",
			"Identify tampering risks.",
			"Identify repudiation and auditability risks.",
			"Identify information disclosure risks.",
			"Identify denial-of-service risks.",
			"Identify elevation-of-privilege risks.",
			"Define realistic abuse cases.",
			"Map threats to concrete mitigations.",
			"Distinguish existing controls from missing controls.",
			"Identify residual risk after mitigation.",
			"Recommend security tests for high-risk paths.",
		},
		DecisionRules: []string{
			"If a trust boundary is crossed, require at least one threat and one control for that boundary.",
			"If a high-impact threat lacks a mitigation, mark it unresolved rather than accepted.",
			"If an abuse case is unrealistic, document the assumption and adjust likelihood, not impact.",
			"If a control is only planned, residual risk remains open.",
			"If entry points are unknown, treat the model as incomplete.",
			"If STRIDE categories are skipped, state why they are not applicable.",
		},
		FindingCategories: []string{
			"Spoofing and identity confusion.",
			"Tampering and integrity failure.",
			"Repudiation and missing audit evidence.",
			"Information disclosure and privacy exposure.",
			"Denial of service and resource exhaustion.",
			"Elevation of privilege and authorization bypass.",
		},
		SeverityGuidance: []string{
			"Critical: likely attack path compromises sensitive assets or admin control without effective mitigation.",
			"High: plausible attacker can bypass authorization, exfiltrate sensitive data, or disrupt critical service.",
			"Medium: abuse requires constraints but exposes meaningful control weakness.",
			"Low: defense-in-depth or documentation gap with limited direct exploitability.",
		},
		OutputRequirements: []string{
			"Scope statement with assets, actors, trust boundaries, and entry points.",
			"Data-flow and attack-surface summary.",
			"Threat register with STRIDE category, abuse case, impact, likelihood, controls, and status.",
			"Mitigation and residual-risk register with owners.",
			"Recommended security tests for high-risk threats.",
			"Assumptions and unresolved-threats list.",
		},
		AcceptanceCriteria: []string{
			"Assets, entry points, trust boundaries, and data flows are named.",
			"STRIDE is applied or explicitly scoped out with rationale.",
			"Each high-risk threat has a mitigation or named residual risk owner.",
			"Abuse cases are concrete and realistic.",
			"Security tests are recommended for high-risk paths.",
			"Assumptions and unresolved threats are explicit.",
		},
		AntiPatterns: []string{
			"Listing STRIDE labels without abuse cases.",
			"Assuming internal networks are trusted.",
			"Marking planned controls as implemented mitigations.",
			"Omitting residual risk owners.",
			"Ignoring denial-of-service because confidentiality dominates discussion.",
			"Removing threats because they are uncomfortable to address.",
		},
	},
	"safe-implementer": {
		Purpose: "Implement changes safely, minimally, and auditable while preserving public APIs, tests, rollback, input validation, error handling, safe defaults, and validation evidence.",
		When: []string{
			"The user asks for code, configuration, tests, documentation, or generated file changes.",
			"A requirement is ready for implementation.",
			"A bug fix needs a minimal, testable change.",
			"A migration or feature flag needs safe rollout treatment.",
			"Validation evidence must accompany the change.",
		},
		Operating: []string{
			"Inspect existing patterns, tests, ownership boundaries, and generated-file flows before editing.",
			"Implement only requested behavior and avoid broad refactoring.",
			"Trace every behavioral change to the requested change, specification, Goal, Contract, or an explicitly documented supporting requirement.",
			"Add or update tests for changed behavior and relevant failure paths.",
			"Validate inputs at trust boundaries and handle errors safely.",
			"Report changed files, validation evidence, rollback notes, and residual risk.",
		},
		ReviewScope: []string{
			"Minimal change principle and no broad refactoring.",
			"Input validation, error handling, safe defaults, and API compatibility.",
			"Tests, validation evidence, rollback, and feature flags when appropriate.",
			"Secrets avoidance, no global side effects, and concurrency safety.",
			"Generated-file synchronization and reviewable diffs.",
		},
		Checklist: []string{
			"Implement only the requested behavior.",
			"Avoid broad refactoring unless explicitly required.",
			"Preserve public APIs unless a breaking change is explicitly requested.",
			"Add or update tests for changed behavior.",
			"Validate inputs at trust boundaries.",
			"Handle errors explicitly and safely.",
			"Avoid hardcoded secrets or credentials.",
			"Avoid global side effects and hidden state changes.",
			"Include rollback or mitigation notes for risky changes.",
			"Summarize changed files and validation performed.",
			"Keep generated outputs synchronized with canonical sources.",
			"Separate formatting-only churn from functional changes.",
			"Identify new outbound requests, privileged paths, persistence mechanisms, bypasses, or tool capabilities and record their requirement provenance.",
		},
		DecisionRules: []string{
			"If required behavior implies unrelated refactoring, ask before expanding scope.",
			"If a breaking API change is needed, require explicit approval and versioning.",
			"If input crosses a trust boundary, validate before use.",
			"If a migration is required, make rollout and rollback explicit.",
			"If tests cannot run, report the reason and residual risk.",
			"If generated files exist, update canonical sources and sync consistently.",
			"If behavior cannot be traced to the task, specification, Goal, Contract, or a documented supporting requirement, do not implement it and require security review.",
		},
		FindingCategories: []string{
			"Scope creep or unrelated change.",
			"Missing tests or validation evidence.",
			"Unsafe input validation or error handling.",
			"API compatibility or migration risk.",
			"Secret exposure or unsafe configuration.",
			"Global side effect or hidden state change.",
			"Unexplained behavior or capability without requirement provenance.",
		},
		SeverityGuidance: []string{
			"Critical: change introduces data loss, secret exposure, or unsafe production behavior.",
			"High: missing validation, rollback, or tests for risky behavior.",
			"Medium: maintainability or compatibility risk needs follow-up.",
			"Low: small cleanup, docs, or validation gap with limited blast radius.",
		},
		OutputRequirements: []string{
			"Changed files and purpose of each change.",
			"Acceptance criteria satisfied by the implementation.",
			"Tests and validation commands run with results.",
			"Rollback or mitigation notes for risky changes.",
			"Known gaps, skipped checks, and residual risks.",
			"Generated files or sync actions performed.",
		},
		AcceptanceCriteria: []string{
			"Requested behavior is implemented without unrelated scope.",
			"Relevant tests or validation pass or failures are explained.",
			"Inputs and errors are handled safely.",
			"Public API compatibility is preserved or approved.",
			"No secrets or unsafe globals are introduced.",
			"Rollback and generated-file consistency are addressed.",
			"Every behavioral change is traceable to an authorized requirement, Goal, Contract declaration, or explicit supporting decision.",
		},
		AntiPatterns: []string{
			"Broad refactoring in a narrow fix.",
			"Changing public APIs without approval.",
			"Skipping tests because the change is small.",
			"Swallowing errors or leaking internal errors externally.",
			"Hardcoding environment-specific secrets or URLs.",
			"Editing generated copies without updating canonical source.",
			"Introducing undocumented network calls, privileged control paths, persistence, bypasses, or other behavior unrelated to the authorized change.",
		},
	},
	"test-strategy-engineer": {
		Purpose: "Design a risk-based test strategy for features, fixes, migrations, and releases across unit, integration, contract, E2E, regression, negative, security, performance, test data, mocking, CI gates, and coverage risk.",
		When: []string{
			"A change needs a test plan before implementation or release.",
			"Acceptance criteria must be mapped to concrete test types and validation evidence.",
			"Risky migrations, external integrations, contracts, or security-sensitive behavior need targeted coverage.",
			"Existing tests are flaky, slow, missing, or not aligned with the changed behavior.",
			"CI gates need a must-have versus optional validation strategy.",
		},
		Operating: []string{
			"Map user-visible behavior, invariants, edge cases, and failure modes to test levels.",
			"Classify test scope as unit, integration, contract, E2E, regression, negative, security, performance, or manual verification.",
			"Prioritize tests by production risk, blast radius, frequency of change, and cost to execute.",
			"Identify fixtures, test data, mocks, stubs, and cleanup needed for deterministic results.",
			"Define CI gating order so fast deterministic checks block before expensive or optional suites.",
		},
		ReviewScope: []string{
			"Unit, integration, contract, E2E, regression, negative, security, and performance coverage.",
			"Test data, fixtures, mocking strategy, determinism, isolation, and cleanup.",
			"CI gates, coverage risks, flaky tests, and validation sequencing.",
			"Boundary cases, abuse cases, migrations, and external dependencies.",
			"Must-have versus optional test scope.",
		},
		Checklist: []string{
			"Map each acceptance criterion to at least one concrete test or explicit manual verification step.",
			"Identify pure logic, validators, parsers, and branching behavior that require unit tests.",
			"Identify database, queue, filesystem, network, cache, auth provider, or external-service boundaries that require integration tests.",
			"Identify public APIs, events, CLI output, schemas, or SDK contracts that require contract tests.",
			"Add negative tests for invalid input, authorization failure, malformed payloads, timeouts, and abuse cases.",
			"Add regression tests that fail on the reported bug before accepting the fix.",
			"Specify fixtures, seed data, factories, mocks, and cleanup needed for deterministic isolated tests.",
			"Identify flaky-test risks caused by time, randomness, ordering, retries, shared state, or external services.",
			"Define CI gates with fast required checks before slow E2E, performance, or exploratory checks.",
			"Separate release-blocking tests from optional confidence-building tests with rationale.",
			"Include migration, rollback, compatibility, and data-loss test scenarios when persistence changes.",
			"Call out coverage gaps that remain after the proposed test plan.",
		},
		DecisionRules: []string{
			"If a changed behavior has acceptance criteria but no automated or named manual verification, classify the strategy as incomplete.",
			"If an external API, event, or schema changes, require contract tests or documented consumer compatibility verification.",
			"If data migration or rollback is involved, require forward migration, rollback, idempotency, and corrupted-input scenarios.",
			"If a test depends on real time, randomness, global state, network services, or shared data, require determinism controls or mark flaky risk.",
			"If security-sensitive behavior changes, require negative tests for authn, authz, validation, and unsafe input paths.",
			"If CI runtime is high, split blocking smoke coverage from scheduled exhaustive coverage instead of dropping critical tests.",
		},
		FindingCategories: []string{
			"Untested acceptance criterion or user-visible behavior.",
			"Missing contract coverage for API, event, schema, CLI, or SDK compatibility.",
			"Missing negative, authorization, validation, abuse-case, or error-path coverage.",
			"Flaky, nondeterministic, order-dependent, or environment-dependent test design.",
			"Missing migration, rollback, idempotency, or data-integrity validation.",
			"Weak CI gate that allows high-risk changes without required validation.",
		},
		SeverityGuidance: []string{
			"Critical: release can corrupt data, bypass security, or break core flows with no blocking validation.",
			"High: key acceptance criteria, contracts, migrations, or auth paths lack required tests before merge or release.",
			"Medium: meaningful edge cases, fixtures, determinism, or CI sequencing gaps reduce confidence but can be tracked.",
			"Low: naming, organization, coverage reporting, or optional confidence checks can improve maintainability.",
		},
		OutputRequirements: []string{
			"Test matrix mapping requirements or changed behaviors to test type, file/location, owner, and gate status.",
			"Must-have versus optional test list with risk rationale.",
			"Negative, regression, contract, migration, and rollback scenarios where applicable.",
			"Fixture, mock, seed data, cleanup, and determinism requirements.",
			"CI gate recommendation with blocking, non-blocking, scheduled, and manual checks separated.",
			"Residual coverage gaps and explicit release risk if tests are deferred.",
		},
		AcceptanceCriteria: []string{
			"Every acceptance criterion has a mapped automated test or named manual verification step.",
			"Changed contracts have compatibility or consumer verification coverage.",
			"Critical negative, auth, validation, migration, and rollback paths are covered or explicitly risk-accepted.",
			"Required tests are deterministic, isolated, and suitable for CI gating.",
			"The strategy distinguishes required release blockers from optional confidence checks.",
			"Residual test gaps include owner, impact, and follow-up recommendation.",
		},
		AntiPatterns: []string{
			"Counting coverage percentage without mapping tests to changed behavior and risk.",
			"Using only happy-path E2E tests while missing unit, contract, and negative coverage.",
			"Relying on live external services or wall-clock timing for blocking CI tests.",
			"Treating manual QA as sufficient without named scenarios and evidence.",
			"Skipping rollback, migration, or compatibility tests because deployment tooling exists.",
			"Adding broad slow tests that make CI unusable instead of targeted gates.",
		},
	},
	"verification-reviewer": {
		Purpose: "Verify implementation results against requirements, changed files, validation commands, test evidence, edge cases, security validation, documentation consistency, and residual risk.",
		When: []string{
			"A change claims completion and needs independent evidence-based verification.",
			"Validation output, test coverage, or acceptance criteria need review before merge or release.",
			"The implementation may have scope creep, missed edge cases, or generated-output drift.",
			"Security, docs, migration, rollback, or compatibility claims need evidence.",
			"A final pass, conditional pass, or block recommendation is required.",
		},
		Operating: []string{
			"Map requested requirements and acceptance criteria to changed files and observable behavior.",
			"Inspect validation commands, test output, logs, diffs, generated files, and documentation updates.",
			"Separate verified facts from assumptions, unrun checks, stale evidence, and missing evidence.",
			"Classify verification gaps by release impact, security impact, and regression likelihood.",
			"Return pass, conditional pass, or block with concrete remaining validation steps.",
		},
		ReviewScope: []string{
			"Requirement-to-implementation traceability.",
			"Acceptance criteria and test evidence credibility.",
			"Changed files, edge cases, scope expansion, and regression risk.",
			"Security validation, documentation consistency, and generated outputs.",
			"Residual risk and pass/fail recommendation.",
		},
		Checklist: []string{
			"Map every requested requirement to changed files, tests, docs, or explicit non-code evidence.",
			"Verify acceptance criteria against actual implementation behavior, not only author claims.",
			"Check validation commands were run in the relevant repo state and report failures honestly.",
			"Identify untested edge cases, negative paths, concurrency paths, migration paths, and rollback paths.",
			"Check changed files for unrelated edits, broad refactors, generated-output drift, or hidden scope expansion.",
			"Verify documentation, changelog, API docs, runbooks, or examples changed when behavior changed.",
			"Verify security-sensitive changes include auth, validation, logging, and secrets checks where relevant.",
			"Distinguish verified facts, inferred facts, assumptions, missing evidence, and stale evidence.",
			"Check generated files, lockfiles, schema files, and platform copies are synchronized with source files.",
			"Assess residual risks and whether they block merge, block release, or can be tracked.",
			"Review failing or skipped tests for relevance before accepting pass status.",
			"Confirm final recommendation matches the evidence level.",
		},
		DecisionRules: []string{
			"If acceptance criteria cannot be traced to implementation or validation evidence, do not mark the change verified.",
			"If tests failed, were skipped, or were not run, classify the result as conditional or blocked based on risk.",
			"If generated outputs or lockfiles are stale, require regeneration before pass.",
			"If implementation changes behavior without docs or changelog updates, flag verification incomplete for user-facing changes.",
			"If a security-sensitive path lacks negative validation, block or conditionally pass with explicit risk.",
			"If evidence conflicts, prefer repository state and command output over summaries.",
		},
		FindingCategories: []string{
			"Unverified or unmapped acceptance criterion.",
			"Missing, stale, failed, skipped, or irrelevant validation evidence.",
			"Scope creep, unrelated refactor, or unintended behavior change.",
			"Missing edge-case, regression, migration, rollback, or negative validation.",
			"Generated file, schema, lockfile, docs, or platform-copy drift.",
			"Unsupported pass recommendation or hidden residual risk.",
		},
		SeverityGuidance: []string{
			"Critical: claimed verification hides failed validation for security, data integrity, destructive, or production-critical behavior.",
			"High: must-have requirement, migration, contract, rollback, or security path lacks credible validation before release.",
			"Medium: important edge case, generated artifact, documentation, or regression evidence is incomplete but trackable.",
			"Low: evidence formatting, traceability clarity, or non-blocking verification detail can be improved.",
		},
		OutputRequirements: []string{
			"Verification table mapping requirement, evidence, status, and residual risk.",
			"Validation commands reviewed or run, with pass/fail/skipped state.",
			"Files and artifacts reviewed, including generated outputs and docs where relevant.",
			"Findings with severity, evidence, impact, and exact remediation.",
			"Final recommendation: pass, conditional pass, or block, with reasons.",
			"Open verification gaps with owner-oriented next steps.",
		},
		AcceptanceCriteria: []string{
			"All must-have requirements are traced to implementation and credible validation evidence.",
			"Validation results are current, honest, and tied to the reviewed repository state.",
			"Generated outputs, docs, schemas, and lockfiles are synchronized or called out.",
			"Security, migration, rollback, and compatibility risks have appropriate evidence.",
			"Residual risks are explicit and reflected in the pass/conditional/block recommendation.",
			"Assumptions are separated from verified facts.",
		},
		AntiPatterns: []string{
			"Accepting an author's summary instead of checking changed files and validation evidence.",
			"Treating green CI as sufficient when it does not cover the changed behavior.",
			"Ignoring skipped tests, flaky failures, stale generated files, or missing docs.",
			"Marking pass while hiding residual risks or unrun checks.",
			"Verifying only happy paths for security-sensitive or migration-heavy changes.",
			"Expanding review scope into unrelated refactoring advice without verification relevance.",
		},
	},
	"security-reviewer": {
		Purpose:            "Review code, design, configuration, and changes for authentication, authorization, input validation, output encoding, injection, SSRF, path traversal, deserialization, XSS, CSRF, secrets, logging, cryptography, errors, dependencies, and least privilege.",
		When:               []string{"Code or configuration changes touch trust boundaries, identity, permissions, inputs, files, URLs, serialization, logging, crypto, or dependencies.", "A review needs exploitability, impact, and concrete remediation, not generic security advice.", "AuthN, AuthZ, tenant isolation, or sensitive-data handling changed.", "User-controlled data reaches interpreters, file paths, network clients, templates, logs, or storage.", "The central agent routes to security review."},
		Operating:          []string{"Identify assets, actors, trust boundaries, entry points, and sensitive data flows.", "Trace attacker-controlled input to sinks such as SQL, shell, templates, file paths, URLs, logs, and deserializers.", "Review authorization checks at object, tenant, role, route, and service boundaries.", "Assess exploitability using reachable code paths, preconditions, privileges, and impact.", "Recommend minimal fixes with tests for bypass, invalid input, and unsafe defaults."},
		ReviewScope:        []string{"Authentication, authorization, tenant isolation, and permission boundaries.", "Injection, SSRF, path traversal, deserialization, XSS, CSRF, and unsafe file handling.", "Secrets exposure, unsafe logging, cryptography, error disclosure, and insecure defaults.", "Dependency and configuration security exposure.", "Exploitability, impact, remediation, and regression tests."},
		Checklist:          []string{"Check whether every sensitive operation enforces authentication and object-level authorization.", "Check tenant, workspace, project, organization, or account isolation on reads and writes.", "Trace user-controlled input into SQL, NoSQL, LDAP, shell, template, regex, expression, and query builders.", "Review path construction, archive extraction, uploads, downloads, and file deletion for traversal and unsafe access.", "Review outbound URL fetches, webhooks, redirects, metadata IP access, and DNS rebinding for SSRF/open redirect risk.", "Check output encoding, content type, CSP assumptions, XSS, CSRF, and browser trust boundaries.", "Check deserialization, parser, YAML/XML/entity, and polymorphic binding behavior.", "Check secrets handling, unsafe logging, error disclosure, and redaction boundaries.", "Check crypto choices, random generation, key storage, token expiry, and signature verification.", "Check insecure defaults, debug modes, broad CORS, permissive headers, and disabled TLS verification.", "Check dependency and container configuration for reachable vulnerable code paths.", "Require negative security tests for confirmed high-risk paths."},
		DecisionRules:      []string{"If authorization depends only on UI, route naming, or client-provided IDs, classify as AuthZ bypass risk.", "If untrusted input reaches an interpreter without parameterization or allowlisting, classify as injection risk.", "If server-side fetch accepts user-controlled URLs without strict allowlist and metadata blocking, classify as SSRF risk.", "If file paths combine user input with filesystem operations without canonicalization and root checks, classify as path traversal risk.", "If secrets or PII can enter logs or errors, require redaction and retention review.", "If a finding has a plausible exploit path and sensitive impact, do not downgrade because exploitation is inconvenient."},
		FindingCategories:  []string{"AuthZ bypass or tenant isolation failure.", "Injection into SQL, shell, template, expression, query, or command sinks.", "Path traversal, unsafe upload/download, archive extraction, or file deletion.", "SSRF, open redirect, webhook abuse, or unsafe outbound request.", "Secrets exposure, unsafe logging, PII leakage, or error disclosure.", "Insecure default, weak crypto, missing token validation, or excessive privilege."},
		SeverityGuidance:   []string{"Critical: unauthenticated or low-privilege attacker can access secrets, regulated data, admin actions, or RCE/destructive behavior.", "High: authenticated attacker can bypass authorization, exfiltrate sensitive data, inject commands, or pivot across tenants.", "Medium: exploit requires constraints but exposes meaningful data, integrity, availability, or defense-in-depth weakness.", "Low: hardening, logging clarity, header, or configuration improvement with limited direct exploitability."},
		OutputRequirements: []string{"Findings ordered by severity with affected asset, code path, trust boundary, and exploit scenario.", "Evidence references to files, routes, configs, sinks, and validation gaps.", "Concrete remediation with safer API, validation rule, permission check, or configuration change.", "Security tests or negative cases needed to prevent regression.", "Residual risk and assumptions, including unreachable or false-positive rationale.", "Clear pass, conditional pass, or block recommendation."},
		AcceptanceCriteria: []string{"AuthN, AuthZ, tenant isolation, and least privilege are explicitly assessed for sensitive operations.", "Untrusted input paths to dangerous sinks are traced and mitigated or documented as safe.", "Secrets, logs, errors, crypto, and defaults are reviewed where touched.", "Critical and High findings include concrete exploitability and remediation guidance.", "Security-sensitive changes include appropriate negative tests or validation steps.", "False positives are justified with repository evidence."},
		AntiPatterns:       []string{"Reporting generic OWASP advice without a reachable code path.", "Assuming middleware protects object-level authorization without checking resource ownership.", "Downgrading injection, SSRF, or traversal because input appears internal.", "Printing suspected secret values in findings.", "Accepting broad allowlists, wildcard permissions, or disabled verification as temporary convenience.", "Treating dependency CVSS as impact without reachability analysis."},
	},
	"secrets-reviewer": {
		Purpose:            "Identify and handle secrets, credentials, tokens, private keys, passwords, `.env` files, CI variables, config files, logs, fixtures, database dumps, rotation, revocation, scanning, and false positives.",
		When:               []string{"A change adds or modifies credentials, tokens, config, logs, CI variables, fixtures, dumps, or secret references.", "A suspected secret exposure needs triage without printing secret values.", "Secret storage, rotation, revocation, scope, or ownership is unclear.", "CI/CD, examples, docs, or tests may leak sensitive values.", "The central agent routes to secrets review."},
		Operating:          []string{"Inspect likely secret locations while redacting values in output.", "Classify each candidate as real secret, placeholder, test fixture, public identifier, or unknown.", "Assess exposure path: repository content, history, logs, artifacts, CI, docs, images, or dumps.", "Recommend containment: remove, rotate, revoke, scope down, and prevent recurrence.", "Report only fingerprints, paths, and safe excerpts."},
		ReviewScope:        []string{"API keys, tokens, private keys, passwords, `.env` files, and CI variables.", "Config files, logs, fixtures, artifacts, database dumps, and history exposure.", "Secret storage, rotation, revocation, scope, and owners.", "Secret scanning, false positives, prevention controls, and follow-up actions.", "Output redaction and safe handling of suspected values."},
		Checklist:          []string{"Search code, configs, docs, examples, tests, logs, artifacts, and CI files for high-entropy or credential-like values.", "Classify each candidate without echoing full secret values.", "Check whether placeholders use safe fake values and clear naming.", "Check CI/CD variables, workflow logs, debug output, and artifact upload paths.", "Check `.env`, sample env files, config maps, secrets manifests, and local setup docs.", "Check database dumps, fixtures, snapshots, screenshots, and generated files for sensitive data.", "Assess token scope, expiry, environment, owner, and blast radius when exposed.", "Recommend rotation and revocation for real or likely real secrets.", "Recommend secret-manager references, environment indirection, or scoped CI variables.", "Identify prevention controls such as scanning, pre-commit hooks, deny patterns, and log redaction.", "Call out history rewrite or artifact deletion when repository history or builds contain secrets.", "Document false-positive rationale safely."},
		DecisionRules:      []string{"If a value could authenticate to a real system, treat it as a secret until proven otherwise.", "If a secret reached git history, logs, artifacts, package images, or third-party systems, require rotation and revocation.", "If a token is broad-scope, long-lived, or production-scoped, raise severity.", "If examples need credentials, use obvious fake placeholders and setup instructions.", "If output would reveal a secret, redact all but a short fingerprint.", "If ownership is unclear, require owner identification before closure."},
		FindingCategories:  []string{"Hardcoded credential, token, private key, password, or connection string.", "Secret leakage through logs, artifacts, screenshots, fixtures, dumps, or generated files.", "Unsafe CI/CD secret scope, masking, debug output, or environment exposure.", "Missing rotation, revocation, owner, expiry, or scope reduction.", "False positive or placeholder requiring safe classification.", "Secret prevention gap in scanning, hooks, docs, or review process."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Redacted findings with path, line/context, secret type, confidence, and fingerprint.", "Exposure assessment covering repository, history, logs, artifacts, CI, docs, and packages.", "Rotation, revocation, removal, and prevention steps with owners.", "False-positive list with safe rationale.", "Validation commands or scanning evidence used.", "Residual risk and whether release or merge should be blocked."},
		AcceptanceCriteria: []string{"No full secret value is printed in output.", "Real or likely real secrets have rotation and revocation guidance.", "Exposure paths and blast radius are assessed.", "False positives are justified without unsafe disclosure.", "Prevention controls are recommended for recurring classes.", "Release recommendation reflects secret severity and containment state."},
		AntiPatterns:       []string{"Printing full credentials to prove a finding.", "Dismissing tokens as test data without evidence.", "Removing a secret from the current file while ignoring git history, logs, or artifacts.", "Rotating without revoking old credentials.", "Using production-looking examples in documentation.", "Treating masked CI variables as safe when debug output or artifacts expose them."},
	},
	"dependency-supply-chain-reviewer": {
		Purpose:            "Review dependencies, lockfiles, package managers, SBOMs, provenance, vulnerability reachability, license risk, transitive dependencies, update strategy, artifact integrity, and supply-chain controls.",
		When:               []string{"Dependencies, lockfiles, package registries, images, build tools, or update bots change.", "A vulnerability, license, provenance, or package integrity question must be triaged.", "Transitive dependency reachability or exploitability is unclear.", "SBOM, signing, checksums, or registry trust needs review.", "The central agent routes to dependency or supply-chain review."},
		Operating:          []string{"Identify ecosystems, manifests, lockfiles, registries, package managers, build steps, and artifacts.", "Compare manifest and lockfile changes for unexpected transitive movement.", "Assess vulnerability reachability, exploit preconditions, and available fixed versions.", "Review package provenance, signatures, checksums, registry source, and maintainer trust.", "Recommend upgrade, pin, replace, isolate, or risk-accept actions."},
		ReviewScope:        []string{"Manifests, lockfiles, container base images, build plugins, and generated dependency metadata.", "Direct and transitive dependencies, vulnerable code paths, exploitability, and fixed versions.", "License, provenance, signing, checksums, SBOM, registry, and artifact integrity.", "Update automation, pinning, version ranges, vendoring, and reproducibility.", "Dependency removal, replacement, or isolation options."},
		Checklist:          []string{"Check manifest and lockfile consistency.", "Identify newly added, removed, upgraded, downgraded, or transitive dependencies.", "Check whether vulnerable dependency code is reachable in this application.", "Check fixed versions, breaking changes, and upgrade notes.", "Check version ranges, floating tags, branch dependencies, and unpinned plugins.", "Check package source, registry, maintainer, download URL, and namespace confusion risk.", "Check license compatibility and policy exceptions.", "Check SBOM, signatures, checksums, provenance, and reproducible build evidence.", "Check container base images and OS packages for update hygiene.", "Check dependency update automation and review gates.", "Check whether dependency removal or replacement is safer than upgrade.", "Document residual risk and owner when risk is accepted."},
		DecisionRules:      []string{"If a vulnerable package is reachable and fix exists, require upgrade or compensating control.", "If a package is unpinned, registry-sourced unexpectedly, or from an unknown maintainer, flag supply-chain risk.", "If lockfile changes are unexplained, require review before merge.", "If license policy is violated, require legal or governance review.", "If artifact integrity cannot be verified for release input, block release readiness.", "If vulnerability is not reachable, document evidence instead of relying only on CVSS."},
		FindingCategories:  []string{"Reachable vulnerable dependency or base image.", "Suspicious, unpinned, typosquatted, abandoned, or provenance-weak package.", "Manifest/lockfile drift or unexpected transitive dependency change.", "License, policy, or attribution violation.", "Missing SBOM, signature, checksum, or artifact integrity evidence.", "Unsafe update automation, registry trust, or reproducibility gap."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Dependency change summary with direct/transitive and lockfile impact.", "Vulnerability table with reachability, fixed version, exploitability, and remediation.", "Supply-chain integrity findings covering source, signatures, checksums, SBOM, and registry trust.", "License and policy findings with required owner review.", "Upgrade, pin, replace, remove, or risk-accept recommendation.", "Validation commands and residual risk."},
		AcceptanceCriteria: []string{"Manifest and lockfile are consistent.", "Reachable vulnerabilities have fixes or documented compensating controls.", "Unpinned or suspicious dependencies are resolved or risk-accepted.", "License and policy risks are reviewed.", "SBOM/provenance/integrity expectations are met for release artifacts.", "Residual risk includes owner and deadline."},
		AntiPatterns:       []string{"Triage by CVSS only without reachability.", "Accepting broad version ranges for production-critical packages.", "Ignoring lockfile diffs because manifest diff is small.", "Trusting package names without checking source or maintainer.", "Treating SBOM generation as artifact integrity.", "Upgrading major versions without compatibility validation."},
	},
	"ci-cd-reviewer": {
		Purpose:            "Review CI/CD pipelines, workflow permissions, secrets handling, runner trust, cache safety, artifact integrity, deployment gates, environments, approvals, provenance, and release automation.",
		When:               []string{"Workflow, pipeline, deployment, release, or runner configuration changes.", "Tokens, permissions, caches, artifacts, or secrets are used in automation.", "A pipeline runs on forks, untrusted branches, self-hosted runners, or privileged environments.", "Deployment gates, approvals, or environment protections need review.", "The central agent routes to CI/CD review."},
		Operating:          []string{"Map triggers, actors, permissions, runners, secrets, caches, artifacts, and deployment targets.", "Trace untrusted inputs from branch names, PR metadata, matrix values, and scripts into shell or actions.", "Review token scopes, job permissions, environment protections, and approval gates.", "Assess artifact and cache integrity across job boundaries.", "Recommend least-privilege pipeline changes with validation steps."},
		ReviewScope:        []string{"Workflow triggers, branch/tag filters, and fork behavior.", "Token permissions, OIDC, secrets, environments, approvals, and deployment gates.", "Runner trust, self-hosted runner exposure, privileged containers, and network access.", "Cache keys, artifact upload/download, provenance, signatures, and checksums.", "Shell injection, third-party actions, pinned versions, and release automation."},
		Checklist:          []string{"Check workflow triggers for pull_request_target, forks, tags, schedules, and manual dispatch risk.", "Check job permissions and default token scopes for least privilege.", "Check secrets availability by branch, environment, fork, and job boundary.", "Check shell commands for untrusted PR, branch, tag, matrix, or commit data.", "Check third-party actions, includes, templates, and images are pinned or trusted.", "Check cache keys for poisoning, privilege boundary crossing, and restore-key abuse.", "Check artifacts are integrity-protected before downstream use or deployment.", "Check self-hosted runner labels, isolation, cleanup, and access to secrets.", "Check deployment environments require approvals, protected branches, and rollback gates.", "Check release publishing requires provenance, signing, changelog, and explicit version inputs.", "Check logs do not reveal secrets or tokens.", "Check failed-job reruns cannot escalate privileges unexpectedly."},
		DecisionRules:      []string{"If untrusted code can access secrets or write tokens, classify as blocking risk.", "If deployment happens without environment approval or protected ref policy, require a gate.", "If artifacts cross trust boundaries without integrity verification, flag artifact tampering risk.", "If caches are shared between untrusted and trusted jobs, flag cache poisoning risk.", "If shell commands include untrusted context without quoting or allowlisting, flag injection risk.", "If third-party actions are unpinned, require pinning or risk acceptance."},
		FindingCategories:  []string{"Excessive token permissions or missing least privilege.", "Secret exposure across fork, branch, job, log, or environment boundary.", "Script injection from untrusted CI context.", "Cache poisoning or artifact integrity failure.", "Untrusted or overprivileged runner execution.", "Missing deployment gate, environment approval, provenance, or rollback control."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Pipeline risk summary with trigger, actor, token, runner, secret, artifact, and deployment boundaries.", "Findings with workflow file, job, permission, evidence, impact, and remediation.", "Recommended permission, trigger, secret, cache, artifact, or environment changes.", "Validation commands or CI checks to prove the fix.", "Release/deployment readiness recommendation.", "Residual risks and required owner approvals."},
		AcceptanceCriteria: []string{"Token permissions are least privilege per job.", "Secrets are unavailable to untrusted code paths.", "Artifacts and caches do not cross trust boundaries without integrity controls.", "Deployment jobs have protected refs, environment gates, and rollback path.", "Third-party actions/images are pinned or trusted.", "Injection risks from CI context are mitigated."},
		AntiPatterns:       []string{"Using repository-wide write tokens by default.", "Running untrusted fork code with secrets.", "Trusting artifacts from earlier jobs without checksums or provenance.", "Sharing cache keys between trusted and untrusted jobs.", "Deploying directly from a build job with no approval gate.", "Pinning by branch name instead of immutable version or digest."},
	},
	"iac-gitops-reviewer": {
		Purpose:            "Review Terraform, Kubernetes, Helm, Kustomize, GitOps, policies, cloud IAM, network exposure, drift, secrets, state, promotion, rollback, and environment safety.",
		When:               []string{"Infrastructure, Kubernetes, Helm, Terraform, GitOps, IAM, networking, or policy files change.", "A change affects production environments, cluster policy, secrets, state, or promotion flow.", "Drift, rollback, plan/apply safety, or GitOps reconciliation behavior is unclear.", "Cloud permissions, public exposure, or workload security needs review.", "The central agent routes to IaC/GitOps review."},
		Operating:          []string{"Map resources, environments, state backends, namespaces, IAM roles, networks, and reconciliation controllers.", "Review plan/apply or diff semantics, not only YAML shape.", "Assess least privilege, network exposure, secret references, and workload security context.", "Check promotion, drift, rollback, and blast radius across environments.", "Recommend minimal policy or manifest changes with validation commands."},
		ReviewScope:        []string{"Terraform/OpenTofu, Kubernetes, Helm, Kustomize, Argo CD, Flux, and policy-as-code.", "IAM, RBAC, service accounts, network rules, ingress, egress, storage, and secrets.", "State backend, drift detection, plan/apply safety, promotion, and rollback.", "Workload security context, resource limits, probes, disruption budgets, and scheduling.", "Environment overlays and generated manifests."},
		Checklist:          []string{"Check Terraform plan or manifest diff for created, changed, replaced, or destroyed resources.", "Check IAM/RBAC for wildcard actions, broad resources, admin roles, and privilege escalation.", "Check network exposure, ingress, public IPs, security groups, and egress controls.", "Check secret handling via secret managers, sealed secrets, external secrets, or unsafe literals.", "Check state backend encryption, locking, access, and workspace/environment separation.", "Check Kubernetes security context, root containers, capabilities, hostPath, privileged mode, and service accounts.", "Check resource requests/limits, probes, PDBs, rollout strategy, and autoscaling.", "Check GitOps sync waves, pruning, drift behavior, promotion flow, and manual override risks.", "Check environment overlays for prod/dev value bleed or missing policy constraints.", "Check rollback and destroy safety for stateful resources.", "Check policy exceptions and approvals.", "Check generated manifests are consistent with source charts or kustomizations."},
		DecisionRules:      []string{"If a change destroys or replaces stateful production resources, block unless migration and rollback are approved.", "If IAM/RBAC uses wildcard privilege without bounded scope, classify as High or Critical by environment.", "If public network exposure reaches sensitive services, require explicit justification and controls.", "If secrets are stored as plaintext in IaC, require secret-management remediation.", "If GitOps pruning or sync can remove live resources unexpectedly, require rollout guardrails.", "If no plan/diff evidence exists for risky IaC, do not approve release readiness."},
		FindingCategories:  []string{"Destructive or unsafe infrastructure change.", "Overbroad IAM/RBAC, service account, or privilege escalation path.", "Public exposure, insecure ingress/egress, or network segmentation gap.", "Plaintext secret, unsafe state backend, or environment separation failure.", "GitOps drift, pruning, promotion, or rollback risk.", "Workload hardening, resource, probe, or availability gap."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Resource and environment diff summary.", "Findings with resource, environment, evidence, blast radius, and remediation.", "Plan/apply, kubectl diff, helm template, policy, or GitOps validation evidence.", "IAM/RBAC, network, secret, state, and workload security review result.", "Rollback, migration, and promotion recommendations.", "Pass, conditional pass, or block decision by environment."},
		AcceptanceCriteria: []string{"Risky IaC changes have plan/diff evidence.", "Secrets, IAM/RBAC, networking, and state backend are reviewed.", "Production-impacting changes include rollback or migration plan.", "GitOps reconciliation behavior is understood.", "Generated manifests match source definitions.", "Policy exceptions have owner and expiry."},
		AntiPatterns:       []string{"Approving IaC by reading only filenames.", "Ignoring Terraform replacement markers or Kubernetes prune behavior.", "Using admin roles because least privilege is tedious.", "Putting secrets directly in values files or manifests.", "Assuming dev overlay safety applies to production.", "Skipping rollback review for stateful resources."},
	},
	"compliance-governance-reviewer": {
		Purpose:            "Review governance, policy, auditability, compliance evidence, control mapping, approvals, audit trails, ownership, segregation of duties, policy exceptions, retention, access reviews, change management, and risk acceptance.",
		When:               []string{"A change affects controls, approvals, audit evidence, access, retention, policy exceptions, or regulated workflows.", "An auditor, compliance owner, or reviewer needs evidence mapped to controls and owners.", "Risk acceptance, exception expiry, segregation of duties, or approval authority is unclear.", "Generated artifacts, logs, tickets, or repository settings must support audit readiness.", "The central agent routes to compliance or governance review."},
		Operating:          []string{"Identify applicable policies, controls, repositories, systems, owners, approvers, and evidence sources.", "Map repository evidence to control objectives without inventing compliance claims.", "Review approval authority, segregation of duties, risk acceptance, exception expiry, and audit trail completeness.", "Assess evidence quality: timestamp, actor, immutable source, linkage, and retention.", "Recommend control remediation, evidence collection, or governance decision with owner and deadline."},
		ReviewScope:        []string{"Control mapping, approvals, audit trails, evidence, and ownership.", "Segregation of duties, access reviews, policy exceptions, and expiry.", "Retention, change management, risk acceptance, and accountability.", "Branch protection, CODEOWNERS, review gates, and governance settings.", "Audit-ready evidence and compliance gaps."},
		Checklist:          []string{"Map each relevant change to policy, control objective, framework requirement, or documented governance rule.", "Check approvals are from authorized owners and are visible in merge requests, tickets, or audit records.", "Check segregation of duties between author, approver, deployer, and risk accepter.", "Check CODEOWNERS, branch protection, required reviews, status checks, and bypass permissions.", "Check risk acceptance records include scope, rationale, owner, expiry date, and compensating controls.", "Check policy exceptions are time-bound and linked to remediation work.", "Check evidence includes timestamp, actor, system of record, artifact link, and tamper-resistant source where needed.", "Check access reviews, role changes, service accounts, and privileged permissions for owner approval.", "Check retention and deletion requirements for logs, audit records, customer data, and generated artifacts.", "Check change-management linkage between ticket, commit, review, deployment, and release evidence.", "Check audit trails for administrative actions, approval changes, and production-impacting events.", "Check compliance documentation stays synchronized with generated platform files and governance docs."},
		DecisionRules:      []string{"If approval authority is missing or approver is also the sole implementer for regulated control, flag segregation-of-duties risk.", "If evidence cannot be tied to actor, timestamp, artifact, and system of record, treat it as weak audit evidence.", "If a policy exception has no owner or expiry, require remediation before calling it accepted risk.", "If repository settings allow bypass of required governance checks, classify by production and data impact.", "If retention or deletion behavior changes without policy mapping, require compliance owner review.", "If control claims are not supported by repository evidence, report them as unverified rather than compliant."},
		FindingCategories:  []string{"Missing or unauthorized approval evidence.", "Segregation-of-duties, bypass permission, or required-review failure.", "Weak audit evidence, missing system of record, or broken traceability.", "Risk acceptance or policy exception without owner, scope, expiry, or compensating control.", "Retention, access review, privileged account, or audit-log gap.", "Control mapping, change-management, or governance documentation mismatch."},
		SeverityGuidance:   []string{"Critical: control failure creates immediate regulatory breach, unauthorized production access, audit falsification, or unrecoverable evidence loss.", "High: missing approval, audit trail, segregation, retention, or risk acceptance can block release or audit readiness.", "Medium: control mapping, ownership, evidence quality, or policy-exception gap should be tracked and remediated.", "Low: formatting, traceability, naming, or evidence-packaging improvement with limited compliance impact."},
		OutputRequirements: []string{"Control evidence table with control, artifact, actor, timestamp, owner, and status.", "Approval and segregation-of-duties findings with repository or ticket evidence.", "Risk acceptance and policy exception summary with scope, expiry, and compensating controls.", "Governance settings reviewed, including CODEOWNERS, branch protection, required checks, and bypasses.", "Retention, access-review, audit-log, and change-management gaps.", "Pass, conditional pass, or block recommendation for audit/release readiness."},
		AcceptanceCriteria: []string{"Every compliance claim is backed by named repository, ticket, log, or governance evidence.", "Required approvals are present and authorized.", "Segregation-of-duties and bypass risks are assessed.", "Risk acceptances and policy exceptions include owner, scope, expiry, and compensating controls.", "Retention, access review, audit trail, and change-management requirements are addressed.", "Unverified controls are explicitly marked and not reported as compliant."},
		AntiPatterns:       []string{"Claiming compliance because a checklist exists without audit evidence.", "Accepting self-approval for regulated production changes without exception.", "Leaving policy exceptions open-ended.", "Using screenshots or chat messages as sole evidence when a system of record exists.", "Ignoring repository bypass permissions and admin overrides.", "Mixing desired governance state with verified current state."},
	},
	"release-readiness-reviewer": {
		Purpose:            "Determine whether a change or system is ready for release by reviewing test status, security findings, known issues, rollback, migrations, monitoring, alerts, runbooks, feature flags, approvals, release notes, support readiness, and go/no-go recommendation.",
		When:               []string{"A release, deployment, version bump, migration, or production rollout needs go/no-go review.", "Known issues, security findings, rollback, monitoring, or support readiness are unclear.", "Feature flags, staged rollout, or compatibility risk must be assessed.", "Release notes, approvals, or operational ownership need validation.", "The central agent routes to release readiness review."},
		Operating:          []string{"Collect release scope, changed artifacts, validation status, deployment plan, and owners.", "Evaluate blockers across tests, security, migrations, rollback, monitoring, docs, support, and approvals.", "Separate go/no-go criteria from follow-up work and known accepted risks.", "Assess rollout strategy, feature flags, blast radius, and recovery time.", "Return go, conditional go, or no-go with explicit blockers."},
		ReviewScope:        []string{"Test status, validation evidence, security findings, and known issues.", "Rollback, migrations, compatibility, feature flags, and staged rollout.", "Monitoring, alerts, runbooks, support readiness, and ownership.", "Approvals, release notes, communication, and go/no-go decision.", "Blockers, exceptions, residual risk, and follow-up actions."},
		Checklist:          []string{"Verify required tests, builds, scans, migrations, and smoke checks are complete.", "Identify unresolved blockers and classify known issues by user impact.", "Check open security findings, exceptions, owners, and expiry.", "Verify rollback plan, rollback trigger, rollback owner, and data rollback constraints.", "Check migration forward/backward compatibility, idempotency, and backup plan.", "Verify monitoring dashboards, alerts, SLO indicators, and post-deploy validation.", "Check runbooks, escalation contacts, on-call coverage, and support readiness.", "Check release notes, changelog, customer communication, and breaking-change guidance.", "Identify feature flag, canary, staged rollout, kill switch, or traffic-shaping options.", "Confirm approvals, release owner, deployment window, and freeze constraints.", "Check dependencies on external services, infra capacity, and version compatibility.", "Produce go/no-go with explicit conditions."},
		DecisionRules:      []string{"If rollback is impossible or untested for high-impact change, no-go unless risk is accepted by owner.", "If Critical/High security findings are open without approved exception, no-go.", "If migrations can corrupt or lose data without backup and validation, no-go.", "If monitoring cannot detect release failure, require conditional go or no-go by impact.", "If known issues affect core user journeys, require mitigation, communication, or staged rollout.", "If release notes omit breaking changes or migrations, block external release readiness."},
		FindingCategories:  []string{"Missing go/no-go evidence or owner.", "Failed, skipped, stale, or insufficient validation gate.", "Open security, compliance, privacy, or known-issue blocker.", "Rollback, migration, compatibility, or data-safety gap.", "Monitoring, alerting, runbook, on-call, or support readiness gap.", "Release notes, communication, approval, or change-management gap."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Go/no-go summary with blockers, conditions, and owner.", "Release checklist covering tests, security, migrations, rollback, monitoring, docs, support, and approvals.", "Known issues table with severity, impact, mitigation, and acceptance owner.", "Rollback and post-deploy validation plan.", "Feature flag, canary, staged rollout, or kill-switch recommendation.", "Residual risks and follow-up actions with deadlines."},
		AcceptanceCriteria: []string{"All release blockers are resolved or explicitly accepted.", "Rollback, migration, monitoring, and support readiness are verified.", "Security findings and known issues have owner-approved disposition.", "Release notes cover breaking changes, migrations, and operational impact.", "Go/no-go recommendation follows evidence.", "Post-release validation and escalation path are defined."},
		AntiPatterns:       []string{"Treating green CI as complete release readiness.", "Approving release with no rollback trigger or owner.", "Ignoring known issues because they are documented elsewhere.", "Skipping monitoring and support readiness until after deployment.", "Accepting security findings without expiry and owner.", "Publishing breaking changes without migration guidance."},
	},
	"observability-reviewer": {
		Purpose:            "Review logging, metrics, tracing, alerting, dashboards, operational visibility, SLOs, SLIs, runbooks, audit logs, sensitive data in logs, on-call usability, and incident detection.",
		When:               []string{"A service, feature, or deployment changes behavior that operators must detect, debug, or support.", "Logs, metrics, traces, alerts, dashboards, or runbooks are added or changed.", "SLOs, SLIs, audit logs, or incident detection coverage is unclear.", "Sensitive data may enter logs or telemetry.", "The central agent routes to observability review."},
		Operating:          []string{"Map critical user journeys, failure modes, dependencies, and operational questions.", "Check whether logs, metrics, traces, alerts, and dashboards answer those questions.", "Review signal quality: labels, cardinality, correlation IDs, thresholds, routing, and ownership.", "Assess privacy and security of telemetry.", "Recommend concrete telemetry, alert, dashboard, or runbook changes."},
		ReviewScope:        []string{"Logs, metrics, traces, alerts, dashboards, and runbooks.", "SLOs, SLIs, critical journeys, audit logs, and incident detection.", "Sensitive data in logs, correlation IDs, retention, and cardinality.", "On-call usability, alert routing, alert fatigue, and ownership.", "Operational decision support and recovery verification."},
		Checklist:          []string{"Identify critical user journeys and failure modes requiring visibility.", "Check error, latency, throughput, saturation, dependency, and queue metrics.", "Check logs include event names, correlation IDs, actor/resource IDs, and safe context.", "Check traces connect ingress, service calls, database, queues, and external dependencies.", "Check alerts have actionable thresholds, severity, runbook, owner, and routing.", "Check dashboards answer deploy health, customer impact, dependency health, and rollback decisions.", "Check sensitive data, secrets, tokens, PII, and payloads are excluded or redacted.", "Check metric label cardinality, retention, cost, and aggregation safety.", "Check SLO/SLI coverage for critical journeys.", "Check audit logs for security-relevant actions and tamper-resistant retention.", "Check runbooks include diagnosis, mitigation, rollback, and verification steps.", "Identify noisy, duplicate, missing, or unactionable alerts."},
		DecisionRules:      []string{"If operators cannot detect failure of a critical journey, require metrics or alerts before release.", "If logs can expose secrets or PII, require redaction before approval.", "If an alert lacks owner, severity, runbook, or action, classify as alert-quality gap.", "If dashboards cannot support rollback decisions, require deployment health panels.", "If metric cardinality can explode from user input, require label redesign.", "If audit-relevant actions lack logs, raise security/compliance severity."},
		FindingCategories:  []string{"Missing signal for critical user journey or dependency failure.", "Unsafe logging of secrets, PII, tokens, or sensitive payloads.", "Unactionable, noisy, duplicate, or ownerless alert.", "Dashboard gap for deploy health, customer impact, or rollback decision.", "Trace, correlation ID, or context propagation gap.", "SLO/SLI, audit log, retention, or cardinality risk."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Observability coverage map for journeys, failure modes, signals, alerts, dashboards, and runbooks.", "Findings with affected signal, evidence, operator impact, and remediation.", "Concrete metric, log, trace, alert, dashboard, or runbook recommendation.", "Sensitive telemetry and cardinality risk assessment.", "Post-deploy monitoring and incident-detection recommendation.", "Residual blind spots and owner actions."},
		AcceptanceCriteria: []string{"Critical journeys have metrics, logs/traces, alerts, and dashboard support.", "Telemetry excludes secrets and sensitive payloads.", "Alerts are actionable, owned, routed, and tied to runbooks.", "Dashboards support diagnosis and rollback decisions.", "SLOs/SLIs or audit logs exist where required.", "Known blind spots are explicitly documented."},
		AntiPatterns:       []string{"Adding logs without deciding what operator question they answer.", "Alerting on every error without severity, owner, or action.", "Using high-cardinality user input as metric labels.", "Logging full payloads to debug production issues.", "Building dashboards that cannot support rollback or incident triage.", "Assuming tracing solves missing metrics or alerting."},
	},
	"incident-postmortem-assistant": {
		Purpose:            "Support incident response, postmortems, and corrective actions across triage, severity, impact, timeline, containment, eradication, recovery, communication, evidence preservation, root cause, contributing factors, corrective actions, and prevention.",
		When:               []string{"An active incident needs structured triage, timeline, impact, containment, or communication support.", "A postmortem needs facts, contributing factors, root cause, and corrective actions.", "Follow-up actions must be owner-assigned and verifiable.", "Stakeholder communication must separate facts from assumptions.", "The central agent routes to incident or postmortem assistance."},
		Operating:          []string{"Separate confirmed facts, assumptions, hypotheses, unknowns, and decisions.", "Build timeline from alerts, logs, deploys, tickets, chats, and customer impact.", "Classify severity, affected services, customer/business/security impact, and current state.", "Guide containment, recovery, and evidence preservation without destructive shortcuts.", "Produce blameless postmortem and corrective actions with owners and due dates."},
		ReviewScope:        []string{"Triage, severity, impact, timeline, containment, eradication, and recovery.", "Communication, evidence preservation, facts, assumptions, and unknowns.", "Root cause, contributing factors, corrective actions, prevention, and owners.", "Security, customer, business, compliance, and operational impact.", "Postmortem readiness and follow-up issue quality."},
		Checklist:          []string{"Separate confirmed facts, assumptions, hypotheses, and unknowns.", "Establish timeline with timestamps, sources, and confidence level.", "Identify customer, business, security, compliance, and operational impact.", "Classify severity and document severity changes over time.", "Recommend containment steps that preserve evidence and avoid unsafe recovery.", "Track detection, mitigation, recovery, and resolution times.", "Identify immediate cause, contributing factors, and systemic causes.", "Distinguish mitigation, eradication, recovery, prevention, and follow-up work.", "Draft stakeholder communication using confirmed facts only.", "Assign corrective actions with owner, deadline, and verification criterion.", "Identify monitoring, runbook, test, process, or architecture gaps that allowed recurrence.", "Verify recovery using telemetry, customer impact, and service health checks."},
		DecisionRules:      []string{"If facts are incomplete, mark them unknown instead of inventing root cause.", "If incident may involve security compromise, preserve evidence before cleanup.", "If containment may worsen impact, state trade-offs and seek owner decision.", "If corrective action lacks owner, due date, and verification, it is not postmortem-ready.", "If customer impact is unknown, require impact assessment before final severity.", "If communication includes hypotheses, label them clearly or remove them from external messaging."},
		FindingCategories:  []string{"Incomplete timeline, missing fact source, or assumption presented as fact.", "Unclear severity, impact, affected service, or customer scope.", "Unsafe containment, recovery, or evidence-destroying action.", "Weak root cause analysis or missing contributing factors.", "Corrective action without owner, deadline, or verification criterion.", "Missing communication, monitoring, runbook, test, or prevention follow-up."},
		SeverityGuidance:   []string{"Critical: immediate exploitability or operational failure can expose secrets, regulated data, production safety, or release integrity.", "High: credible security, reliability, compliance, rollback, or user-impact risk requires owner action before merge or release.", "Medium: meaningful maintainability, validation, documentation, or process gap should be tracked and resolved.", "Low: advisory improvement, clarity issue, or hardening opportunity with limited immediate impact."},
		OutputRequirements: []string{"Incident summary with severity, impact, affected systems, status, and confidence level.", "Timeline with timestamped events, sources, and gaps.", "Facts, assumptions, hypotheses, unknowns, and decisions separated.", "Containment, recovery, and evidence-preservation recommendations.", "Root cause and contributing factors with supporting evidence.", "Corrective action table with owner, due date, verification, and priority."},
		AcceptanceCriteria: []string{"Timeline covers detection, mitigation, recovery, and resolution with sources.", "Impact and severity are justified and updated when evidence changes.", "Root cause analysis distinguishes immediate cause from contributing factors.", "Corrective actions are specific, owned, dated, and verifiable.", "Communications avoid speculation and expose open questions.", "Recovery is verified by telemetry or customer-impact evidence."},
		AntiPatterns:       []string{"Declaring root cause before facts support it.", "Blaming individuals instead of analyzing systems and contributing factors.", "Deleting logs or changing systems before preserving evidence.", "Writing corrective actions like “be more careful”.", "Omitting customer impact because service health recovered.", "Closing postmortem without verifying follow-up completion criteria."},
	},
	"documentation-maintainer": {
		Purpose: "Keep technical and operational documentation accurate and useful across README, architecture docs, runbooks, API docs, changelogs, setup instructions, configuration docs, examples, troubleshooting, ownership, freshness, and consistency.",
		When: []string{
			"A code, configuration, API, CLI, deployment, or workflow change requires documentation updates.",
			"README files, ADRs, runbooks, setup guides, API docs, examples, or changelogs may be stale.",
			"Users or operators need accurate install, upgrade, rollback, troubleshooting, or support instructions.",
			"Generated docs or platform-specific copies must stay aligned with canonical documentation.",
			"The central agent routes documentation freshness, completeness, or consistency work to this skill.",
		},
		Operating: []string{
			"Identify the changed behavior, interface, setup step, operational procedure, or decision that documentation must describe.",
			"Map the change to concrete documentation artifacts: README, ADR, runbook, API reference, setup guide, example, changelog, or release note.",
			"Compare documentation claims against repository evidence such as commands, flags, config keys, API schemas, workflow files, and generated outputs.",
			"Classify stale, missing, contradictory, or unsafe documentation by user impact and operational risk.",
			"Recommend exact documentation edits, owners, and validation commands instead of broad documentation advice.",
		},
		ReviewScope: []string{
			"README, architecture docs, ADRs, API docs, setup guides, and examples.",
			"Runbooks, troubleshooting, ownership, support contacts, and operational docs.",
			"Changelogs, release notes, freshness, consistency, and source-of-truth rules.",
			"Configuration docs, CLI docs, platform docs, and generated outputs.",
			"Secret-safe documentation and actionable task orientation.",
		},
		Checklist: []string{
			"Check whether README quickstart, install, build, test, and usage commands match current repository behavior.",
			"Check CLI docs for added, removed, renamed, or changed flags, defaults, examples, and exit behavior.",
			"Check API documentation against route definitions, request and response schemas, auth requirements, error codes, and versioning notes.",
			"Check ADRs for architecture decisions that changed module boundaries, data ownership, interfaces, dependencies, or deployment topology.",
			"Check runbooks for current service names, dashboards, alerts, escalation paths, validation steps, rollback steps, and recovery commands.",
			"Check setup guides for required tools, environment variables, credentials handling, seed data, local services, and troubleshooting notes.",
			"Check examples and snippets by comparing paths, package names, config keys, commands, and expected output to the repository.",
			"Check changelog and release notes for user-visible behavior, breaking changes, migrations, deprecations, and operational changes.",
			"Check generated documentation and platform copies against the canonical source-of-truth instructions.",
			"Check ownership metadata, support contacts, codeowners, and escalation references for stale teams or links.",
			"Check docs for secret leakage, unsafe credential examples, production URLs, tokens, private keys, or misleading placeholder values.",
			"Check diagrams, tables, screenshots, and links for outdated references, broken anchors, missing alt text, or inaccessible context.",
		},
		DecisionRules: []string{
			"If a command, flag, config key, API field, route, or workflow changed, require the matching README, setup, CLI, API, or runbook update.",
			"If documentation contradicts code or configuration, treat repository behavior as evidence and flag the stale doc claim.",
			"If an operational procedure changed, require runbook validation, rollback guidance, and escalation ownership before release readiness.",
			"If a behavior change affects users, require changelog or release-note coverage with migration or breaking-change notes where applicable.",
			"If an architecture decision changes boundaries or trade-offs, require an ADR update or an explicit note that no ADR is needed.",
			"If documentation validation cannot be run, state the unvalidated artifact and the exact command or manual check still needed.",
		},
		FindingCategories: []string{
			"Stale README, quickstart, install, build, test, or usage instruction.",
			"Incorrect CLI, API, schema, configuration, or example documentation.",
			"Missing ADR, architecture note, migration guide, or decision rationale.",
			"Stale runbook, troubleshooting, rollback, dashboard, alert, or escalation procedure.",
			"Missing changelog, release note, deprecation, breaking-change, or upgrade guidance.",
			"Broken link, outdated diagram, inaccessible screenshot, or inconsistent generated documentation copy.",
			"Unsafe documentation that exposes secrets, encourages insecure credential handling, or misstates production risk.",
		},
		SeverityGuidance: []string{
			"Critical: documentation exposes secrets or gives instructions likely to cause production outage, data loss, or unsafe rollback.",
			"High: missing or wrong runbook, migration, breaking-change, API, setup, or credential guidance blocks safe release or operation.",
			"Medium: stale README, ADR, example, config, changelog, or ownership detail is likely to mislead users or maintainers.",
			"Low: wording, formatting, link text, diagram freshness, or clarity issue with limited immediate operational impact.",
		},
		OutputRequirements: []string{
			"List each documentation artifact reviewed: README, ADR, runbook, API doc, setup guide, example, changelog, or generated copy.",
			"For each finding, include severity, affected artifact, repository evidence, user or operator impact, and exact suggested edit.",
			"Report validation performed, such as command checks, link checks, schema comparisons, or manual artifact review.",
			"Identify missing documentation artifacts and the owner or decision needed to create them.",
			"Separate confirmed stale documentation from assumptions, open questions, and validation gaps.",
			"End with a documentation readiness recommendation: pass, conditional pass, or block release until docs are fixed.",
		},
		AcceptanceCriteria: []string{
			"README quickstart, setup, build, test, and usage instructions match current commands, paths, and prerequisites.",
			"ADRs or architecture docs capture changed decisions, trade-offs, boundaries, ownership, and alternatives when architecture changes.",
			"Runbooks include current alerts, dashboards, escalation contacts, diagnosis steps, rollback steps, and recovery validation.",
			"API docs describe current routes, auth, request and response schemas, error cases, examples, and compatibility notes.",
			"Setup guides document required tools, environment variables, local services, credentials handling, seed data, and common failures.",
			"Changelog and release notes cover user-visible changes, breaking changes, migrations, deprecations, and operational impact.",
			"Generated or platform-specific documentation copies are synchronized with the canonical source or explicitly marked generated.",
		},
		AntiPatterns: []string{
			"Updating only the README while leaving runbooks, API docs, setup guides, examples, or changelog stale.",
			"Copying command examples without checking flags, paths, environment variables, or expected output.",
			"Documenting secrets, real credentials, production tokens, or unsafe credential handling examples.",
			"Treating generated documentation copies as canonical without source-of-truth guidance.",
			"Writing vague release notes that omit breaking changes, migrations, deprecations, or operational impact.",
			"Leaving ADRs unchanged after architecture, ownership, boundary, or trade-off decisions change.",
			"Approving documentation freshness without naming artifacts reviewed and validation gaps.",
		},
	},
	"universal-skill-creator": {
		Purpose:            "Create new production-ready skills and prevent generic copy-paste skills by enforcing full frontmatter, SemVer, dates, authors, stability, min_platform_version, changelog, domain-specific scope, checklist, decision rules, finding categories, severity guidance, outputs, acceptance criteria, anti-patterns, and no generic body reuse.",
		When:               []string{"A user asks to create or upgrade a skill.", "A skill body must be checked for generator smell or copy-paste generic content.", "Skill metadata, versioning, compatibility, or changelog rules need enforcement.", "A skill needs domain-specific review logic, not only name and description changes.", "The central agent routes to universal skill creation."},
		Operating:          []string{"Identify the skill domain, users, trigger conditions, outputs, risks, and non-goals.", "Write domain-specific review scope, checklist, decisions, categories, severity, outputs, acceptance, and anti-patterns.", "Reject bodies that only differ by name, purpose, or generic operating text.", "Ensure frontmatter, body changelog, compatibility metadata, and versioning are consistent.", "Validate generated examples or tests that prove the skill is not generic."},
		ReviewScope:        []string{"YAML frontmatter, SemVer, since, last_modified, authors, stability, and min_platform_version.", "Body changelog, purpose, review scope, checklist, decision rules, and finding categories.", "Severity guidance, output requirements, acceptance criteria, and anti-patterns.", "Generic body reuse detection and platform compatibility honesty.", "Skill routing, generated copies, validation, and governance preservation."},
		Checklist:          []string{"Validate full YAML frontmatter and required metadata fields.", "Synchronize frontmatter changelog and body changelog version/date/message.", "Require min_platform_version entries for all supported platforms and mark unvalidated platforms honestly.", "Write a purpose that names the domain and concrete work products.", "Write When-to-use triggers that are specific enough for routing decisions.", "Write at least 10 checklist items that mention domain artifacts, risks, and evidence.", "Write at least 5 decision rules that decide real domain trade-offs.", "Write finding categories that are domain failure types, not generic evidence/control phrases.", "Write severity guidance with domain-specific Critical, High, Medium, and Low criteria.", "Write output requirements naming concrete artifacts the agent must produce.", "Write acceptance criteria that are testable for the domain.", "Write anti-patterns that describe misuse of this exact skill.", "Reject “structured analysis or review”, “<skill> evidence”, and “<skill> control” boilerplate.", "Verify generated platform copies use the shared renderer and stay synchronized."},
		DecisionRules:      []string{"Never create a skill that only differs by name and description. Every generated skill must include domain-specific review scope, checklist items, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns. Generic operating-model text is allowed only as shared baseline, never as the complete skill body.", "If a checklist item could apply unchanged to most skills, rewrite it with domain artifacts and failure modes.", "If a finding category contains the skill name plus “evidence” or “control”, reject it as generator smell.", "If severity guidance does not say what is Critical/High/Medium/Low in this domain, reject production readiness.", "If compatibility versions are concrete without validation evidence, use unknown or mark compatibility unverified.", "If frontmatter and body changelogs disagree, block the skill."},
		FindingCategories:  []string{"Generic copy-paste body or name-only variation.", "Missing or inconsistent versioning, changelog, or compatibility metadata.", "Non-domain checklist, decision rule, finding category, severity guidance, output, or acceptance criterion.", "Missing trigger clarity or routing ambiguity.", "Unsafe governance, secrets, release, or validation instruction.", "Generated output drift across platform copies."},
		SeverityGuidance:   []string{"Critical: skill instructs unsafe actions, fabricates validation, leaks secrets, or falsely claims production compatibility.", "High: skill is generic enough to misroute or produce low-quality domain work despite valid structure.", "Medium: domain content exists but lacks testable acceptance, output artifacts, or severity precision.", "Low: wording, examples, or metadata clarity needs improvement without blocking basic use."},
		OutputRequirements: []string{"Complete SKILL.md body with all required sections and synchronized changelog.", "Domain-specific checklist, decision rules, finding categories, severity, outputs, acceptance, and anti-patterns.", "Compatibility metadata state and validation evidence or unverified marker.", "Generator-smell review result with any rejected generic phrases.", "Tests or validation commands that enforce structure and non-generic content.", "Generated-copy synchronization notes where applicable."},
		AcceptanceCriteria: []string{"Skill cannot be reduced to name, description, and shared boilerplate.", "Every required section contains domain-specific, testable content.", "Frontmatter and body changelog are synchronized.", "Compatibility metadata is honest and centrally sourced.", "Generic phrases are absent or explicitly rejected.", "Validation/tests cover required structure and non-genericness."},
		AntiPatterns:       []string{"Creating a skill by search-and-replace from another skill.", "Using “structured analysis or review” as a trigger.", "Writing finding categories like “missing <skill> evidence”.", "Using generic severity guidance unrelated to domain impact.", "Claiming production-ready because required headings exist.", "Setting concrete platform versions without validation evidence."},
	},
}

type additionalSkillSeed struct {
	Name        string
	Description string
	Domain      string
	Artifacts   []string
	Risks       []string
	Signals     []string
	Outputs     []string
}

var additionalSkillSeeds = []additionalSkillSeed{
	{Name: "dora-readiness-reviewer", Description: "Review DORA readiness for ICT risk management, resilience testing, ICT incidents, third-party risk, roles, policies, evidence, and auditability.", Domain: "DORA readiness and ICT auditability", Artifacts: []string{"ICT risk framework", "resilience test plan", "incident procedure", "third-party register", "role matrix", "policy set"}, Risks: []string{"unowned DORA capability", "untested critical function", "missing incident evidence", "outsourcing blind spot", "policy without approval", "audit trail gap"}, Signals: []string{"risk acceptance record", "test protocol", "incident timeline", "contract inventory", "management approval", "evidence package"}, Outputs: []string{"DORA readiness gap list", "auditability matrix", "evidence request list", "policy and role remediation plan", "residual readiness risk summary"}},
	{Name: "ict-risk-management-reviewer", Description: "Review ICT risks, protection needs, criticality, controls, residual risks, risk treatment, and recurring reassessment.", Domain: "ICT risk management", Artifacts: []string{"asset inventory", "protection needs analysis", "criticality rating", "control catalogue", "risk register", "reassessment schedule"}, Risks: []string{"unrated critical asset", "weak residual risk rationale", "stale risk treatment", "missing control owner", "unsupported protection level", "outdated reassessment"}, Signals: []string{"asset classification", "control test", "risk decision", "treatment task", "owner approval", "review cadence"}, Outputs: []string{"ICT risk findings", "control-to-risk map", "residual risk decision log", "reassessment backlog", "risk treatment recommendation"}},
	{Name: "ict-third-party-risk-reviewer", Description: "Review cloud, SaaS, outsourcing, subcontractors, contracts, exit strategies, concentration risks, and DORA information-register readiness.", Domain: "ICT third-party risk", Artifacts: []string{"provider inventory", "cloud contract", "SaaS data-flow", "subcontractor list", "exit plan", "information register fields"}, Risks: []string{"unlisted subcontractor", "weak exit clause", "concentration exposure", "missing data location", "untracked outsourcing dependency", "contract evidence gap"}, Signals: []string{"contract clause", "provider assessment", "service criticality", "suboutsourcing notice", "exit-test result", "register update"}, Outputs: []string{"third-party risk report", "DORA register gap list", "exit readiness findings", "concentration risk summary", "contract remediation items"}},
	{Name: "ict-incident-reporting-reviewer", Description: "Review classification, escalation, documentation, reportability, reporting timelines, responsibilities, templates, and communication chains for ICT incidents.", Domain: "ICT incident reporting", Artifacts: []string{"incident classification matrix", "escalation procedure", "reporting template", "communications tree", "timeline log", "responsibility matrix"}, Risks: []string{"missed reporting threshold", "late escalation", "incomplete incident record", "unclear owner", "template mismatch", "broken communication path"}, Signals: []string{"severity decision", "timestamped escalation", "draft authority report", "contact list", "incident ticket", "post-incident review"}, Outputs: []string{"incident reporting readiness assessment", "classification gap list", "timeline evidence review", "template remediation plan", "communications chain findings"}},
	{Name: "operational-resilience-tester", Description: "Review backup and restore, failover, disaster recovery, restart procedures, crisis exercises, scenario tests, and lessons learned.", Domain: "operational resilience testing", Artifacts: []string{"backup policy", "restore test log", "failover runbook", "DR plan", "crisis exercise report", "lessons-learned register"}, Risks: []string{"unproven restore", "manual failover bottleneck", "RTO mismatch", "unrehearsed crisis role", "scenario gap", "unclosed lesson"}, Signals: []string{"RPO evidence", "RTO measurement", "failover transcript", "exercise participant record", "defect ticket", "retest proof"}, Outputs: []string{"resilience test findings", "RPO/RTO evidence table", "scenario coverage map", "recovery gap backlog", "lessons-learned closure plan"}},
	{Name: "audit-evidence-reviewer", Description: "Review evidence, approvals, tickets, logs, test protocols, risk decisions, versioning, and accountable owners.", Domain: "audit evidence quality", Artifacts: []string{"approval record", "ticket trail", "system log", "test protocol", "risk decision", "version history"}, Risks: []string{"orphaned approval", "missing log retention", "unlinked ticket", "unsigned test result", "stale evidence", "unclear accountable owner"}, Signals: []string{"review timestamp", "change reference", "approver identity", "log extract", "test result", "version tag"}, Outputs: []string{"audit evidence quality report", "missing-evidence list", "approval traceability map", "retention risk note", "evidence remediation checklist"}},
	{Name: "control-mapping-reviewer", Description: "Map technical measures to DORA, VAIT or BAIT migration needs, ISO 27001, BSI, internal policies, or MaRisk review expectations.", Domain: "control mapping", Artifacts: []string{"technical control", "DORA chapter map", "VAIT or BAIT mapping", "ISO 27001 annex", "BSI baseline", "internal policy"}, Risks: []string{"unmapped technical measure", "duplicated control claim", "obsolete VAIT reference", "weak MaRisk linkage", "missing implementation proof", "policy conflict"}, Signals: []string{"control owner", "implementation ticket", "test evidence", "policy clause", "audit note", "exception record"}, Outputs: []string{"control mapping matrix", "framework coverage gap list", "migration notes", "duplicate-control cleanup list", "evidence alignment report"}},
	{Name: "outsourcing-exit-strategy-reviewer", Description: "Review exit plans, data return, provider transitions, emergency operations, suboutsourcing, cloud dependencies, and business impact.", Domain: "outsourcing exit strategy", Artifacts: []string{"exit plan", "data return clause", "provider transition runbook", "emergency operating model", "suboutsourcing register", "business impact assessment"}, Risks: []string{"unrecoverable data", "provider lock-in", "untested transition", "critical suboutsourcing dependency", "cloud portability gap", "understated business impact"}, Signals: []string{"exit test", "data deletion proof", "alternate provider assessment", "critical process map", "contract clause", "transition timeline"}, Outputs: []string{"exit strategy findings", "data return gap list", "provider transition risk register", "business impact summary", "exit test plan"}},
	{Name: "documentation-governance-reviewer", Description: "Review documentation freshness, ownership, review cycles, approvals, versioning, validity, and traceability.", Domain: "documentation governance", Artifacts: []string{"document inventory", "owner matrix", "review schedule", "approval workflow", "version history", "validity metadata"}, Risks: []string{"stale procedure", "missing owner", "expired review", "unapproved change", "unversioned policy", "broken traceability"}, Signals: []string{"last review date", "approver record", "document status", "change request", "publication location", "source reference"}, Outputs: []string{"documentation governance report", "staleness backlog", "ownership gap list", "approval remediation plan", "traceability findings"}},
	{Name: "runbook-playbook-maintainer", Description: "Create and review runbooks, operating instructions, incident playbooks, escalation paths, restart procedures, and checklists.", Domain: "runbooks and playbooks", Artifacts: []string{"runbook", "operating instruction", "incident playbook", "escalation path", "restart checklist", "validation step"}, Risks: []string{"missing diagnosis step", "unsafe recovery command", "obsolete contact", "unverified restart", "ambiguous escalation", "checklist gap"}, Signals: []string{"operator dry run", "alert link", "dashboard reference", "ticket example", "rollback step", "post-action verification"}, Outputs: []string{"updated runbook", "playbook gap report", "operator checklist", "escalation correction list", "validation notes"}},
	{Name: "architecture-decision-recorder", Description: "Create and maintain ADRs with context, decision, alternatives, risks, security impact, compliance relation, and review points.", Domain: "architecture decision records", Artifacts: []string{"ADR", "decision context", "alternative analysis", "risk section", "security impact", "review date"}, Risks: []string{"undocumented decision", "missing alternative", "unowned risk", "security impact omission", "compliance ambiguity", "stale review point"}, Signals: []string{"decision owner", "status marker", "linked issue", "architecture diagram", "control reference", "review trigger"}, Outputs: []string{"ADR draft or update", "decision gap list", "alternative trade-off summary", "security and compliance note", "review schedule"}},
	{Name: "audit-traceability-maintainer", Description: "Link requirements, controls, implementation, tests, tickets, and evidence into an auditable trace.", Domain: "audit traceability", Artifacts: []string{"requirement", "control", "implementation change", "test case", "ticket", "evidence item"}, Risks: []string{"unlinked requirement", "control without test", "ticket without evidence", "implementation drift", "missing owner", "trace break"}, Signals: []string{"requirement ID", "control ID", "commit reference", "test report", "approval", "evidence link"}, Outputs: []string{"traceability matrix", "broken-link report", "evidence coverage map", "owner action list", "audit trail summary"}},
	{Name: "policy-documentation-maintainer", Description: "Create and update policies, standards, procedures, and control descriptions.", Domain: "policy documentation", Artifacts: []string{"policy", "standard", "procedure", "control description", "exception process", "approval record"}, Risks: []string{"policy conflict", "missing normative language", "unapproved exception", "unclear applicability", "stale control text", "weak procedure"}, Signals: []string{"policy owner", "effective date", "approval board", "control objective", "exception expiry", "review result"}, Outputs: []string{"policy update", "standard gap list", "procedure correction plan", "control description changes", "approval checklist"}},
	{Name: "evidence-package-creator", Description: "Create auditable evidence packages from tickets, pipeline results, test reports, approvals, scans, and architecture information.", Domain: "evidence packages", Artifacts: []string{"ticket export", "pipeline result", "test report", "approval", "scan report", "architecture reference"}, Risks: []string{"missing source link", "unverifiable package", "sensitive data exposure", "incomplete scan context", "unapproved change", "timestamp mismatch"}, Signals: []string{"package manifest", "checksum", "redaction note", "reviewer sign-off", "retention location", "source URL"}, Outputs: []string{"evidence package manifest", "collection checklist", "redaction summary", "gap list", "auditor handoff notes"}},
	{Name: "devsecops-maturity-reviewer", Description: "Assess maturity across plan, code, build, test, release, deploy, and operate: shift left, shield right, automation, security gates, ownership, and feedback loops.", Domain: "DevSecOps maturity", Artifacts: []string{"SDLC workflow", "security gate", "pipeline policy", "ownership model", "feedback loop", "operational metric"}, Risks: []string{"manual security bottleneck", "late vulnerability discovery", "weak gate enforcement", "unclear ownership", "missing runtime feedback", "immature automation"}, Signals: []string{"maturity score", "gate result", "team responsibility", "defect trend", "incident feedback", "improvement roadmap"}, Outputs: []string{"DevSecOps maturity assessment", "capability heatmap", "gap backlog", "ownership recommendations", "target-state roadmap"}},
	{Name: "pipeline-security-architect", Description: "Design and review secure CI/CD pipelines with isolated runners, minimal rights, OIDC, signed artifacts, protected environments, approval gates, and safe deployments.", Domain: "pipeline security architecture", Artifacts: []string{"workflow file", "runner pool", "OIDC trust policy", "artifact signing step", "protected environment", "approval gate"}, Risks: []string{"overprivileged token", "untrusted runner", "artifact tampering", "environment bypass", "secret sprawl", "unsafe deployment job"}, Signals: []string{"token permission block", "runner isolation evidence", "signature verification", "environment protection", "approval record", "deployment audit log"}, Outputs: []string{"pipeline security design", "CI/CD risk findings", "permission reduction plan", "artifact integrity controls", "deployment gate recommendations"}},
	{Name: "software-supply-chain-architect", Description: "Review SLSA, provenance, SBOM, signatures, attestations, build integrity, artifact promotion, and trusted builders.", Domain: "software supply chain architecture", Artifacts: []string{"SLSA target", "provenance attestation", "SBOM", "signature", "trusted builder", "artifact promotion rule"}, Risks: []string{"untrusted build", "missing provenance", "unsigned artifact", "SBOM drift", "promotion bypass", "attestation gap"}, Signals: []string{"builder identity", "attestation predicate", "signature verification", "dependency graph", "promotion approval", "artifact digest"}, Outputs: []string{"supply-chain architecture review", "SLSA gap map", "provenance findings", "SBOM and signing plan", "trusted-builder recommendations"}},
	{Name: "policy-as-code-engineer", Description: "Create and review policies for OPA/Rego, Kyverno, GitLab Policies, Conftest, Checkov, Terraform, Kubernetes, and CI/CD gates.", Domain: "policy-as-code engineering", Artifacts: []string{"Rego policy", "Kyverno rule", "GitLab policy", "Conftest suite", "Checkov check", "CI gate"}, Risks: []string{"overbroad deny rule", "missing test fixture", "policy bypass", "false-positive flood", "unversioned exception", "weak enforcement mode"}, Signals: []string{"policy test", "violation example", "exception record", "enforcement mode", "coverage report", "gate result"}, Outputs: []string{"policy-as-code changes", "test fixture set", "enforcement recommendation", "exception handling notes", "policy rollout plan"}},
	{Name: "secure-developer-platform-reviewer", Description: "Review Internal Developer Platforms for secure golden paths, self-service with guardrails, templates, permission models, secrets handling, and auditability.", Domain: "secure developer platform", Artifacts: []string{"golden path", "self-service workflow", "template catalog", "permission model", "secrets workflow", "audit log"}, Risks: []string{"unsafe template default", "privilege escalation", "secret leakage", "guardrail bypass", "untracked self-service action", "poor platform adoption"}, Signals: []string{"template scan", "RBAC review", "audit event", "developer feedback", "exception flow", "platform metric"}, Outputs: []string{"platform security review", "golden-path gap list", "guardrail recommendations", "permission model findings", "developer-experience risk notes"}},
	{Name: "vulnerability-management-coordinator", Description: "Assess CVE triage, prioritization, SLAs, exploitability, asset criticality, exceptions, risk acceptance, and tracking through remediation.", Domain: "vulnerability management coordination", Artifacts: []string{"CVE queue", "asset inventory", "SLA policy", "exploitability assessment", "exception record", "remediation ticket"}, Risks: []string{"missed critical CVE", "SLA breach", "weak risk acceptance", "unowned remediation", "false priority", "stale exception"}, Signals: []string{"EPSS or KEV signal", "asset criticality", "fix version", "owner assignment", "exception expiry", "closure proof"}, Outputs: []string{"vulnerability triage report", "SLA breach list", "risk acceptance review", "remediation tracker", "prioritization recommendation"}},
	{Name: "cloud-landing-zone-reviewer", Description: "Review accounts or subscriptions, networks, IAM, logging, policies, baselines, guardrails, encryption, tagging, and tenant separation.", Domain: "cloud landing zones", Artifacts: []string{"account structure", "network topology", "IAM baseline", "logging baseline", "policy assignment", "encryption standard"}, Risks: []string{"flat account model", "network exposure", "overprivileged IAM", "missing central logs", "unenforced policy", "tenant isolation gap"}, Signals: []string{"organization policy", "VPC or VNet design", "KMS configuration", "tag policy", "guardrail result", "audit log"}, Outputs: []string{"landing-zone review", "baseline gap list", "IAM and network findings", "logging and encryption recommendations", "tenant separation notes"}},
	{Name: "cloud-governance-reviewer", Description: "Review naming, tags, ownership, cost centers, allowed services, regions, data classification, policy enforcement, and audit evidence.", Domain: "cloud governance", Artifacts: []string{"naming standard", "tag policy", "ownership register", "cost center map", "allowed-service list", "region policy"}, Risks: []string{"unowned resource", "untagged cost", "forbidden region", "unapproved service", "data classification gap", "policy drift"}, Signals: []string{"resource inventory", "policy compliance report", "tag coverage", "budget owner", "classification label", "audit export"}, Outputs: []string{"cloud governance findings", "tagging backlog", "ownership gap report", "policy enforcement plan", "audit evidence summary"}},
	{Name: "finops-reviewer", Description: "Review cloud costs, budgets, rightsizing, reserved or committed usage, cost anomalies, showback or chargeback, and cost transparency per team or service.", Domain: "FinOps", Artifacts: []string{"cloud bill", "budget", "rightsizing report", "commitment plan", "anomaly alert", "showback view"}, Risks: []string{"unexplained cost spike", "idle resource waste", "wrong commitment", "missing budget owner", "opaque team spend", "chargeback dispute"}, Signals: []string{"cost allocation tag", "utilization metric", "forecast", "reservation coverage", "anomaly record", "unit cost"}, Outputs: []string{"FinOps review", "savings opportunities", "budget risk list", "rightsizing actions", "showback improvement plan"}},
	{Name: "sre-reliability-reviewer", Description: "Assess SLOs, SLIs, error budgets, capacity, degradation, timeouts, retries, circuit breakers, load shedding, and operational risks.", Domain: "SRE reliability", Artifacts: []string{"SLO", "SLI query", "error budget", "capacity plan", "degradation mode", "timeout policy"}, Risks: []string{"missing SLO", "bad SLI proxy", "budget burn blind spot", "retry storm", "capacity cliff", "load shedding gap"}, Signals: []string{"burn-rate alert", "latency percentile", "saturation metric", "incident trend", "chaos test result", "runbook"}, Outputs: []string{"reliability findings", "SLO and SLI recommendations", "error-budget analysis", "capacity risk notes", "resilience control backlog"}},
	{Name: "kubernetes-platform-reviewer", Description: "Review clusters, namespaces, RBAC, NetworkPolicies, Pod Security, admission controllers, resource limits, secrets, ingress, multi-tenancy, and upgrade readiness.", Domain: "Kubernetes platform", Artifacts: []string{"cluster baseline", "namespace model", "RBAC binding", "NetworkPolicy", "Pod Security setting", "admission controller"}, Risks: []string{"cluster-admin sprawl", "namespace escape", "missing network isolation", "privileged pod", "unbounded resource use", "stale cluster version"}, Signals: []string{"kubectl manifest", "policy report", "secret reference", "ingress rule", "resource quota", "upgrade plan"}, Outputs: []string{"Kubernetes platform review", "RBAC and tenancy findings", "policy remediation list", "upgrade readiness notes", "runtime hardening recommendations"}},
	{Name: "gitops-operations-reviewer", Description: "Review Argo CD or Flux setups, sync policies, drift detection, promotion, rollback, app-of-apps, secrets, cluster access, and deployment governance.", Domain: "GitOps operations", Artifacts: []string{"Argo CD application", "Flux Kustomization", "sync policy", "drift report", "promotion workflow", "rollback procedure"}, Risks: []string{"auto-sync blast radius", "drift ignored", "secret exposure", "cluster credential overreach", "promotion bypass", "rollback gap"}, Signals: []string{"reconciliation log", "health status", "commit signature", "environment branch", "sealed secret", "access audit"}, Outputs: []string{"GitOps operations review", "sync and drift findings", "promotion governance map", "rollback recommendations", "cluster access risk list"}},
	{Name: "aiops-signal-correlation-reviewer", Description: "Assess correlation of logs, metrics, traces, events, and incidents to reduce noise, improve root-cause analysis, and lower alert fatigue.", Domain: "AIOps signal correlation", Artifacts: []string{"log stream", "metric series", "trace span", "event feed", "incident record", "correlation rule"}, Risks: []string{"alert noise", "missing service context", "weak root-cause hint", "duplicate incident", "bad correlation key", "hidden failure pattern"}, Signals: []string{"correlation ID", "topology map", "incident cluster", "suppression rule", "trace exemplar", "noise ratio"}, Outputs: []string{"signal correlation review", "noise reduction findings", "root-cause context gaps", "deduplication recommendations", "AIOps rollout notes"}},
	{Name: "alert-quality-reviewer", Description: "Review alerts for actionability, clear symptoms, runbook links, severity, ownership, SLO relation, deduplication, escalation, and auto-remediation suitability.", Domain: "alert quality", Artifacts: []string{"alert rule", "runbook link", "severity model", "owner mapping", "SLO reference", "escalation policy"}, Risks: []string{"non-actionable alert", "missing owner", "severity inflation", "runbook dead link", "duplicate page", "unsafe remediation trigger"}, Signals: []string{"alert firing history", "page volume", "acknowledgement time", "burn-rate link", "dedupe key", "operator feedback"}, Outputs: []string{"alert quality report", "actionability fixes", "severity corrections", "runbook gap list", "auto-remediation suitability notes"}},
	{Name: "auto-remediation-reviewer", Description: "Review automated repair actions for safe limits, dry runs, approval modes, rollback, audit logs, blast radius, and protection against endless loops.", Domain: "auto-remediation safety", Artifacts: []string{"remediation workflow", "dry-run mode", "approval gate", "rollback action", "audit log", "rate limit"}, Risks: []string{"runaway loop", "wide blast radius", "unsafe automatic change", "missing rollback", "approval bypass", "audit blind spot"}, Signals: []string{"execution history", "safety threshold", "manual override", "change ticket", "rollback result", "loop breaker"}, Outputs: []string{"auto-remediation safety review", "blast-radius findings", "approval and rollback recommendations", "auditability notes", "guardrail backlog"}},
	{Name: "mlops-governance-reviewer", Description: "Review model versioning, training data, bias, drift, monitoring, approvals, reproducibility, model registry, and deployment gates.", Domain: "MLOps governance", Artifacts: []string{"model registry entry", "training dataset", "feature pipeline", "bias evaluation", "drift monitor", "deployment gate"}, Risks: []string{"unreproducible model", "biased dataset", "silent drift", "unapproved promotion", "missing lineage", "weak rollback"}, Signals: []string{"model version", "dataset hash", "evaluation report", "approval record", "serving metric", "registry transition"}, Outputs: []string{"MLOps governance review", "lineage and reproducibility findings", "bias and drift recommendations", "deployment gate gaps", "model approval notes"}},
	{Name: "llmops-security-reviewer", Description: "Review GenAI workloads for prompt injection, tool permissions, data exfiltration, RAG sources, sensitive prompt logging, eval sets, guardrails, and model access.", Domain: "LLMOps security", Artifacts: []string{"prompt template", "tool permission", "RAG source", "prompt log", "eval set", "guardrail policy"}, Risks: []string{"prompt injection", "tool abuse", "data exfiltration", "untrusted RAG content", "sensitive prompt logging", "uncontrolled model access"}, Signals: []string{"red-team eval", "retrieval allowlist", "tool audit", "DLP finding", "guardrail result", "access policy"}, Outputs: []string{"LLMOps security review", "prompt-injection findings", "tool-permission recommendations", "RAG source risk notes", "eval and guardrail gaps"}},
	{Name: "ai-change-risk-reviewer", Description: "Review AI-assisted changes before execution: automation boundaries, human approval, affected-system criticality, and audit evidence.", Domain: "AI-assisted change risk", Artifacts: []string{"AI change proposal", "automation boundary", "approval record", "system criticality", "execution plan", "audit trail"}, Risks: []string{"unapproved autonomous change", "critical-system impact", "ambiguous human oversight", "untracked AI rationale", "unsafe tool use", "missing rollback"}, Signals: []string{"human approval", "change classification", "tool call log", "diff summary", "validation result", "rollback plan"}, Outputs: []string{"AI change risk assessment", "approval gap list", "automation-boundary recommendation", "criticality findings", "audit evidence request"}},
	{Name: "privacy-data-protection-reviewer", Description: "Review privacy, personal data, data classification, deletion concepts, purpose limitation, GDPR risks, and sensitive-data logging.", Domain: "privacy and data protection", Artifacts: []string{"personal-data inventory", "classification label", "deletion concept", "purpose statement", "DPIA note", "log sample"}, Risks: []string{"personal data overcollection", "purpose creep", "missing deletion path", "sensitive logging", "unlawful retention", "weak data minimization"}, Signals: []string{"data-flow map", "retention setting", "consent or legal basis", "redaction rule", "access log", "privacy review"}, Outputs: []string{"privacy review findings", "data-classification gaps", "deletion and retention actions", "sensitive logging recommendations", "GDPR risk notes"}},
	{Name: "api-contract-reviewer", Description: "Review REST, GraphQL, OpenAPI, and gRPC contracts, breaking changes, versioning, AuthN/AuthZ, error formats, and compatibility.", Domain: "API contracts", Artifacts: []string{"OpenAPI spec", "GraphQL schema", "gRPC proto", "REST route", "error schema", "version policy"}, Risks: []string{"breaking change", "auth bypass", "incompatible error format", "missing version", "schema drift", "client compatibility gap"}, Signals: []string{"contract test", "schema diff", "consumer fixture", "deprecation note", "auth matrix", "compatibility report"}, Outputs: []string{"API contract review", "breaking-change findings", "versioning recommendations", "auth and error-format notes", "consumer compatibility checklist"}},
	{Name: "secure-design-reviewer", Description: "Review secure-by-design decisions, least privilege, Zero Trust, tenant separation, secure defaults, and abuse scenarios.", Domain: "secure design", Artifacts: []string{"design proposal", "trust boundary", "permission model", "tenant model", "default configuration", "abuse case"}, Risks: []string{"overtrusted network", "privilege creep", "tenant data leak", "insecure default", "missing abuse case", "weak isolation"}, Signals: []string{"architecture diagram", "threat model", "access decision", "configuration baseline", "data-flow review", "control rationale"}, Outputs: []string{"secure design review", "abuse-case findings", "least-privilege recommendations", "tenant-isolation notes", "secure-default backlog"}},
	{Name: "policy-as-code-reviewer", Description: "Review GitLab Security Policies, OPA/Rego, Kyverno, Conftest, Sentinel, admission policies, compliance pipelines, and central guardrails.", Domain: "policy-as-code review", Artifacts: []string{"GitLab security policy", "OPA/Rego rule", "Kyverno policy", "Sentinel policy", "admission policy", "compliance pipeline"}, Risks: []string{"policy bypass", "untested deny path", "central guardrail drift", "weak exception process", "pipeline enforcement gap", "noisy violation"}, Signals: []string{"policy test result", "violation fixture", "exception approval", "enforcement mode", "pipeline evidence", "audit event"}, Outputs: []string{"policy-as-code review", "guardrail gap list", "test coverage findings", "exception governance notes", "enforcement recommendations"}},
	{Name: "container-security-reviewer", Description: "Review Dockerfiles, base images, user rights, capabilities, SBOM, image signing, distroless or slim images, CVEs, and runtime hardening.", Domain: "container security", Artifacts: []string{"Dockerfile", "base image", "runtime user", "Linux capability", "image SBOM", "signature"}, Risks: []string{"root container", "bloated vulnerable image", "unsigned image", "excess capability", "secret in layer", "missing runtime hardening"}, Signals: []string{"image scan", "SBOM digest", "cosign verification", "USER directive", "seccomp profile", "CVE triage"}, Outputs: []string{"container security findings", "base-image recommendations", "runtime hardening checklist", "SBOM and signing gaps", "CVE remediation notes"}},
	{Name: "identity-access-reviewer", Description: "Review IAM, roles, service accounts, groups, tokens, OIDC federation, GitLab or GitHub permissions, cloud rights, and privilege-escalation paths.", Domain: "identity and access", Artifacts: []string{"IAM policy", "role binding", "service account", "group membership", "OIDC federation", "repository permission"}, Risks: []string{"privilege escalation", "stale access", "overbroad role", "long-lived token", "weak federation trust", "unreviewed group"}, Signals: []string{"access review", "policy simulator", "token age", "assume-role path", "audit log", "permission diff"}, Outputs: []string{"identity access review", "privilege-escalation findings", "least-privilege plan", "token and federation notes", "access cleanup list"}},
	{Name: "risk-acceptance-reviewer", Description: "Document and assess conscious risk decisions, impact and likelihood, expiry dates, and compensating measures.", Domain: "risk acceptance", Artifacts: []string{"risk acceptance record", "impact assessment", "likelihood assessment", "expiry date", "compensating measure", "approval"}, Risks: []string{"expired acceptance", "unsupported impact", "missing approver", "weak compensating measure", "open-ended exception", "untracked review"}, Signals: []string{"risk owner", "approval timestamp", "review date", "control exception", "residual risk", "closure condition"}, Outputs: []string{"risk acceptance assessment", "expiry and owner gap list", "compensating-control findings", "approval evidence request", "residual risk recommendation"}},
	{Name: "secure-code-reviewer", Description: "Review code vulnerabilities such as injection, path traversal, SSRF, XSS, deserialization, crypto misuse, and race conditions.", Domain: "secure code review", Artifacts: []string{"source diff", "input parser", "file access path", "HTTP client", "template rendering", "crypto use"}, Risks: []string{"injection", "path traversal", "SSRF", "XSS", "unsafe deserialization", "race condition"}, Signals: []string{"tainted input", "authorization check", "fuzz case", "unit test", "static-analysis warning", "exploitability note"}, Outputs: []string{"secure code findings", "exploitability assessment", "fix guidance", "negative-test recommendations", "residual risk notes"}},
	{Name: "performance-scalability-reviewer", Description: "Review load behavior, bottlenecks, caching, database access, queue behavior, scaling, timeouts, and resource limits.", Domain: "performance and scalability", Artifacts: []string{"load test", "database query", "cache strategy", "queue consumer", "autoscaling rule", "resource limit"}, Risks: []string{"N+1 query", "cache stampede", "queue backlog", "timeout cascade", "CPU saturation", "scaling bottleneck"}, Signals: []string{"latency percentile", "throughput chart", "query plan", "cache hit rate", "queue depth", "capacity forecast"}, Outputs: []string{"performance review", "bottleneck findings", "scaling recommendations", "timeout and resource notes", "load-test plan"}},
	{Name: "migration-change-reviewer", Description: "Review database migrations, schema changes, breaking changes, rollback ability, backward compatibility, and zero-downtime deployments.", Domain: "migration and change safety", Artifacts: []string{"database migration", "schema diff", "compatibility plan", "rollback script", "feature flag", "deployment sequence"}, Risks: []string{"irreversible migration", "breaking schema change", "downtime window", "old client incompatibility", "data-loss path", "rollback gap"}, Signals: []string{"expand-contract plan", "migration test", "backup checkpoint", "rollout metric", "dual-write note", "deprecation timeline"}, Outputs: []string{"migration risk review", "rollback readiness findings", "compatibility checklist", "zero-downtime recommendations", "validation plan"}},
	{Name: "sbom-vulnerability-management-reviewer", Description: "Review SBOM generation, CVE triage, VEX, exception processes, patch SLAs, and the vulnerability lifecycle.", Domain: "SBOM vulnerability management", Artifacts: []string{"SBOM", "CVE report", "VEX document", "exception process", "patch SLA", "vulnerability ticket"}, Risks: []string{"missing component", "untriaged CVE", "invalid VEX claim", "expired exception", "SLA breach", "lifecycle blind spot"}, Signals: []string{"CycloneDX or SPDX file", "scanner output", "reachability note", "fix version", "risk acceptance", "closure evidence"}, Outputs: []string{"SBOM vulnerability review", "CVE triage gap list", "VEX quality findings", "patch SLA report", "lifecycle remediation plan"}},
	{Name: "developer-experience-reviewer", Description: "Review setup, local development, error messages, Makefiles or scripts, onboarding, tooling consistency, and practicality for teams.", Domain: "developer experience", Artifacts: []string{"setup guide", "local dev script", "Makefile target", "error message", "onboarding doc", "tool version file"}, Risks: []string{"broken setup", "unclear failure", "tool mismatch", "slow feedback loop", "hidden prerequisite", "fragile script"}, Signals: []string{"fresh-clone run", "command output", "developer feedback", "CI parity", "dependency check", "troubleshooting path"}, Outputs: []string{"developer-experience review", "setup friction list", "tooling consistency findings", "onboarding fixes", "feedback-loop recommendations"}},
	{Name: "resilience-reviewer", Description: "Review timeouts, retries, circuit breakers, failover, backpressure, degraded modes, and resilience behavior.", Domain: "application resilience", Artifacts: []string{"timeout setting", "retry policy", "circuit breaker", "failover design", "backpressure mechanism", "degraded mode"}, Risks: []string{"retry storm", "missing timeout", "failover surprise", "unbounded queue", "no degraded path", "cascading failure"}, Signals: []string{"chaos test", "load-shed metric", "dependency error rate", "fallback log", "SLO impact", "incident history"}, Outputs: []string{"resilience review", "failure-mode findings", "timeout and retry recommendations", "degraded-mode gaps", "resilience test plan"}},
	{Name: "backup-restore-reviewer", Description: "Review restore tests, RPO/RTO, data integrity, backup protection, recoverability, and disaster recovery.", Domain: "backup and restore", Artifacts: []string{"backup schedule", "restore test", "RPO target", "RTO target", "integrity check", "DR procedure"}, Risks: []string{"untested backup", "corrupt restore", "RPO miss", "RTO miss", "unprotected backup", "DR gap"}, Signals: []string{"restore transcript", "checksum", "retention policy", "immutability setting", "recovery metric", "backup alert"}, Outputs: []string{"backup-restore review", "RPO/RTO evidence table", "data-integrity findings", "backup protection recommendations", "DR remediation plan"}},
	{Name: "penetration-test-coordinator", Description: "Coordinate and review penetration test scoping, execution evidence, findings, severity ratings, remediation tracking, retest results, and disclosure readiness.", Domain: "penetration testing coordination", Artifacts: []string{"test scope document", "pentest report", "finding severity rating", "retest evidence", "remediation ticket", "disclosure record"}, Risks: []string{"scope creep or gap", "critical finding without retest", "unowned remediation", "incomplete coverage", "stale pentest", "undisclosed critical finding"}, Signals: []string{"tester credential and authorization", "scope approval", "CVSSv3 or CVSS4 rating", "exploitation proof-of-concept", "retest confirmation", "closure ticket"}, Outputs: []string{"pentest coordination review", "scope and coverage assessment", "remediation tracking report", "retest readiness findings", "disclosure preparation notes"}},
	{Name: "dast-reviewer", Description: "Review dynamic application security testing for web applications and APIs including scan coverage, authenticated paths, false positive triage, and CI/CD pipeline integration.", Domain: "dynamic application security testing", Artifacts: []string{"scan configuration", "authentication token or session", "scan report", "false-positive record", "CI pipeline integration step", "coverage baseline"}, Risks: []string{"unauthenticated scan gap", "missing API endpoint coverage", "false-positive flood", "scan environment contamination", "CI gate bypass", "stale coverage baseline"}, Signals: []string{"scan credential evidence", "authenticated request log", "finding triage note", "deduplication rule", "pipeline gate result", "coverage metric"}, Outputs: []string{"DAST review findings", "authenticated coverage assessment", "false-positive triage report", "CI integration recommendations", "scan configuration improvements"}},
	{Name: "secrets-scanning-reviewer", Description: "Review secrets detection across pre-commit hooks, CI pipelines, repository history, container images, and IaC for hardcoded credentials, tokens, keys, and unsafe secret handling.", Domain: "secrets scanning and credential hygiene", Artifacts: []string{"pre-commit hook configuration", "CI scan result", "git history scan output", "detection rule set", "credential rotation record", "exception allow-list"}, Risks: []string{"bypassed pre-commit hook", "secret committed to git history", "hardcoded token in IaC or image", "stale rotation record", "missing detection rule coverage", "allow-list abuse or scope creep"}, Signals: []string{"scan tool output with finding detail", "detection rule coverage report", "rotation timestamp and evidence", "CI gate block event", "hook bypass attempt log", "exception approval record"}, Outputs: []string{"secrets scanning review", "detection coverage gap report", "exposed-secret remediation plan", "rotation evidence findings", "allow-list governance recommendations"}},
	{Name: "license-compliance-reviewer", Description: "Review open-source license types, copyleft obligations, license compatibility, attribution requirements, SBOM license fields, and internal policy adherence across direct and transitive dependencies.", Domain: "open-source license compliance", Artifacts: []string{"dependency inventory", "SBOM with license fields", "copyleft obligation record", "attribution file", "license policy", "exception approval record"}, Risks: []string{"copyleft contagion", "missing attribution file", "license incompatibility between components", "unlicensed or unknown-license dependency", "expired or unapproved exception", "policy drift across teams"}, Signals: []string{"SPDX identifier", "license scan tool output", "attribution file completeness", "policy exception approval", "compatibility matrix", "transitive dependency graph"}, Outputs: []string{"license compliance review", "copyleft risk list", "attribution gap report", "license policy exception findings", "remediation or procurement guidance"}},
	{Name: "runtime-security-monitor-reviewer", Description: "Review runtime threat detection coverage using eBPF, Falco, or equivalent tools including behavioral profiles, detection rules, alert fidelity, response playbooks, and operational integration.", Domain: "runtime security monitoring", Artifacts: []string{"Falco or eBPF rule set", "behavioral baseline profile", "alert definition", "response playbook", "sensor deployment configuration", "detection coverage map"}, Risks: []string{"detection rule coverage gap", "behavioral baseline drift", "alert fatigue from noisy rules", "undetected lateral movement or privilege escalation", "response playbook gap for runtime events", "sensor deployment gap in workloads"}, Signals: []string{"rule test and validation result", "alert firing history and triage", "baseline deviation record", "incident correlation evidence", "sensor health metric", "playbook exercise outcome"}, Outputs: []string{"runtime security monitoring review", "detection coverage findings", "alert fidelity and noise report", "response playbook gaps", "sensor deployment recommendations"}},
	{Name: "sast-reviewer", Description: "Review static analysis tool configuration, rule sets, CI gate calibration, false-positive triage workflows, noise reduction, and SAST coverage across Semgrep, SonarQube, CodeQL, and equivalent tools.", Domain: "static application security testing", Artifacts: []string{"SAST tool configuration", "rule set or query pack", "CI quality gate definition", "false-positive triage record", "coverage baseline", "suppression allow-list"}, Risks: []string{"rule set coverage gap for the language or framework", "CI gate too noisy to enforce", "false-positive suppression abused to hide real findings", "scan excluded from critical paths", "stale rule set missing recent CVE patterns", "findings not routed to owning team"}, Signals: []string{"scan tool output with severity breakdown", "false-positive rate over time", "CI gate enforcement evidence", "suppression approval record", "coverage report by code path", "finding-to-ticket assignment rate"}, Outputs: []string{"SAST tool review", "rule set coverage and tuning recommendations", "CI gate calibration findings", "false-positive triage guidance", "suppression governance notes"}},
	{Name: "security-incident-responder", Description: "Guide and review active security incident response including detection, triage, containment, eradication, recovery, stakeholder communication, and evidence preservation throughout the incident lifecycle.", Domain: "security incident response", Artifacts: []string{"incident ticket or war-room record", "triage classification", "containment action log", "eradication evidence", "recovery validation", "stakeholder communication record"}, Risks: []string{"delayed containment allowing attacker persistence", "evidence destruction during remediation", "incomplete eradication leaving backdoors", "stakeholder communication gap during active incident", "premature recovery before root cause confirmed", "scope underestimation missing affected systems"}, Signals: []string{"initial detection alert or report", "affected system inventory", "attacker TTP evidence", "containment completion confirmation", "clean-state validation", "communication timeline"}, Outputs: []string{"incident response coordination log", "containment and eradication action plan", "affected scope and impact assessment", "stakeholder communication drafts", "evidence preservation checklist", "recovery readiness decision"}},
	{Name: "cryptography-reviewer", Description: "Review key management, PKI, TLS configuration, cipher suite selection, certificate lifecycle, algorithm choices, HSM usage, and cryptographic migration readiness.", Domain: "cryptography and key management", Artifacts: []string{"key management policy", "PKI or CA configuration", "TLS configuration baseline", "cipher suite list", "certificate inventory", "algorithm inventory"}, Risks: []string{"weak or deprecated algorithm in use", "private key stored without HSM or secret manager", "certificate expiry without automated renewal", "TLS downgrade or insecure default", "missing certificate pinning for critical clients", "cryptographic migration plan absent for post-quantum readiness"}, Signals: []string{"TLS scan result", "certificate validity and renewal automation evidence", "key storage audit", "cipher suite negotiation log", "algorithm usage grep or scan output", "HSM or KMS access policy"}, Outputs: []string{"cryptography review findings", "algorithm and cipher deprecation list", "key management gap report", "TLS hardening recommendations", "certificate lifecycle remediation plan", "post-quantum migration readiness note"}},
	{Name: "red-team-coordinator", Description: "Plan and review red team and purple team exercises including ATT&CK scenario design, adversarial simulation scoping, detection gap analysis, findings-to-controls mapping, and defensive improvement feedback loops.", Domain: "red team and purple team operations", Artifacts: []string{"ATT&CK scenario plan", "rules of engagement", "exercise scope boundary", "detection gap finding", "findings-to-controls map", "purple team debrief"}, Risks: []string{"scenario outside agreed scope boundary", "detection gap without assigned control owner", "findings not fed back into detection rules or controls", "exercise evidence not retained for compliance", "red team TTP overlap with production incident indicators", "debrief skipped leaving improvement backlog unactioned"}, Signals: []string{"ATT&CK technique coverage map", "rules of engagement signature", "detection hit or miss log", "control improvement ticket", "purple team session recording or notes", "before-and-after detection coverage comparison"}, Outputs: []string{"red team exercise review", "ATT&CK coverage gap map", "detection findings and evasion techniques used", "findings-to-controls assignment", "purple team improvement backlog", "debrief and detection uplift summary"}},
	{Name: "security-metrics-reviewer", Description: "Review DevSecOps program effectiveness through MTTR, MTTD, vulnerability aging, mean time to patch, SLA adherence, security debt trends, and KPI dashboard quality.", Domain: "security metrics and program effectiveness", Artifacts: []string{"vulnerability aging report", "MTTR and MTTD measurement", "SLA adherence dashboard", "security debt backlog", "KPI definition", "trend report"}, Risks: []string{"metric gaming hiding real risk exposure", "SLA breach unreported due to measurement gap", "security debt growing faster than remediation capacity", "KPI definition misaligned with actual risk", "MTTD inflated by detection blind spots", "dashboard not reviewed or actioned regularly"}, Signals: []string{"vulnerability ticket timestamps", "SLA breach record", "remediation closure rate", "open-risk aging chart", "detection timestamp from SIEM", "backlog growth trend"}, Outputs: []string{"security metrics review", "SLA adherence report", "vulnerability aging and debt findings", "KPI definition recommendations", "MTTR and MTTD measurement gaps", "program effectiveness improvement plan"}},
	{Name: "network-security-reviewer", Description: "Review firewall rule sets, network segmentation, DNS security, WAF configuration, DDoS protection, east-west traffic controls, and network access control lists for compliance with least-privilege network principles.", Domain: "network security", Artifacts: []string{"firewall rule set", "network segmentation diagram", "DNS security configuration", "WAF rule set", "DDoS protection policy", "network ACL"}, Risks: []string{"overly permissive firewall rule", "missing east-west traffic isolation", "DNS hijacking or cache poisoning exposure", "WAF rule bypass or false-positive flood", "DDoS protection misconfiguration", "implicit allow in default ACL"}, Signals: []string{"firewall rule review output", "network flow log", "DNS query anomaly", "WAF block and allow log", "DDoS simulation result", "ACL effective-permission trace"}, Outputs: []string{"network security review", "firewall and ACL findings", "segmentation gap map", "DNS and WAF hardening recommendations", "DDoS protection assessment", "east-west traffic control plan"}},
	{Name: "patch-management-reviewer", Description: "Review OS and middleware patch workflows, patch window scheduling, emergency patching escalation, rollback readiness, SLA tracking per asset class, and patch compliance reporting.", Domain: "patch management", Artifacts: []string{"patch policy", "asset inventory with patch status", "patch window schedule", "emergency patching procedure", "rollback plan", "patch compliance report"}, Risks: []string{"critical patch SLA breach on internet-exposed asset", "missing emergency escalation path for zero-day", "rollback procedure untested or absent", "patch window too infrequent for asset criticality", "third-party or OEM component excluded from patch scope", "compliance report masking actual patch status"}, Signals: []string{"patch age per asset class", "SLA compliance rate", "emergency patch response time", "rollback test evidence", "scan tool patch gap output", "exception approval record"}, Outputs: []string{"patch management review", "SLA breach and gap list", "asset-class remediation priority", "emergency patching process findings", "rollback readiness assessment", "compliance reporting accuracy notes"}},
	{Name: "security-awareness-reviewer", Description: "Review phishing simulation campaigns, developer security training programs, security champions effectiveness, awareness metrics, and the measurability of human-factor risk reduction.", Domain: "security awareness and human-factor risk", Artifacts: []string{"phishing simulation campaign result", "security training completion record", "security champions roster and activity", "awareness metric dashboard", "developer security curriculum", "click and report rate trend"}, Risks: []string{"phishing click rate not improving over time", "training completion without knowledge retention measure", "security champions without mandate or recognition", "awareness program not tailored to developer threat model", "metric showing improvement while real incidents increase", "no feedback loop from incidents to training content"}, Signals: []string{"phishing simulation click and report rate", "training completion and quiz score", "champions activity log", "incident correlation to human error", "curriculum review date", "developer security feedback"}, Outputs: []string{"security awareness review", "phishing simulation trend analysis", "training program gap list", "security champions effectiveness findings", "metric quality recommendations", "curriculum improvement plan"}},
	{Name: "endpoint-security-reviewer", Description: "Review EDR configuration, agent deployment coverage, detection rule quality, endpoint hygiene including encryption and patch status, and incident response capability at the endpoint level.", Domain: "endpoint security", Artifacts: []string{"EDR policy configuration", "agent deployment coverage report", "detection rule set", "endpoint hygiene baseline", "disk encryption status", "endpoint incident response playbook"}, Risks: []string{"EDR agent not deployed on critical workloads", "detection rule gap for living-off-the-land techniques", "endpoint encryption not enforced or escrowed", "stale OS or software increasing attack surface", "EDR exclusion abused to hide malicious activity", "endpoint IR playbook untested or absent"}, Signals: []string{"agent health and coverage metric", "detection rule test result", "encryption compliance report", "patch status per endpoint class", "exclusion list audit", "EDR alert and response time"}, Outputs: []string{"endpoint security review", "agent coverage gap list", "detection rule quality findings", "encryption and hygiene remediation plan", "exclusion governance recommendations", "IR playbook readiness assessment"}},
	{Name: "ux-research-reviewer", Description: "Review user research methodology, participant recruitment, usability study design, persona validity, journey map quality, and synthesis rigour to ensure design decisions are grounded in evidence rather than assumption.", Domain: "UX research", Artifacts: []string{"usability study plan", "persona set", "user journey map", "research synthesis report", "participant recruitment criteria", "interview or survey guide"}, Risks: []string{"biased or non-representative participant sample", "leading interview questions skewing findings", "synthesis conclusions unsupported by observed data", "persona not updated after recent research", "journey map missing significant pain points or emotions", "research findings not traceable to design decisions"}, Signals: []string{"participant diversity and sample size record", "session recording or transcript", "affinity diagram or synthesis artefact", "design decision linked to specific finding", "persona revision date", "research backlog priority"}, Outputs: []string{"UX research review", "methodology quality findings", "persona and journey map gap list", "synthesis rigour recommendations", "research-to-design traceability notes"}},
	{Name: "ui-design-reviewer", Description: "Review visual design quality covering typography scales, colour theory, spacing systems, layout grids, iconography, visual hierarchy, modern design patterns, and brand consistency across product surfaces.", Domain: "UI visual design", Artifacts: []string{"design mockup or Figma file", "typography scale definition", "colour palette and token set", "spacing and grid system", "icon set and usage guide", "component visual inventory"}, Risks: []string{"inconsistent typography scale breaking visual rhythm", "colour contrast failing accessibility thresholds", "spacing applied ad hoc without system reference", "broken visual hierarchy burying primary actions", "icon style mismatch across product surfaces", "design diverging from brand guidelines without rationale"}, Signals: []string{"design tool visual inspection", "contrast ratio measurement", "grid and spacing overlay audit", "component visual audit", "brand guideline comparison", "peer design critique record"}, Outputs: []string{"UI design review", "typography and colour findings", "spacing and grid gap list", "visual hierarchy recommendations", "brand and icon consistency notes"}},
	{Name: "responsive-design-reviewer", Description: "Review responsive design implementation covering mobile-first strategy, breakpoint definitions, fluid grids, flexbox and CSS Grid usage, viewport handling, image scaling, and cross-device layout correctness.", Domain: "responsive design", Artifacts: []string{"breakpoint definition set", "fluid grid or CSS Grid and flexbox configuration", "viewport meta tag", "CSS media query set", "responsive image directive", "cross-device test result"}, Risks: []string{"content overflow at intermediate viewport widths", "fixed-width element breaking mobile layout", "missing or incorrect viewport meta tag", "image not scaling or loading at wrong resolution", "breakpoint gap leaving unsupported viewport range", "touch target too small at mobile breakpoint"}, Signals: []string{"browser DevTools responsive mode inspection", "real-device test on iOS and Android", "CSS Grid and flexbox audit", "image srcset and sizes inspection", "Lighthouse mobile score", "cross-browser compatibility test result"}, Outputs: []string{"responsive design review", "breakpoint coverage assessment", "overflow and layout gap list", "image scaling and srcset recommendations", "cross-device test findings"}},
	{Name: "accessibility-reviewer", Description: "Review WCAG 2.1 and 2.2 conformance covering colour contrast, keyboard navigation, screen reader compatibility, ARIA semantics, focus management, skip links, and inclusive design patterns against AA and AAA criteria.", Domain: "web accessibility", Artifacts: []string{"WCAG audit report", "colour contrast measurement", "keyboard navigation flow map", "ARIA attribute and landmark set", "focus indicator definition", "screen reader test result"}, Risks: []string{"colour contrast below WCAG AA 4.5:1 threshold", "keyboard trap preventing sequential navigation", "missing or semantically incorrect ARIA role or label", "focus indicator invisible or absent on interactive element", "image without descriptive alt text", "form field without programmatically associated label"}, Signals: []string{"automated accessibility scan output", "manual keyboard traversal session", "screen reader session recording", "colour contrast ratio per element", "ARIA validator and landmark check", "WCAG criterion pass or fail mapping"}, Outputs: []string{"accessibility review", "WCAG conformance gap list", "colour and contrast findings", "keyboard and focus management recommendations", "ARIA correction plan", "screen reader compatibility notes"}},
	{Name: "design-system-reviewer", Description: "Review design system health covering component library coverage, design token definitions, documentation quality, versioning governance, design-to-code parity, and adoption consistency across product teams.", Domain: "design systems", Artifacts: []string{"component library inventory", "design token set", "component documentation", "versioning and changelog policy", "design-to-code parity report", "governance and ownership record"}, Risks: []string{"component absent from library forcing ad hoc one-off implementations", "design token drifting between design tool and codebase", "component documentation out of sync with current behaviour", "breaking change shipped without version notice or migration guide", "component used inconsistently across product areas", "no accountable owner for governance and deprecation decisions"}, Signals: []string{"component usage audit across codebase", "token sync status between Figma and code", "documentation last-reviewed date", "version changelog completeness", "design-to-Storybook visual parity check", "governance decision log and owner"}, Outputs: []string{"design system review", "component coverage gap list", "token and parity findings", "documentation quality report", "versioning and governance recommendations"}},
	{Name: "interaction-design-reviewer", Description: "Review interaction design quality covering micro-interactions, animation principles, easing and timing, motion accessibility, state transition consistency, loading states, empty states, and error feedback patterns.", Domain: "interaction design", Artifacts: []string{"interaction specification", "animation timing and easing definition", "state transition diagram", "loading state design", "empty state design", "error and success feedback pattern"}, Risks: []string{"animation too slow degrading perceived performance", "motion causing discomfort for users with vestibular disorders", "loading state absent leaving user without progress feedback", "empty state missing guidance or call to action", "state transition inconsistent across similar interactions", "error feedback absent or appearing too late after user action"}, Signals: []string{"animation duration and easing audit", "prefers-reduced-motion media query implementation check", "loading state coverage across async operations", "empty state inventory per data-driven view", "transition consistency review across flows", "user action-to-feedback latency measurement"}, Outputs: []string{"interaction design review", "animation and timing findings", "motion accessibility compliance report", "loading and empty state gap list", "state transition consistency notes", "feedback pattern recommendations"}},
	{Name: "information-architecture-reviewer", Description: "Review information architecture quality covering navigation structure, content hierarchy, findability, sitemap organisation, labelling clarity, search effectiveness, and alignment with user mental models.", Domain: "information architecture", Artifacts: []string{"sitemap or navigation tree", "navigation label set", "content hierarchy model", "search configuration and index", "card sort or tree test result", "wayfinding element inventory"}, Risks: []string{"navigation label ambiguous to target user", "primary content buried beyond three interaction levels", "search returning irrelevant or no results for common queries", "sitemap not reflecting user mental model from research", "duplicate content paths creating navigation confusion", "missing breadcrumb or wayfinding on deep pages"}, Signals: []string{"tree test task success rate", "card sort result and cluster analysis", "search query log and zero-result rate", "click-depth analytics", "navigation heatmap or session recording", "findability usability test result"}, Outputs: []string{"information architecture review", "navigation structure and labelling findings", "findability gap list", "sitemap improvement recommendations", "search quality notes"}},
	{Name: "content-strategy-reviewer", Description: "Review microcopy, tone of voice consistency, error messages, onboarding copy, button labels, empty state text, and content hierarchy to ensure language reduces friction and supports user goals at every touchpoint.", Domain: "content strategy and UX writing", Artifacts: []string{"microcopy inventory", "tone of voice guideline", "error message set", "onboarding and first-run copy", "button and CTA label convention", "empty state copy"}, Risks: []string{"error message blaming user without actionable next step", "tone inconsistent across product surfaces and channels", "button label too generic to communicate outcome", "onboarding copy overwhelming first-time user with information", "empty state copy missing guidance or call to action", "content hierarchy mismatched with user goal priority"}, Signals: []string{"copy audit across critical user flows", "tone of voice checklist comparison", "error message review against plain-language standard", "A/B test result on copy variants", "user comprehension test or five-second test", "content last-reviewed date"}, Outputs: []string{"content strategy review", "microcopy gap list", "tone and voice consistency findings", "error message improvement plan", "onboarding and empty-state copy recommendations"}},
	{Name: "performance-ux-reviewer", Description: "Review perceived performance and Core Web Vitals covering LCP, CLS, INP, skeleton screen usage, lazy loading, font loading strategies, optimistic UI patterns, and the user experience impact of performance regressions.", Domain: "performance UX", Artifacts: []string{"Core Web Vitals report", "skeleton screen or placeholder implementation", "lazy loading configuration", "font loading strategy", "optimistic UI pattern inventory", "Lighthouse or PageSpeed Insights report"}, Risks: []string{"high LCP caused by render-blocking resource or unoptimised hero image", "CLS from late-loading image or dynamic content without reserved dimensions", "missing skeleton screen leaving blank viewport during data load", "web font causing FOUT or FOIT on first load", "INP degraded by heavy main-thread task blocking interaction", "performance regression undetected before production release"}, Signals: []string{"Lighthouse performance score", "CrUX field data for real users", "LCP element and resource waterfall trace", "CLS shift source and element", "font loading waterfall and swap behaviour", "Interaction to Next Paint trace"}, Outputs: []string{"performance UX review", "Core Web Vitals findings and priorities", "skeleton and loading state gap list", "font loading and rendering recommendations", "CLS and LCP remediation plan"}},
	{Name: "design-patterns-reviewer", Description: "Review adherence to established UI design patterns and best practices covering atomic design methodology, component composition, progressive disclosure, Gestalt principles, Nielsen heuristics, dark mode support, and cross-platform consistency.", Domain: "UI design patterns and best practices", Artifacts: []string{"component composition model", "atomic design structure", "progressive disclosure implementation", "dark mode and theming configuration", "heuristic evaluation result", "cross-platform pattern inventory"}, Risks: []string{"UI reinventing a solved pattern rather than using established convention", "progressive disclosure hiding critical information users expect up front", "dark mode colour not derived from semantic tokens causing hardcoded overrides", "Gestalt principle violated causing grouping or figure-ground confusion", "heuristic violation such as no undo, no system status, or inconsistent affordance", "mobile and desktop pattern diverging without justified platform rationale"}, Signals: []string{"pattern library or convention reference", "atomic design layer audit", "dark mode visual regression test", "heuristic evaluation score per principle", "Gestalt grouping review", "cross-platform pattern comparison"}, Outputs: []string{"design patterns review", "pattern adherence and deviation findings", "atomic design structure recommendations", "dark mode and theming gap list", "heuristic violation report", "cross-platform consistency notes"}},
	{Name: "usability-testing-coordinator", Description: "Plan and review usability testing sessions covering test objective definition, task scenario design, participant recruitment, moderation quality, finding synthesis, severity rating, and design improvement traceability.", Domain: "usability testing", Artifacts: []string{"usability test plan", "task scenario set", "participant screening and recruitment criteria", "moderation guide", "session recording or transcript", "synthesis and severity report"}, Risks: []string{"task scenario too narrow or leading to surface real usability problems", "participant sample not representative of target user population", "moderator bias influencing session direction or outcomes", "findings not prioritised by frequency and severity", "synthesis conclusions not connected to specific design changes", "testing conducted too late in the design cycle to influence decisions"}, Signals: []string{"task completion rate per scenario", "time-on-task measurement", "error count and recovery observation per task", "think-aloud transcript excerpt", "finding severity rating", "design change ticket linked to finding"}, Outputs: []string{"usability test review", "task scenario quality assessment", "participant sample findings", "synthesis and severity report quality notes", "design improvement traceability map"}},
	{Name: "mobile-ux-reviewer", Description: "Review mobile UX patterns covering touch target sizing, gesture design, thumb-zone layout optimisation, iOS and Android platform convention adherence, offline state handling, and mobile-specific interaction design.", Domain: "mobile UX design", Artifacts: []string{"touch target size specification", "gesture definition and conflict map", "thumb-zone layout analysis", "iOS and Android guideline compliance check", "offline and degraded state design", "mobile interaction pattern inventory"}, Risks: []string{"touch target below 44x44pt or 48x48dp minimum causing missed taps", "custom gesture conflicting with system or browser gesture", "primary action placed outside comfortable single-handed thumb reach", "Android and iOS pattern inconsistency confusing cross-platform users", "no offline or degraded connectivity state designed", "bottom navigation unreachable on large-screen devices without adaptation"}, Signals: []string{"touch target size audit against platform guidelines", "gesture conflict test on real device", "thumb-zone reachability heatmap", "platform guideline compliance checklist", "offline mode connectivity simulation test", "device and OS version coverage report"}, Outputs: []string{"mobile UX review", "touch target and gesture findings", "thumb-zone layout recommendations", "iOS and Android platform convention gap list", "offline state design assessment"}},
	{Name: "design-handoff-reviewer", Description: "Review design-to-development handoff quality covering Figma specification completeness, design token export accuracy, component annotation, asset export readiness, and implementation fidelity against design intent.", Domain: "design handoff and design-to-code fidelity", Artifacts: []string{"Figma or Sketch design specification file", "design token export", "component annotation and state documentation", "asset export specification", "visual regression or implementation review", "design-to-code parity check"}, Risks: []string{"specification missing spacing, interactive states, or edge cases the developer needs", "design token not exported or mismatched with codebase token values", "asset exported at wrong resolution, format, or missing dark mode variant", "component annotation absent for error, hover, focus, disabled, or loading state", "implementation diverging from design without documented rationale or design approval", "handoff occurring too late for developer to raise questions before build begins"}, Signals: []string{"specification completeness checklist", "token parity diff between design tool and codebase", "asset export validation for all required formats and resolutions", "interactive state coverage audit", "developer question and clarification log", "visual regression test result"}, Outputs: []string{"design handoff review", "specification gap list", "token and asset export findings", "interactive state coverage report", "implementation fidelity notes", "handoff process improvement recommendations"}},
	{Name: "java-reviewer", Description: "Review Java code for Effective Java best practices, SOLID principles, idiomatic use of Streams, Optional, records, sealed classes, pattern matching, generics, exception handling, concurrency patterns, and JVM memory model correctness.", Domain: "Java development best practices", Artifacts: []string{"Java source file or diff", "exception handling strategy", "concurrency construct", "generic type declaration", "Stream or Optional usage", "JVM configuration or tuning parameter"}, Risks: []string{"checked exception swallowed or wrapped without context", "raw type or unchecked cast hiding type safety violation", "mutable state shared across threads without synchronisation", "Stream misuse causing unnecessary boxing or intermediate collection", "Optional used as method parameter or field instead of return value", "equals and hashCode contract broken causing incorrect Collection behaviour"}, Signals: []string{"static analysis output from SpotBugs or ErrorProne", "compiler warning on unchecked cast or raw type", "concurrent test with race condition or deadlock detector", "Stream pipeline profiling result", "code review comment on Effective Java item", "JVM heap or GC log"}, Outputs: []string{"Java code review findings", "concurrency and thread-safety recommendations", "type safety and generics gap list", "Stream and Optional usage corrections", "exception handling improvement plan", "JVM and memory model notes"}},
	{Name: "spring-boot-reviewer", Description: "Review Spring Boot applications for dependency injection correctness, bean lifecycle, transaction boundary design, JPA and Hibernate pitfalls, Spring Security configuration, externalised configuration, actuator security, and Spring Boot 3.x migration readiness.", Domain: "Spring Boot best practices", Artifacts: []string{"Spring Boot application configuration", "bean definition and injection point", "transaction boundary annotation", "JPA entity and repository", "Spring Security configuration", "actuator endpoint exposure setting"}, Risks: []string{"circular bean dependency causing startup failure", "@Transactional on private method or wrong proxy type being bypassed", "N+1 query from lazy-loaded JPA association without fetch join", "Spring Security permitting all by default after configuration mistake", "sensitive actuator endpoint exposed without authentication", "application secret stored in application.properties committed to version control"}, Signals: []string{"Spring Boot startup log and bean graph", "transaction rollback or commit log", "SQL query count from Hibernate statistics", "Spring Security filter chain audit", "actuator endpoint list with access control", "externalized configuration review against 12-factor principles"}, Outputs: []string{"Spring Boot review findings", "bean and transaction boundary recommendations", "JPA and query optimisation gap list", "Spring Security hardening notes", "actuator and configuration security findings", "Spring Boot 3.x migration readiness assessment"}},
	{Name: "angular-reviewer", Description: "Review Angular applications for component architecture, smart and dumb component separation, RxJS subscription management, change detection strategy, lazy loading, Angular 17 standalone components and signals, reactive forms, and NgRx or state management correctness.", Domain: "Angular best practices", Artifacts: []string{"Angular component and template", "RxJS observable and subscription", "Angular module or standalone component configuration", "reactive form definition", "lazy-loaded route configuration", "state management store or signal"}, Risks: []string{"memory leak from unsubscribed Observable in component lifecycle", "change detection running too frequently due to missing OnPush strategy", "business logic embedded in component instead of service", "form validation missing on submit allowing invalid data submission", "eager loading of feature module increasing initial bundle size", "signal or store mutation bypassing immutability causing unpredictable state"}, Signals: []string{"Angular DevTools change detection cycle profiler", "RxJS subscription audit with takeUntilDestroyed or async pipe", "bundle size analysis with source-map-explorer", "reactive form validator coverage", "lazy route chunk boundary in build output", "NgRx state trace or signal debugger"}, Outputs: []string{"Angular code review findings", "RxJS subscription and memory leak recommendations", "change detection optimisation plan", "component architecture and responsibility findings", "lazy loading and bundle size notes", "state management correctness assessment"}},
	{Name: "react-reviewer", Description: "Review React applications for hooks correctness, component composition, prop drilling elimination, memoisation discipline, Context API usage, concurrent features, Server Components, accessibility, and React 18 best practices.", Domain: "React best practices", Artifacts: []string{"React component and JSX", "hook implementation and dependency array", "Context or state management definition", "React.memo or useMemo or useCallback usage", "React Server Component boundary", "accessibility attribute and ARIA usage in JSX"}, Risks: []string{"stale closure in useEffect dependency array causing incorrect behaviour", "unnecessary re-render from object or array literal created inline in JSX", "useEffect used for data fetching without cleanup or cancellation", "Context triggering full subtree re-render on every value change", "React.memo or useMemo applied without profiling confirming render cost", "key prop using array index causing DOM reconciliation errors on list reorder"}, Signals: []string{"React DevTools Profiler flame graph", "ESLint react-hooks plugin exhaustive-deps warning", "render count measurement with Why Did You Render", "bundle analysis with Webpack Bundle Analyzer or Vite rollup output", "Accessibility audit with axe-core or React Testing Library queries", "React Server Component serialisation error or client boundary audit"}, Outputs: []string{"React code review findings", "hooks correctness and dependency array recommendations", "re-render and memoisation optimisation notes", "Context and state management findings", "Server Component boundary assessment", "accessibility gap list"}},
	{Name: "typescript-reviewer", Description: "Review TypeScript code for strict mode compliance, type narrowing correctness, discriminated union usage, generic constraints, utility type application, declaration file accuracy, satisfies operator usage, and compiler option alignment with project risk tolerance.", Domain: "TypeScript best practices", Artifacts: []string{"TypeScript source file or diff", "tsconfig.json compiler options", "generic type definition", "discriminated union or type guard", "declaration file or module augmentation", "utility type usage"}, Risks: []string{"any type disabling type checking and spreading unsafely through codebase", "type assertion with as bypassing compiler checks without runtime validation", "non-null assertion operator used without confirming value presence", "generic constraint too loose allowing unintended types at call site", "strict mode disabled in tsconfig allowing implicit any and loose null checks", "declaration file out of sync with runtime module exports"}, Signals: []string{"tsc output with strict and noUncheckedIndexedAccess flags", "ESLint @typescript-eslint no-explicit-any and no-non-null-assertion warnings", "type coverage report from type-coverage tool", "generic usage at call site with inferred type trace", "module resolution audit with moduleResolution and paths configuration", "declaration file parity check with runtime exports"}, Outputs: []string{"TypeScript review findings", "strict mode and compiler option recommendations", "type safety gap list", "generic and utility type improvement plan", "declaration file accuracy findings", "any and assertion cleanup recommendations"}},
	{Name: "javascript-reviewer", Description: "Review modern JavaScript for ES2022 and later idiomatic usage, async and Promise correctness, module system consistency, event loop and microtask awareness, prototype chain and closure correctness, and runtime performance patterns.", Domain: "modern JavaScript best practices", Artifacts: []string{"JavaScript source file or diff", "async function and Promise chain", "ES module or CommonJS module definition", "closure and scope usage", "prototype or class definition", "runtime performance measurement"}, Risks: []string{"unhandled Promise rejection causing silent failure", "async function called without await discarding returned Promise", "mixing ES module import and CommonJS require causing resolution failure", "closure over loop variable capturing reference instead of value", "blocking main thread with synchronous computation delaying rendering or I/O", "prototype mutation causing unexpected behaviour across module boundaries"}, Signals: []string{"ESLint no-floating-promises and no-async-promise-executor warnings", "Node.js unhandledRejection event or browser console unhandled rejection", "module bundler resolution warning", "performance profile showing long task on main thread", "memory leak detection from retained closure reference", "code path analysis for unreachable Promise rejection handler"}, Outputs: []string{"JavaScript review findings", "async and Promise correctness recommendations", "module system consistency findings", "closure and scope gap list", "main thread performance notes", "ES2022 and later idiomatic usage improvements"}},
	{Name: "software-design-patterns-reviewer", Description: "Review software design for GoF pattern application, SOLID principles, DRY, KISS, and YAGNI adherence, Clean Architecture and Hexagonal Architecture boundaries, DDD building blocks, CQRS, and coupling and cohesion quality.", Domain: "software design patterns and principles", Artifacts: []string{"class and module design", "dependency graph", "domain model or aggregate", "CQRS command and query separation", "architectural layer boundary", "design pattern implementation"}, Risks: []string{"God class or God module accumulating unrelated responsibilities violating SRP", "anemic domain model with logic in services instead of entities", "Singleton misused creating hidden global state and coupling", "abstraction added for anticipated future use violating YAGNI", "architectural layer boundary crossed directly creating tight coupling", "CQRS write model leaking into read path causing inconsistent state"}, Signals: []string{"dependency graph cycle detection", "class responsibility and method count audit", "domain logic location analysis", "pattern identification in code review", "architectural fitness function or ArchUnit test result", "coupling and cohesion metric from static analysis"}, Outputs: []string{"software design review", "SOLID violation findings", "design pattern misuse and missing pattern recommendations", "Clean Architecture boundary gap list", "DDD aggregate and domain model findings", "coupling and cohesion improvement plan"}},
	{Name: "code-quality-reviewer", Description: "Review code quality for readability, naming conventions, function and class size, cyclomatic and cognitive complexity, code smells, technical debt quantification, refactoring opportunities, and test coverage adequacy.", Domain: "code quality and maintainability", Artifacts: []string{"source code file or diff", "complexity metric report", "test coverage report", "static analysis output", "code smell catalogue", "technical debt backlog"}, Risks: []string{"function exceeding single responsibility with high cyclomatic complexity", "naming failing to communicate intent requiring comment to explain", "magic number or string literal without named constant", "code duplication spreading bug fix surface across multiple locations", "test coverage metric passing while critical paths remain untested", "technical debt accepted without owner expiry or remediation plan"}, Signals: []string{"cyclomatic and cognitive complexity score per function", "test coverage percentage with branch and mutation coverage", "duplication percentage from static analysis", "naming clarity review in code walkthrough", "code smell category count from SonarQube or equivalent", "technical debt ratio and remediation estimate"}, Outputs: []string{"code quality review", "complexity and naming findings", "duplication and dead code list", "test coverage gap assessment", "refactoring priority recommendations", "technical debt quantification and plan"}},
	// Testing disciplines
	{Name: "unit-testing-reviewer", Description: "Review unit tests for correctness, isolation, test double usage, Arrange-Act-Assert structure, naming conventions, boundary and negative case coverage, and avoidance of anti-patterns such as testing implementation details or writing tests that always pass.", Domain: "unit testing best practices", Artifacts: []string{"unit test file or suite", "mock or stub or fake implementation", "test coverage report with branch data", "test naming convention guide", "assertion library configuration", "test runner output"}, Risks: []string{"test asserting on internal implementation detail coupling test to refactoring", "mock returning incorrect behaviour making test pass for wrong reason", "test missing boundary values leaving off-by-one defect undetected", "single test covering multiple behaviours making failure diagnosis difficult", "test with no assertion or trivially passing assertion giving false confidence", "flaky test due to time dependency or shared mutable state causing intermittent CI failure"}, Signals: []string{"line and branch coverage report", "mutation testing score from Pitest or Stryker", "test naming clarity audit against should-when or given-when-then pattern", "flaky test detection from repeated CI run history", "mock verification call count and argument assertion review", "test isolation check for shared state between test cases"}, Outputs: []string{"unit test review findings", "test isolation and mock correctness recommendations", "boundary and negative case gap list", "anti-pattern and false-confidence finding list", "naming convention improvement notes", "mutation coverage improvement plan"}},
	{Name: "integration-testing-reviewer", Description: "Review integration tests for correct component boundary verification, database and external service interaction, transaction management, test data setup and teardown, container or embedded service usage, and realistic failure scenario coverage.", Domain: "integration testing best practices", Artifacts: []string{"integration test file or suite", "database migration or seed script", "test container or embedded service configuration", "external service stub or WireMock mapping", "transaction boundary test", "test data lifecycle management"}, Risks: []string{"integration test sharing database state causing order-dependent test failure", "external service stubbed incorrectly masking real integration incompatibility", "test not rolling back transaction leaving database in dirty state for subsequent tests", "missing failure scenario test for network timeout or external service error", "integration test running too slowly from unnecessary real service call", "test data hardcoded with production-like sensitive values"}, Signals: []string{"test execution order independence check", "database state inspection before and after each test", "WireMock or stub mapping parity with real service contract", "test execution time per suite", "transaction rollback or cleanup log", "external service error injection test result"}, Outputs: []string{"integration test review findings", "test isolation and state management recommendations", "stub and contract parity gap list", "failure scenario coverage assessment", "test data lifecycle improvement plan", "performance and scope optimisation notes"}},
	{Name: "e2e-testing-reviewer", Description: "Review end-to-end tests written with Playwright, Cypress, or Selenium for user journey completeness, selector stability, test isolation, environment parity, retry and flakiness handling, and meaningful assertion coverage across critical user flows.", Domain: "end-to-end testing best practices", Artifacts: []string{"E2E test file or suite", "page object or component abstraction", "test environment configuration", "CI pipeline E2E stage definition", "visual regression snapshot", "test run report with failure log"}, Risks: []string{"test relying on fragile CSS class or position-based selector breaking on UI refactor", "missing page object abstraction coupling test code to raw DOM structure", "test not waiting correctly for async operation causing intermittent failure", "E2E test suite covering happy path only missing critical error and edge flows", "test running against non-production-parity environment masking deployment issues", "sensitive test credential or fixture data committed to repository"}, Signals: []string{"selector audit for data-testid or aria-label versus fragile class selectors", "page object coverage of all tested views", "CI retry count and flakiness rate per test", "user journey coverage map against product requirements", "environment diff between E2E and production configuration", "test execution time and parallelisation strategy"}, Outputs: []string{"E2E test review findings", "selector stability and page object recommendations", "flakiness root cause analysis", "user journey coverage gap list", "environment parity findings", "CI pipeline optimisation notes"}},
	{Name: "bdd-reviewer", Description: "Review Behaviour-Driven Development scenarios written in Gherkin for clarity, appropriate abstraction level, step definition reuse, living documentation quality, three amigos alignment, and correct mapping from business requirement to executable specification.", Domain: "BDD and Gherkin best practices", Artifacts: []string{"Gherkin feature file", "step definition implementation", "BDD framework configuration (Cucumber, SpecFlow, Behave)", "scenario outline with examples table", "living documentation report", "three amigos session output or acceptance criteria"}, Risks: []string{"scenario written at implementation level exposing UI detail instead of business behaviour", "step definition doing too much or too little causing brittle and hard-to-reuse steps", "Given-When-Then misused with multiple Whens in one scenario testing multiple behaviours", "scenario outline example table incomplete missing boundary and negative values", "Gherkin not understood by non-technical stakeholder defeating living documentation purpose", "step definition duplicated across feature files diverging over time"}, Signals: []string{"step definition reuse ratio across feature files", "stakeholder readability review of feature file without technical explanation", "scenario count per feature and step per scenario audit", "acceptance criteria traceability from feature file to business requirement", "step definition coupling to UI selector or internal identifier audit", "living documentation generation from Cucumber or SpecFlow report"}, Outputs: []string{"BDD scenario review findings", "Gherkin clarity and abstraction recommendations", "step definition reuse and duplication findings", "three amigos alignment gap assessment", "living documentation quality notes", "scenario outline completeness improvement plan"}},
	{Name: "tdd-reviewer", Description: "Review Test-Driven Development discipline for red-green-refactor cycle adherence, test-first commit history, emergent design quality, incremental step size, coverage as outcome not goal, and use of TDD to drive API and interface design.", Domain: "TDD best practices and discipline", Artifacts: []string{"commit history showing test-first progression", "failing test added before implementation", "refactoring step with no behaviour change", "API or interface design driven by test", "coverage report as outcome metric", "design quality metric before and after TDD cycle"}, Risks: []string{"implementation committed before failing test violating test-first discipline", "refactoring step changing behaviour instead of structure invalidating green phase", "test written to pass existing code retrofitting tests without design benefit", "TDD step size too large making red-to-green transition non-trivial", "coverage chased as target rather than emerging from feature completeness", "API design not improved by test-first approach due to test written after contract fixed"}, Signals: []string{"commit order analysis for test-before-implementation pattern", "diff of test commit versus implementation commit for same feature", "API shape change history driven by test feedback", "cyclomatic complexity trend across TDD cycles", "refactoring commit purity check for behaviour equivalence", "coverage trend correlated with feature delivery not metric target"}, Outputs: []string{"TDD discipline review findings", "red-green-refactor cycle adherence assessment", "test-first commit history analysis", "emergent design quality observations", "API and interface design improvement driven by TDD", "recommendations for TDD step size calibration"}},
	{Name: "api-testing-reviewer", Description: "Review API tests for REST and GraphQL endpoints covering request and response schema validation, status code correctness, authentication and authorisation boundary testing, error response format, idempotency, and contract test alignment with consumers.", Domain: "API testing best practices", Artifacts: []string{"API test file or Postman or Bruno collection", "OpenAPI or GraphQL schema", "contract test definition", "authentication and authorisation test case", "error response format specification", "idempotency and retry test case"}, Risks: []string{"test asserting only on HTTP 200 missing error path and status code correctness", "authentication bypass not tested allowing unauthorised access in production", "response schema not validated allowing breaking contract change to reach consumer", "idempotency not tested for POST or PATCH endpoint that should be idempotent", "test using hardcoded production URL or credential committed to repository", "GraphQL query not tested for N+1 resolver issue under load"}, Signals: []string{"status code coverage across 2xx, 4xx, and 5xx paths", "OpenAPI schema validation report against actual responses", "authentication and authorisation boundary test matrix", "contract test result from Pact or Spring Cloud Contract", "idempotency test with repeated request and state comparison", "error response structure conformance check"}, Outputs: []string{"API test review findings", "status code and error path coverage gap list", "schema and contract conformance findings", "authentication and authorisation test gap assessment", "idempotency and retry behaviour recommendations", "GraphQL resolver and performance test notes"}},
	{Name: "performance-testing-reviewer", Description: "Review performance and load tests written with k6, JMeter, Gatling, or Locust for realistic user scenario modelling, ramp-up strategy, SLO threshold definition, bottleneck identification, environment parity, and result baseline comparison.", Domain: "performance and load testing best practices", Artifacts: []string{"performance test script or scenario", "load profile and ramp-up configuration", "SLO and SLA threshold definition", "performance test result report", "baseline comparison data", "environment specification matching production scale"}, Risks: []string{"load test using unrealistic think time producing traffic spike not representative of real users", "SLO threshold not defined causing test to pass without meaningful performance assertion", "environment under-provisioned compared to production invalidating results", "test not isolating component under test mixing network latency into application metric", "memory leak not detected due to insufficient soak test duration", "result baseline not recorded preventing regression detection across releases"}, Signals: []string{"p50 p95 p99 and p999 latency distribution under target load", "throughput and error rate at each ramp-up stage", "CPU memory and I/O utilisation on application and database tier", "SLO breach count and breach percentage", "comparison delta against recorded baseline", "JVM or runtime GC pause and heap pressure under sustained load"}, Outputs: []string{"performance test review findings", "load profile and ramp-up strategy recommendations", "SLO threshold gap list", "bottleneck and saturation point analysis", "environment parity assessment", "baseline and regression tracking improvement plan"}},
	{Name: "contract-testing-reviewer", Description: "Review consumer-driven contract tests using Pact or Spring Cloud Contract for provider-consumer parity, contract version management, broker integration, verification pipeline placement, and handling of breaking versus non-breaking schema changes.", Domain: "contract testing and consumer-driven contracts", Artifacts: []string{"Pact or Spring Cloud Contract definition", "consumer test generating contract", "provider verification test", "Pact broker or contract registry configuration", "CI pipeline contract verification stage", "breaking change analysis between contract versions"}, Risks: []string{"contract not verified against real provider allowing consumer assumption to diverge", "provider verification running only in consumer repository missing provider-side regression", "contract breaking change merged without consumer notification or version negotiation", "Pact broker not integrated into CI preventing automatic verification on provider change", "optional field removed from provider response breaking consumer that assumed presence", "consumer test mocking provider behaviour incorrectly producing invalid contract"}, Signals: []string{"Pact broker verification status for each consumer-provider pair", "contract version history and compatibility matrix", "provider verification test result in provider CI pipeline", "breaking versus compatible change classification for schema diff", "consumer notification flow on provider contract change", "Pact pending and work in progress flag usage audit"}, Outputs: []string{"contract test review findings", "consumer-provider parity gap list", "broker integration and pipeline placement recommendations", "breaking change handling process assessment", "version management improvement plan", "contract accuracy and mock fidelity findings"}},
	{Name: "mutation-testing-reviewer", Description: "Review mutation testing configuration and results from Pitest, Stryker, or mutmut to assess test suite effectiveness, identify surviving mutants indicating undertested logic, prioritise high-value mutation operators, and set meaningful mutation score thresholds.", Domain: "mutation testing and test suite effectiveness", Artifacts: []string{"mutation testing configuration file", "mutation test result report", "surviving mutant list with code location", "mutation score per module or class", "mutation operator selection", "CI pipeline mutation gate configuration"}, Risks: []string{"surviving mutant in critical business logic indicating logic branch not exercised by any test", "mutation score threshold set too low allowing severely undertested module to pass gate", "mutation testing excluded from CI due to long runtime hiding test suite degradation over time", "equivalent mutant counted as surviving inflating apparent weakness in test suite", "mutation operator set too narrow missing important fault classes", "mutation results not reviewed causing accumulated surviving mutants to grow unnoticed"}, Signals: []string{"mutation score percentage per class and module", "surviving mutant count in business-critical code paths", "mutation test run time and CI integration feasibility", "equivalent mutant identification and exclusion rate", "mutation score trend over recent releases", "operator coverage across arithmetic relational logical and boundary shift mutations"}, Outputs: []string{"mutation testing review findings", "surviving mutant priority list by business criticality", "mutation score threshold recommendations", "CI integration feasibility assessment", "mutation operator coverage gap list", "test suite effectiveness improvement plan"}},
	{Name: "snapshot-testing-reviewer", Description: "Review snapshot and visual regression tests for meaningful snapshot content, stale or overly large snapshot management, intentional update discipline, visual baseline accuracy, and integration of tools such as Percy, Chromatic, or jest-image-snapshot.", Domain: "snapshot and visual regression testing", Artifacts: []string{"snapshot file or visual baseline image", "snapshot test definition", "visual regression tool configuration (Percy, Chromatic, jest-image-snapshot)", "snapshot update history and commit log", "CI visual diff review step", "component or page screenshot under test"}, Risks: []string{"snapshot updated blindly without reviewing diff masking unintentional visual regression", "snapshot file too large capturing unrelated markup making meaningful diff impossible", "visual baseline captured from non-production-parity environment embedding environment artefact", "snapshot test covering dynamic content (timestamp, random id) causing perpetual mismatch", "snapshot never updated after intentional design change causing permanent CI failure", "visual regression review step skipped in PR flow allowing visual defect to merge"}, Signals: []string{"snapshot size and churn rate per component", "snapshot update frequency correlated with intentional design change versus accident", "dynamic content exclusion or masking configuration audit", "visual diff approval rate and review time in PR", "baseline environment parity check", "false positive rate from dynamic or third-party content in snapshot"}, Outputs: []string{"snapshot test review findings", "snapshot scope and size recommendations", "dynamic content masking improvement plan", "visual regression review process assessment", "baseline parity and environment findings", "update discipline and PR flow recommendations"}},
	{Name: "property-based-testing-reviewer", Description: "Review property-based tests written with fast-check, QuickCheck, Hypothesis, or jqwik for meaningful property definition, generator design, shrinking effectiveness, stateful model testing, integration with unit test suite, and edge case discovery validation.", Domain: "property-based testing best practices", Artifacts: []string{"property-based test definition", "custom generator or arbitrary implementation", "shrinking configuration and output", "stateful model test", "property-based test runner configuration", "discovered edge case report from previous test run"}, Risks: []string{"property defined as round-trip identity test only missing meaningful invariant verification", "generator producing values too narrowly failing to explore interesting edge cases", "shrinking not configured or disabled making counterexample diagnosis difficult", "stateful model not matching real state machine allowing invalid command sequences", "property-based test run count set too low reducing probability of finding rare defect", "discovered counterexample not added to unit test suite as regression"}, Signals: []string{"property definition review for meaningful invariant versus tautology", "generator value distribution audit for boundary and edge values", "shrunk counterexample quality and reproducibility", "run count and seed configuration", "stateful model command sequence coverage", "correlation between property test runs and discovered regression"}, Outputs: []string{"property-based test review findings", "property definition and invariant quality assessment", "generator coverage and distribution recommendations", "shrinking and reproducibility improvement plan", "stateful model accuracy findings", "integration with regression suite recommendations"}},
	{Name: "load-testing-reviewer", Description: "Review load, stress, soak, spike, and chaos testing strategies for completeness of non-functional requirement coverage, realistic traffic modelling, infrastructure failure injection, auto-scaling verification, and production incident prevention.", Domain: "load stress and resilience testing", Artifacts: []string{"load test plan and scenario", "stress test ramp-up profile", "soak test duration and monitoring configuration", "spike test traffic injection definition", "chaos experiment definition", "infrastructure auto-scaling or circuit breaker configuration"}, Risks: []string{"stress test stopping at expected peak load not discovering breaking point above capacity", "soak test too short to reveal slow memory leak or connection pool exhaustion", "spike test not matching real traffic spike shape causing unrealistic result", "chaos experiment run in production without feature flag or blast radius control", "auto-scaling not verified under load allowing cold-start latency to breach SLO", "test not monitoring all relevant tiers leaving database or cache saturation undetected"}, Signals: []string{"breaking point under stress test in requests per second or concurrent users", "resource exhaustion signal during soak test over minimum 4 hour window", "recovery time after spike and auto-scaling event", "chaos experiment steady state hypothesis verification result", "SLO breach rate during each test type", "cross-tier resource utilisation including application database cache and network"}, Outputs: []string{"load and resilience test review findings", "breaking point and capacity headroom assessment", "soak and memory leak risk findings", "chaos experiment blast radius and safety recommendations", "auto-scaling and recovery behaviour assessment", "non-functional requirement coverage gap list"}},
}

var paymentSkillSeeds = []additionalSkillSeed{
	{Name: "payment-integration-engineer", Description: "Design and implement provider-neutral payment integrations with explicit state machines, amount and currency handling, idempotency, asynchronous confirmation, failure recovery, and auditable order linkage.", Domain: "payment integration engineering", Artifacts: []string{"payment state machine", "provider adapter", "idempotency-key strategy", "order-to-payment mapping", "amount and currency value object", "failure recovery runbook"}, Risks: []string{"duplicate charge", "incorrect currency or minor-unit conversion", "client-authoritative payment status", "ambiguous timeout outcome", "provider lock-in", "missing order-to-payment trace"}, Signals: []string{"sandbox transaction trace", "idempotent retry test", "state-transition test", "provider response and webhook correlation", "ledger entry", "failure-injection result"}, Outputs: []string{"payment integration design", "provider adapter implementation plan", "state-transition matrix", "retry and recovery policy", "integration test suite"}},
	{Name: "payment-security-reviewer", Description: "Review payment data flows, tokenization, credential boundaries, PCI DSS scope reduction, authorization, sensitive logging, encryption, and abuse-resistant checkout design.", Domain: "payment security", Artifacts: []string{"cardholder-data flow diagram", "tokenization boundary", "payment credential policy", "checkout authorization rule", "sensitive-log inventory", "PCI DSS scope assessment"}, Risks: []string{"PAN or CVV storage", "secret-key exposure", "payment amount tampering", "broken tenant authorization", "sensitive payload logging", "unnecessary PCI DSS scope"}, Signals: []string{"tokenized test payload", "secret-manager reference", "authorization test", "log-redaction test", "data-flow review", "security control evidence"}, Outputs: []string{"payment security findings", "cardholder-data flow review", "PCI scope-reduction recommendations", "credential boundary assessment", "abuse-case test plan"}},
	{Name: "payment-webhook-reviewer", Description: "Review payment webhook endpoints for signature verification on the raw payload, replay and duplicate handling, event ordering, durable queues, fast acknowledgement, retries, and reconciliation.", Domain: "payment webhook processing", Artifacts: []string{"webhook endpoint", "signature verification routine", "event deduplication store", "durable event queue", "event-ordering policy", "webhook replay test"}, Risks: []string{"forged webhook", "replayed event", "duplicate fulfillment", "out-of-order state regression", "acknowledgement timeout", "dropped event"}, Signals: []string{"invalid-signature test", "duplicate-event test", "out-of-order event test", "queue persistence evidence", "provider delivery log", "reconciliation result"}, Outputs: []string{"webhook security review", "event-processing state model", "deduplication and ordering findings", "retry test matrix", "operational recovery runbook"}},
	{Name: "payment-flow-tester", Description: "Design and review payment tests across authorization, capture, settlement, cancellation, refund, timeout, decline, authentication challenge, webhook delay, and provider outage paths.", Domain: "payment flow testing", Artifacts: []string{"payment scenario matrix", "sandbox fixture", "provider test clock or simulator", "failure-injection test", "webhook fixture", "ledger assertion"}, Risks: []string{"happy-path-only coverage", "untested ambiguous timeout", "missing authentication challenge", "double-submit regression", "sandbox-production mismatch", "incorrect ledger side effect"}, Signals: []string{"authorization and capture trace", "decline-code assertion", "delayed-webhook test", "idempotent retry result", "refund test", "state-machine coverage report"}, Outputs: []string{"payment test strategy", "provider sandbox scenario suite", "negative-path coverage matrix", "state-transition coverage report", "production smoke-test plan"}},
	{Name: "refund-dispute-handler", Description: "Guide controlled refund, reversal, chargeback, and dispute workflows with eligibility checks, evidence packaging, deadlines, segregation of duties, customer communication, and explicit approval.", Domain: "refund and dispute handling", Artifacts: []string{"refund request", "captured-payment record", "dispute case", "evidence package", "approval record", "customer communication"}, Risks: []string{"refund above captured amount", "duplicate refund", "missed dispute deadline", "unsupported evidence submission", "approval bypass", "unsafe autonomous money movement"}, Signals: []string{"refund eligibility decision", "provider refund status", "dispute due date", "approver identity", "ledger adjustment", "case audit trail"}, Outputs: []string{"refund decision record", "dispute response package", "approval checklist", "customer communication draft", "case reconciliation note"}},
	{Name: "payment-reconciliation-reviewer", Description: "Reconcile orders, provider transactions, captures, refunds, fees, chargebacks, settlements, payouts, and ledger entries with deterministic matching and exception ownership.", Domain: "payment reconciliation", Artifacts: []string{"internal payment ledger", "provider balance transaction export", "settlement report", "payout report", "fee record", "reconciliation exception queue"}, Risks: []string{"unmatched payment", "missing fee", "duplicate ledger entry", "payout variance", "currency mismatch", "unowned reconciliation break"}, Signals: []string{"provider transaction identifier", "merchant order reference", "gross-net-fee equation", "payout bank reference", "daily reconciliation total", "exception closure evidence"}, Outputs: []string{"reconciliation report", "unmatched-item queue", "settlement variance analysis", "ledger correction proposal", "control evidence package"}},
	{Name: "subscription-billing-engineer", Description: "Design and implement subscription billing for plans, trials, invoices, proration, usage, renewals, dunning, cancellation, entitlements, and asynchronous billing events.", Domain: "subscription billing", Artifacts: []string{"subscription state machine", "plan and price catalogue", "invoice lifecycle", "proration rule", "usage-meter record", "dunning workflow"}, Risks: []string{"incorrect proration", "entitlement-payment drift", "duplicate renewal", "unbounded retry or dunning", "retroactive price change", "missing cancellation effective date"}, Signals: []string{"billing-cycle test", "invoice calculation", "test-clock scenario", "usage aggregation result", "renewal webhook", "entitlement reconciliation"}, Outputs: []string{"subscription billing design", "plan-change matrix", "invoice and proration tests", "dunning policy", "entitlement synchronization plan"}},
	{Name: "sca-3ds-reviewer", Description: "Review Strong Customer Authentication and 3-D Secure flows for challenge and frictionless paths, exemptions, liability signals, fallback behavior, accessibility, and payment-state correctness.", Domain: "SCA and 3-D Secure", Artifacts: []string{"3DS authentication flow", "challenge return handler", "exemption policy", "authentication result", "liability-shift signal", "fallback UX"}, Risks: []string{"SCA bypass", "incorrect exemption use", "challenge callback forgery", "authentication-payment state mismatch", "abandoned challenge loop", "inaccessible challenge flow"}, Signals: []string{"3DS test card result", "challenge completion trace", "exemption decision record", "authentication status mapping", "fallback test", "conversion and failure metric"}, Outputs: []string{"SCA and 3DS review", "challenge-flow test matrix", "exemption governance findings", "state-mapping corrections", "fallback and accessibility recommendations"}},
	{Name: "payment-fraud-risk-reviewer", Description: "Review payment fraud controls, velocity and behavioral signals, risk rules, step-up actions, manual review, false-positive impact, account takeover, and post-transaction feedback loops.", Domain: "payment fraud risk", Artifacts: []string{"fraud rule set", "risk score", "velocity control", "manual-review queue", "step-up policy", "chargeback feedback dataset"}, Risks: []string{"card testing", "account takeover", "friendly fraud", "rule evasion", "false-positive customer blocking", "biased or opaque decisioning"}, Signals: []string{"authorization decline trend", "chargeback ratio", "rule hit rate", "manual-review outcome", "device and account signal", "false-positive measurement"}, Outputs: []string{"fraud risk review", "rule effectiveness report", "manual-review design", "step-up recommendation", "false-positive and fairness assessment"}},
	{Name: "payment-observability-reviewer", Description: "Review payment telemetry for conversion, authorization, authentication, capture, webhook, refund, reconciliation, latency, provider health, and customer-impact signals without leaking sensitive data.", Domain: "payment observability", Artifacts: []string{"payment funnel dashboard", "authorization metric", "webhook backlog alert", "provider latency SLI", "reconciliation exception metric", "redacted payment trace"}, Risks: []string{"silent conversion drop", "provider outage blind spot", "webhook backlog", "high-cardinality identifiers", "PAN or customer data in telemetry", "unactionable payment alert"}, Signals: []string{"authorization success rate", "checkout conversion rate", "provider error code distribution", "webhook processing lag", "reconciliation break count", "redaction test"}, Outputs: []string{"payment observability review", "payment SLI and alert catalogue", "funnel dashboard recommendations", "provider-health runbook", "telemetry privacy findings"}},
	{Name: "payment-provider-migration-reviewer", Description: "Plan and review payment-provider migrations with capability parity, token portability, dual processing, routing, reconciliation, rollback, customer impact, and decommissioning controls.", Domain: "payment provider migration", Artifacts: []string{"provider capability matrix", "token migration plan", "dual-processing design", "traffic-routing rule", "rollback plan", "provider decommission checklist"}, Risks: []string{"non-portable payment token", "double charge during cutover", "capability regression", "settlement reconciliation gap", "unsafe big-bang migration", "premature provider shutdown"}, Signals: []string{"sandbox parity test", "shadow or canary result", "token migration evidence", "route allocation metric", "dual-ledger reconciliation", "rollback exercise"}, Outputs: []string{"provider migration plan", "capability gap matrix", "cutover and rollback runbook", "dual-processing controls", "decommission readiness decision"}},
	{Name: "payment-compliance-reviewer", Description: "Review payment controls and evidence for PCI DSS, privacy, SCA obligations, retention, access, auditability, outsourcing, and policy exceptions without presenting legal conclusions.", Domain: "payment compliance", Artifacts: []string{"PCI DSS responsibility matrix", "cardholder-data inventory", "SCA control mapping", "retention schedule", "access review", "provider responsibility record"}, Risks: []string{"unscoped cardholder data", "missing service-provider responsibility", "unreviewed privileged access", "retention violation", "unsupported compliance claim", "expired policy exception"}, Signals: []string{"approved scope assessment", "attestation or provider evidence", "access-review record", "data deletion test", "control test result", "exception expiry"}, Outputs: []string{"payment compliance gap report", "responsibility matrix", "evidence request list", "control remediation plan", "residual-risk summary"}},
	{Name: "payment-operations-agent", Description: "Support tightly controlled payment operations such as capture, cancellation, refund, resend, and case lookup using least privilege, explicit human approval, amount limits, idempotency, and immutable audit records.", Domain: "payment operations", Artifacts: []string{"operations request", "human approval", "payment status lookup", "amount and currency confirmation", "idempotency key", "immutable audit event"}, Risks: []string{"unauthorized money movement", "approval replay", "wrong payment target", "amount or currency mutation", "duplicate operation", "live credential exposure"}, Signals: []string{"authenticated requester", "fresh approval token", "before-and-after status", "provider operation identifier", "ledger confirmation", "audit-log correlation ID"}, Outputs: []string{"operation plan", "approval request", "dry-run or status preview", "execution receipt", "post-operation reconciliation record"}},
	{Name: "stripe-integration-engineer", Description: "Design, implement, and review Stripe integrations using PaymentIntents or Checkout, idempotent requests, signed webhooks, Connect or Billing boundaries, API versioning, and test-mode validation.", Domain: "Stripe payment integration", Artifacts: []string{"Stripe PaymentIntent or Checkout Session", "Stripe idempotency key", "Stripe-Signature verification", "Connect account boundary", "Billing subscription or invoice", "Stripe API version configuration"}, Risks: []string{"duplicate PaymentIntent or capture", "unverified Stripe webhook", "secret-key or client-secret exposure", "wrong connected account", "API-version drift", "client-authoritative fulfillment"}, Signals: []string{"Stripe test-mode request log", "PaymentIntent state trace", "signed webhook fixture", "idempotent retry result", "Connect account assertion", "Stripe CLI or sandbox test"}, Outputs: []string{"Stripe integration design", "PaymentIntent or Checkout implementation plan", "webhook verification review", "API-version migration notes", "Stripe test matrix"}},
	{Name: "paypal-integration-engineer", Description: "Design, implement, and review PayPal REST integrations using Orders and Captures, PayPal-Request-Id idempotency, OAuth credential boundaries, verified webhooks, refunds, and sandbox validation.", Domain: "PayPal payment integration", Artifacts: []string{"PayPal Order", "authorization or capture request", "PayPal-Request-Id", "OAuth access-token flow", "webhook signature verification", "PayPal refund record"}, Risks: []string{"duplicate capture or refund", "unverified PayPal webhook", "client-secret exposure", "order-capture state mismatch", "unsupported idempotency assumption", "sandbox-live credential mix-up"}, Signals: []string{"PayPal sandbox order trace", "idempotent capture retry", "verified webhook event", "OAuth scope review", "refund status", "Orders API error simulation"}, Outputs: []string{"PayPal integration design", "Orders and Captures state matrix", "webhook verification review", "idempotency and retry policy", "PayPal sandbox test suite"}},
	{Name: "adyen-integration-engineer", Description: "Design, implement, and review Adyen Checkout integrations using merchant references, idempotency-key requests, HMAC-verified webhooks, asynchronous result handling, captures, refunds, and test-environment validation.", Domain: "Adyen payment integration", Artifacts: []string{"Adyen Checkout payment request", "merchantReference mapping", "Adyen idempotency-key", "HMAC webhook validator", "pspReference state mapping", "capture or refund modification"}, Risks: []string{"duplicate Adyen payment modification", "invalid HMAC acceptance", "merchantReference collision", "out-of-order webhook regression", "regional idempotency assumption", "test-live HMAC key mix-up"}, Signals: []string{"Adyen test payment trace", "HMAC validation result", "eventCode and pspReference deduplication", "idempotent retry response", "capture or refund reconciliation", "Customer Area test webhook"}, Outputs: []string{"Adyen integration design", "Checkout state mapping", "HMAC webhook review", "idempotency and retry policy", "Adyen test matrix"}},
}

type technologySkillDefinition struct {
	Name        string
	Description string
	Domain      string
	Focus       []string
}

var technologySkillDefinitions = []technologySkillDefinition{
	{Name: "golang-reviewer", Description: "Review Go code for idioms, concurrency, context propagation, errors, interfaces, modules, tests, and performance.", Domain: "Go engineering", Focus: []string{"goroutines and channels", "context cancellation", "error wrapping", "interfaces", "modules", "race detection"}},
	{Name: "python-reviewer", Description: "Review Python code for typing, packaging, async behavior, resource safety, tests, security, and maintainability.", Domain: "Python engineering", Focus: []string{"type annotations", "packaging", "asyncio", "context managers", "pytest", "dependency hygiene"}},
	{Name: "ruby-reviewer", Description: "Review Ruby code for idioms, object design, metaprogramming boundaries, Bundler hygiene, tests, and performance.", Domain: "Ruby engineering", Focus: []string{"object design", "blocks and enumerables", "metaprogramming", "Bundler", "RSpec", "runtime performance"}},
	{Name: "rust-reviewer", Description: "Review Rust ownership, lifetimes, unsafe boundaries, concurrency, error handling, Cargo dependencies, and tests.", Domain: "Rust engineering", Focus: []string{"ownership and borrowing", "lifetimes", "unsafe code", "Send and Sync", "Result handling", "Cargo"}},
	{Name: "csharp-reviewer", Description: "Review C# and .NET code for async correctness, dependency injection, resource disposal, nullable types, tests, and performance.", Domain: "C# and .NET engineering", Focus: []string{"async and await", "dependency injection", "IDisposable", "nullable reference types", "LINQ", "NuGet"}},
	{Name: "kotlin-reviewer", Description: "Review Kotlin code for null safety, coroutines, sealed models, Java interop, Gradle configuration, and tests.", Domain: "Kotlin engineering", Focus: []string{"null safety", "coroutines", "sealed classes", "Java interop", "Gradle", "testing"}},
	{Name: "php-reviewer", Description: "Review modern PHP code for type safety, Composer hygiene, framework boundaries, security, tests, and performance.", Domain: "PHP engineering", Focus: []string{"strict types", "Composer", "PSR standards", "request validation", "PHPUnit", "runtime performance"}},
	{Name: "quarkus-reviewer", Description: "Review Quarkus build-time behavior, CDI, reactive paths, native images, configuration, security, and tests.", Domain: "Quarkus applications", Focus: []string{"CDI", "build-time initialization", "Mutiny", "GraalVM native images", "configuration", "Dev Services"}},
	{Name: "vuejs-reviewer", Description: "Review Vue.js Composition API, reactivity, components, state, routing, accessibility, tests, and performance.", Domain: "Vue.js applications", Focus: []string{"Composition API", "reactivity", "Pinia", "Vue Router", "accessibility", "Vitest"}},
	{Name: "nodejs-reviewer", Description: "Review Node.js services for event-loop safety, async behavior, modules, streams, HTTP security, dependencies, and operations.", Domain: "Node.js services", Focus: []string{"event loop", "Promises", "ESM and CommonJS", "streams", "HTTP lifecycle", "npm dependencies"}},
	{Name: "nextjs-reviewer", Description: "Review Next.js routing, Server and Client Components, caching, data fetching, security, deployment, and performance.", Domain: "Next.js applications", Focus: []string{"App Router", "Server Components", "caching", "data fetching", "route handlers", "deployment runtime"}},
	{Name: "django-reviewer", Description: "Review Django models, migrations, ORM usage, authentication, middleware, settings, tests, and deployment safety.", Domain: "Django applications", Focus: []string{"models and migrations", "ORM queries", "authentication", "middleware", "settings", "deployment checks"}},
	{Name: "fastapi-reviewer", Description: "Review FastAPI schemas, dependency injection, async paths, authentication, validation, OpenAPI, tests, and operations.", Domain: "FastAPI services", Focus: []string{"Pydantic models", "dependency injection", "async endpoints", "authentication", "OpenAPI", "ASGI operations"}},
	{Name: "ruby-on-rails-reviewer", Description: "Review Rails models, controllers, jobs, migrations, Active Record behavior, security, tests, and deployment safety.", Domain: "Ruby on Rails applications", Focus: []string{"Active Record", "migrations", "controllers", "background jobs", "security defaults", "RSpec"}},
	{Name: "aws-cloud-reviewer", Description: "Review AWS accounts, IAM, networking, compute, storage, databases, observability, security, cost, and resilience.", Domain: "AWS cloud architecture", Focus: []string{"Organizations and accounts", "IAM", "VPC", "compute", "data services", "CloudTrail and CloudWatch"}},
	{Name: "azure-cloud-reviewer", Description: "Review Azure tenants, subscriptions, identities, networks, compute, data services, Policy, monitoring, cost, and resilience.", Domain: "Azure cloud architecture", Focus: []string{"tenants and subscriptions", "Entra ID and RBAC", "virtual networks", "compute", "data services", "Azure Policy and Monitor"}},
	{Name: "gcp-cloud-reviewer", Description: "Review GCP organizations, projects, IAM, networking, compute, data services, observability, cost, and resilience.", Domain: "Google Cloud architecture", Focus: []string{"organizations and projects", "IAM", "VPC", "compute", "data services", "Cloud Audit Logs and Monitoring"}},
	{Name: "terraform-reviewer", Description: "Review Terraform modules, providers, plans, state, drift, lifecycle, IAM, tests, and safe apply or destroy boundaries.", Domain: "Terraform infrastructure as code", Focus: []string{"modules", "providers", "plan changes", "remote state", "lifecycle", "tests"}},
	{Name: "opentofu-reviewer", Description: "Review OpenTofu modules, providers, plans, state, drift, lifecycle, tests, and safe apply or destroy boundaries.", Domain: "OpenTofu infrastructure as code", Focus: []string{"modules", "providers", "plan changes", "state encryption", "lifecycle", "tests"}},
	{Name: "ansible-reviewer", Description: "Review Ansible inventories, roles, playbooks, idempotency, secrets, privilege escalation, testing, and rollout safety.", Domain: "Ansible automation", Focus: []string{"inventories", "roles", "playbooks", "idempotency", "Ansible Vault", "Molecule"}},
	{Name: "vagrant-reviewer", Description: "Review Vagrant environments, providers, networking, provisioning, shared folders, reproducibility, and isolation.", Domain: "Vagrant development environments", Focus: []string{"Vagrantfile", "providers", "networking", "provisioners", "shared folders", "base boxes"}},
	{Name: "virtualization-reviewer", Description: "Review VM and virtualization architecture, images, isolation, networking, storage, snapshots, patching, and capacity.", Domain: "virtualization platforms", Focus: []string{"hypervisors", "VM images", "isolation", "virtual networking", "storage and snapshots", "capacity"}},
	{Name: "helm-reviewer", Description: "Review Helm charts, templates, values, dependencies, hooks, secrets, schema validation, upgrades, and rollbacks.", Domain: "Helm package management", Focus: []string{"Chart.yaml", "templates", "values schema", "dependencies", "hooks", "upgrade and rollback"}},
}

func init() {
	for _, seed := range additionalSkillSeeds {
		sdlcSkillContent[seed.Name] = generatedAdditionalSkillContent(seed)
	}
	for _, seed := range paymentSkillSeeds {
		sdlcSkillContent[seed.Name] = generatedPaymentSkillContent(seed)
	}
	for _, definition := range technologySkillDefinitions {
		sdlcSkillContent[definition.Name] = generatedTechnologySkillContent(definition)
	}
	for _, entry := range cncf.MustEntries() {
		sdlcSkillContent[entry.SkillName] = generatedCNCFSkillContent(entry)
	}
}

func generatedTechnologySkillContent(definition technologySkillDefinition) skillContent {
	focus := strings.Join(definition.Focus, ", ")
	content := skillContent{
		Purpose: definition.Description,
		When: []string{
			"Source code, configuration, dependencies, tests, deployment artifacts, or runtime behavior in " + definition.Domain + " change.",
			"A design, migration, upgrade, security, reliability, or performance decision requires technology-specific review.",
		},
		Operating: []string{
			"Inspect the repository version, dependency lockfiles, build configuration, runtime target, and deployment environment before applying version-sensitive guidance.",
			"Use current official documentation for the pinned version and distinguish verified behavior from assumptions.",
			"Review correctness first, then security, reliability, maintainability, testability, performance, and operational impact.",
		},
		ReviewScope: []string{focus, "dependency and supply-chain boundaries", "configuration, secrets, identity, networking, data, observability, deployment, upgrade, and rollback behavior"},
		Checklist: []string{
			"Identify concrete versions, supported runtimes, dependency managers, build tools, and deployment targets.",
			"Check idiomatic API use, lifecycle boundaries, error handling, concurrency or asynchronous behavior, resource cleanup, and data consistency.",
			"Check input validation, authorization, secret handling, dependency provenance, insecure defaults, and sensitive logging.",
			"Check unit, integration, failure-path, upgrade, rollback, and operational test evidence.",
			"Check observability, capacity, performance, compatibility, deprecations, and migration impact.",
		},
		DecisionRules: []string{
			"Do not recommend an API, option, or migration from memory when behavior is version-sensitive; verify it against official documentation for the detected version.",
			"Do not run deployments, applies, destroys, migrations, or production mutations without explicit authorization for the exact target.",
			"Treat unsupported versions, ambiguous state changes, missing rollback, broad credentials, and untested destructive transitions as release blockers by impact.",
		},
		FindingCategories: []string{"Correctness or lifecycle defect.", "Security, identity, secret, dependency, or isolation weakness.", "Reliability, data integrity, upgrade, rollback, or operational gap.", "Testing, observability, maintainability, compatibility, or performance gap."},
		SeverityGuidance: []string{
			"Critical: enables remote compromise, unrestricted privilege, secret disclosure, destructive production change, or material data loss.",
			"High: causes exploitable authorization or isolation failure, sustained outage, corrupt state, or unsafe upgrade and rollback behavior.",
			"Medium: creates a bounded correctness, reliability, compatibility, test, observability, or performance risk.",
			"Low: improves clarity, idiomatic usage, maintainability, or evidence without changing material behavior.",
		},
		OutputRequirements: []string{"List detected versions and evidence sources.", "Return prioritized findings with affected files or resources, impact, evidence, and concrete remediation.", "Include validation commands, migration and rollback notes, and unresolved version-sensitive assumptions."},
		AcceptanceCriteria: []string{"Guidance matches the detected technology version and official documentation.", "Material correctness, security, reliability, test, upgrade, and rollback risks have evidence-backed disposition.", "No production mutation or destructive action is performed without exact-scope approval."},
		AntiPatterns:       []string{"Giving generic advice without inspecting versions or artifacts.", "Inventing configuration keys, APIs, defaults, test results, or compatibility claims.", "Treating a successful build as proof of security, runtime correctness, upgrade safety, or operability."},
	}
	completeGeneratedSkillContent(&content, definition.Domain)
	return content
}

func generatedCNCFSkillContent(entry cncf.Entry) skillContent {
	placements := make([]string, 0, len(entry.Placements))
	member := false
	for _, placement := range entry.Placements {
		placements = append(placements, placement.Category+" / "+placement.Subcategory)
		if placement.Category == "CNCF Members" {
			member = true
		}
	}
	source := entry.HomepageURL
	if entry.RepoURL != "" {
		source = entry.RepoURL
	}
	if source == "" {
		source = cncf.SourceURL
	}
	if member {
		content := skillContent{
			Purpose:            "Review " + entry.Name + " as an organization listed in the CNCF Landscape using current first-party evidence and explicit procurement, governance, security, portability, and operational criteria.",
			When:               []string{"The organization, its cloud-native products, support, partnership, procurement, outsourcing, or exit strategy is being assessed.", "Claims about CNCF membership or capabilities need current verification."},
			Operating:          []string{"Treat CNCF Landscape inclusion as classification metadata, not endorsement, certification, security assurance, or proof of product fitness.", "Verify current claims against the organization's first-party material and the official CNCF Landscape entry.", "Separate organization-level membership from the maturity or CNCF status of individual projects."},
			ReviewScope:        []string{"CNCF classification: " + strings.Join(placements, "; "), "First-party source: " + source, "product scope, ownership, support, security, compliance, portability, commercial dependencies, concentration risk, and exit strategy"},
			Checklist:          []string{"Confirm legal and product identity, current CNCF classification, offered services, support boundaries, regions, data handling, and shared responsibilities.", "Check security documentation, incident handling, vulnerability disclosure, support SLAs, audit evidence, subcontractors, portability, and termination assistance.", "Identify proprietary dependencies, lock-in, pricing or licensing constraints, migration paths, and operational ownership."},
			DecisionRules:      []string{"Do not infer CNCF project status, certification, security, or endorsement from membership.", "Do not make procurement or legal conclusions without the applicable current contract and accountable reviewers.", "Require an exit path and evidence for material capability or compliance claims."},
			FindingCategories:  []string{"Identity, scope, classification, or claim mismatch.", "Security, compliance, support, responsibility, or incident-management gap.", "Portability, concentration, licensing, cost, contract, or exit-strategy risk."},
			OutputRequirements: []string{"State the Landscape snapshot classification and sources checked.", "Return an evidence-backed capability and risk assessment with open questions, owners, and expiry dates."},
			AcceptanceCriteria: []string{"Membership is not presented as endorsement or project maturity.", "Material claims, responsibilities, dependencies, and exit assumptions are traceable to current evidence."},
			AntiPatterns:       []string{"Equating a Landscape logo with CNCF certification or technical approval.", "Inventing product features, contract terms, regions, certifications, or support commitments."},
		}
		completeGeneratedSkillContent(&content, entry.Name)
		return content
	}
	projectStatus := entry.Project
	if projectStatus == "" {
		projectStatus = "not specified"
	}
	content := skillContent{
		Purpose:            "Review and operate " + entry.Name + " using current official documentation, repository evidence, and its CNCF Landscape classification without treating Landscape inclusion as endorsement.",
		When:               []string{"" + entry.Name + " code, configuration, APIs, deployment, integrations, dependencies, upgrades, or operations are in scope.", "A technology choice or migration involving " + entry.Name + " requires evidence-backed assessment."},
		Operating:          []string{"Detect the installed or proposed version and deployment model before giving version-sensitive guidance.", "Use current official documentation and repository evidence; use the CNCF Landscape only for classification metadata.", "Inspect boundaries with identity, secrets, network, storage, dependencies, observability, upgrades, rollback, and disaster recovery."},
		ReviewScope:        []string{"CNCF classification: " + strings.Join(placements, "; "), "CNCF project maturity field: " + projectStatus, "First-party source: " + source, "configuration, APIs, architecture, security, reliability, performance, lifecycle, deployment, upgrade, rollback, and operations"},
		Checklist:          []string{"Confirm product identity, version, edition, license, repository, deployment model, and supported dependencies.", "Check secure defaults, authentication, authorization, secret handling, network exposure, tenant isolation, data protection, and supply-chain provenance.", "Check resource limits, scaling, availability, state consistency, backup and restore, failure handling, observability, upgrades, compatibility, and rollback.", "Check official tests, project-specific validation, negative paths, operational runbooks, and deprecation or migration notices."},
		DecisionRules:      []string{"Do not infer CNCF project maturity, security, compatibility, or endorsement when the Landscape project field is absent.", "Do not invent commands, configuration, APIs, defaults, or compatibility; verify version-sensitive behavior in first-party sources.", "Do not deploy, mutate clusters or cloud resources, migrate data, or change production state without exact-scope authorization and rollback evidence."},
		FindingCategories:  []string{"Version, identity, API, configuration, or lifecycle mismatch.", "Security, identity, secret, network, isolation, data, or supply-chain weakness.", "Reliability, scaling, state, backup, upgrade, rollback, observability, or operational gap.", "Landscape classification, maturity, licensing, maintenance, adoption, or migration assumption."},
		SeverityGuidance:   []string{"Critical: enables broad compromise, secret disclosure, destructive production action, cluster or tenant escape, or material data loss.", "High: causes exploitable access, sustained outage, corrupt state, unsafe upgrade, or loss of recovery capability.", "Medium: creates a bounded correctness, reliability, compatibility, observability, maintenance, or performance risk.", "Low: improves documentation, classification, idiomatic configuration, or validation evidence."},
		OutputRequirements: []string{"State detected versions, deployment model, Landscape classification, project field, and first-party sources checked.", "Return prioritized findings with affected artifacts, evidence, impact, remediation, validation, upgrade, and rollback guidance.", "Distinguish verified facts, source-derived inferences, and unresolved assumptions."},
		AcceptanceCriteria: []string{"Recommendations match the detected version and deployment model.", "Landscape inclusion is not presented as endorsement, certification, or security assurance.", "Material security, reliability, data, upgrade, rollback, and operational risks have evidence-backed disposition."},
		AntiPatterns:       []string{"Generating generic product advice without checking the installed version or repository artifacts.", "Copying marketing claims into architecture or security conclusions.", "Running production commands or destructive examples merely to validate a recommendation."},
	}
	completeGeneratedSkillContent(&content, entry.Name)
	return content
}

func completeGeneratedSkillContent(content *skillContent, subject string) {
	appendUntil := func(target *[]string, minimum int, candidates []string) {
		for _, candidate := range candidates {
			if len(*target) >= minimum {
				return
			}
			*target = append(*target, candidate)
		}
	}
	appendUntil(&content.Checklist, 10, []string{
		"Trace " + subject + " inputs, outputs, identities, trust boundaries, external dependencies, and persistent state.",
		"Check " + subject + " defaults, configuration precedence, environment separation, and drift from reviewed source.",
		"Check " + subject + " least privilege, credential rotation, audit events, policy enforcement, and break-glass behavior.",
		"Check " + subject + " resource ownership, cleanup, quotas, rate limits, timeouts, retries, and backpressure.",
		"Check " + subject + " release notes, supported upgrade paths, schema or API compatibility, and rollback constraints.",
		"Check " + subject + " dashboards, alerts, health signals, incident procedures, backup evidence, and recovery exercises.",
		"Check " + subject + " license, maintenance state, vulnerability handling, provenance, signatures, and dependency pinning.",
		"Check " + subject + " documentation, ownership, support boundaries, known limitations, and decommissioning plan.",
	})
	appendUntil(&content.DecisionRules, 5, []string{
		"If " + subject + " identity, version, edition, or deployment model cannot be established, report the uncertainty before recommending a change.",
		"If a " + subject + " change can affect availability, security boundaries, persistent data, or external consumers, require staged validation and explicit rollback criteria.",
		"If first-party evidence conflicts with generated guidance for " + subject + ", follow the pinned first-party documentation and report catalog drift.",
	})
	appendUntil(&content.FindingCategories, 5, []string{
		"" + subject + " documentation, ownership, maintenance, licensing, provenance, or evidence gap.",
		"" + subject + " integration, dependency, compatibility, migration, or decommissioning risk.",
	})
	appendUntil(&content.SeverityGuidance, 4, []string{
		"Critical: " + subject + " can enable broad compromise, destructive production mutation, tenant escape, secret exposure, or material data loss.",
		"High: " + subject + " can cause exploitable access, sustained outage, corrupt state, unsafe upgrade, or loss of recovery capability.",
		"Medium: " + subject + " has a bounded correctness, reliability, compatibility, observability, maintenance, or performance risk.",
		"Low: " + subject + " needs documentation, classification, idiomatic configuration, or validation-evidence improvement.",
	})
	appendUntil(&content.OutputRequirements, 5, []string{
		"Include a " + subject + " architecture and dependency summary with trust boundaries and persistent state.",
		"Include " + subject + " validation commands or tests that are safe for the stated environment.",
		"Include " + subject + " ownership, rollout, rollback, monitoring, and follow-up actions.",
	})
	appendUntil(&content.AcceptanceCriteria, 5, []string{
		"" + subject + " identity, version, configuration, dependencies, and deployment assumptions are explicit.",
		"" + subject + " failure, upgrade, rollback, observability, and recovery paths have proportionate evidence.",
		"" + subject + " security boundaries, credentials, network exposure, data handling, and supply-chain risks are addressed.",
	})
	appendUntil(&content.AntiPatterns, 5, []string{
		"Assuming Landscape inclusion or popularity proves that " + subject + " is secure, supported, compatible, or suitable.",
		"Using an unpinned latest version of " + subject + " as the basis for migration or production guidance.",
		"Ignoring " + subject + " operational ownership, data lifecycle, upgrade constraints, rollback, and decommissioning.",
	})
}

func generatedPaymentSkillContent(seed additionalSkillSeed) skillContent {
	content := generatedAdditionalSkillContent(seed)
	content.Operating = append(content.Operating,
		"Model payment changes as explicit, monotonic state transitions and keep provider state, order state, fulfillment, and ledger effects independently reconcilable.",
		"Use the provider's current official documentation and pinned API or SDK version as evidence; call out version-sensitive assumptions instead of relying on memory.",
	)
	content.Checklist = append(content.Checklist,
		"Represent monetary values with explicit currency and integer minor units or a decimal type that cannot introduce binary floating-point rounding.",
		"Test retries with a stable operation-level idempotency key and verify that ambiguous timeouts cannot create duplicate charges, captures, refunds, or fulfillment.",
		"Verify webhook authenticity before parsing business fields, deduplicate events, tolerate out-of-order delivery, acknowledge promptly, and reconcile missed events.",
		"Keep secret keys, cardholder data, authentication values, access tokens, and raw provider payloads out of source control, prompts, logs, traces, fixtures, and screenshots.",
		"Require server-side price, amount, currency, order ownership, refund eligibility, and final payment-state validation; never trust client-supplied payment status.",
	)
	content.DecisionRules = append(content.DecisionRules,
		"Never create, capture, cancel, refund, dispute, or otherwise move real funds without explicit authorization for the exact environment, provider account, operation, payment, amount, and currency.",
		"If a write outcome is ambiguous after a timeout, reconcile by idempotency key, provider identifier, webhook, and status lookup before issuing any new mutation.",
		"If official provider documentation conflicts with a template instruction, follow the pinned provider documentation and report the template drift.",
		"If raw PAN, CVV, full bank credentials, production secrets, or unrestricted live credentials enter scope, stop and require an approved secure handling path.",
	)
	content.FindingCategories = append(content.FindingCategories,
		"Duplicate or ambiguous monetary mutation caused by missing idempotency or unsafe retry behavior.",
		"Payment state, fulfillment, provider balance, payout, or internal ledger inconsistency.",
		"Unauthorized live operation, missing human approval, excessive credential scope, or absent immutable audit trail.",
	)
	content.SeverityGuidance = []string{
		"Critical: a defect can cause unauthorized or duplicate money movement, expose cardholder authentication data or live credentials, bypass approval, or fulfill an unpaid order.",
		"High: payment state, webhook authenticity, amount, currency, tenant authorization, refund, settlement, or reconciliation is incorrect for a material flow.",
		"Medium: recovery, observability, test coverage, provider-version pinning, or evidence is incomplete without a demonstrated monetary or sensitive-data impact.",
		"Low: naming, metadata, documentation, or minor traceability improvements are needed without affecting payment correctness or security.",
	}
	content.DevSecOpsGuardrails = []string{
		"Use sandbox or test mode by default and use only synthetic payment data; never place real cardholder or bank-account data in tests.",
		"Treat live payment mutations as high-impact external actions requiring exact-scope approval and an immutable audit record.",
		"Do not expose, retrieve, rotate, or copy provider secrets unless the user explicitly authorizes the exact credential-management task.",
		"Never recommend bypassing SCA, 3-D Secure, webhook verification, provider risk controls, approval gates, or reconciliation to make a test pass.",
	}
	content.OutputRequirements = append(content.OutputRequirements,
		"State whether evidence came from sandbox, fixtures, recorded responses, or production-safe read-only inspection; never imply a live transaction was tested when it was not.",
		"List provider API or SDK versions, environment assumptions, approval boundaries, idempotency strategy, and reconciliation path.",
	)
	content.AcceptanceCriteria = append(content.AcceptanceCriteria,
		"All monetary mutations are server-authorized, idempotent, auditable, and covered by duplicate, timeout, webhook, and reconciliation tests.",
		"No real payment data or live credential is required for routine development, review, or automated validation.",
	)
	content.AntiPatterns = append(content.AntiPatterns,
		"Treating a browser redirect or client callback as proof that a payment succeeded.",
		"Retrying a timed-out charge, capture, or refund with a new idempotency key before reconciling the original operation.",
		"Using production credentials, real cards, or live money movement to compensate for incomplete sandbox tests.",
	)
	addProviderSpecificPaymentContent(seed.Name, &content)
	return content
}

func addProviderSpecificPaymentContent(name string, content *skillContent) {
	switch name {
	case "stripe-integration-engineer":
		content.Operating = append(content.Operating,
			"Choose PaymentIntents, Checkout Sessions, Billing, or Connect from the merchant and platform requirements; do not mix their ownership models accidentally.",
			"Pin and test the Stripe API version, SDK version, webhook endpoint version, and connected-account context before changing object or event handling.",
		)
		content.Checklist = append(content.Checklist,
			"Map PaymentIntent statuses and next_action handling to the local order state without treating client confirmation as fulfillment authority.",
			"Verify Stripe-Signature against the unmodified request body and the correct endpoint secret before enqueueing an event.",
			"Separate publishable keys, client secrets, restricted keys, and secret keys across browser, server, Connect account, and environment boundaries.",
			"Test Checkout Session completion, asynchronous payment success or failure, expired sessions, refunds, disputes, and delayed webhook delivery.",
			"For Connect, assert the platform versus connected-account owner on every request, event, fee, transfer, refund, and reconciliation record.",
		)
		content.DecisionRules = append(content.DecisionRules,
			"Use a PaymentIntent or Checkout Session as the payment lifecycle anchor; do not create independent Charges as an informal retry mechanism.",
			"Do not fulfill from success_url, return_url, or client-side status alone; require server-side Stripe state plus durable webhook or reconciliation evidence.",
			"If an API-version upgrade changes fields, events, or expansion behavior, require a fixture diff, migration test, and rollback decision.",
		)
		content.FindingCategories = append(content.FindingCategories,
			"Incorrect PaymentIntent, Checkout Session, Charge, Invoice, Subscription, Transfer, or connected-account ownership model.",
			"Stripe API-version, endpoint-version, SDK-version, event-shape, or test-clock drift.",
		)
		content.OutputRequirements = append(content.OutputRequirements, "Stripe object and event ownership map covering PaymentIntent or Checkout, Connect context, fulfillment authority, and reconciliation identifiers.")
		content.AcceptanceCriteria = append(content.AcceptanceCriteria, "Stripe-Signature validation, PaymentIntent lifecycle, Connect scoping when used, delayed events, disputes, and API-version changes have test evidence.")
		content.AntiPatterns = append(content.AntiPatterns,
			"Logging a Stripe secret key, webhook secret, PaymentIntent client_secret, raw payment-method details, or unredacted event payload.",
			"Assuming every payment method reaches a final state synchronously during confirm or Checkout redirect.",
		)
	case "paypal-integration-engineer":
		content.Operating = append(content.Operating,
			"Model PayPal Orders separately from authorization, capture, refund, and webhook resources, following their HATEOAS links and current REST API reference.",
			"Obtain OAuth access tokens server-side with environment-specific client credentials and constrain partner or multiparty headers to the intended merchant context.",
		)
		content.Checklist = append(content.Checklist,
			"Map Order CREATED, APPROVED, COMPLETED, VOIDED, PAYER_ACTION_REQUIRED, authorization, and capture outcomes to local states without collapsing them.",
			"Use PayPal-Request-Id only on API operations that document support, preserve it across an ambiguous retry, and scope it to one operation type.",
			"Verify REST webhook transmission headers, certificate or verification response, webhook ID, and original body before accepting the event.",
			"Keep client ID and client secret roles distinct and ensure OAuth access tokens never enter browser code, fixtures, URLs, or normal application logs.",
			"Test buyer approval cancellation, instrument decline, duplicate invoice ID, capture timeout, partial refund, webhook resend, and sandbox/live isolation.",
		)
		content.DecisionRules = append(content.DecisionRules,
			"Do not capture an Order merely because the buyer returned from approval; retrieve or advance the server-side Order using the documented link and validate the result.",
			"If PayPal-Request-Id retention or support differs by API, follow that endpoint reference and do not generalize a retention period across APIs.",
			"For multiparty payments, require explicit merchant attribution, payee ownership, fee, disbursement, refund, and webhook-routing decisions.",
		)
		content.FindingCategories = append(content.FindingCategories,
			"PayPal Order, payer approval, authorization, capture, refund, partner attribution, or payee ownership mismatch.",
			"OAuth token, PayPal-Request-Id, webhook ID, transmission-signature, or sandbox/live boundary failure.",
		)
		content.OutputRequirements = append(content.OutputRequirements, "PayPal Order-to-capture lifecycle map with HATEOAS transitions, OAuth boundary, request-ID scope, webhook verification, and refund reconciliation.")
		content.AcceptanceCriteria = append(content.AcceptanceCriteria, "PayPal buyer approval, Orders and Captures transitions, supported idempotency, verified webhooks, negative testing, and sandbox isolation have evidence.")
		content.AntiPatterns = append(content.AntiPatterns,
			"Treating buyer approval as capture completion or trusting a browser return parameter as the final PayPal transaction state.",
			"Reusing one PayPal-Request-Id across create-order, authorize, capture, and refund operations.",
		)
	case "adyen-integration-engineer":
		content.Operating = append(content.Operating,
			"Choose Adyen Sessions or Advanced flow deliberately and map paymentMethods, payments, payments/details, modifications, and webhooks to one merchant reference model.",
			"Keep company account, merchant account, regional endpoint, live prefix, API credential, client key, and HMAC key boundaries explicit across test and live environments.",
		)
		content.Checklist = append(content.Checklist,
			"Handle action objects and payments/details results without treating a front-end resultCode as final fulfillment authority.",
			"Validate standard webhook HMAC fields or the raw-body header signature appropriate to the webhook type before processing eventCode.",
			"Deduplicate standard payment events by eventCode and pspReference, compare eventDate, and prevent an older event from regressing local state.",
			"Scope idempotency keys to the Adyen company account and account for regional endpoints when evaluating duplicate protection.",
			"Test AUTHORISATION, CAPTURE, REFUND, CANCELLATION, chargeback, refusal, pending action, duplicate event, and asynchronous modification paths.",
		)
		content.DecisionRules = append(content.DecisionRules,
			"A front-end resultCode or redirect return is not sufficient for fulfillment; use authenticated server response and AUTHORISATION or reconciliation evidence.",
			"A transient-error response may be retried with the same idempotency key and backoff; do not silently fall back to a non-idempotent mutation.",
			"If standard and non-standard webhook formats differ, select the documented signature algorithm and payload form for that exact webhook type.",
		)
		content.FindingCategories = append(content.FindingCategories,
			"Adyen merchantReference, pspReference, merchant account, regional endpoint, resultCode, action, or modification mismatch.",
			"Wrong standard-event HMAC fields, non-standard raw-body signature, HMAC-key environment, or event ordering logic.",
		)
		content.OutputRequirements = append(content.OutputRequirements, "Adyen Sessions or Advanced flow map covering action handling, merchantReference and pspReference correlation, HMAC format, modifications, and reconciliation.")
		content.AcceptanceCriteria = append(content.AcceptanceCriteria, "Adyen actions, AUTHORISATION and modification events, HMAC validation, duplicate and ordering behavior, regional idempotency, and test/live separation have evidence.")
		content.AntiPatterns = append(content.AntiPatterns,
			"Using Adyen resultCode or redirect completion alone to release goods before an authenticated server-side authorization decision.",
			"Applying the standard NotificationRequestItem HMAC construction to a non-standard webhook that signs the raw body in headers.",
		)
	}
}

func generatedAdditionalSkillContent(seed additionalSkillSeed) skillContent {
	artifacts := strings.Join(seed.Artifacts, ", ")
	risks := strings.Join(seed.Risks, ", ")
	signals := strings.Join(seed.Signals, ", ")
	outputs := strings.Join(seed.Outputs, ", ")
	return skillContent{
		Purpose: seed.Description + " Treat regulatory, security, and operational references as review and evidence guidance, not legal advice.",
		When: []string{
			fmt.Sprintf("%s decisions, controls, or operating practices need independent review.", seed.Domain),
			fmt.Sprintf("A change affects %s artifacts such as %s.", seed.Domain, artifacts),
			fmt.Sprintf("The user needs evidence-oriented findings for risks such as %s.", risks),
			"Audit, security, operations, or platform stakeholders need a concise readiness position.",
			"Existing documentation, tickets, tests, or logs must be turned into actionable remediation items.",
		},
		Operating: []string{
			fmt.Sprintf("Identify the relevant %s artifacts, owners, systems, environments, and review boundary.", seed.Domain),
			fmt.Sprintf("Compare the available artifacts against expected signals such as %s.", signals),
			"Separate confirmed gaps from assumptions, missing evidence, and advisory improvement opportunities.",
			"Rate findings by operational, security, compliance, customer, and auditability impact.",
			"Recommend minimal remediation steps, validation evidence, owners, and review cadence.",
		},
		ReviewScope: []string{
			fmt.Sprintf("Primary artifacts: %s.", artifacts),
			fmt.Sprintf("Risk themes: %s.", risks),
			fmt.Sprintf("Evidence signals: %s.", signals),
			"Ownership, approvals, review cadence, exception handling, and residual-risk decisions.",
			"Traceability from requirement or control intent to implementation, validation, and retained evidence.",
		},
		Checklist: []string{
			fmt.Sprintf("Confirm the review boundary covers the right %s systems, teams, and environments.", seed.Domain),
			fmt.Sprintf("Inventory and inspect the current %s.", seed.Artifacts[0]),
			fmt.Sprintf("Check whether %s is current, approved, versioned, and owned.", seed.Artifacts[1]),
			fmt.Sprintf("Verify that %s has test, ticket, log, or approval support.", seed.Artifacts[2]),
			fmt.Sprintf("Look for %s and record concrete repository or process evidence.", seed.Risks[0]),
			fmt.Sprintf("Look for %s and identify affected assets, services, or stakeholders.", seed.Risks[1]),
			fmt.Sprintf("Look for %s and classify the operational or audit impact.", seed.Risks[2]),
			fmt.Sprintf("Use %s to validate that the control or practice is operating.", seed.Signals[0]),
			fmt.Sprintf("Use %s to confirm ownership, timing, and reproducibility.", seed.Signals[1]),
			fmt.Sprintf("Check exception, risk-acceptance, and expiry handling for %s.", seed.Domain),
			"Confirm remediation items have owners, due dates, validation steps, and evidence expectations.",
			"Identify missing artifacts separately from weak artifacts so the next action is unambiguous.",
			"Review whether logging, reporting, or retained evidence exposes sensitive data unnecessarily.",
		},
		DecisionRules: []string{
			fmt.Sprintf("If %s is missing for a critical service, raise at least a high-severity readiness gap.", seed.Artifacts[0]),
			fmt.Sprintf("If %s cannot be tied to an owner and approval, treat the outcome as unauditable until corrected.", seed.Signals[1]),
			fmt.Sprintf("If %s is present but expired or untested, require validation before accepting residual risk.", seed.Artifacts[3]),
			"If the only support is verbal or chat-only context, request durable ticket, document, log, or test evidence.",
			"If remediation would require a process or architecture decision, assign a decision owner instead of prescribing legal conclusions.",
			"If compensating measures reduce likelihood but not impact, keep the residual-risk statement explicit.",
		},
		FindingCategories: []string{
			fmt.Sprintf("Missing or stale %s artifact.", seed.Domain),
			"Unclear ownership, approval, review cadence, or accountability.",
			"Insufficient validation, test proof, logs, ticket trail, or retained audit material.",
			"Unreviewed exception, residual risk, expiry, or compensating measure.",
			"Policy, architecture, operational, or platform implementation drift.",
			"Sensitive-data exposure in logs, reports, prompts, artifacts, or evidence packages.",
		},
		SeverityGuidance: []string{
			fmt.Sprintf("Critical: a gap in %s creates immediate outage, data-loss, privilege, regulatory-reporting, or irreversible business risk.", seed.Domain),
			fmt.Sprintf("High: %s is missing, unowned, untested, or unauditable for a critical service or material change.", seed.Artifacts[0]),
			fmt.Sprintf("Medium: %s exists but is stale, incomplete, inconsistently enforced, or weakly evidenced.", seed.Artifacts[1]),
			"Low: wording, metadata, formatting, link freshness, or minor traceability improvements are needed.",
		},
		OutputRequirements: []string{
			fmt.Sprintf("Findings ordered by severity with affected %s artifacts and evidence references.", seed.Domain),
			fmt.Sprintf("Coverage note for reviewed artifacts: %s.", artifacts),
			fmt.Sprintf("Risk note covering relevant themes: %s.", risks),
			fmt.Sprintf("Evidence request list using expected signals: %s.", signals),
			fmt.Sprintf("Deliverables or updates needed: %s.", outputs),
			"Residual-risk, assumptions, missing-context, and validation-gap summary.",
		},
		AcceptanceCriteria: []string{
			fmt.Sprintf("Relevant %s artifacts are identified, current, owned, and versioned where applicable.", seed.Domain),
			"Each high-impact finding includes evidence, impact, likelihood, owner, and remediation guidance.",
			"Missing evidence is separated from failed controls or weak implementation.",
			"Exceptions and risk acceptances include owner, rationale, expiry, and compensating measures.",
			"Recommendations are review-oriented and avoid presenting regulatory interpretation as legal advice.",
			"Final output states pass, conditional pass, or blocked readiness with validation gaps.",
		},
		AntiPatterns: []string{
			"Treating a policy title or control name as proof that the practice operates effectively.",
			"Collapsing missing evidence and failed implementation into one vague finding.",
			"Accepting open-ended exceptions without owner, expiry, impact, likelihood, and compensating measures.",
			"Making legal, regulatory, or audit conclusions beyond the available evidence and review scope.",
			"Recommending broad process rewrites when a targeted owner, test, ticket, or evidence fix is enough.",
			"Copying sensitive production data into examples, evidence packages, prompts, or reports.",
		},
	}
}

var skillBodies = buildSkillBodies()

func buildSkillBodies() map[string]string {
	bodies := map[string]string{}
	for _, seed := range additionalSkillSeeds {
		sdlcSkillContent[seed.Name] = generatedAdditionalSkillContent(seed)
	}
	for _, seed := range paymentSkillSeeds {
		sdlcSkillContent[seed.Name] = generatedPaymentSkillContent(seed)
	}
	for _, definition := range technologySkillDefinitions {
		sdlcSkillContent[definition.Name] = generatedTechnologySkillContent(definition)
	}
	for _, entry := range cncf.MustEntries() {
		sdlcSkillContent[entry.SkillName] = generatedCNCFSkillContent(entry)
	}
	for name, content := range agenticSecuritySkillContent {
		sdlcSkillContent[name] = content
	}
	for _, name := range SDLCSkillNames {
		bodies[name] = buildSkillBody(name, sdlcSkillContent[name])
	}
	return bodies
}

func buildSkillBody(name string, content skillContent) string {
	var b strings.Builder
	b.WriteString("# {{.Title}}\n\n")
	writeParagraph(&b, "Purpose", content.Purpose)
	writeParagraph(&b, "Goal and behavioral contract", "The authoritative Goal and artifact references are defined in `descriptor.yaml`. Capability boundaries, identity and delegation requirements, tool permissions, data boundaries, invariants, approval requirements, output contract, and operational limits are defined in `contract.yaml`. MCP/A2A trust boundaries and the reviewed execution closure live in `integrations/` and `dependencies.yaml`; ASPS and assurance requirements live in `assurance.yaml`.\n\nTreat those declarations as mandatory execution constraints. `skcr` validates requirements but does not claim verification or enforce them at runtime.")
	writeBullets(&b, "When to use", content.When, false)
	writeNumbered(&b, "Operating model", content.Operating)
	writeBullets(&b, "Spec-Driven Change Context", sharedSpecDrivenChangeContext, false)
	writeBullets(&b, "Skill-Specific Review Scope", content.ReviewScope, false)
	writeBullets(&b, "Skill-Specific Checklist", content.Checklist, true)
	writeBullets(&b, "Decision Rules", content.DecisionRules, false)
	writeBullets(&b, "Finding Categories", content.FindingCategories, false)
	writeBullets(&b, "Severity Guidance", content.SeverityGuidance, false)
	guardrails := append(append([]string{}, sharedDevSecOpsGuardrails...), content.DevSecOpsGuardrails...)
	writeBullets(&b, "DevSecOps Guardrails", guardrails, false)
	writeBullets(&b, "Output Requirements", content.OutputRequirements, false)
	writeBullets(&b, "Acceptance Criteria", content.AcceptanceCriteria, false)
	writeBullets(&b, "Anti-Patterns", content.AntiPatterns, false)
	b.WriteString("## Changelog\n\n")
	b.WriteString("### {{.Version}} - {{.LastModified}}\n\n")
	b.WriteString("- Initial generated production-ready SDLC / DevSecOps skill.\n")
	return b.String()
}

func writeParagraph(b *strings.Builder, heading, text string) {
	fmt.Fprintf(b, "## %s\n\n%s\n\n", heading, text)
}

func writeNumbered(b *strings.Builder, heading string, items []string) {
	fmt.Fprintf(b, "## %s\n\n", heading)
	for i, item := range items {
		fmt.Fprintf(b, "%d. %s\n", i+1, item)
	}
	b.WriteString("\n")
}

func writeBullets(b *strings.Builder, heading string, items []string, checklist bool) {
	fmt.Fprintf(b, "## %s\n\n", heading)
	for _, item := range items {
		if checklist {
			fmt.Fprintf(b, "- [ ] %s\n", item)
		} else {
			fmt.Fprintf(b, "- %s\n", item)
		}
	}
	b.WriteString("\n")
}

func renderSkillTemplate(name string, data skillTemplateData) (string, error) {
	body, ok := skillBodies[name]
	if !ok {
		return "", nil
	}
	if len(data.MinPlatforms) == 0 {
		data.MinPlatforms = platformcompat.AllMinVersions()
	}
	full := skillFrontmatter + body
	tmpl, err := template.New(name).Funcs(template.FuncMap{"quote": strconv.Quote}).Parse(full)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func skillTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
