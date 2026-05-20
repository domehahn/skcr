from __future__ import annotations

from copy import deepcopy
from pathlib import Path
from typing import Any

import yaml

from .catalog import BASE_RULES, CORE_SKILLS, DEVSECOPS_FLOWS
from .models import BakeConfig, TargetConfig


def _merge_unique(base: list[str], add: list[str]) -> list[str]:
    result = list(base)
    for item in add:
        if item not in result:
            result.append(item)
    return result


def _deep_merge(base: dict[str, Any], add: dict[str, Any]) -> dict[str, Any]:
    result = deepcopy(base)
    for key, value in add.items():
        if isinstance(value, dict) and isinstance(result.get(key), dict):
            result[key] = _deep_merge(result[key], value)
        else:
            result[key] = deepcopy(value)
    return result


def load_bake_file(path: Path) -> BakeConfig:
    raw = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
    return BakeConfig.model_validate(raw)


def dump_bake_file(config: BakeConfig, path: Path) -> None:
    data = config.model_dump(mode="json", exclude_none=True)
    path.write_text(yaml.safe_dump(data, sort_keys=False, allow_unicode=True), encoding="utf-8")


def resolve_target(config: BakeConfig, target_name: str) -> TargetConfig:
    if target_name not in config.targets:
        raise KeyError(f"Unknown target '{target_name}'. Available: {', '.join(config.targets)}")

    active_stack: list[str] = []

    def resolve(name: str) -> TargetConfig:
        if name in active_stack:
            chain = " -> ".join(active_stack + [name])
            raise ValueError(f"Circular target inheritance detected: {chain}")

        active_stack.append(name)
        target = config.targets[name]
        merged = TargetConfig(description=target.description)

        for parent in target.inherits:
            parent_target = resolve(parent)
            merged.platforms = _merge_unique(merged.platforms, parent_target.platforms)
            merged.profiles = _merge_unique(merged.profiles, parent_target.profiles)
            merged.skills = _merge_unique(merged.skills, parent_target.skills)
            merged.flows = _merge_unique(merged.flows, parent_target.flows)
            merged.rules = _deep_merge(merged.rules, parent_target.rules)
            merged.model = _deep_merge(merged.model, parent_target.model)

        merged.platforms = _merge_unique(merged.platforms, target.platforms)
        merged.profiles = _merge_unique(merged.profiles, target.profiles)
        merged.skills = _merge_unique(merged.skills, target.skills)
        merged.flows = _merge_unique(merged.flows, target.flows)
        merged.rules = _deep_merge(merged.rules, target.rules)
        merged.model = _deep_merge(merged.model, target.model)

        active_stack.pop()
        return merged

    return resolve(target_name)


def build_initial_config(
    platforms: list[str],
    project_name: str,
    owner_team: str,
    language: str,
    governance_level: str,
    preset: str | None = None,
) -> BakeConfig:
    if preset:
        if preset == "minimal":
            platforms = ["codex"]
        elif preset == "gitlab":
            platforms = ["gitlab-duo"]
        elif preset == "enterprise":
            platforms = ["gitlab-duo", "codex", "github-copilot"]
        elif preset == "local-ai":
            platforms = ["opencode", "openhands", "ollama"]
        elif preset == "all":
            platforms = [
                "codex",
                "github-copilot",
                "claude",
                "gitlab-duo",
                "opencode",
                "openhands",
                "ollama",
            ]
        else:
            raise ValueError(f"Unsupported preset: {preset}")

    if not platforms:
        platforms = ["codex", "github-copilot"]

    variables = {
        "project_name": project_name,
        "owner_team": owner_team,
        "default_language": language,
        "governance_level": governance_level,
    }

    targets: dict[str, TargetConfig] = {}

    if "codex" in platforms:
        targets["codex"] = TargetConfig(
            description="Codex AGENTS.md and project skills",
            platforms=["codex"],
            profiles=["base", "devsecops"],
            skills=CORE_SKILLS,
            rules=BASE_RULES,
        )

    if "github-copilot" in platforms:
        targets["copilot"] = TargetConfig(
            description="GitHub Copilot repository instructions and prompt files",
            platforms=["github-copilot"],
            profiles=["base", "devsecops", "documentation"],
            skills=CORE_SKILLS,
            rules=BASE_RULES,
        )

    if "claude" in platforms:
        targets["claude"] = TargetConfig(
            description="Claude Code CLAUDE.md and project skills",
            platforms=["claude"],
            profiles=["base", "devsecops"],
            skills=CORE_SKILLS,
            rules=BASE_RULES,
        )

    if "gitlab-duo" in platforms:
        targets["gitlab"] = TargetConfig(
            description="GitLab Duo Agent Platform setup with AGENTS.md, project-level skills, custom rules, and flow templates",
            platforms=["gitlab-duo"],
            profiles=["base", "gitlab-governance", "devsecops", "documentation"],
            skills=CORE_SKILLS,
            rules=BASE_RULES,
            flows=DEVSECOPS_FLOWS,
        )

    local_platforms = [p for p in ["opencode", "openhands", "ollama"] if p in platforms]
    if local_platforms:
        targets["local-ai"] = TargetConfig(
            description="Local Ollama/OpenCode/OpenHands setup",
            platforms=local_platforms,
            profiles=["base", "local-models"],
            skills=CORE_SKILLS,
            rules=BASE_RULES,
            model={
                "provider": "ollama",
                "default_model": "qwen2.5-coder:7b",
                "base_url": "http://localhost:11434",
            },
        )

    target_names = list(targets.keys())
    targets["default"] = TargetConfig(
        description="Default target for this repository",
        inherits=target_names,
    )
    targets["all"] = TargetConfig(
        description="Generate all configured platform artifacts",
        inherits=target_names,
    )

    return BakeConfig(version="1", variables=variables, targets=targets)
