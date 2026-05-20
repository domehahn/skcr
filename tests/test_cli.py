from pathlib import Path

from typer.testing import CliRunner

from agentic_template_kit.cli import app

runner = CliRunner()


def test_init_with_comma_platforms(tmp_path: Path):
    result = runner.invoke(
        app,
        [
            "init",
            "--target",
            str(tmp_path),
            "--platform",
            "gitlab-duo,codex,github-copilot",
            "--project-name",
            "Demo",
        ],
    )
    assert result.exit_code == 0
    text = (tmp_path / "agentic.bake.yaml").read_text(encoding="utf-8")
    assert "gitlab-duo" in text
    assert "codex" in text
    assert "github-copilot" in text
    assert "{{ project_name }}" not in text


def test_gitlab_bake_writes_expected_paths(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "gitlab-duo", "--project-name", "Demo"],
    )

    result = runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])

    assert result.exit_code == 0
    assert (tmp_path / "AGENTS.md").exists()
    assert (tmp_path / "skills" / "security-reviewer" / "SKILL.md").exists()
    assert (tmp_path / "skills" / "threat-modeler" / "SKILL.md").exists()
    assert (tmp_path / ".gitlab" / "duo" / "chat-rules.md").exists()
    assert (tmp_path / ".gitlab" / "duo" / "flows" / "secure-code-change.yaml").exists()


def test_gitlab_skill_enables_slash_command(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "gitlab-duo", "--project-name", "Demo"],
    )
    runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])

    text = (tmp_path / "skills" / "security-reviewer" / "SKILL.md").read_text(encoding="utf-8")
    assert "slash-command: enabled" in text


def test_gitlab_flow_passes_required_context(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "gitlab-duo", "--project-name", "Demo"],
    )
    runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])

    text = (tmp_path / ".gitlab" / "duo" / "flows" / "secure-code-change.yaml").read_text(
        encoding="utf-8"
    )
    assert "context:inputs.user_rule" in text
    assert "context:inputs.workspace_agent_skills" in text


def test_gitlab_validation_passes(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "gitlab-duo", "--project-name", "Demo"],
    )
    runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])
    result = runner.invoke(app, ["validate", "--target", str(tmp_path)])
    assert result.exit_code == 0


def test_all_preset_contains_local_ai(tmp_path: Path):
    result = runner.invoke(app, ["init", "--target", str(tmp_path), "--preset", "all", "--project-name", "Demo"])

    assert result.exit_code == 0
    text = (tmp_path / "agentic.bake.yaml").read_text(encoding="utf-8")
    assert "ollama" in text
    assert "openhands" in text
    assert "opencode" in text


def test_codex_bake_writes_codex_skills(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "codex", "--project-name", "Demo"],
    )
    result = runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])
    assert result.exit_code == 0
    assert (tmp_path / ".agents" / "skills" / "safe-implementer" / "SKILL.md").exists()


def test_copilot_bake_writes_prompt_files(tmp_path: Path):
    runner.invoke(
        app,
        ["init", "--target", str(tmp_path), "--platform", "github-copilot", "--project-name", "Demo"],
    )
    result = runner.invoke(app, ["bake", "default", "--target", str(tmp_path), "--write"])
    assert result.exit_code == 0
    assert (tmp_path / ".github" / "copilot-instructions.md").exists()
    assert (tmp_path / ".github" / "prompts" / "security-reviewer.prompt.md").exists()
