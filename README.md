# Skill Creator (`skcr`)

`skcr` is a Go CLI for creating versioned AI agent skill structures and rendering agentic project and platform files across multiple agent platforms.

## Purpose

This repository contains versioned Agent Skills and the `skcr` CLI used to scaffold, render, sync, validate, and package those skills across Codex, GitLab Duo, Claude Code, GitHub Copilot, OpenHands, OpenCode, Ollama, Cursor, Roo Code, Kiro, Junie, Gemini CLI, Windsurf, and other agent platforms.

```text
skcr  = init / add / remove / rename / list / bake / sync / status / doctor / export / validate / clean
skpm  = validate / version / package / publish / install / update / lock / verify
```

`skcr` creates and renders. `skpm` manages the skill lifecycle.

## Architecture

```text
skcr init
  → creates agentic.bake.yaml

skcr add skill <name>
  → adds skill to all bakefile targets and scaffolds immediately

skcr remove skill <name>
  → removes skill from bakefile targets; --delete-dirs also removes directories

skcr rename skill <old> <new>
  → renames skill in bakefile targets and moves directories across all platform dirs

skcr bake [target] --write
  → scaffolds skill directories in all platform dirs + renders platform files

skcr list skills [--with-targets] [--in-target <name>]
  → lists all skills defined across bakefile targets (one per line, pipeable)

skcr sync
  → propagates SKILL.md edits from .agents/skills/ to all platform directories

skcr status
  → shows which skills are scaffolded and in sync across platform directories

skcr doctor
  → checks bakefile, skill files, platform sync, and toolchain without modifying anything

skcr export [--out SKILLS.md] [--in-target <name>] [--skill <name>]
  → concatenates all SKILL.md files into one document for LLM context or documentation
```

### Conceptual model

Skills are declared once in `targets.*.skills`. `skcr bake --write` scaffolds the full directory structure in every platform-specific skill directory:

```text
.agents/skills/<name>/     ← canonical source + universal fallback
  ├── SKILL.md             ← edit this to update the skill content
  ├── skill.yaml
  ├── VERSION
  ├── CHANGELOG.md
  ├── README.md
  ├── LICENSE
  └── tests/
      └── README.md

.claude/skills/<name>/     ← Claude Code  (full scaffold)
.github/skills/<name>/     ← GitHub Copilot  (full scaffold)
.cursor/skills/<name>/     ← Cursor  (full scaffold)
.roo/skills/<name>/        ← Roo Code  (full scaffold)
.kiro/skills/<name>/       ← Kiro  (full scaffold)
...
```

All platform directories receive the complete scaffold. `.agents/skills/` is the canonical source — after editing it, run `skcr sync` to propagate `SKILL.md` changes to the other directories.

Platform-specific instruction files (AGENTS.md, CLAUDE.md, etc.) are tracked in `.agentic-template.lock` and managed by `skcr bake`.

## Repository Structure

- `.agents/skills/` contains canonical skill sources for this repository.
- `skills/` contains the GitLab Duo platform copy generated or synchronized from `.agents/skills/`.
- `.claude/skills/`, `.github/skills/`, `.opencode/skills/`, `.openhands/skills/`, `.ollama/skills/`, `.cursor/skills/`, `.roo/skills/`, `.kiro/skills/`, `.junie/skills/`, `.gemini/skills/`, and `.windsurf/skills/` contain platform-specific synchronized copies.
- `agentic.bake.yaml` defines platform rendering and governance rules.
- `AGENTS.md` defines agent routing and usage rules.
- `docs/skill-content-readiness.md` defines when a skill is framework-ready versus content-ready.

## Skill Authoring Rules

- Every skill must contain complete YAML frontmatter.
- Every skill must contain a `## Changelog` section.
- Every skill must contain skill-specific checklists, decision rules, acceptance criteria, and output requirements.
- Generic copy-paste skill bodies are not production-ready.
- Edit canonical skill content in `.agents/skills/<name>/SKILL.md`, then run `skcr sync`.
- Changes must go through merge requests.
- Security-sensitive changes require review.

## Production Readiness

A skill is production-ready only when it has:

- complete versioning metadata,
- validated platform compatibility from the central compatibility matrix,
- concrete operating model,
- concrete checklist,
- concrete acceptance criteria,
- changelog,
- security guardrails,
- clear output requirements.

## Platform Compatibility

Built-in production skills use concrete `min_platform_version` values from `internal/platforms/compatibility.go`. `unknown` is reserved for incomplete custom metadata and should be treated as a validation warning, not as a production-ready compatibility claim.

## Supported platforms

`skcr` recognises all platforms from the [agentskills.io](https://agentskills.io) open standard.

**Platforms with dedicated skill directories:**

| Platform | Alias(es) | Skill directory |
| --- | --- | --- |
| `gitlab-duo` | `gitlab` | `skills/` |
| `claude-code` | `claude` | `.claude/skills/` |
| `github-copilot` | `copilot`, `github` | `.github/skills/` |
| `cursor` | | `.cursor/skills/` |
| `junie` | `jetbrains` | `.junie/skills/` |
| `gemini-cli` | `gemini` | `.gemini/skills/` |
| `roo-code` | `roo` | `.roo/skills/` |
| `kiro` | | `.kiro/skills/` |
| `opencode` | | `.opencode/skills/` |
| `openhands` | | `.openhands/skills/` |
| `ollama` | | `.ollama/skills/` |
| `windsurf` | | `.windsurf/skills/` |

**All other platforms** (including `codex`, `amp`, `goose`, `trae`, and 30+ more) use `.agents/skills/` as the universal fallback.

Full platform list: `sklib/spec/platform.go`.

## Prerequisites

- Go `>= 1.22`
- `make` (optional, for `make build` and `make install`)

## Installation

```bash
# From source
make install

# Directly with Go
go install ./cmd/skcr

# From GitHub
go install github.com/domehahn/skcr/cmd/skcr@latest
```

## Quick start

```bash
# 1. Initialise
skcr init --target . --project-name MyProject --platform codex,claude-code,github-copilot

# 2. Add skills (writes to bakefile + scaffolds immediately)
skcr add skill requirements-analyst
skcr add skill architecture-reviewer

# 3. Render platform files
skcr bake --write

# 4. Check status
skcr status

# 5. Edit .agents/skills/requirements-analyst/SKILL.md, then propagate
skcr sync

# 6. Validate
skcr validate
```

## Commands

| Command | Description |
| --- | --- |
| `skcr init` | Create `agentic.bake.yaml` |
| `skcr add skill <name>` | Add a skill to all bakefile targets and scaffold its directories |
| `skcr remove skill <name>` | Remove a skill from bakefile targets, optionally deleting directories |
| `skcr rename skill <old> <new>` | Rename a skill across bakefile targets and all platform directories |
| `skcr list skills` | List all skills defined across bakefile targets |
| `skcr list targets` | List available bake targets |
| `skcr bake [target]` | Scaffold skill directories and render platform-specific output |
| `skcr sync` | Propagate `SKILL.md` edits from `.agents/skills/` to all platform dirs |
| `skcr status` | Show skill scaffold status across all platform directories |
| `skcr doctor` | Check project health: bakefile, skills, platform sync, and toolchain |
| `skcr export` | Export all skill content as a single Markdown document |
| `skcr validate` | Validate configuration and generated state |
| `skcr clean` | Remove skcr-managed files listed in `.agentic-template.lock` |
| `skcr list-targets` | List available bake targets |
| `skcr scaffold skill <name>` | Create a standalone skill skeleton from CLI flags |
| `skcr version` | Show version, commit, and build date |

## Bakefile format (`agentic.bake.yaml`)

Skills are declared in `targets.*.skills`. `skcr bake` reads these lists and scaffolds the corresponding directories — no separate `skill_sources.skills` enumeration needed.

```yaml
version: "1"

variables:
  project_name: SkillDemo
  owner_team: platform-engineering
  default_language: de
  governance_level: standard

skill_sources:
  defaults:
    version: 0.1.0
    owner: platform-engineering
    license: MIT
    compatible_with:
      - codex
      - claude-code

skills:
  source: agent-skills.lock
  mode: reference

targets:
  codex:
    description: Codex AGENTS.md and project skills
    platforms:
      - codex
    profiles:
      - base
      - devsecops
    skills:
      - requirements-analyst
      - architecture-reviewer

  claude:
    description: Claude Code CLAUDE.md and project skills
    platforms:
      - claude-code
    profiles:
      - base
    skills:
      - requirements-analyst
      - architecture-reviewer

  all:
    description: All configured platforms
    inherits:
      - codex
      - claude
```

### `skill_sources` block

Controls default metadata for scaffolded skill directories.

- `defaults.version` — initial `VERSION` file content (default: `0.1.0`)
- `defaults.owner` — owner written into `skill.yaml`
- `defaults.license` — license written into `skill.yaml` and `LICENSE` (default: `MIT`)
- `defaults.compatible_with` — default platform list for `skill.yaml`

### `skills` block

Controls integration of installed/locked skills from `skpm` state.

- `source` — path to `agent-skills.lock` (default: `agent-skills.lock`)
- `mode` — `reference` | `copy` | `link` | `embed` (default: `reference`)
- `platforms` — limit skill integration to specific platforms

## `skcr add skill`

Adds a skill to all bakefile targets (or selected targets) and immediately scaffolds the full directory structure.

```bash
skcr add skill <name> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--in-target` | all | Bake target(s) to add the skill to (repeatable) |
| `--no-scaffold` | | Update bakefile only, skip scaffolding |

```bash
# Add to all targets
skcr add skill threat-modeler

# Add only to the codex target
skcr add skill threat-modeler --in-target codex
```

## `skcr remove skill`

Removes a skill from bakefile targets and optionally deletes its directories in all platform dirs.

```bash
skcr remove skill <name> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--in-target` | all | Bake target(s) to remove skill from (repeatable) |
| `--delete-dirs` | | Also delete skill directories from all platform dirs |
| `--dry-run` | | Preview changes without writing |

```bash
# Remove from bakefile only (directories preserved)
skcr remove skill deprecated-skill

# Remove from bakefile and delete all directories
skcr remove skill deprecated-skill --delete-dirs

# Preview first
skcr remove skill deprecated-skill --delete-dirs --dry-run
```

## `skcr rename skill`

Renames a skill in all bakefile targets and moves the corresponding directories in every platform dir. Skips directories that are already absent; aborts on naming conflicts.

```bash
skcr rename skill <old-name> <new-name> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--dry-run` | | Preview changes without writing |

```bash
# Preview
skcr rename skill policy-reviewer security-policy-reviewer --dry-run

# Apply
skcr rename skill policy-reviewer security-policy-reviewer
```

Output:

```text
Renamed "policy-reviewer" → "security-policy-reviewer" in targets: all, codex, gitlab
moved  .agents/skills/policy-reviewer/  →  .agents/skills/security-policy-reviewer/
moved  .claude/skills/policy-reviewer/  →  .claude/skills/security-policy-reviewer/
moved  .github/skills/policy-reviewer/  →  .github/skills/security-policy-reviewer/

Done: 3 director(ies) moved.
```

## `skcr bake`

Scaffolds skill directories in all configured platform locations and renders platform-specific instruction files.

```bash
skcr bake [target] [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--write` | | Write files to disk |
| `--plan` | | Show plan without writing (default when `--write` is omitted) |
| `--detailed-diff` | | Show full diffs in plan |
| `--platform` | | Render only one platform |
| `--skills-from` | | Read locked skills from a lock file |
| `--skills-mode` | | Override skill integration mode |

If no target name is given, `skcr bake` falls back in order: `default` → `all` → sole target → error listing available targets.

`skcr bake` does **not** call `skpm`, write `agent-skills.lock`, or talk to any registry.

## `skcr sync`

Reads `SKILL.md` from the canonical `.agents/skills/<name>/SKILL.md` and writes it to every other platform directory where the skill is already scaffolded. Directories that have not been scaffolded yet are skipped — run `skcr bake --write` first.

```bash
skcr sync [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--dry-run` | | Preview changes without writing |
| `--skill` | | Sync only a specific skill |

```bash
# After editing .agents/skills/requirements-analyst/SKILL.md:
skcr sync
skcr sync --skill requirements-analyst   # single skill only
skcr sync --dry-run                      # preview
```

## `skcr status`

Prints a skill × platform-directory matrix.

```bash
skcr status [--target .]
```

```text
Skill                           agents            claude            github
────────────────────────────────────────────────────────────────────────────
requirements-analyst            ✓                 ✓                 ✓
architecture-reviewer           ✓                 ✓                 ~
threat-modeler                  ✓                 ✗                 ✗

2 ✓ in sync  ·  1 ~ differs (run skcr sync)  ·  2 ✗ missing (run skcr bake --write)
```

- `✓` — directory exists and `SKILL.md` matches canonical source
- `~` — directory exists but `SKILL.md` differs from `.agents/skills/` (run `skcr sync`)
- `✗` — directory not yet scaffolded (run `skcr bake --write`)

## `skcr list skills`

Lists all unique skills defined across bakefile targets. One skill per line — designed for piping.

```bash
skcr list skills [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `--in-target` | | Filter to a single bake target |
| `--with-targets` | | Show which bake targets each skill belongs to |

```bash
# Plain list (pipeable)
skcr list skills
skcr list skills | xargs -I{} skpm validate .agents/skills/{}

# With target annotation
skcr list skills --with-targets

# Only skills from one target
skcr list skills --in-target codex
```

## `skcr doctor`

Checks project health without modifying anything. Useful as a CI pre-flight gate.

```bash
skcr doctor [--target .]
```

Checks performed:

| Check | What it verifies |
| --- | --- |
| `toolchain` | `skpm` is available in `PATH` |
| `bakefile` | `agentic.bake.yaml` parses without errors |
| `targets` | At least one target defined; no duplicate skill names per target |
| `skills` | Each skill has `SKILL.md`, `skill.yaml`, `VERSION`; `VERSION` is valid semver; `SKILL.md` has complete versioning frontmatter and a body `## Changelog` |
| `sync` | All platform `SKILL.md` files match the canonical `.agents/skills/` source |
| `lockfile` | `.agentic-template.lock` is present |

Exit code is non-zero when any `error`-level finding is present.

```text
  ✓  [toolchain ]  skpm found
  ✓  [bakefile  ]  agentic.bake.yaml is valid
  ✓  [targets   ]  3 target(s) defined
  ✓  [skills    ]  .agents/skills/requirements-analyst/SKILL.md valid
  !  [sync      ]  .claude/skills/architecture-reviewer/SKILL.md differs from canonical — run: skcr sync
  ✓  [lockfile  ]  .agentic-template.lock present

1 warning(s) found.
```

## `skcr export`

Reads `SKILL.md` from every skill in `.agents/skills/` and concatenates them into a single Markdown document. YAML frontmatter is stripped by default so the output is clean prose suitable for LLM context injection.

```bash
skcr export [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--target` | `.` | Repository path |
| `-o`, `--out` | stdout | Write output to file |
| `--skill` | | Export only this skill |
| `--in-target` | | Export only skills from this bake target |
| `--keep-frontmatter` | | Retain YAML frontmatter in output |

```bash
# Pipe directly into an LLM prompt or context file
skcr export > SKILLS.md

# Write to file (progress printed to stderr, content to file)
skcr export --out docs/SKILLS.md

# Only skills from the gitlab target
skcr export --in-target gitlab --out docs/gitlab-skills.md

# Single skill with frontmatter retained
skcr export --skill requirements-analyst --keep-frontmatter
```

Output format:

```markdown
# Agent Skills

> Generated by `skcr export` on 2026-06-10
> Source: `.agents/skills/`
> Skills: 18

---

# Requirements Analyst

## Purpose
...

---

# Architecture Reviewer
...
```

## `skcr scaffold skill`

Creates a standalone skill skeleton from CLI flags, independent of any bakefile.

```bash
skcr scaffold skill <name> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--output-dir` | `.` | Directory where the skill directory is created |
| `--version` | `0.1.0` | Initial semver version |
| `--description` | | Skill description for `skill.yaml` |
| `--owner` | | Skill owner |
| `--platform` | `claude-code`, `codex` | Compatible platform (repeatable) |
| `--license` | `MIT` | License |
| `--force` | | Overwrite existing files |
| `--dry-run` | | Preview without writing |

## Validation

```bash
skcr validate
skcr validate --platform codex
skcr validate --against-lock agent-skills.lock
skcr validate --skills
```

`skcr validate` checks:

- `agentic.bake.yaml` structure and platform names
- Platform names are valid (delegates to sklib)
- Generated platform files match the lockfile state
- `skill.yaml` and `SKILL.md` presence in discovered skill directories

## Skill lifecycle (with `skpm`)

`skcr` creates and scaffolds. `skpm` handles versioning, packaging, publishing, and installation.

```bash
# After editing a skill:
skpm validate .agents/skills/requirements-analyst
skpm package  .agents/skills/requirements-analyst
skpm publish  .agents/skills/requirements-analyst --source myregistry

# Install skills from a registry:
skpm add requirements-analyst@^1.0.0
skpm lock && skpm install
skcr bake --skills-from agent-skills.lock --write
```

## Boundary with `skpm`

`skcr` **may**:

- Create `agentic.bake.yaml`
- Scaffold skill directories (full structure)
- Render and sync platform-specific output files
- Track generated files in `.agentic-template.lock`
- Validate generated project state

`skcr` **must not**:

- Bump skill versions
- Package or publish skills
- Resolve registry versions or download artifacts
- Write `agent-skills.lock`
- Install, update, or verify skills from registries
- Talk to SkillForge, Artifactory, GitLab, GitHub, or other registries

## Clean

Removes only files tracked in `.agentic-template.lock`. Never deletes scaffolded skill directories.

```bash
skcr clean --plan
skcr clean --write
```
