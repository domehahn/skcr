from __future__ import annotations

from pathlib import Path
from typing import Any

from pydantic import BaseModel, Field


SUPPORTED_PLATFORMS = {
    "codex",
    "gitlab-duo",
    "github-copilot",
    "claude",
    "openhands",
    "opencode",
    "ollama",
    "generic",
}

PLATFORM_ALIASES = {
    "gitlab": "gitlab-duo",
    "duo": "gitlab-duo",
    "gitlab-duo-agent-platform": "gitlab-duo",
    "copilot": "github-copilot",
    "github": "github-copilot",
    "github-copilot-chat": "github-copilot",
    "claude-code": "claude",
    "open-code": "opencode",
    "open-hands": "openhands",
}


def normalize_platform(value: str) -> str:
    normalized = value.strip().lower()
    normalized = PLATFORM_ALIASES.get(normalized, normalized)
    if normalized not in SUPPORTED_PLATFORMS:
        raise ValueError(f"Unsupported platform: {value}")
    return normalized


def parse_platforms(value: str | None) -> list[str]:
    if not value:
        return []

    platforms: list[str] = []
    for item in value.split(","):
        item = item.strip()
        if not item:
            continue
        platform = normalize_platform(item)
        if platform not in platforms:
            platforms.append(platform)

    return platforms


class TargetConfig(BaseModel):
    description: str = ""
    inherits: list[str] = Field(default_factory=list)
    platforms: list[str] = Field(default_factory=list)
    profiles: list[str] = Field(default_factory=list)
    skills: list[str] = Field(default_factory=list)
    flows: list[str] = Field(default_factory=list)
    rules: dict[str, Any] = Field(default_factory=dict)
    model: dict[str, Any] = Field(default_factory=dict)


class BakeConfig(BaseModel):
    version: str = "1"
    variables: dict[str, Any] = Field(default_factory=dict)
    targets: dict[str, TargetConfig] = Field(default_factory=dict)


class RenderedFile(BaseModel):
    source: str
    destination: Path
    content: str
    platform: str
