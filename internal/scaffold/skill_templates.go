package scaffold

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	platformcompat "github.com/domehahn/skcr/internal/platforms"
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
		MinPlatforms: platformcompat.AllMinVersions(),
	}
	rendered, err := renderSkillTemplate(name, data)
	if err != nil {
		return "", true, err
	}
	return rendered, true, nil
}

const skillFrontmatter = `---
name: "{{.Name}}"
description: "{{.Description}}"
version: "{{.Version}}"
since: "{{.Since}}"
last_modified: "{{.LastModified}}"
authors:
  - "{{.Owner}}"
stability: "{{.Stability}}"
min_platform_version:
{{- range .MinPlatforms }}
  {{.Name}}: "{{.MinVersion}}"
{{- end }}
deprecated_since:
replaces:
supersedes: []
changelog:
  - version: "{{.Version}}"
    date: "{{.Since}}"
    change: "Initial release"
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

var SDLCSkillNames = []string{
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
var sharedDevSecOpsGuardrails = []string{
	"Do not read secrets, `.env` files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.",
	"Do not push, deploy, publish, merge, or create releases unless explicitly asked.",
	"Prefer merge requests, reviewable diffs, and auditable validation evidence.",
	"Prefer least privilege, minimal changes, and explicit rollback notes.",
	"Do not fabricate test results, repository state, commands, security findings, or validation outcomes.",
	"Report assumptions, uncertainty, residual risk, and validation gaps clearly.",
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
		},
		DecisionRules: []string{
			"If required behavior implies unrelated refactoring, ask before expanding scope.",
			"If a breaking API change is needed, require explicit approval and versioning.",
			"If input crosses a trust boundary, validate before use.",
			"If a migration is required, make rollout and rollback explicit.",
			"If tests cannot run, report the reason and residual risk.",
			"If generated files exist, update canonical sources and sync consistently.",
		},
		FindingCategories: []string{
			"Scope creep or unrelated change.",
			"Missing tests or validation evidence.",
			"Unsafe input validation or error handling.",
			"API compatibility or migration risk.",
			"Secret exposure or unsafe configuration.",
			"Global side effect or hidden state change.",
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
		},
		AntiPatterns: []string{
			"Broad refactoring in a narrow fix.",
			"Changing public APIs without approval.",
			"Skipping tests because the change is small.",
			"Swallowing errors or leaking internal errors externally.",
			"Hardcoding environment-specific secrets or URLs.",
			"Editing generated copies without updating canonical source.",
		},
	},
	"test-strategy-engineer": {
		Purpose: "Design test strategy for features, fixes, migrations, and releases across unit, integration, contract, E2E, regression, negative, security, performance, test data, mocking, CI gates, and coverage risk.",
		When: []string{
			"Test Strategy Engineer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the test strategy engineer scope and gather minimum relevant repository evidence.",
			"Apply test strategy engineer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Unit, integration, contract, E2E, regression, negative, security, and performance coverage.",
			"Test data, fixtures, mocking strategy, determinism, isolation, and cleanup.",
			"CI gates, coverage risks, flaky tests, and validation sequencing.",
			"Boundary cases, abuse cases, migrations, and external dependencies.",
			"Must-have versus optional test scope.",
		},
		Checklist: []string{
			"Identify critical behavior that requires unit tests.",
			"Identify integration boundaries that require integration tests.",
			"Identify APIs or contracts that require contract tests.",
			"Add negative tests for invalid input and abuse cases.",
			"Add regression tests for bug fixes.",
			"Identify missing test data or fixtures.",
			"Check whether tests are deterministic.",
			"Identify flaky-test risks.",
			"Propose CI gates for critical paths.",
			"Separate must-have tests from optional tests.",
			"Map acceptance criteria to test types.",
			"Define test data cleanup and isolation.",
		},
		DecisionRules: []string{
			"If test strategy engineer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak test strategy engineer evidence.",
			"Incorrect or unsafe test strategy engineer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The test strategy engineer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"verification-reviewer": {
		Purpose: "Verify implementation results against requirements, changed files, validation commands, test evidence, edge cases, security validation, documentation consistency, and residual risk.",
		When: []string{
			"Verification Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the verification reviewer scope and gather minimum relevant repository evidence.",
			"Apply verification reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Requirement-to-implementation traceability.",
			"Acceptance criteria and test evidence credibility.",
			"Changed files, edge cases, scope expansion, and regression risk.",
			"Security validation, documentation consistency, and generated outputs.",
			"Residual risk and pass/fail recommendation.",
		},
		Checklist: []string{
			"Map each requirement to changed behavior.",
			"Verify acceptance criteria against implementation evidence.",
			"Check whether validation commands were actually run or are missing.",
			"Identify untested edge cases.",
			"Check changed files for unintended scope expansion.",
			"Identify documentation or changelog gaps.",
			"Verify security-relevant behavior where applicable.",
			"Distinguish verified facts from assumptions.",
			"Identify residual risks.",
			"Produce a clear pass/fail or conditional recommendation.",
			"Check generated files and synchronized outputs.",
			"Confirm test evidence is credible and current.",
		},
		DecisionRules: []string{
			"If verification reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak verification reviewer evidence.",
			"Incorrect or unsafe verification reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The verification reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"security-reviewer": {
		Purpose: "Review code, design, configuration, and changes for authentication, authorization, input validation, output encoding, injection, SSRF, path traversal, deserialization, XSS, CSRF, secrets, logging, cryptography, errors, dependencies, and least privilege.",
		When: []string{
			"Security Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the security reviewer scope and gather minimum relevant repository evidence.",
			"Apply security reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Authentication, authorization, and tenant isolation.",
			"Input validation, injection, SSRF, path traversal, deserialization, XSS, and CSRF.",
			"Secrets, logging, cryptography, errors, dependencies, and permissions.",
			"Configuration, CI/CD, runtime, and least-privilege boundaries.",
			"Exploitability, impact, and remediation quality.",
		},
		Checklist: []string{
			"Check authentication and authorization boundaries.",
			"Identify attacker-controlled inputs.",
			"Check validation and sanitization at trust boundaries.",
			"Identify injection risks.",
			"Identify path traversal or unsafe file access.",
			"Identify SSRF or unsafe outbound requests.",
			"Check secrets handling and logging.",
			"Review cryptographic usage and key handling.",
			"Identify unsafe error disclosure.",
			"Prioritize findings by exploitability and impact.",
			"Check dependency and configuration security exposure.",
			"Verify least-privilege and permission boundaries.",
		},
		DecisionRules: []string{
			"If security reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak security reviewer evidence.",
			"Incorrect or unsafe security reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The security reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"secrets-reviewer": {
		Purpose: "Identify and handle secrets, credentials, tokens, private keys, passwords, `.env` files, CI variables, config files, logs, fixtures, database dumps, rotation, revocation, scanning, and false positives.",
		When: []string{
			"Secrets Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the secrets reviewer scope and gather minimum relevant repository evidence.",
			"Apply secrets reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"API keys, tokens, private keys, passwords, `.env` files, and CI variables.",
			"Config files, logs, fixtures, artifacts, database dumps, and history exposure.",
			"Secret storage, rotation, revocation, scope, and owners.",
			"Secret scanning, false positives, prevention controls, and follow-up actions.",
			"Output redaction and safe handling of suspected values.",
		},
		Checklist: []string{
			"Search for potential secrets without unnecessarily exposing their values.",
			"Identify hardcoded credentials or tokens.",
			"Check whether secrets are stored in config or code.",
			"Check logs for accidental sensitive data disclosure.",
			"Distinguish test fixtures from real credentials where possible.",
			"Recommend rotation and revocation when a real secret is exposed.",
			"Avoid printing full secret values in output.",
			"Check CI/CD secret handling.",
			"Verify that secret references use secure mechanisms.",
			"Identify follow-up actions and owners.",
			"Check history or artifacts when exposure is suspected.",
			"Classify false positives with evidence.",
		},
		DecisionRules: []string{
			"If secrets reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak secrets reviewer evidence.",
			"Incorrect or unsafe secrets reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The secrets reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"dependency-supply-chain-reviewer": {
		Purpose: "Review dependencies, packages, lockfiles, container images, direct and transitive CVEs, license risk, maintainer health, pinning, SBOM, provenance, dependency confusion, typosquatting, package scripts, and base images.",
		When: []string{
			"Dependency Supply Chain Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the dependency supply chain reviewer scope and gather minimum relevant repository evidence.",
			"Apply dependency supply chain reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Direct and transitive dependencies, manifests, and lockfiles.",
			"CVEs, licenses, maintainer health, version pinning, and SBOM.",
			"Provenance, dependency confusion, typosquatting, package scripts, and registries.",
			"Container base images, third-party actions, and generated artifacts.",
			"Blocking versus advisory supply-chain risk.",
		},
		Checklist: []string{
			"Identify newly added or upgraded dependencies.",
			"Check whether dependencies are pinned or locked.",
			"Identify known vulnerability or license risks where data is available.",
			"Identify suspicious packages or typosquatting risks.",
			"Check package install scripts and build hooks.",
			"Review transitive dependency risk.",
			"Check whether SBOM or provenance exists.",
			"Recommend safer alternatives or updates.",
			"Identify unnecessary dependencies.",
			"Separate blocking risks from advisory risks.",
			"Review container base images and third-party actions.",
			"Check dependency confusion and registry source controls.",
		},
		DecisionRules: []string{
			"If dependency supply chain reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak dependency supply chain reviewer evidence.",
			"Incorrect or unsafe dependency supply chain reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The dependency supply chain reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"ci-cd-reviewer": {
		Purpose: "Review CI/CD pipelines, build definitions, release workflows, token permissions, secrets, protected branches, merge gates, approvals, reproducibility, artifact integrity, cache poisoning, dependency installation, runner trust, deployment gates, and environment protection.",
		When: []string{
			"Ci Cd Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the ci cd reviewer scope and gather minimum relevant repository evidence.",
			"Apply ci cd reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Token permissions, secrets, protected branches, merge gates, and approvals.",
			"Build reproducibility, dependency installation, artifact integrity, and caches.",
			"Runner trust, fork-trigger safety, OIDC, cloud credentials, and shell execution.",
			"Deployment gates, environment protection, and release workflow controls.",
			"Pipeline logs, artifacts, caches, and privilege boundaries.",
		},
		Checklist: []string{
			"Check workflow token permissions.",
			"Verify least-privilege for CI jobs.",
			"Identify unsafe shell execution.",
			"Check secrets exposure in logs and steps.",
			"Review dependency installation and caching.",
			"Identify cache poisoning risks.",
			"Check artifact signing or checksum generation.",
			"Verify branch protection and merge gates.",
			"Check deployment approval controls.",
			"Identify untrusted runner risks.",
			"Review OIDC and cloud credential scoping.",
			"Check fork and pull-request trigger safety.",
		},
		DecisionRules: []string{
			"If ci cd reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak ci cd reviewer evidence.",
			"Incorrect or unsafe ci cd reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The ci cd reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"iac-gitops-reviewer": {
		Purpose: "Review infrastructure-as-code and GitOps changes across Terraform, Helm, Kubernetes, GitOps manifests, IAM, network exposure, public access, secrets, encryption, logging, backups, least privilege, security contexts, resource limits, drift, and policy enforcement.",
		When: []string{
			"Iac Gitops Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the iac gitops reviewer scope and gather minimum relevant repository evidence.",
			"Apply iac gitops reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Terraform, Helm, Kubernetes, Kustomize, and GitOps manifests.",
			"IAM, network exposure, public access, secrets, encryption, logging, and backups.",
			"Least privilege, security contexts, resource limits, and network policies.",
			"Drift, reconciliation, promotion gates, rollback, and policy enforcement.",
			"Environment separation and production-safety controls.",
		},
		Checklist: []string{
			"Identify public exposure of services, buckets, databases, or ingress.",
			"Review IAM privileges for least privilege.",
			"Check encryption settings for storage and transport.",
			"Identify secrets in manifests or Terraform variables.",
			"Check Kubernetes security contexts.",
			"Check resource requests and limits.",
			"Check network policies or segmentation.",
			"Check logging, monitoring, and backup configuration.",
			"Identify drift or manual-change risks.",
			"Recommend safe rollout and rollback steps.",
			"Review GitOps reconciliation and promotion gates.",
			"Check policy enforcement with OPA, Kyverno, or equivalent.",
		},
		DecisionRules: []string{
			"If iac gitops reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak iac gitops reviewer evidence.",
			"Incorrect or unsafe iac gitops reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The iac gitops reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"compliance-governance-reviewer": {
		Purpose: "Review governance, policy, auditability, compliance evidence, control mapping, approvals, audit trails, ownership, segregation of duties, policy exceptions, retention, access reviews, change management, and risk acceptance.",
		When: []string{
			"Compliance Governance Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the compliance governance reviewer scope and gather minimum relevant repository evidence.",
			"Apply compliance governance reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Control mapping, approvals, audit trails, evidence, and ownership.",
			"Segregation of duties, access reviews, policy exceptions, and expiry.",
			"Retention, change management, risk acceptance, and accountability.",
			"Branch protection, CODEOWNERS, review gates, and governance settings.",
			"Audit-ready evidence and compliance gaps.",
		},
		Checklist: []string{
			"Map changes to relevant controls or policies.",
			"Verify approval and review evidence.",
			"Check whether risk acceptance is documented.",
			"Identify missing audit trails.",
			"Check ownership and accountability.",
			"Identify segregation-of-duties issues.",
			"Check policy exceptions and expiry.",
			"Verify evidence is timestamped and attributable.",
			"Identify compliance gaps.",
			"Produce audit-ready recommendations.",
			"Check retention and access-review evidence.",
			"Review change-management evidence.",
		},
		DecisionRules: []string{
			"If compliance governance reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak compliance governance reviewer evidence.",
			"Incorrect or unsafe compliance governance reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The compliance governance reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"release-readiness-reviewer": {
		Purpose: "Determine whether a change or system is ready for release by reviewing test status, security findings, known issues, rollback, migrations, monitoring, alerts, runbooks, feature flags, approvals, release notes, support readiness, and go/no-go recommendation.",
		When: []string{
			"Release Readiness Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the release readiness reviewer scope and gather minimum relevant repository evidence.",
			"Apply release readiness reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Test status, validation evidence, security findings, and known issues.",
			"Rollback, migrations, compatibility, feature flags, and staged rollout.",
			"Monitoring, alerts, runbooks, support readiness, and ownership.",
			"Approvals, release notes, communication, and go/no-go decision.",
			"Blockers, exceptions, residual risk, and follow-up actions.",
		},
		Checklist: []string{
			"Verify required tests and validations.",
			"Identify unresolved blockers.",
			"Check security findings and exceptions.",
			"Verify rollback plan.",
			"Check migration and compatibility risks.",
			"Verify monitoring and alerting readiness.",
			"Check runbooks and operational ownership.",
			"Check release notes and support communication.",
			"Identify feature flag or staged rollout options.",
			"Provide clear go/no-go recommendation.",
			"Confirm approvals and release owner.",
			"Check known issues and residual risks.",
		},
		DecisionRules: []string{
			"If release readiness reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak release readiness reviewer evidence.",
			"Incorrect or unsafe release readiness reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The release readiness reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"observability-reviewer": {
		Purpose: "Review logging, metrics, tracing, alerting, dashboards, operational visibility, SLOs, SLIs, runbooks, audit logs, sensitive data in logs, on-call usability, and incident detection.",
		When: []string{
			"Observability Reviewer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the observability reviewer scope and gather minimum relevant repository evidence.",
			"Apply observability reviewer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Logs, metrics, traces, alerts, dashboards, and runbooks.",
			"SLOs, SLIs, critical journeys, audit logs, and incident detection.",
			"Sensitive data in logs, correlation IDs, retention, and cardinality.",
			"On-call usability, alert routing, alert fatigue, and ownership.",
			"Operational decision support and recovery verification.",
		},
		Checklist: []string{
			"Identify critical user journeys and required signals.",
			"Check whether errors are observable.",
			"Check whether latency and saturation metrics exist.",
			"Check whether logs include useful correlation IDs.",
			"Check whether sensitive data is excluded from logs.",
			"Verify alert quality and ownership.",
			"Check dashboards for operational decisions.",
			"Review runbook coverage.",
			"Identify missing audit logs.",
			"Recommend concrete metrics and alerts.",
			"Check SLO and SLI coverage.",
			"Assess on-call usability and alert fatigue.",
		},
		DecisionRules: []string{
			"If observability reviewer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak observability reviewer evidence.",
			"Incorrect or unsafe observability reviewer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The observability reviewer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"incident-postmortem-assistant": {
		Purpose: "Support incident response, postmortems, and corrective actions across triage, severity, impact, timeline, containment, eradication, recovery, communication, evidence preservation, root cause, contributing factors, corrective actions, and prevention.",
		When: []string{
			"Incident Postmortem Assistant work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the incident postmortem assistant scope and gather minimum relevant repository evidence.",
			"Apply incident postmortem assistant specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"Triage, severity, impact, timeline, containment, eradication, and recovery.",
			"Communication, evidence preservation, facts, assumptions, and unknowns.",
			"Root cause, contributing factors, corrective actions, prevention, and owners.",
			"Security, customer, business, compliance, and operational impact.",
			"Postmortem readiness and follow-up issue quality.",
		},
		Checklist: []string{
			"Separate facts, assumptions, and unknowns.",
			"Establish incident timeline.",
			"Identify customer, business, security, and operational impact.",
			"Recommend safe containment steps.",
			"Preserve evidence and avoid destructive actions.",
			"Identify root cause and contributing factors.",
			"Distinguish immediate mitigations from long-term fixes.",
			"Identify owners and deadlines for corrective actions.",
			"Prepare stakeholder communication.",
			"Produce postmortem-ready structure.",
			"Classify severity and escalation needs.",
			"Verify recovery and prevention actions.",
		},
		DecisionRules: []string{
			"If incident postmortem assistant evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak incident postmortem assistant evidence.",
			"Incorrect or unsafe incident postmortem assistant control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The incident postmortem assistant review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"documentation-maintainer": {
		Purpose: "Keep technical and operational documentation accurate and useful across README, architecture docs, runbooks, API docs, changelogs, setup instructions, configuration docs, examples, troubleshooting, ownership, freshness, and consistency.",
		When: []string{
			"Documentation Maintainer work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the documentation maintainer scope and gather minimum relevant repository evidence.",
			"Apply documentation maintainer specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"README, architecture docs, ADRs, API docs, setup guides, and examples.",
			"Runbooks, troubleshooting, ownership, support contacts, and operational docs.",
			"Changelogs, release notes, freshness, consistency, and source-of-truth rules.",
			"Configuration docs, CLI docs, platform docs, and generated outputs.",
			"Secret-safe documentation and actionable task orientation.",
		},
		Checklist: []string{
			"Identify docs affected by code or config changes.",
			"Check whether setup instructions still work.",
			"Check API or CLI documentation consistency.",
			"Update changelog or release notes where needed.",
			"Verify examples are current.",
			"Check ownership and support contacts.",
			"Identify missing troubleshooting guidance.",
			"Ensure docs avoid leaking secrets.",
			"Prefer concise, task-oriented documentation.",
			"Mark assumptions and outdated sections.",
			"Verify source-of-truth and generated-file guidance.",
			"Check runbooks for validation and rollback steps.",
		},
		DecisionRules: []string{
			"If documentation maintainer evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak documentation maintainer evidence.",
			"Incorrect or unsafe documentation maintainer control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The documentation maintainer review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
	"universal-skill-creator": {
		Purpose: "Create new production-ready skills and prevent generic copy-paste skills by enforcing full frontmatter, SemVer, dates, authors, stability, min_platform_version, changelog, domain-specific scope, checklist, decision rules, finding categories, severity guidance, outputs, acceptance criteria, anti-patterns, and no generic body reuse.",
		When: []string{
			"Universal Skill Creator work needs structured analysis or review.",
			"The request touches production, security, compliance, release, or operational risk.",
			"Repository evidence must be turned into actionable findings.",
			"A reviewer needs a clear pass, conditional pass, or block recommendation.",
			"The central agent routes to this skill.",
		},
		Operating: []string{
			"Define the universal skill creator scope and gather minimum relevant repository evidence.",
			"Apply universal skill creator specific checks before using generic DevSecOps guidance.",
			"Classify findings by severity and impact.",
			"Map each recommendation to a concrete owner action or validation step.",
			"Report assumptions, gaps, and residual risk clearly.",
		},
		ReviewScope: []string{
			"YAML frontmatter, SemVer, since, last_modified, authors, stability, and min_platform_version.",
			"Body changelog, purpose, review scope, checklist, decision rules, and finding categories.",
			"Severity guidance, output requirements, acceptance criteria, and anti-patterns.",
			"Generic body reuse detection and platform compatibility honesty.",
			"Skill routing, generated copies, validation, and governance preservation.",
		},
		Checklist: []string{
			"Validate full YAML frontmatter before body approval.",
			"Check SemVer, since, last_modified, authors, stability, and changelog.",
			"Require min_platform_version entries from the central compatibility matrix for all supported platforms.",
			"Require body-level changelog.",
			"Write skill-specific purpose and review scope.",
			"Write at least 10 domain-specific checklist items.",
			"Write at least 5 domain-specific decision rules.",
			"Write finding categories and severity guidance.",
			"Write output requirements and acceptance criteria.",
			"Write anti-patterns that prevent misuse.",
			"Reject skills that only differ by name and description.",
			"Verify generated copies or platform outputs stay synchronized.",
		},
		DecisionRules: []string{
			"Never create a skill that only differs by name and description. Every generated skill must include domain-specific review scope, checklist items, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns. Generic operating-model text is allowed only as shared baseline, never as the complete skill body.",
			"If universal skill creator evidence is missing, report the gap instead of assuming success.",
			"If a finding affects production safety, security, compliance, or user impact, raise severity and require owner action.",
			"If repository evidence contradicts the request, document the conflict and ask for a decision owner.",
			"If validation cannot be run, state why and identify residual risk.",
			"If a recommendation changes governance, release, or security posture, require explicit review.",
			"If scope is unclear, separate confirmed scope from assumptions before giving a recommendation.",
		},
		FindingCategories: []string{
			"Missing or weak universal skill creator evidence.",
			"Incorrect or unsafe universal skill creator control.",
			"Unclear ownership, approval, or decision record.",
			"Validation, test, or monitoring gap.",
			"Security, compliance, privacy, or operational risk.",
			"Documentation or traceability gap.",
		},
		SeverityGuidance: []string{
			"Critical: immediate risk to production safety, secrets, regulated data, or release integrity.",
			"High: credible security, compliance, reliability, or rollback risk needs owner action before merge or release.",
			"Medium: meaningful quality, maintainability, observability, or process gap should be planned and tracked.",
			"Low: advisory improvement, documentation gap, or cleanup with limited immediate impact.",
		},
		OutputRequirements: []string{
			"Lead with findings ordered by severity and tied to repository evidence.",
			"List files, systems, workflows, controls, or artifacts reviewed.",
			"State validation performed, missing validation, and residual risk.",
			"Provide concrete remediation or next action for each significant finding.",
			"Separate confirmed facts, assumptions, and open questions.",
			"End with a clear recommendation: pass, conditional pass, or block where applicable.",
		},
		AcceptanceCriteria: []string{
			"The universal skill creator review scope is explicit and skill-specific.",
			"At least one concrete evidence source or explicit evidence gap is recorded.",
			"Critical and High findings include owner-oriented remediation guidance.",
			"Validation evidence or validation gaps are reported honestly.",
			"Security, governance, and DevSecOps guardrails remain intact.",
			"The final recommendation is actionable without generic filler.",
		},
		AntiPatterns: []string{
			"Creating a skill that only changes name and description while reusing generic body content.",
			"Using generic checklist language that does not mention the skill domain.",
			"Claiming success without repository evidence or validation.",
			"Mixing assumptions with verified facts.",
			"Downgrading security, compliance, or release risk for convenience.",
			"Producing findings without concrete remediation guidance.",
			"Ignoring ownership, follow-up, or residual risk.",
		},
	},
}
var skillBodies = buildSkillBodies()

func buildSkillBodies() map[string]string {
	bodies := map[string]string{}
	for _, name := range SDLCSkillNames {
		bodies[name] = buildSkillBody(name, sdlcSkillContent[name])
	}
	return bodies
}

func buildSkillBody(name string, content skillContent) string {
	var b strings.Builder
	b.WriteString("# {{.Title}}\n\n")
	writeParagraph(&b, "Purpose", content.Purpose)
	writeBullets(&b, "When to use", content.When, false)
	writeNumbered(&b, "Operating model", content.Operating)
	writeBullets(&b, "Skill-Specific Review Scope", content.ReviewScope, false)
	writeBullets(&b, "Skill-Specific Checklist", content.Checklist, true)
	writeBullets(&b, "Decision Rules", content.DecisionRules, false)
	writeBullets(&b, "Finding Categories", content.FindingCategories, false)
	writeBullets(&b, "Severity Guidance", content.SeverityGuidance, false)
	writeBullets(&b, "DevSecOps Guardrails", sharedDevSecOpsGuardrails, false)
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

func skillTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
