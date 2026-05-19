from pathlib import Path

from agentic_template_kit.validator import validate_target


def test_validate_empty_target(tmp_path: Path):
    assert validate_target(tmp_path) == []


def test_validate_gitlab_old_rules_path_is_rejected(tmp_path: Path):
    old_rules = tmp_path / ".gitlab" / "duo" / "custom-rules.md"
    old_rules.parent.mkdir(parents=True)
    old_rules.write_text("old", encoding="utf-8")

    issues = validate_target(tmp_path)

    assert any("chat-rules.md" in issue for issue in issues)


def test_validate_gitlab_flow_requires_context_inputs(tmp_path: Path):
    flow = tmp_path / ".gitlab" / "duo" / "flows" / "bad.yaml"
    flow.parent.mkdir(parents=True)
    flow.write_text(
        """version: "1"
environment: ambient
components:
  - name: agent
    type: AgentComponent
    prompt_id: p
prompts:
  - id: p
    messages: []
flow:
  - component: agent
""",
        encoding="utf-8",
    )

    issues = validate_target(tmp_path)

    assert any("user_rule" in issue for issue in issues)
    assert any("workspace_agent_skills" in issue for issue in issues)
