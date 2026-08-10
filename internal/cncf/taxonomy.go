package cncf

import "sort"

var semanticCategorySet = map[string]bool{
	"security":                   true,
	"identity-secrets":           true,
	"supply-chain":               true,
	"infrastructure-as-code":     true,
	"storage":                    true,
	"runtime-containers":         true,
	"networking-service-mesh":    true,
	"orchestration-scheduling":   true,
	"api-integration":            true,
	"databases":                  true,
	"messaging-streaming":        true,
	"developer-platforms":        true,
	"cicd-gitops":                true,
	"kubernetes":                 true,
	"cloud-platform":             true,
	"serverless":                 true,
	"release-feature-management": true,
	"reliability-chaos":          true,
	"performance-finops":         true,
	"observability":              true,
	"organizations-members":      true,
	"service-providers":          true,
	"training-certification":     true,
	"wasm":                       true,
	"languages":                  true,
	"frameworks":                 true,
	"edge-iot":                   true,
	"ai-ml-data":                 true,
	"developer-tools":            true,
	"distributed-systems":        true,
	"ai-agents":                  true,
	"governance-compliance":      true,
}

func SemanticCategoryNames() []string {
	names := make([]string, 0, len(semanticCategorySet))
	for name := range semanticCategorySet {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func semanticCategories(placement Placement) []string {
	category, subcategory := placement.Category, placement.Subcategory
	result := []string{}
	add := func(names ...string) {
		for _, name := range names {
			if !containsString(result, name) {
				result = append(result, name)
			}
		}
	}

	switch category {
	case "Provisioning":
		switch subcategory {
		case "Automation & Configuration":
			add("infrastructure-as-code", "developer-tools")
		case "Container Registry":
			add("supply-chain", "cicd-gitops")
		case "Security & Compliance":
			add("security", "governance-compliance")
		case "Key Management":
			add("identity-secrets", "security")
		}
	case "Runtime":
		switch subcategory {
		case "Cloud Native Storage":
			add("storage")
		case "Container Runtime":
			add("runtime-containers")
		case "Cloud Native Network":
			add("networking-service-mesh")
		}
	case "Orchestration & Management":
		switch subcategory {
		case "Scheduling & Orchestration":
			add("orchestration-scheduling", "kubernetes")
		case "Coordination & Service Discovery":
			add("distributed-systems", "networking-service-mesh")
		case "Remote Procedure Call":
			add("api-integration", "distributed-systems")
		case "Service Proxy", "Service Mesh":
			add("networking-service-mesh")
		case "API Gateway":
			add("api-integration", "networking-service-mesh")
		}
	case "App Definition and Development":
		switch subcategory {
		case "Database":
			add("databases")
		case "Streaming & Messaging":
			add("messaging-streaming", "distributed-systems")
		case "Application Definition & Image Build":
			add("developer-platforms", "supply-chain")
		case "Continuous Integration & Delivery":
			add("cicd-gitops", "supply-chain")
		}
	case "Platform":
		add("kubernetes", "cloud-platform")
		if subcategory == "PaaS/Container Service" {
			add("developer-platforms")
		}
	case "Serverless":
		add("serverless")
		switch subcategory {
		case "Security":
			add("security")
		case "Tools":
			add("developer-tools")
		case "Framework":
			add("frameworks")
		case "Hosted Platform", "Installable Platform":
			add("cloud-platform")
		}
	case "Observability and Analysis":
		switch subcategory {
		case "Feature Flagging":
			add("release-feature-management", "developer-tools")
		case "Chaos Engineering":
			add("reliability-chaos")
		case "Continuous Optimization":
			add("performance-finops")
		case "Observability":
			add("observability")
		}
	case "Special":
		switch subcategory {
		case "Kubernetes Certified Service Provider":
			add("service-providers")
		case "Kubernetes and Cloud Native Training Partner":
			add("training-certification")
		case "Certified CNFs":
			add("training-certification", "networking-service-mesh")
		}
	case "CNCF Members":
		add("organizations-members")
	case "Wasm":
		add("wasm")
		switch subcategory {
		case "Languages":
			add("languages")
		case "Runtimes", "Embedded Functions":
			add("runtime-containers")
		case "Application Frameworks":
			add("frameworks")
		case "Edge/Bare metal":
			add("edge-iot")
		case "AI/Machine Learning":
			add("ai-ml-data")
		case "Tooling":
			add("developer-tools")
		case "Orchestration & Management":
			add("orchestration-scheduling")
		case "Hosted Platforms":
			add("cloud-platform")
		case "Decentralized Platforms":
			add("distributed-systems")
		case "Debugging & Observability":
			add("observability")
		case "Packaging, Registries & Application Delivery":
			add("supply-chain", "cicd-gitops")
		}
	case "AI Agent":
		add("ai-agents", "ai-ml-data")
		switch subcategory {
		case "Guardrail":
			add("security", "governance-compliance")
		case "Evaluation":
			add("reliability-chaos")
		case "Knowledge Graph", "RAG", "State and Memory", "Vector Database":
			add("databases")
		case "Protocol", "Structured Output":
			add("api-integration")
		case "Workflow Orchestration":
			add("orchestration-scheduling")
		case "Agent Framework", "Agent Tool":
			add("frameworks", "developer-tools")
		}
	case "Inference":
		add("ai-ml-data")
		if subcategory == "Runtime" {
			add("runtime-containers")
		} else {
			add("frameworks")
		}
	case "Data":
		add("ai-ml-data", "databases")
	case "Training":
		add("ai-ml-data")
		if subcategory == "Distributed Training" {
			add("distributed-systems", "orchestration-scheduling")
		} else if subcategory == "Evaluation" {
			add("reliability-chaos")
		}
	case "AI Native Infra":
		add("ai-ml-data")
		switch subcategory {
		case "Orchestration and Scheduling":
			add("orchestration-scheduling")
		case "Workload Runtime", "Accelerator and SuperPod":
			add("runtime-containers")
		case "Storage", "Model Asset and Registry":
			add("storage", "supply-chain")
		case "Network", "Gateway":
			add("networking-service-mesh", "api-integration")
		case "Governance, Policy and Security":
			add("security", "governance-compliance")
		case "Observability":
			add("observability")
		case "Continuous Integration and Delivery":
			add("cicd-gitops")
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
