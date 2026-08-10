package catalog

import (
	"sort"

	"github.com/domehahn/skcr/v2/internal/cncf"
)

var builtInSemanticCategories = map[string][]string{
	"planning-architecture": {
		"requirements-analyst", "cost-based-planner", "architecture-reviewer", "threat-modeler",
		"architecture-decision-recorder", "api-contract-reviewer", "secure-design-reviewer", "migration-change-reviewer",
	},
	"software-development": {
		"safe-implementer", "secure-code-reviewer", "developer-experience-reviewer", "universal-skill-creator",
	},
	"testing-quality": {
		"test-strategy-engineer", "verification-reviewer", "operational-resilience-tester",
		"security-invariant-test-engineer", "agent-behavior-eval-engineer", "payment-flow-tester",
	},
	"security": {
		"security-reviewer", "threat-modeler", "secrets-reviewer", "dependency-supply-chain-reviewer",
		"ci-cd-reviewer", "iac-gitops-reviewer", "pipeline-security-architect", "software-supply-chain-architect",
		"policy-as-code-engineer", "vulnerability-management-coordinator", "llmops-security-reviewer",
		"agent-containment-reviewer", "agent-runtime-enforcement-reviewer", "backdoor-persistence-reviewer",
		"agentic-threat-modeler", "security-invariant-test-engineer", "privacy-data-protection-reviewer",
		"secure-design-reviewer", "policy-as-code-reviewer", "container-security-reviewer",
		"identity-access-reviewer", "secure-code-reviewer", "sbom-vulnerability-management-reviewer",
		"payment-security-reviewer", "payment-webhook-reviewer", "payment-fraud-risk-reviewer",
		"payment-compliance-reviewer", "sca-3ds-reviewer",
	},
	"identity-secrets": {
		"secrets-reviewer", "identity-access-reviewer", "cloud-landing-zone-reviewer",
		"payment-security-reviewer", "agent-runtime-enforcement-reviewer",
	},
	"supply-chain": {
		"dependency-supply-chain-reviewer", "software-supply-chain-architect", "pipeline-security-architect",
		"container-security-reviewer", "sbom-vulnerability-management-reviewer", "vulnerability-management-coordinator",
	},
	"governance-compliance": {
		"compliance-governance-reviewer", "dora-readiness-reviewer", "ict-risk-management-reviewer",
		"ict-third-party-risk-reviewer", "ict-incident-reporting-reviewer", "audit-evidence-reviewer",
		"control-mapping-reviewer", "outsourcing-exit-strategy-reviewer", "documentation-governance-reviewer",
		"audit-traceability-maintainer", "policy-documentation-maintainer", "evidence-package-creator",
		"devsecops-maturity-reviewer", "cloud-governance-reviewer", "mlops-governance-reviewer",
		"ai-change-risk-reviewer", "privacy-data-protection-reviewer", "risk-acceptance-reviewer",
		"payment-compliance-reviewer",
	},
	"privacy-risk": {
		"privacy-data-protection-reviewer", "risk-acceptance-reviewer", "ict-risk-management-reviewer",
		"ict-third-party-risk-reviewer", "payment-compliance-reviewer", "llmops-security-reviewer",
	},
	"documentation-evidence": {
		"documentation-maintainer", "incident-postmortem-assistant", "documentation-governance-reviewer",
		"runbook-playbook-maintainer", "architecture-decision-recorder", "audit-traceability-maintainer",
		"policy-documentation-maintainer", "evidence-package-creator", "audit-evidence-reviewer",
	},
	"cloud-platform": {
		"cloud-landing-zone-reviewer", "cloud-governance-reviewer", "finops-reviewer",
		"sre-reliability-reviewer", "kubernetes-platform-reviewer", "secure-developer-platform-reviewer",
		"aws-cloud-reviewer", "azure-cloud-reviewer", "gcp-cloud-reviewer",
	},
	"kubernetes": {
		"kubernetes-platform-reviewer", "gitops-operations-reviewer", "iac-gitops-reviewer",
		"container-security-reviewer", "policy-as-code-engineer", "policy-as-code-reviewer", "helm-reviewer",
	},
	"infrastructure-as-code": {
		"iac-gitops-reviewer", "policy-as-code-engineer", "policy-as-code-reviewer", "terraform-reviewer",
		"opentofu-reviewer", "ansible-reviewer", "vagrant-reviewer", "helm-reviewer",
	},
	"runtime-containers": {
		"container-security-reviewer", "kubernetes-platform-reviewer", "virtualization-reviewer",
		"resilience-reviewer", "performance-scalability-reviewer",
	},
	"networking-service-mesh": {
		"cloud-landing-zone-reviewer", "kubernetes-platform-reviewer", "resilience-reviewer",
	},
	"storage": {
		"backup-restore-reviewer", "migration-change-reviewer", "kubernetes-platform-reviewer",
		"aws-cloud-reviewer", "azure-cloud-reviewer", "gcp-cloud-reviewer",
	},
	"databases": {
		"migration-change-reviewer", "backup-restore-reviewer", "performance-scalability-reviewer",
	},
	"messaging-streaming": {
		"resilience-reviewer", "performance-scalability-reviewer", "observability-reviewer",
	},
	"cicd-gitops": {
		"ci-cd-reviewer", "pipeline-security-architect", "software-supply-chain-architect",
		"policy-as-code-engineer", "gitops-operations-reviewer", "iac-gitops-reviewer",
		"release-readiness-reviewer", "migration-change-reviewer",
	},
	"observability": {
		"observability-reviewer", "incident-postmortem-assistant", "aiops-signal-correlation-reviewer",
		"alert-quality-reviewer", "sre-reliability-reviewer", "payment-observability-reviewer",
	},
	"reliability-operations": {
		"release-readiness-reviewer", "incident-postmortem-assistant", "operational-resilience-tester",
		"runbook-playbook-maintainer", "sre-reliability-reviewer", "alert-quality-reviewer",
		"auto-remediation-reviewer", "resilience-reviewer", "backup-restore-reviewer",
		"payment-reconciliation-reviewer", "payment-operations-agent",
	},
	"backup-disaster-recovery": {
		"operational-resilience-tester", "backup-restore-reviewer", "resilience-reviewer",
		"runbook-playbook-maintainer", "release-readiness-reviewer",
	},
	"performance-finops": {
		"finops-reviewer", "performance-scalability-reviewer", "sre-reliability-reviewer",
		"payment-observability-reviewer",
	},
	"ai-ml-data": {
		"aiops-signal-correlation-reviewer", "mlops-governance-reviewer", "llmops-security-reviewer",
		"ai-change-risk-reviewer", "agent-behavior-eval-engineer", "agentic-threat-modeler",
	},
	"ai-agents": {
		"agent-containment-reviewer", "agent-runtime-enforcement-reviewer", "agent-behavior-eval-engineer",
		"backdoor-persistence-reviewer", "agentic-threat-modeler", "security-invariant-test-engineer",
	},
	"developer-platforms": {
		"secure-developer-platform-reviewer", "developer-experience-reviewer", "universal-skill-creator",
		"gitops-operations-reviewer", "kubernetes-platform-reviewer",
	},
	"api-integration": {
		"api-contract-reviewer", "payment-integration-engineer", "payment-webhook-reviewer",
		"stripe-integration-engineer", "paypal-integration-engineer", "adyen-integration-engineer",
	},
	"provider-risk": {
		"ict-third-party-risk-reviewer", "outsourcing-exit-strategy-reviewer", "payment-provider-migration-reviewer",
	},
	"release-feature-management": {
		"release-readiness-reviewer", "migration-change-reviewer", "ci-cd-reviewer",
	},
}

var staticSemanticCategorySet = map[string]bool{
	"payments": true, "languages": true, "frameworks": true,
}

func registerSemanticCategories() {
	for category, skills := range builtInSemanticCategories {
		SkillCategories[category] = mergeUniqueSkills(SkillCategories[category], skills)
		staticSemanticCategorySet[category] = true
	}
	for _, category := range cncf.SemanticCategoryNames() {
		staticSemanticCategorySet[category] = true
	}
	SkillCategories["software-development"] = mergeUniqueSkills(SkillCategories["software-development"], languageSkills)
	SkillCategories["software-development"] = mergeUniqueSkills(SkillCategories["software-development"], frameworkSkills)
}

func SemanticCategoryNames() []string {
	names := make([]string, 0, len(staticSemanticCategorySet))
	for name := range staticSemanticCategorySet {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func mergeUniqueSkills(target []string, values []string) []string {
	seen := make(map[string]bool, len(target)+len(values))
	for _, value := range target {
		seen[value] = true
	}
	for _, value := range values {
		if !seen[value] {
			target = append(target, value)
			seen[value] = true
		}
	}
	return target
}
