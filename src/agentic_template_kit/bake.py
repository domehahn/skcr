from __future__ import annotations

from pathlib import Path

import yaml

from agentic_template_kit.models import BakeFile

DEFAULT_BAKE_FILE = "agentic.bake.yaml"


def load_bake_file(path: Path) -> BakeFile:
    if not path.exists():
        raise FileNotFoundError(f"Bake file not found: {path}")
    data = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
    return BakeFile.model_validate(data)


def write_default_bake_file(target: Path, overwrite: bool = False) -> Path:
    bake_file = target / DEFAULT_BAKE_FILE
    if bake_file.exists() and not overwrite:
        return bake_file

    bake_file.write_text(
        """version: "1"

variables:
  project_name: "{{ project_name }}"
  owner_team: platform-engineering
  default_language: de
  governance_level: standard

targets:
  codex:
    description: Codex AGENTS.md and project skills
    platforms:
      - codex
    profiles:
      - base
      - devsecops
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer

  copilot:
    description: GitHub Copilot repository instructions and prompt files
    platforms:
      - github-copilot
    profiles:
      - base
      - devsecops
      - documentation

  claude:
    description: Claude Code CLAUDE.md and project skills
    platforms:
      - claude
    profiles:
      - base
      - devsecops
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer

  gitlab:
    description: GitLab Duo Agent Platform setup with AGENTS.md, project-level skills, custom rules, and flow templates
    platforms:
      - gitlab-duo
    profiles:
      - base
      - gitlab-governance
      - devsecops
      - documentation
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer
      - universal-skill-creator
    rules:
      no_direct_push: true
      require_merge_request: true
      require_tests: true
      require_security_review: true
      forbid_secret_files: true
      forbid_env_file_access: true
      require_diff_summary: true
      require_validation_summary: true
      allow_autonomous_changes: false
    flows:
      - secure-code-change
      - documentation-review
      - ci-cd-review
      - dependency-review
      - security-policy-review

  local-ai:
    description: Local Ollama/OpenCode/OpenHands setup
    platforms:
      - opencode
      - openhands
      - ollama
    profiles:
      - base
      - local-models
    model:
      provider: ollama
      default_model: qwen2.5-coder:7b
      base_url: http://localhost:11434

  default:
    description: Standard daily-development setup
    inherits:
      - codex
      - copilot

  all:
    description: Generate all supported platform artifacts
    inherits:
      - codex
      - copilot
      - claude
      - gitlab
      - local-ai
""",
        encoding="utf-8",
    )
    return bake_file
