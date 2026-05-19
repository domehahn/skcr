from __future__ import annotations

from pathlib import Path

import yaml


def _check_skill_dir(skills_dir: Path, issues: list[str]) -> None:
    if not skills_dir.exists():
        return
    for skill_dir in sorted(p for p in skills_dir.iterdir() if p.is_dir()):
        skill_file = skill_dir / "SKILL.md"
        if not skill_file.exists():
            issues.append(f"Missing SKILL.md in {skill_dir}")
            continue
        text = skill_file.read_text(encoding="utf-8")
        if not text.startswith("---"):
            issues.append(f"{skill_file} is missing YAML frontmatter")
        if "name:" not in text:
            issues.append(f"{skill_file} is missing 'name' metadata")
        if "description:" not in text:
            issues.append(f"{skill_file} is missing 'description' metadata")


def _validate_gitlab_flows(target: Path, issues: list[str]) -> None:
    flows_dir = target / ".gitlab" / "duo" / "flows"
    if not flows_dir.exists():
        return
    for flow in sorted(flows_dir.glob("*.yaml")):
        try:
            data = yaml.safe_load(flow.read_text(encoding="utf-8")) or {}
        except Exception as exc:  # noqa: BLE001
            issues.append(f"Invalid GitLab flow YAML {flow}: {exc}")
            continue

        for forbidden in ("name", "description", "product_group"):
            if forbidden in data:
                issues.append(f"{flow} contains forbidden top-level field '{forbidden}' for GitLab custom flows")

        if data.get("environment") != "ambient":
            issues.append(f"{flow} should set environment: ambient")

        for prompt in data.get("prompts", []) or []:
            if isinstance(prompt, dict) and "model" in prompt:
                issues.append(f"{flow} prompt '{prompt.get('id', '<unknown>')}' must not define a model field")

        agent_components = [
            component
            for component in (data.get("components", []) or [])
            if isinstance(component, dict) and component.get("type") == "AgentComponent"
        ]
        for component in agent_components:
            inputs = component.get("inputs", []) or []
            sources = {item.get("from") for item in inputs if isinstance(item, dict)}
            if "context:inputs.user_rule" not in sources:
                issues.append(f"{flow} AgentComponent '{component.get('name')}' does not receive context:inputs.user_rule")
            if "context:inputs.workspace_agent_skills" not in sources:
                issues.append(f"{flow} AgentComponent '{component.get('name')}' does not receive context:inputs.workspace_agent_skills")


def validate_target(target: Path) -> list[str]:
    issues: list[str] = []

    if not target.exists():
        return [f"Target does not exist: {target}"]

    bake = target / "agentic.bake.yaml"
    if bake.exists():
        try:
            yaml.safe_load(bake.read_text(encoding="utf-8"))
        except Exception as exc:  # noqa: BLE001
            issues.append(f"Invalid agentic.bake.yaml: {exc}")

    agents = target / "AGENTS.md"
    if agents.exists() and not agents.read_text(encoding="utf-8").strip():
        issues.append("AGENTS.md exists but is empty")

    claude = target / "CLAUDE.md"
    if claude.exists() and not claude.read_text(encoding="utf-8").strip():
        issues.append("CLAUDE.md exists but is empty")

    _check_skill_dir(target / ".agents" / "skills", issues)
    _check_skill_dir(target / ".claude" / "skills", issues)
    _check_skill_dir(target / "skills", issues)

    old_gitlab_rules = target / ".gitlab" / "duo" / "rules"
    if old_gitlab_rules.exists():
        issues.append("GitLab custom rules should render to .gitlab/duo/chat-rules.md, not .gitlab/duo/rules/**")

    old_gitlab_custom_rules = target / ".gitlab" / "duo" / "custom-rules.md"
    if old_gitlab_custom_rules.exists():
        issues.append("GitLab custom rules should render to .gitlab/duo/chat-rules.md, not .gitlab/duo/custom-rules.md")

    gitlab_chat_rules = target / ".gitlab" / "duo" / "chat-rules.md"
    if gitlab_chat_rules.exists() and not gitlab_chat_rules.read_text(encoding="utf-8").strip():
        issues.append(".gitlab/duo/chat-rules.md exists but is empty")

    _validate_gitlab_flows(target, issues)

    copilot = target / ".github" / "copilot-instructions.md"
    if copilot.exists() and "Agentic" not in copilot.read_text(encoding="utf-8"):
        issues.append(".github/copilot-instructions.md does not look like an agentic-template output")

    modelfile = target / ".ollama" / "Modelfile"
    if modelfile.exists() and "FROM" not in modelfile.read_text(encoding="utf-8"):
        issues.append(".ollama/Modelfile is missing a FROM instruction")

    return issues
