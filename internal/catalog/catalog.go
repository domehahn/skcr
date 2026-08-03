package catalog

import (
	"fmt"
	"sort"
	"strings"
)

var CoreSkills = []string{
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

var additionalCoreSkills = []string{
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

var SkillCategories = map[string][]string{
	"dora-vait": {
		"dora-readiness-reviewer",
		"ict-risk-management-reviewer",
		"ict-third-party-risk-reviewer",
		"ict-incident-reporting-reviewer",
		"operational-resilience-tester",
		"audit-evidence-reviewer",
		"control-mapping-reviewer",
		"outsourcing-exit-strategy-reviewer",
	},
	"documentation-evidence": {
		"documentation-governance-reviewer",
		"runbook-playbook-maintainer",
		"architecture-decision-recorder",
		"audit-traceability-maintainer",
		"policy-documentation-maintainer",
		"evidence-package-creator",
	},
	"devsecops": {
		"devsecops-maturity-reviewer",
		"pipeline-security-architect",
		"software-supply-chain-architect",
		"policy-as-code-engineer",
		"secure-developer-platform-reviewer",
		"vulnerability-management-coordinator",
	},
	"cloudops-platformops": {
		"cloud-landing-zone-reviewer",
		"cloud-governance-reviewer",
		"finops-reviewer",
		"sre-reliability-reviewer",
		"kubernetes-platform-reviewer",
		"gitops-operations-reviewer",
	},
	"aiops-mlops-llmops": {
		"aiops-signal-correlation-reviewer",
		"alert-quality-reviewer",
		"auto-remediation-reviewer",
		"mlops-governance-reviewer",
		"llmops-security-reviewer",
		"ai-change-risk-reviewer",
	},
	"agentic-security": {
		"agent-containment-reviewer",
		"agent-runtime-enforcement-reviewer",
		"agent-behavior-eval-engineer",
		"backdoor-persistence-reviewer",
		"agentic-threat-modeler",
		"security-invariant-test-engineer",
	},
	"security-governance": {
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
	},
}

var SkillCategoryAliases = map[string]string{
	"dora":                       "dora-vait",
	"vait":                       "dora-vait",
	"audit":                      "dora-vait",
	"regulatory":                 "dora-vait",
	"documentation":              "documentation-evidence",
	"docs":                       "documentation-evidence",
	"evidence":                   "documentation-evidence",
	"cloudops":                   "cloudops-platformops",
	"platformops":                "cloudops-platformops",
	"cloud":                      "cloudops-platformops",
	"platform":                   "cloudops-platformops",
	"aiops":                      "aiops-mlops-llmops",
	"mlops":                      "aiops-mlops-llmops",
	"llmops":                     "aiops-mlops-llmops",
	"agent-security":             "agentic-security",
	"agentic":                    "agentic-security",
	"agentic-security":           "agentic-security",
	"security":                   "security-governance",
	"governance":                 "security-governance",
	"security-compliance":        "security-governance",
	"security-and-governance":    "security-governance",
	"documentation-and-evidence": "documentation-evidence",
	"cloudops-platformops":       "cloudops-platformops",
	"aiops-mlops-llmops":         "aiops-mlops-llmops",
	"dora-vait":                  "dora-vait",
	"devsecops":                  "devsecops",
}

func CategoryNames() []string {
	names := make([]string, 0, len(SkillCategories))
	for name := range SkillCategories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func NormalizeCategory(category string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(category))
	key = strings.ReplaceAll(key, "/", "-")
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.ReplaceAll(key, " ", "-")
	if canonical, ok := SkillCategoryAliases[key]; ok {
		return canonical, nil
	}
	if _, ok := SkillCategories[key]; ok {
		return key, nil
	}
	return "", fmt.Errorf("unknown skill category %q (available: %s)", category, strings.Join(CategoryNames(), ", "))
}

func SkillsForCategory(category string) ([]string, error) {
	canonical, err := NormalizeCategory(category)
	if err != nil {
		return nil, err
	}
	return append([]string(nil), SkillCategories[canonical]...), nil
}

var SkillDescriptions = map[string]string{
	"requirements-analyst":             "Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.",
	"cost-based-planner":               "Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.",
	"architecture-reviewer":            "Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.",
	"threat-modeler":                   "Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.",
	"safe-implementer":                 "Create or modify code, tests, configuration, and project files safely with real file changes.",
	"test-strategy-engineer":           "Design and generate unit, integration, regression, security, and end-to-end test strategies.",
	"verification-reviewer":            "Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.",
	"security-reviewer":                "Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.",
	"secrets-reviewer":                 "Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.",
	"dependency-supply-chain-reviewer": "Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.",
	"ci-cd-reviewer":                   "Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.",
	"iac-gitops-reviewer":              "Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.",
	"compliance-governance-reviewer":   "Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.",
	"release-readiness-reviewer":       "Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.",
	"observability-reviewer":           "Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.",
	"incident-postmortem-assistant":    "Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.",
	"documentation-maintainer":         "Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.",
	"universal-skill-creator":          "Create, adapt, validate, and optimize reusable agent skills across agentic platforms.",
}

var additionalSkillDescriptions = map[string]string{
	"dora-readiness-reviewer":                "Review DORA readiness for ICT risk management, resilience testing, incidents, third-party risk, roles, policies, evidence, and auditability.",
	"ict-risk-management-reviewer":           "Review ICT risks, protection needs, criticality, controls, residual risks, treatment, and recurring reassessment.",
	"ict-third-party-risk-reviewer":          "Review cloud, SaaS, outsourcing, subcontractors, contracts, exit strategies, concentration risks, and DORA information-register readiness.",
	"ict-incident-reporting-reviewer":        "Review ICT incident classification, escalation, documentation, reportability, timelines, responsibilities, templates, and communication chains.",
	"operational-resilience-tester":          "Review backup and restore, failover, disaster recovery, restart procedures, crisis exercises, scenario tests, and lessons learned.",
	"audit-evidence-reviewer":                "Review evidence, approvals, tickets, logs, test protocols, risk decisions, versioning, and accountable owners.",
	"control-mapping-reviewer":               "Map technical measures to DORA, VAIT or BAIT migration needs, ISO 27001, BSI, internal policies, or MaRisk review expectations.",
	"outsourcing-exit-strategy-reviewer":     "Review exit plans, data return, provider transitions, emergency operations, suboutsourcing, cloud dependencies, and business impact.",
	"documentation-governance-reviewer":      "Review documentation freshness, ownership, review cycles, approvals, versioning, validity, and traceability.",
	"runbook-playbook-maintainer":            "Create and review runbooks, operating instructions, incident playbooks, escalation paths, restart procedures, and checklists.",
	"architecture-decision-recorder":         "Create and maintain ADRs with context, decisions, alternatives, risks, security impact, compliance relation, and review points.",
	"audit-traceability-maintainer":          "Link requirements, controls, implementation, tests, tickets, and evidence into an auditable trace.",
	"policy-documentation-maintainer":        "Create and update policies, standards, procedures, and control descriptions.",
	"evidence-package-creator":               "Create auditable evidence packages from tickets, pipeline results, test reports, approvals, scans, and architecture information.",
	"devsecops-maturity-reviewer":            "Assess maturity across plan, code, build, test, release, deploy, and operate with automation, security gates, ownership, and feedback loops.",
	"pipeline-security-architect":            "Design and review secure CI/CD pipelines with isolated runners, minimal rights, OIDC, signed artifacts, protected environments, and approval gates.",
	"software-supply-chain-architect":        "Review SLSA, provenance, SBOM, signatures, attestations, build integrity, artifact promotion, and trusted builders.",
	"policy-as-code-engineer":                "Create and review policies for OPA/Rego, Kyverno, GitLab Policies, Conftest, Checkov, Terraform, Kubernetes, and CI/CD gates.",
	"secure-developer-platform-reviewer":     "Review Internal Developer Platforms for secure golden paths, self-service guardrails, templates, permission models, secrets handling, and auditability.",
	"vulnerability-management-coordinator":   "Assess CVE triage, prioritization, SLAs, exploitability, asset criticality, exceptions, risk acceptance, and remediation tracking.",
	"cloud-landing-zone-reviewer":            "Review cloud accounts or subscriptions, networks, IAM, logging, policies, baselines, guardrails, encryption, tagging, and tenant separation.",
	"cloud-governance-reviewer":              "Review cloud naming, tags, ownership, cost centers, allowed services, regions, data classification, policy enforcement, and audit evidence.",
	"finops-reviewer":                        "Review cloud costs, budgets, rightsizing, reserved or committed usage, anomalies, showback or chargeback, and team cost transparency.",
	"sre-reliability-reviewer":               "Assess SLOs, SLIs, error budgets, capacity, degradation, timeouts, retries, circuit breakers, load shedding, and operational risks.",
	"kubernetes-platform-reviewer":           "Review Kubernetes clusters, namespaces, RBAC, NetworkPolicies, Pod Security, admission controllers, resource limits, secrets, ingress, tenancy, and upgrades.",
	"gitops-operations-reviewer":             "Review Argo CD or Flux setups, sync policies, drift detection, promotion, rollback, app-of-apps, secrets, cluster access, and deployment governance.",
	"aiops-signal-correlation-reviewer":      "Assess correlation of logs, metrics, traces, events, and incidents to reduce noise, improve root-cause analysis, and lower alert fatigue.",
	"alert-quality-reviewer":                 "Review alerts for actionability, clear symptoms, runbook links, severity, ownership, SLO relation, deduplication, escalation, and remediation suitability.",
	"auto-remediation-reviewer":              "Review automated repair actions for safe limits, dry runs, approval modes, rollback, audit logs, blast radius, and loop protection.",
	"mlops-governance-reviewer":              "Review model versioning, training data, bias, drift, monitoring, approvals, reproducibility, model registry, and deployment gates.",
	"llmops-security-reviewer":               "Review GenAI workloads for prompt injection, tool permissions, data exfiltration, RAG sources, sensitive prompt logging, evals, guardrails, and model access.",
	"ai-change-risk-reviewer":                "Review AI-assisted changes before execution for automation boundaries, human approval, affected-system criticality, and audit evidence.",
	"agent-containment-reviewer":             "Review agent sandboxes, transitive egress, privilege escalation, lateral movement, shared infrastructure, monitoring, and kill switches.",
	"agent-runtime-enforcement-reviewer":     "Compare declared agent contracts with independent runtime enforcement across tools, files, processes, networks, secrets, approvals, and limits.",
	"agent-behavior-eval-engineer":           "Design trajectory-level agent security evals for Goal compliance, contract boundaries, goal hacking, unsafe tools, and long-horizon behavior.",
	"backdoor-persistence-reviewer":          "Review changes for hidden privileged paths, triggers, covert egress, persistence, security-control tampering, and unexplained behavior.",
	"agentic-threat-modeler":                 "Threat-model agents as untrusted principals across prompts, tools, MCP, memory, delegation, runtime infrastructure, and external systems.",
	"security-invariant-test-engineer":       "Derive negative tests and declarative evals from contract capabilities, tools, data flows, approvals, limits, and structured invariants.",
	"privacy-data-protection-reviewer":       "Review privacy, personal data, data classification, deletion concepts, purpose limitation, GDPR risks, and sensitive-data logging.",
	"api-contract-reviewer":                  "Review REST, GraphQL, OpenAPI, and gRPC contracts, breaking changes, versioning, AuthN/AuthZ, error formats, and compatibility.",
	"secure-design-reviewer":                 "Review secure-by-design decisions, least privilege, Zero Trust, tenant separation, secure defaults, and abuse scenarios.",
	"policy-as-code-reviewer":                "Review GitLab Security Policies, OPA/Rego, Kyverno, Conftest, Sentinel, admission policies, compliance pipelines, and central guardrails.",
	"container-security-reviewer":            "Review Dockerfiles, base images, user rights, capabilities, SBOM, image signing, distroless or slim images, CVEs, and runtime hardening.",
	"identity-access-reviewer":               "Review IAM, roles, service accounts, groups, tokens, OIDC federation, GitLab or GitHub permissions, cloud rights, and privilege-escalation paths.",
	"risk-acceptance-reviewer":               "Document and assess conscious risk decisions, impact and likelihood, expiry dates, and compensating measures.",
	"secure-code-reviewer":                   "Review code vulnerabilities such as injection, path traversal, SSRF, XSS, deserialization, crypto misuse, and race conditions.",
	"performance-scalability-reviewer":       "Review load behavior, bottlenecks, caching, database access, queue behavior, scaling, timeouts, and resource limits.",
	"migration-change-reviewer":              "Review database migrations, schema changes, breaking changes, rollback ability, backward compatibility, and zero-downtime deployments.",
	"sbom-vulnerability-management-reviewer": "Review SBOM generation, CVE triage, VEX, exception processes, patch SLAs, and the vulnerability lifecycle.",
	"developer-experience-reviewer":          "Review setup, local development, error messages, Makefiles or scripts, onboarding, tooling consistency, and practicality for teams.",
	"resilience-reviewer":                    "Review timeouts, retries, circuit breakers, failover, backpressure, degraded modes, and resilience behavior.",
	"backup-restore-reviewer":                "Review restore tests, RPO/RTO, data integrity, backup protection, recoverability, and disaster recovery.",
}

func init() {
	CoreSkills = append(CoreSkills, additionalCoreSkills...)
	for name, desc := range additionalSkillDescriptions {
		SkillDescriptions[name] = desc
	}
}

var BaseRules = map[string]any{
	"no_direct_push":             true,
	"require_merge_request":      true,
	"require_tests":              true,
	"require_security_review":    true,
	"forbid_secret_files":        true,
	"forbid_env_file_access":     true,
	"require_diff_summary":       true,
	"require_validation_summary": true,
	"allow_autonomous_changes":   false,
}

var DevsecopsFlows = []string{
	"secure-code-change",
	"documentation-review",
	"ci-cd-review",
	"dependency-review",
	"security-policy-review",
	"iac-gitops-review",
	"release-readiness-review",
	"incident-postmortem",
}

type ClaudeSubagent struct {
	Name           string
	Description    string
	Tools          string
	Model          string
	PermissionMode string
	MaxTurns       int
	Skills         []string
	Prompt         string
}

var ClaudeSubagents = []ClaudeSubagent{
	{
		Name:           "requirements-analyst",
		Description:    "Use proactively to analyze issues, user stories, acceptance criteria, constraints, risks, and missing requirements before implementation.",
		Tools:          "Read, Glob, Grep",
		Model:          "sonnet",
		PermissionMode: "plan",
		MaxTurns:       8,
		Skills:         []string{"requirements-analyst", "cost-based-planner"},
		Prompt:         "You are a requirements analyst.\n\nWhen invoked:\n1. Inspect the issue, task description, README, and relevant project files.\n2. Extract functional requirements, non-functional requirements, acceptance criteria, constraints, and open questions.\n3. Identify ambiguity, missing edge cases, and dependency on external systems.\n4. Do not modify files.\n5. Return a concise requirements brief with recommended next steps.",
	},
	{
		Name:           "architecture-reviewer",
		Description:    "Use proactively for architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical design risks.",
		Tools:          "Read, Glob, Grep",
		Model:          "sonnet",
		PermissionMode: "plan",
		MaxTurns:       10,
		Skills:         []string{"architecture-reviewer", "threat-modeler", "documentation-maintainer"},
		Prompt:         "You are an architecture reviewer.\n\nWhen invoked:\n1. Inspect only architecture-relevant docs and source files.\n2. Identify module boundaries, data flows, dependencies, integration points, and ownership concerns.\n3. Review coupling, cohesion, scalability, resilience, and maintainability.\n4. Do not modify files.\n5. Return findings by severity and include actionable recommendations.",
	},
	{
		Name:           "devsecops-reviewer",
		Description:    "Use proactively after code, CI/CD, dependency, IaC, GitOps, or security-sensitive changes to review DevSecOps risk and merge readiness.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       12,
		Skills: []string{
			"security-reviewer",
			"secrets-reviewer",
			"dependency-supply-chain-reviewer",
			"ci-cd-reviewer",
			"iac-gitops-reviewer",
			"compliance-governance-reviewer",
			"release-readiness-reviewer",
		},
		Prompt: "You are a DevSecOps reviewer.\n\nWhen invoked:\n1. Inspect the current git diff and relevant adjacent files.\n2. Review for security, CI/CD, secrets, dependencies, IaC/GitOps, compliance, and release risks.\n3. Do not modify files unless explicitly asked.\n4. Run only safe read-only commands unless validation commands are clearly repository-native.\n5. Report findings by severity:\n   - CRITICAL\n   - HIGH\n   - MEDIUM\n   - LOW\n6. Include required fixes, recommended fixes, validation evidence, and merge-readiness recommendation.",
	},
	{
		Name:           "security-reviewer",
		Description:    "Use proactively for authentication, authorization, input validation, file handling, permissions, secrets, and security-sensitive code changes.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       10,
		Skills:         []string{"security-reviewer", "secrets-reviewer", "threat-modeler"},
		Prompt:         "You are a security reviewer.\n\nWhen invoked:\n1. Inspect the diff and security-relevant adjacent files.\n2. Focus on auth, authorization, injection, path traversal, unsafe file handling, unsafe logging, secrets, and permission boundaries.\n3. Do not modify files unless explicitly asked.\n4. Do not print secrets.\n5. Return findings by severity with concrete remediation.",
	},
	{
		Name:           "ci-cd-reviewer",
		Description:    "Use proactively for GitLab CI, GitHub Actions, runners, deployment jobs, caches, artifacts, tokens, and pipeline governance.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       10,
		Skills:         []string{"ci-cd-reviewer", "secrets-reviewer", "security-reviewer"},
		Prompt:         "You are a CI/CD reviewer.\n\nWhen invoked:\n1. Inspect pipeline files, workflow files, scripts, includes, and deployment definitions.\n2. Review runner permissions, token exposure, artifacts, caches, environment protection, branch/MR gates, and deployment safety.\n3. Do not modify files unless explicitly asked.\n4. Return findings, required fixes, and safe validation commands.",
	},
	{
		Name:           "iac-gitops-reviewer",
		Description:    "Use proactively for Terraform, Kubernetes, Helm, Kustomize, GitOps, environment promotion, reconciliation, and infrastructure changes.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       10,
		Skills:         []string{"iac-gitops-reviewer", "security-reviewer", "compliance-governance-reviewer"},
		Prompt:         "You are an IaC and GitOps reviewer.\n\nWhen invoked:\n1. Inspect Terraform, Kubernetes, Helm, Kustomize, Argo CD, Flux, and GitOps-related files.\n2. Review environment separation, drift/reconciliation, permissions, secrets, rollout safety, and rollback.\n3. Do not run destructive infrastructure commands.\n4. Return findings, validation suggestions, and release safety notes.",
	},
	{
		Name:           "test-runner",
		Description:    "Use proactively to run relevant tests, analyze failures, and summarize validation results after implementation.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       10,
		Skills:         []string{"test-strategy-engineer", "verification-reviewer"},
		Prompt:         "You are a test execution and failure-analysis agent.\n\nWhen invoked:\n1. Detect the repository's test framework and package manager.\n2. Run the smallest relevant safe test command first.\n3. If tests fail, summarize failing tests, likely root cause, and minimal fix direction.\n4. Do not hide failing tests.\n5. Do not perform broad refactors.\n6. Return exact commands run and their results.",
	},
	{
		Name:           "release-readiness-reviewer",
		Description:    "Use proactively before release readiness claims to review tests, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.",
		Tools:          "Read, Glob, Grep",
		Model:          "sonnet",
		PermissionMode: "plan",
		MaxTurns:       8,
		Skills:         []string{"release-readiness-reviewer", "observability-reviewer", "documentation-maintainer"},
		Prompt:         "You are a release readiness reviewer.\n\nWhen invoked:\n1. Inspect the diff, release notes, migrations, config changes, tests, and operational docs.\n2. Review rollback, feature flags, observability, breaking changes, and deployment safety.\n3. Do not modify files.\n4. Return release readiness status, blockers, risks, and recommended validation.",
	},
	{
		Name:           "incident-postmortem-assistant",
		Description:    "Use for incident analysis, log summaries, timeline reconstruction, root cause analysis, corrective actions, and follow-up issues.",
		Tools:          "Read, Glob, Grep, Bash",
		Model:          "sonnet",
		PermissionMode: "default",
		MaxTurns:       12,
		Skills:         []string{"incident-postmortem-assistant", "observability-reviewer", "documentation-maintainer"},
		Prompt:         "You are an incident and postmortem assistant.\n\nWhen invoked:\n1. Analyze provided logs, timelines, symptoms, and related repository context.\n2. Build a factual timeline and separate facts from hypotheses.\n3. Identify likely root cause, contributing factors, impact, detection gaps, and corrective actions.\n4. Do not expose secrets or sensitive customer data.\n5. Return a postmortem-ready summary with action items.",
	},
}

func SkillTitle(skill string) string {
	parts := strings.Split(skill, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func SkillDescription(skill string) string {
	if desc, ok := SkillDescriptions[skill]; ok {
		return desc
	}
	return "Reusable agent skill for " + SkillTitle(skill) + " tasks."
}
