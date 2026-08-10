# DevSecOps SDLC Skills

Skill Creator (`skcr`) includes a full SDLC-oriented skill set plus review-oriented governance, regulatory, platform, and operations skills.

Built-in categories are available through the bake filter:

```text
skcr bake --category dora-vait --write
skcr bake --category devsecops --write
skcr bake --category llmops --write
skcr bake --category agent-security --write
skcr bake --category payments --write
skcr bake --category languages --write
skcr bake --category frameworks --write
skcr bake --category infrastructure-as-code --write
skcr bake --category storage --write
skcr bake --category runtime-containers --write
skcr bake --category cncf --write
skcr bake --category cncf-landscape --write
```

Discover available categories with:

```text
skcr list categories
skcr list categories --scope semantic
skcr list categories --scope cncf
```

The default scope is `all`. Use `semantic` for the focused cross-source
taxonomy and `cncf` for CNCF maturity groups and the exact official Landscape
hierarchy.

## Programming languages

- `java-reviewer`
- `golang-reviewer`
- `python-reviewer`
- `ruby-reviewer`
- `javascript-reviewer`
- `typescript-reviewer`
- `rust-reviewer`
- `csharp-reviewer`
- `kotlin-reviewer`
- `php-reviewer`

## Frameworks and runtimes

- `spring-boot-reviewer`
- `quarkus-reviewer`
- `angular-reviewer`
- `react-reviewer`
- `vuejs-reviewer`
- `nodejs-reviewer`
- `nextjs-reviewer`
- `django-reviewer`
- `fastapi-reviewer`
- `ruby-on-rails-reviewer`

## Infrastructure technologies

- `kubernetes-platform-reviewer`
- `aws-cloud-reviewer`
- `azure-cloud-reviewer`
- `gcp-cloud-reviewer`
- `terraform-reviewer`
- `opentofu-reviewer`
- `ansible-reviewer`
- `vagrant-reviewer`
- `virtualization-reviewer`
- `helm-reviewer`

## CNCF Landscape

SKCR embeds the official CNCF Landscape snapshot and deterministically exposes
every unique entry as a `cncf-…-reviewer` skill. The 2,413 source rows currently
produce 2,407 unique skills. Six entries occur in multiple Landscape locations;
their single skill is included in every matching category.

Available filters include:

- `cncf` — the 227 active graduated, incubating, and sandbox CNCF projects
- `cncf-landscape` — every unique Landscape entry, including vendors,
  providers, training partners, and CNCF Members
- `cncf-graduated`, `cncf-incubating`, `cncf-sandbox`, and `cncf-archived`
- `cncf-provisioning`, `cncf-runtime`, `cncf-orchestration-and-management`,
  `cncf-app-definition-and-development`, `cncf-platform`, `cncf-serverless`,
  `cncf-observability-and-analysis`, `cncf-special`, `cncf-members`,
  `cncf-wasm`, `cncf-ai-agent`, `cncf-inference`, `cncf-data`, `cncf-training`,
  and `cncf-ai-native-infra`
- subcategory filters such as
  `cncf-provisioning-automation-and-configuration`

For day-to-day selection, Landscape entries are also assigned to focused
semantic categories. These avoid mixing infrastructure technologies with
members, providers, and training directories:

- infrastructure: `cloud-platform`, `kubernetes`, `infrastructure-as-code`,
  `runtime-containers`, `networking-service-mesh`, `orchestration-scheduling`,
  `serverless`, `storage`, `databases`, and `messaging-streaming`
- delivery and operations: `supply-chain`, `cicd-gitops`, `observability`,
  `reliability-operations`, `reliability-chaos`, `backup-disaster-recovery`,
  `performance-finops`, and `release-feature-management`
- security and governance: `security`, `identity-secrets`,
  `governance-compliance`, `privacy-risk`, and `provider-risk`
- software and data: `languages`, `frameworks`, `software-development`,
  `developer-tools`, `developer-platforms`, `api-integration`, `ai-ml-data`,
  `ai-agents`, `distributed-systems`, `wasm`, and `edge-iot`
- ecosystem: `organizations-members`, `service-providers`, and
  `training-certification`

Landscape inclusion is classification metadata, not CNCF endorsement,
certification, maturity, security assurance, or proof of fitness. Member skills
therefore use procurement and governance guidance; technology skills use
version-aware engineering and operational review guidance.

The embedded source, attribution, snapshot digest, and refresh command are
documented in `internal/cncf/SOURCE.md`.

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

## Agentic Security

- `agent-containment-reviewer`
- `agent-runtime-enforcement-reviewer`
- `agent-behavior-eval-engineer`
- `backdoor-persistence-reviewer`
- `agentic-threat-modeler`
- `security-invariant-test-engineer`

These skills treat an agent as a potentially untrusted principal. They design,
review, and test containment, Contract enforcement, behavior trajectories,
security invariants, hidden control paths, and persistence. They do not turn
instructions or Contract declarations into runtime guarantees: independent
sandboxes, tool gateways, firewalls, monitoring, and kill switches remain
responsible for enforcement.

## Delivery and operations

- `release-readiness-reviewer`
- `observability-reviewer`
- `incident-postmortem-assistant`

## Payments

- `payment-integration-engineer`
- `payment-security-reviewer`
- `payment-webhook-reviewer`
- `payment-flow-tester`
- `refund-dispute-handler`
- `payment-reconciliation-reviewer`
- `subscription-billing-engineer`
- `sca-3ds-reviewer`
- `payment-fraud-risk-reviewer`
- `payment-observability-reviewer`
- `payment-provider-migration-reviewer`
- `payment-compliance-reviewer`
- `payment-operations-agent`
- `stripe-integration-engineer`
- `paypal-integration-engineer`
- `adyen-integration-engineer`

Payment skills default to sandbox or test mode. Live captures, cancellations,
refunds, disputes, or other monetary mutations require exact-scope human
approval, least-privilege credentials, idempotency, reconciliation, and an
immutable audit record.

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

Agentic system:

```text
agentic-threat-modeler -> agent-containment-reviewer -> agent-runtime-enforcement-reviewer -> security-invariant-test-engineer -> agent-behavior-eval-engineer -> backdoor-persistence-reviewer
```

Payment integration:

```text
payment-integration-engineer -> payment-security-reviewer -> payment-webhook-reviewer -> payment-flow-tester -> payment-reconciliation-reviewer -> payment-observability-reviewer
```
