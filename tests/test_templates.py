from __future__ import annotations

import zipfile
from pathlib import Path

from agentic_template_kit.renderer import TEMPLATE_ROOT


EXPECTED_TEMPLATES = [
    # Shared
    "shared/SKILL.md.j2",

    # Codex
    "codex/AGENTS.md.j2",

    # GitLab Duo
    "gitlab-duo/AGENTS.md.j2",
    "gitlab-duo/chat-rules.md.j2",
    "gitlab-duo/flows/README.md.j2",
    "gitlab-duo/flows/flow.yaml.j2",

    # Claude Code
    "claude/CLAUDE.md.j2",

    # GitHub Copilot
    "github-copilot/copilot-instructions.md.j2",
    "github-copilot/prompt.prompt.md.j2",

    # OpenHands
    "openhands/AGENTS.md.j2",
    "openhands/instructions.md.j2",

    # OpenCode
    "opencode/AGENTS.md.j2",
    "opencode/instructions.md.j2",

    # Ollama
    "ollama/Modelfile.j2",
    "ollama/README.md.j2",
]


def test_all_referenced_templates_exist() -> None:
    missing = [
        template
        for template in EXPECTED_TEMPLATES
        if not (TEMPLATE_ROOT / template).exists()
    ]

    assert missing == []


def test_template_root_is_inside_package() -> None:
    assert TEMPLATE_ROOT.exists()
    assert TEMPLATE_ROOT.is_dir()
    assert TEMPLATE_ROOT.name == "templates"


def test_gitlab_templates_are_available() -> None:
    assert (TEMPLATE_ROOT / "gitlab-duo" / "AGENTS.md.j2").exists()
    assert (TEMPLATE_ROOT / "gitlab-duo" / "chat-rules.md.j2").exists()
    assert (TEMPLATE_ROOT / "gitlab-duo" / "flows" / "README.md.j2").exists()
    assert (TEMPLATE_ROOT / "gitlab-duo" / "flows" / "flow.yaml.j2").exists()


def test_shared_skill_template_is_available() -> None:
    assert (TEMPLATE_ROOT / "shared" / "SKILL.md.j2").exists()


def test_built_wheel_contains_templates_when_dist_exists() -> None:
    """
    This test is intentionally passive.

    In normal `pytest` runs, `dist/*.whl` may not exist yet.
    In CI, run `uv build` before pytest to turn this into a packaging check.
    """
    dist = Path("dist")
    wheels = sorted(dist.glob("*.whl"))

    if not wheels:
        return

    with zipfile.ZipFile(wheels[-1]) as wheel:
        names = set(wheel.namelist())

    expected = {
        "agentic_template_kit/templates/shared/SKILL.md.j2",
        "agentic_template_kit/templates/codex/AGENTS.md.j2",
        "agentic_template_kit/templates/gitlab-duo/AGENTS.md.j2",
        "agentic_template_kit/templates/gitlab-duo/chat-rules.md.j2",
        "agentic_template_kit/templates/gitlab-duo/flows/README.md.j2",
        "agentic_template_kit/templates/gitlab-duo/flows/flow.yaml.j2",
        "agentic_template_kit/templates/claude/CLAUDE.md.j2",
        "agentic_template_kit/templates/github-copilot/copilot-instructions.md.j2",
        "agentic_template_kit/templates/github-copilot/prompt.prompt.md.j2",
        "agentic_template_kit/templates/openhands/AGENTS.md.j2",
        "agentic_template_kit/templates/openhands/instructions.md.j2",
        "agentic_template_kit/templates/opencode/AGENTS.md.j2",
        "agentic_template_kit/templates/opencode/instructions.md.j2",
        "agentic_template_kit/templates/ollama/Modelfile.j2",
        "agentic_template_kit/templates/ollama/README.md.j2",
    }

    missing = sorted(expected - names)
    assert missing == []