# DevSecOps SDLC Skills

Skill Creator (`skcr`) includes a full SDLC-oriented skill set plus review-oriented governance, regulatory, platform, and operations skills.

Category targets can be generated directly:

```text
skcr bake dora-vait --write
skcr bake devsecops --write
skcr bake cloudops-platformops --write
```

The same categories are available as a bake filter:

```text
skcr bake --category dora-vait --write
skcr bake --category devsecops --write
skcr bake --category llmops --write
```

Discover available categories with:

```text
skcr list categories
```

## Planning

- `requirements-analyst`
- `cost-based-planner`
- `architecture-reviewer`
- `threat-modeler`

## Implementation and testing

- `safe-implementer`
- `test-strategy-engineer`
- `verification-reviewer`

## Security and governance

- `security-reviewer`
- `secrets-reviewer`
- `dependency-supply-chain-reviewer`
- `ci-cd-reviewer`
- `iac-gitops-reviewer`
- `compliance-governance-reviewer`
- `privacy-data-protection-reviewer`
- `api-contract-reviewer`
- `secure-design-reviewer`
- `policy-as-code-reviewer`
- `container-security-reviewer`
- `identity-access-reviewer`
- `risk-acceptance-reviewer`
- `secure-code-reviewer`
- `performance-scalability-reviewer`
- `migration-change-reviewer`
- `sbom-vulnerability-management-reviewer`
- `developer-experience-reviewer`
- `resilience-reviewer`
- `backup-restore-reviewer`

## DORA / Regulatory / Audit

- `dora-readiness-reviewer`
- `ict-risk-management-reviewer`
- `ict-third-party-risk-reviewer`
- `ict-incident-reporting-reviewer`
- `operational-resilience-tester`
- `audit-evidence-reviewer`
- `control-mapping-reviewer`
- `outsourcing-exit-strategy-reviewer`

These skills provide review and evidence orientation for regulatory readiness. They do not provide legal advice.

## Documentation and evidence

- `documentation-governance-reviewer`
- `runbook-playbook-maintainer`
- `architecture-decision-recorder`
- `audit-traceability-maintainer`
- `policy-documentation-maintainer`
- `evidence-package-creator`

## DevSecOps

- `devsecops-maturity-reviewer`
- `pipeline-security-architect`
- `software-supply-chain-architect`
- `policy-as-code-engineer`
- `secure-developer-platform-reviewer`
- `vulnerability-management-coordinator`

## CloudOps / PlatformOps

- `cloud-landing-zone-reviewer`
- `cloud-governance-reviewer`
- `finops-reviewer`
- `sre-reliability-reviewer`
- `kubernetes-platform-reviewer`
- `gitops-operations-reviewer`

## AIOps / MLOps / LLMOps

- `aiops-signal-correlation-reviewer`
- `alert-quality-reviewer`
- `auto-remediation-reviewer`
- `mlops-governance-reviewer`
- `llmops-security-reviewer`
- `ai-change-risk-reviewer`

## Delivery and operations

- `release-readiness-reviewer`
- `observability-reviewer`
- `incident-postmortem-assistant`

## Knowledge and reuse

- `documentation-maintainer`
- `universal-skill-creator`

## Recommended chains

Feature:

```text
requirements-analyst -> cost-based-planner -> safe-implementer -> test-strategy-engineer -> verification-reviewer -> security-reviewer -> documentation-maintainer -> release-readiness-reviewer
```

CI/CD:

```text
cost-based-planner -> ci-cd-reviewer -> secrets-reviewer -> security-reviewer -> verification-reviewer
```

IaC/GitOps:

```text
cost-based-planner -> iac-gitops-reviewer -> security-reviewer -> compliance-governance-reviewer -> verification-reviewer
```
