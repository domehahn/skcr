from __future__ import annotations

from pathlib import Path

import yaml

from .models import SUPPORTED_PLATFORMS


def validate_project(target: Path) -> list[str]:
    errors: list[str] = []

    bake = target / "agentic.bake.yaml"
    if not bake.exists():
        errors.append("Missing agentic.bake.yaml")
        return errors

    data = yaml.safe_load(bake.read_text(encoding="utf-8")) or {}
    targets = data.get("targets", {})
    if not targets:
        errors.append("No targets configured in agentic.bake.yaml")

    for name, cfg in targets.items():
        for platform in cfg.get("platforms", []) or []:
            if platform not in SUPPORTED_PLATFORMS:
                errors.append(f"Target {name}: unsupported platform {platform}")

    if (target / "skills").exists():
        for skill_dir in sorted((target / "skills").iterdir()):
            if not skill_dir.is_dir():
                continue
            skill_file = skill_dir / "SKILL.md"
            if not skill_file.exists():
                errors.append(f"GitLab skill missing SKILL.md: {skill_dir}")
                continue
            text = skill_file.read_text(encoding="utf-8")
            if "name:" not in text or "description:" not in text:
                errors.append(f"Skill missing name/description: {skill_file}")
            if "slash-command: enabled" not in text:
                errors.append(f"GitLab skill should enable slash command: {skill_file}")

    gitlab_duo = target / ".gitlab" / "duo"
    if gitlab_duo.exists():
        if not (gitlab_duo / "chat-rules.md").exists():
            errors.append("GitLab Duo output missing .gitlab/duo/chat-rules.md")

        flow_dir = gitlab_duo / "flows"
        if flow_dir.exists():
            for flow_file in sorted(flow_dir.glob("*.yaml")):
                text = flow_file.read_text(encoding="utf-8")
                forbidden_top_level = ["name:", "description:", "product_group:"]
                for forbidden in forbidden_top_level:
                    if text.startswith(forbidden):
                        errors.append(f"GitLab custom flow contains forbidden top-level field {forbidden}: {flow_file}")
                if "workspace_agent_skills" not in text:
                    errors.append(f"Flow does not pass workspace_agent_skills: {flow_file}")
                if "user_rule" not in text:
                    errors.append(f"Flow does not pass user_rule: {flow_file}")

    return errors
