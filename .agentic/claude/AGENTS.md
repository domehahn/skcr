# Claude Code Integration

The Claude Code adapter generates both skills and subagents.

## Generated structure

```text
CLAUDE.md
.claude/
├── skills/
│   └── <skill-name>/
│       └── SKILL.md
└── agents/
    └── <subagent-name>.md
Skills vs. Subagents

Skills are reusable instruction bundles.

.claude/skills/<skill-name>/SKILL.md

Subagents are specialized Claude Code workers with their own context, tool access, model, permissions, and system prompt.

.claude/agents/<subagent-name>.md
Generated subagents

The DevSecOps SDLC edition generates these Claude Code subagents when the claude platform is enabled:

requirements-analyst
architecture-reviewer
devsecops-reviewer
security-reviewer
ci-cd-reviewer
iac-gitops-reviewer
test-runner
release-readiness-reviewer
incident-postmortem-assistant
Invocation

Claude Code can delegate automatically when a task matches a subagent description.

You can also ask explicitly:

Use the security-reviewer subagent to review this change.

or:

Use the test-runner subagent to run relevant tests and summarize failures.
Design rules
Keep review subagents mostly read-only.
Use explicit tools.
Use skills: to preload relevant skill instructions into subagent context.
Prefer bounded, specialized subagents over generic workers.
Do not use subagents as a replacement for human review on security-sensitive changes.

---

## README-Ergänzung

Diesen Block in die Plattformübersicht deiner `README.md` einfügen:

```md
### Claude Code

Generated files:

```text
CLAUDE.md
.claude/skills/<skill-name>/SKILL.md
.claude/agents/<subagent-name>.md

Claude Code skills provide reusable instructions. Claude Code subagents provide specialized workers with their own context, tools, model, permissions, and preloaded skills.

The Claude adapter generates both:

skills under .claude/skills/<skill-name>/SKILL.md
subagents under .claude/agents/<subagent-name>.md

Subagents are useful for bounded review, testing, analysis, and DevSecOps tasks that would otherwise flood the main Claude Code conversation with logs, grep output, test output, or detailed review findings.


---

## `src/agentic_template_kit/templates/claude/CLAUDE.md.j2`

```md
# CLAUDE.md

## Claude Code Instructions

Project: agentic-template-kit  
Owner team: platform-engineering  
Governance level: strict

## Required behavior

- Make real file changes when implementation is requested.
- Do not only explain how to do the work.
- Do not commit, push, deploy, publish, or merge unless explicitly asked.
- Avoid secrets and sensitive files.
- Keep changes minimal, validated, and reviewable.
- Summarize changed files, validation results, and residual risks.

## Project Skills

Claude Code project skills are stored under:

```text
.claude/skills/<skill-name>/SKILL.md

Use slash commands where supported, or apply the matching skill instructions directly.



/requirements-analyst — Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.


/cost-based-planner — Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.


/architecture-reviewer — Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.


/threat-modeler — Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.


/safe-implementer — Create or modify code, tests, configuration, and project files safely with real file changes.


/test-strategy-engineer — Design and generate unit, integration, regression, security, and end-to-end test strategies.


/verification-reviewer — Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.


/security-reviewer — Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.


/secrets-reviewer — Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.


/dependency-supply-chain-reviewer — Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.


/ci-cd-reviewer — Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.


/iac-gitops-reviewer — Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.


/compliance-governance-reviewer — Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.


/release-readiness-reviewer — Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.


/observability-reviewer — Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.


/incident-postmortem-assistant — Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.


/documentation-maintainer — Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.


/universal-skill-creator — Create, adapt, validate, and optimize reusable agent skills across agentic platforms.



Project Subagents

Claude Code project subagents are stored under:

.claude/agents/<subagent-name>.md

Use subagents for bounded side tasks that would otherwise flood the main conversation with file reads, grep results, logs, test output, or security review detail.



requirements-analyst — Use proactively to analyze issues, user stories, acceptance criteria, constraints, risks, and missing requirements before implementation.


architecture-reviewer — Use proactively for architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical design risks.


devsecops-reviewer — Use proactively after code, CI/CD, dependency, IaC, GitOps, or security-sensitive changes to review DevSecOps risk and merge readiness.


security-reviewer — Use proactively for authentication, authorization, input validation, file handling, permissions, secrets, and security-sensitive code changes.


ci-cd-reviewer — Use proactively for GitLab CI, GitHub Actions, runners, deployment jobs, caches, artifacts, tokens, and pipeline governance.


iac-gitops-reviewer — Use proactively for Terraform, Kubernetes, Helm, Kustomize, GitOps, environment promotion, reconciliation, and infrastructure changes.


test-runner — Use proactively to run relevant tests, analyze failures, and summarize validation results after implementation.


release-readiness-reviewer — Use proactively before release readiness claims to review tests, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.


incident-postmortem-assistant — Use for incident analysis, log summaries, timeline reconstruction, root cause analysis, corrective actions, and follow-up issues.

Subagent routing

Use the matching subagent proactively when the task fits its description.

Recommended routing:

Requirements or acceptance criteria unclear → requirements-analyst
Architecture or module boundaries affected → architecture-reviewer
Security-sensitive behavior affected → security-reviewer
CI/CD or deployment logic changed → ci-cd-reviewer
Terraform, Kubernetes, Helm, Kustomize, or GitOps changed → iac-gitops-reviewer
Tests should be executed or failures analyzed → test-runner
DevSecOps merge-readiness review needed → devsecops-reviewer
Release readiness must be assessed → release-readiness-reviewer
Incident, outage, logs, or postmortem analysis needed → incident-postmortem-assistant
SDLC skill routing

For feature work, prefer:

/requirements-analyst
→ /cost-based-planner
→ /architecture-reviewer when architecture is affected
→ /threat-modeler when security-sensitive behavior is affected
→ /safe-implementer
→ /test-strategy-engineer
→ /verification-reviewer
→ /security-reviewer when needed
→ /documentation-maintainer when needed
→ /release-readiness-reviewer before release readiness claims

For CI/CD, dependency, IaC, GitOps, secrets, compliance, observability, or incident tasks, use the matching specialized skill or subagent before claiming completion.

Safety model
Prefer read-only subagents for analysis and review.
Use implementation through the main Claude Code session unless a subagent is explicitly designed to modify files.
Do not let review subagents perform broad rewrites.
Treat subagent output as review input; the main session remains responsible for final user-visible conclusions.

---

## `tests/test_templates.py` Ergänzung

In `EXPECTED_TEMPLATES` muss diese Datei enthalten sein:

```python
"claude/agent.md.j2",
