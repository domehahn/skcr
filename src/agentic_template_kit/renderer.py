from __future__ import annotations

from importlib.resources import files
from pathlib import Path
from typing import Any

from jinja2 import Environment, StrictUndefined

from agentic_template_kit.detector import detect_project
from agentic_template_kit.lockfile import sha256_text
from agentic_template_kit.models import PlannedChange, RenderedFile, TargetConfig

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

SKILL_TEMPLATES = {
    "cost-based-planner": "skills/cost-based-planner/SKILL.md.j2",
    "safe-implementer": "skills/safe-implementer/SKILL.md.j2",
    "verification-reviewer": "skills/verification-reviewer/SKILL.md.j2",
    "security-reviewer": "skills/security-reviewer/SKILL.md.j2",
    "documentation-maintainer": "skills/documentation-maintainer/SKILL.md.j2",
    "universal-skill-creator": "skills/universal-skill-creator/SKILL.md.j2",
}

FLOW_TEMPLATES = {
    "secure-code-change": "gitlab-duo/flows/secure-code-change.yaml.j2",
    "documentation-review": "gitlab-duo/flows/documentation-review.yaml.j2",
    "ci-cd-review": "gitlab-duo/flows/ci-cd-review.yaml.j2",
    "dependency-review": "gitlab-duo/flows/dependency-review.yaml.j2",
    "security-policy-review": "gitlab-duo/flows/security-policy-review.yaml.j2",
}


def template_root() -> Path:
    return Path(str(files("agentic_template_kit").joinpath("templates")))


def environment() -> Environment:
    return Environment(
        loader=None,
        undefined=StrictUndefined,
        trim_blocks=True,
        lstrip_blocks=True,
        autoescape=False,
    )


def render_template(relative_template: str, context: dict[str, Any]) -> str:
    path = template_root() / relative_template
    if not path.exists():
        raise FileNotFoundError(f"Template not found: {relative_template}")
    return Environment(
        undefined=StrictUndefined,
        trim_blocks=True,
        lstrip_blocks=True,
        autoescape=False,
    ).from_string(path.read_text(encoding="utf-8")).render(**context)


def build_context(target_dir: Path, target: TargetConfig, target_name: str) -> dict[str, Any]:
    detected = detect_project(target_dir)
    variables = {
        **detected,
        **target.variables,
        "target_name": target_name,
        "platforms": target.platforms,
        "profiles": target.profiles,
        "skills": target.skills,
        "flows": target.flows,
        "rules": target.rules,
        "model": target.model.model_dump() if target.model else {},
    }

    # Resolve placeholder created by default init.
    if variables.get("project_name") == "{{ project_name }}":
        variables["project_name"] = detected["repo_name"]

    variables.setdefault("project_name", detected["repo_name"])
    variables.setdefault("owner_team", "unknown")
    variables.setdefault("default_language", "en")
    variables.setdefault("governance_level", "standard")
    return variables


def render_files(target_dir: Path, target: TargetConfig, target_name: str) -> list[RenderedFile]:
    unknown = sorted(set(target.platforms) - SUPPORTED_PLATFORMS)
    if unknown:
        raise ValueError(f"Unsupported platform(s): {', '.join(unknown)}")

    context = build_context(target_dir, target, target_name)
    rendered: list[RenderedFile] = []

    def add(platform: str, template: str, destination: str) -> None:
        rendered.append(
            RenderedFile(
                source=template,
                destination=Path(destination),
                content=render_template(template, context),
                platform=platform,
            )
        )

    if "codex" in target.platforms:
        add("codex", "codex/AGENTS.md.j2", "AGENTS.md")
        for skill in target.skills:
            skill_template = SKILL_TEMPLATES.get(skill)
            if skill_template:
                add("codex", skill_template, f".agents/skills/{skill}/SKILL.md")

    if "gitlab-duo" in target.platforms:
        # GitLab Duo project-level customization paths are intentionally GitLab-specific:
        # - AGENTS.md at repository root
        # - skills/<skill-name>/SKILL.md for project-level Agent Skills
        # - .gitlab/duo/chat-rules.md for project-level Custom Rules
        # - .gitlab/duo/flows/*.yaml as source-of-truth templates for Custom Flows
        add("gitlab-duo", "gitlab-duo/AGENTS.md.j2", "AGENTS.md")
        add("gitlab-duo", "gitlab-duo/chat-rules.md.j2", ".gitlab/duo/chat-rules.md")
        for skill in target.skills:
            skill_template = SKILL_TEMPLATES.get(skill)
            if skill_template:
                add("gitlab-duo", skill_template, f"skills/{skill}/SKILL.md")
        if target.flows:
            add("gitlab-duo", "gitlab-duo/flows/README.md.j2", ".gitlab/duo/flows/README.md")
        for flow in target.flows:
            flow_template = FLOW_TEMPLATES.get(flow)
            if flow_template:
                add("gitlab-duo", flow_template, f".gitlab/duo/flows/{flow}.yaml")

    if "github-copilot" in target.platforms:
        add(
            "github-copilot",
            "github-copilot/copilot-instructions.md.j2",
            ".github/copilot-instructions.md",
        )
        add(
            "github-copilot",
            "github-copilot/default.prompt.md.j2",
            ".github/prompts/agentic-default.prompt.md",
        )


    if "claude" in target.platforms:
        add("claude", "claude/CLAUDE.md.j2", "CLAUDE.md")
        for skill in target.skills:
            skill_template = SKILL_TEMPLATES.get(skill)
            if skill_template:
                add("claude", skill_template, f".claude/skills/{skill}/SKILL.md")

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
        add("generic", "generic/SKILL.md.j2", ".agentic/generic/SKILL.md")

    return merge_same_destination(rendered)


def merge_same_destination(files: list[RenderedFile]) -> list[RenderedFile]:
    """Merge multiple platform contributions to common files like AGENTS.md."""
    by_dest: dict[Path, list[RenderedFile]] = {}
    for file in files:
        by_dest.setdefault(file.destination, []).append(file)

    merged: list[RenderedFile] = []
    for dest, group in by_dest.items():
        if len(group) == 1:
            merged.append(group[0])
            continue

        content = "\n\n".join(
            [
                "<!-- BEGIN agentic-template-kit platform:"
                + item.platform
                + " source:"
                + item.source
                + " -->\n"
                + item.content.strip()
                + "\n<!-- END agentic-template-kit platform:"
                + item.platform
                + " -->"
                for item in group
            ]
        )
        merged.append(
            RenderedFile(
                source="+".join(item.source for item in group),
                destination=dest,
                content=content + "\n",
                platform="+".join(item.platform for item in group),
            )
        )

    return sorted(merged, key=lambda item: item.destination.as_posix())


def plan_changes(
    target_dir: Path,
    rendered_files: list[RenderedFile],
    managed_checksums: dict[str, str],
    force: bool = False,
) -> list[PlannedChange]:
    changes: list[PlannedChange] = []

    for rendered in rendered_files:
        destination = target_dir / rendered.destination
        rel = rendered.destination.as_posix()
        new_checksum = sha256_text(rendered.content)

        if not destination.exists():
            changes.append(
                PlannedChange(
                    action="create",
                    destination=rendered.destination,
                    source=rendered.source,
                    new_checksum=new_checksum,
                )
            )
            continue

        current_content = destination.read_text(encoding="utf-8")
        current_checksum = sha256_text(current_content)

        if current_checksum == new_checksum:
            changes.append(
                PlannedChange(
                    action="unchanged",
                    destination=rendered.destination,
                    source=rendered.source,
                    old_checksum=current_checksum,
                    new_checksum=new_checksum,
                )
            )
            continue

        is_managed = managed_checksums.get(rel) == current_checksum
        if is_managed or force:
            changes.append(
                PlannedChange(
                    action="update",
                    destination=rendered.destination,
                    source=rendered.source,
                    old_checksum=current_checksum,
                    new_checksum=new_checksum,
                    reason="managed file" if is_managed else "force enabled",
                )
            )
        else:
            changes.append(
                PlannedChange(
                    action="conflict",
                    destination=rendered.destination,
                    source=rendered.source,
                    old_checksum=current_checksum,
                    new_checksum=new_checksum,
                    reason="existing unmanaged or locally modified file; use --force to overwrite",
                )
            )

    return changes


def apply_files(target_dir: Path, rendered_files: list[RenderedFile], changes: list[PlannedChange]) -> None:
    change_by_path = {change.destination.as_posix(): change for change in changes}
    for rendered in rendered_files:
        change = change_by_path[rendered.destination.as_posix()]
        if change.action not in {"create", "update"}:
            continue
        destination = target_dir / rendered.destination
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(rendered.content, encoding="utf-8")
