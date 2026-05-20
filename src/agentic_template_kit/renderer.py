from __future__ import annotations

from pathlib import Path

from jinja2 import Environment, FileSystemLoader

from .catalog import skill_description, skill_title
from .models import BakeConfig, RenderedFile, TargetConfig

TEMPLATE_ROOT = Path(__file__).parent / "templates"


def _render(env: Environment, template: str, context: dict) -> str:
    return env.get_template(template).render(**context)


def _skill_meta(skill_names: list[str]) -> list[dict[str, str]]:
    return [
        {
            "name": skill,
            "title": skill_title(skill),
            "description": skill_description(skill),
        }
        for skill in skill_names
    ]


def render_files(config: BakeConfig, target: TargetConfig) -> list[RenderedFile]:
    env = Environment(
        loader=FileSystemLoader(TEMPLATE_ROOT),
        autoescape=False,
        trim_blocks=True,
        lstrip_blocks=True,
    )

    skills = _skill_meta(target.skills)
    context = {
        "variables": config.variables,
        "target": target,
        "skills": skills,
        "flows": target.flows,
        "rules": target.rules,
        "model": target.model,
    }

    files: list[RenderedFile] = []

    def add(platform: str, template: str, destination: str, extra: dict | None = None) -> None:
        ctx = context if extra is None else {**context, **extra}
        files.append(
            RenderedFile(
                source=template,
                destination=Path(destination),
                content=_render(env, template, ctx),
                platform=platform,
            )
        )

    if "codex" in target.platforms:
        add("codex", "codex/AGENTS.md.j2", "AGENTS.md")
        for skill in skills:
            add(
                "codex",
                "shared/SKILL.md.j2",
                f".agents/skills/{skill['name']}/SKILL.md",
                {"skill": skill, "invocation_prefix": "$", "slash_command": False},
            )

    if "gitlab-duo" in target.platforms:
        add("gitlab-duo", "gitlab-duo/AGENTS.md.j2", "AGENTS.md")
        add("gitlab-duo", "gitlab-duo/chat-rules.md.j2", ".gitlab/duo/chat-rules.md")
        for skill in skills:
            add(
                "gitlab-duo",
                "shared/SKILL.md.j2",
                f"skills/{skill['name']}/SKILL.md",
                {"skill": skill, "invocation_prefix": "/", "slash_command": True},
            )
        add("gitlab-duo", "gitlab-duo/flows/README.md.j2", ".gitlab/duo/flows/README.md")
        for flow in target.flows:
            add(
                "gitlab-duo",
                "gitlab-duo/flows/flow.yaml.j2",
                f".gitlab/duo/flows/{flow}.yaml",
                {"flow_name": flow, "flow_title": skill_title(flow)},
            )

    if "claude" in target.platforms:
        add("claude", "claude/CLAUDE.md.j2", "CLAUDE.md")
        for skill in skills:
            add(
                "claude",
                "shared/SKILL.md.j2",
                f".claude/skills/{skill['name']}/SKILL.md",
                {"skill": skill, "invocation_prefix": "/", "slash_command": False},
            )

    if "github-copilot" in target.platforms:
        add("github-copilot", "github-copilot/copilot-instructions.md.j2", ".github/copilot-instructions.md")
        for skill in skills:
            add(
                "github-copilot",
                "github-copilot/prompt.prompt.md.j2",
                f".github/prompts/{skill['name']}.prompt.md",
                {"skill": skill},
            )

    if "openhands" in target.platforms:
        add("openhands", "openhands/AGENTS.md.j2", "AGENTS.md")
        add("openhands", "openhands/instructions.md.j2", ".openhands/instructions.md")

    if "opencode" in target.platforms:
        add("opencode", "opencode/AGENTS.md.j2", "AGENTS.md")
        add("opencode", "opencode/instructions.md.j2", ".opencode/instructions.md")

    if "ollama" in target.platforms:
        add("ollama", "ollama/Modelfile.j2", ".ollama/Modelfile")
        add("ollama", "ollama/README.md.j2", ".ollama/README.md")

    if "generic" in target.platforms:
        for skill in skills:
            add(
                "generic",
                "shared/SKILL.md.j2",
                f".agentic/skills/{skill['name']}/SKILL.md",
                {"skill": skill, "invocation_prefix": "$", "slash_command": False},
            )

    return files
