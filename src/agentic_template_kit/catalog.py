from __future__ import annotations


CORE_SKILLS = [
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
]

SKILL_DESCRIPTIONS = {
    "requirements-analyst": "Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.",
    "cost-based-planner": "Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.",
    "architecture-reviewer": "Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.",
    "threat-modeler": "Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.",
    "safe-implementer": "Create or modify code, tests, configuration, and project files safely with real file changes.",
    "test-strategy-engineer": "Design and generate unit, integration, regression, security, and end-to-end test strategies.",
    "verification-reviewer": "Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.",
    "security-reviewer": "Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.",
    "secrets-reviewer": "Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.",
    "dependency-supply-chain-reviewer": "Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.",
    "ci-cd-reviewer": "Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.",
    "iac-gitops-reviewer": "Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.",
    "compliance-governance-reviewer": "Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.",
    "release-readiness-reviewer": "Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.",
    "observability-reviewer": "Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.",
    "incident-postmortem-assistant": "Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.",
    "documentation-maintainer": "Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.",
    "universal-skill-creator": "Create, adapt, validate, and optimize reusable agent skills across agentic platforms.",
}

BASE_RULES = {
    "no_direct_push": True,
    "require_merge_request": True,
    "require_tests": True,
    "require_security_review": True,
    "forbid_secret_files": True,
    "forbid_env_file_access": True,
    "require_diff_summary": True,
    "require_validation_summary": True,
    "allow_autonomous_changes": False,
}

DEVSECOPS_FLOWS = [
    "secure-code-change",
    "documentation-review",
    "ci-cd-review",
    "dependency-review",
    "security-policy-review",
    "iac-gitops-review",
    "release-readiness-review",
    "incident-postmortem",
]


def skill_title(skill: str) -> str:
    return skill.replace("-", " ").title()


def skill_description(skill: str) -> str:
    return SKILL_DESCRIPTIONS.get(skill, f"Reusable agent skill for {skill_title(skill)} tasks.")
