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
	writeBullets(&b, "Spec-Driven Change Context", sharedSpecDrivenChangeContext, false)
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
