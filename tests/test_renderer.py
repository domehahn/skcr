from pathlib import Path

from agentic_template_kit.models import TargetConfig
from agentic_template_kit.renderer import plan_changes, render_files


def test_render_codex_files(tmp_path: Path):
    target = TargetConfig(
        platforms=["codex"],
        skills=["safe-implementer"],
        variables={"project_name": "demo"},
    )

    files = render_files(tmp_path, target, "default")
    paths = {f.destination.as_posix() for f in files}

    assert "AGENTS.md" in paths
    assert ".agents/skills/safe-implementer/SKILL.md" in paths


def test_render_gitlab_duo_uses_gitlab_native_paths(tmp_path: Path):
    target = TargetConfig(
        platforms=["gitlab-duo"],
        skills=["security-reviewer"],
        flows=["secure-code-change"],
        variables={"project_name": "demo"},
    )

    files = render_files(tmp_path, target, "gitlab")
    paths = {f.destination.as_posix() for f in files}

    assert "AGENTS.md" in paths
    assert "skills/security-reviewer/SKILL.md" in paths
    assert ".gitlab/duo/chat-rules.md" in paths
    assert ".gitlab/duo/flows/secure-code-change.yaml" in paths
    assert ".gitlab/duo/flows/README.md" in paths
    assert ".agents/skills/security-reviewer/SKILL.md" not in paths
    assert ".gitlab/duo/custom-rules.md" not in paths


def test_render_claude_files(tmp_path: Path):
    target = TargetConfig(
        platforms=["claude"],
        skills=["safe-implementer"],
        variables={"project_name": "demo"},
    )

    files = render_files(tmp_path, target, "claude")
    paths = {f.destination.as_posix() for f in files}

    assert "CLAUDE.md" in paths
    assert ".claude/skills/safe-implementer/SKILL.md" in paths


def test_render_ollama_files(tmp_path: Path):
    target = TargetConfig(
        platforms=["ollama"],
        variables={"project_name": "demo"},
    )

    files = render_files(tmp_path, target, "local-ai")
    paths = {f.destination.as_posix() for f in files}

    assert ".ollama/Modelfile" in paths
    assert ".ollama/README.md" in paths


def test_plan_create(tmp_path: Path):
    target = TargetConfig(
        platforms=["github-copilot"],
        variables={"project_name": "demo"},
    )
    files = render_files(tmp_path, target, "default")
    changes = plan_changes(tmp_path, files, managed_checksums={})

    assert all(change.action == "create" for change in changes)
