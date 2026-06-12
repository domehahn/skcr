package platforms

type DeliveryMode string

const (
	DeliverySkills   DeliveryMode = "skills"
	DeliveryCommands DeliveryMode = "commands"
	DeliveryBoth     DeliveryMode = "both"
)

type ToolCapability struct {
	Name               string
	SkillPathPattern   string
	CommandPathPattern string
	Delivery           DeliveryMode
	Status             string
	Source             string
}

var ToolCapabilities = []ToolCapability{
	{Name: "amazon-q", SkillPathPattern: ".amazonq/skills/%s/SKILL.md", CommandPathPattern: ".amazonq/prompts/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "antigravity", SkillPathPattern: ".agent/skills/%s/SKILL.md", CommandPathPattern: ".agent/workflows/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "auggie", SkillPathPattern: ".augment/skills/%s/SKILL.md", CommandPathPattern: ".augment/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "bob", SkillPathPattern: ".bob/skills/%s/SKILL.md", CommandPathPattern: ".bob/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "claude-code", SkillPathPattern: ".claude/skills/%s/SKILL.md", CommandPathPattern: ".claude/commands/opsx/%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "cline", SkillPathPattern: ".cline/skills/%s/SKILL.md", CommandPathPattern: ".clinerules/workflows/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "codebuddy", SkillPathPattern: ".codebuddy/skills/%s/SKILL.md", CommandPathPattern: ".codebuddy/commands/opsx/%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "codex", SkillPathPattern: ".codex/skills/%s/SKILL.md", CommandPathPattern: "$CODEX_HOME/prompts/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "continue", SkillPathPattern: ".continue/skills/%s/SKILL.md", CommandPathPattern: ".continue/prompts/opsx-%s.prompt", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "costrict", SkillPathPattern: ".cospec/skills/%s/SKILL.md", CommandPathPattern: ".cospec/openspec/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "crush", SkillPathPattern: ".crush/skills/%s/SKILL.md", CommandPathPattern: ".crush/commands/opsx/%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "cursor", SkillPathPattern: ".cursor/skills/%s/SKILL.md", CommandPathPattern: ".cursor/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "factory", SkillPathPattern: ".factory/skills/%s/SKILL.md", CommandPathPattern: ".factory/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "forgecode", SkillPathPattern: ".forge/skills/%s/SKILL.md", CommandPathPattern: "", Delivery: DeliverySkills, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "gemini-cli", SkillPathPattern: ".gemini/skills/%s/SKILL.md", CommandPathPattern: ".gemini/commands/opsx/%s.toml", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "github-copilot", SkillPathPattern: ".github/skills/%s/SKILL.md", CommandPathPattern: ".github/prompts/opsx-%s.prompt.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "iflow", SkillPathPattern: ".iflow/skills/%s/SKILL.md", CommandPathPattern: ".iflow/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "junie", SkillPathPattern: ".junie/skills/%s/SKILL.md", CommandPathPattern: ".junie/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "kilocode", SkillPathPattern: ".kilocode/skills/%s/SKILL.md", CommandPathPattern: ".kilocode/workflows/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "kimi", SkillPathPattern: ".kimi/skills/%s/SKILL.md", CommandPathPattern: "", Delivery: DeliverySkills, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "kiro", SkillPathPattern: ".kiro/skills/%s/SKILL.md", CommandPathPattern: ".kiro/prompts/opsx-%s.prompt.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "lingma", SkillPathPattern: ".lingma/skills/%s/SKILL.md", CommandPathPattern: ".lingma/commands/opsx/%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "opencode", SkillPathPattern: ".opencode/skills/%s/SKILL.md", CommandPathPattern: ".opencode/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "pi", SkillPathPattern: ".pi/skills/%s/SKILL.md", CommandPathPattern: ".pi/prompts/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "qoder", SkillPathPattern: ".qoder/skills/%s/SKILL.md", CommandPathPattern: ".qoder/commands/opsx/%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "qwen", SkillPathPattern: ".qwen/skills/%s/SKILL.md", CommandPathPattern: ".qwen/commands/opsx-%s.toml", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "roo-code", SkillPathPattern: ".roo/skills/%s/SKILL.md", CommandPathPattern: ".roo/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
	{Name: "windsurf", SkillPathPattern: ".windsurf/skills/%s/SKILL.md", CommandPathPattern: ".windsurf/commands/opsx-%s.md", Delivery: DeliveryBoth, Status: "unverified", Source: "openspec-supported-tools"},
}

func CapabilityFor(name string) (ToolCapability, bool) {
	for _, capability := range ToolCapabilities {
		if capability.Name == name {
			return capability, true
		}
	}
	return ToolCapability{}, false
}
