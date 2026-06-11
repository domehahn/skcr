# Skill Content Readiness

## SDLC / DevSecOps Skill Library

The built-in SDLC and DevSecOps skill library is registered as inline Go templates. The current production-ready library contains:

- `requirements-analyst`
- `cost-based-planner`
- `architecture-reviewer`
- `threat-modeler`
- `safe-implementer`
- `test-strategy-engineer`
- `verification-reviewer`
- `security-reviewer`
- `secrets-reviewer`
- `dependency-supply-chain-reviewer`
- `ci-cd-reviewer`
- `iac-gitops-reviewer`
- `compliance-governance-reviewer`
- `release-readiness-reviewer`
- `observability-reviewer`
- `incident-postmortem-assistant`
- `documentation-maintainer`
- `universal-skill-creator`

## Framework-Ready

A skill is framework-ready when it has the required file structure and machine-readable metadata, but its body is still mostly generic. Framework-ready skills can be scaffolded, synced, packaged, and routed, but they are not sufficient for production use.

Framework-ready indicators:

- complete YAML frontmatter,
- valid `## Changelog`,
- platform compatibility fields present,
- generic operating model and guardrails,
- no domain-specific checklist, decision rules, acceptance criteria, or output requirements.

## Content-Ready

A skill is content-ready when it can guide real work in its domain without relying on placeholder text or copy-pasted boilerplate.

Content-ready skills must include:

- `## Purpose`,
- `## When to use`,
- `## Operating model`,
- `## Skill-Specific Review Scope`,
- `## Skill-Specific Checklist`,
- `## Decision Rules`,
- `## Finding Categories`,
- `## Severity Guidance`,
- `## DevSecOps Guardrails`,
- `## Output Requirements`,
- `## Acceptance Criteria`,
- `## Anti-Patterns`,
- `## Changelog`,
- complete frontmatter and body changelog,
- security and governance guardrails,
- concrete validation and output expectations.

Minimum content counts:

- Skill-specific checklist: at least 10 concrete items.
- Decision rules: at least 5 concrete rules.
- Finding categories: at least 5 concrete categories.
- Severity guidance: at least 4 severity levels.
- Output requirements: at least 5 required output elements.
- Acceptance criteria: at least 5 criteria.
- Anti-patterns: at least 5 anti-patterns.

Generic copy-paste is not acceptable. Two skills must not have identical skill-specific content blocks. Shared governance language is allowed as a baseline, but the review scope, checklist, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns must be domain-specific.

The `universal-skill-creator` is held to an additional rule: it must never create a skill that only differs by name and description, and every generated skill must include domain-specific review scope, checklist items, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns.

## Source Of Truth

For built-in SDLC and DevSecOps skills, the source of truth is `internal/scaffold/skill_templates.go`.

The inline registry contains:

- `SDLCSkillNames`: the ordered list of built-in skill names.
- `sdlcSkillContent`: the domain-specific content for each skill.
- `skillBodies`: the rendered template map used by scaffolding.
- `renderSkillTemplate`: the renderer used by `PlanSkill`.
- `skillFrontmatter`: the shared metadata baseline for all built-in skills.
- `internal/platforms/compatibility.go`: the central production compatibility matrix used for `min_platform_version`.

`internal/scaffold/skill.go` contains the fallback template used for ad hoc skill names that are not in the built-in registry. That fallback must also satisfy the required section structure.

`internal/renderer/templates/shared/SKILL.md.j2` is the renderer-facing shared template and must stay aligned with the same quality bar. Do not introduce package-manager, registry, installer, lockfile, SkillForge, or skpm behavior into these SDLC skill templates.

Generated platform directories such as `.agents/skills/`, `.claude/skills/`, `.cursor/skills/`, `.windsurf/skills/`, and top-level `skills/` are generated outputs or platform copies. They should not be treated as the authoritative source for built-in SDLC template content.

## Platform Compatibility

Built-in production skills must use concrete `min_platform_version` values from `internal/platforms/compatibility.go`.

Do not hand-edit per-skill platform versions. Update the central matrix after compatibility validation, then let the shared renderer propagate the value. `unknown` is reserved for incomplete custom-skill drafts and must be treated as a warning, not as verified production compatibility.

## Adding A Skill Template

To add a new built-in skill template:

1. Add the skill name to `SDLCSkillNames`.
2. Add a matching `skillContent` entry in `sdlcSkillContent`.
3. Fill every required section with domain-specific content.
4. Keep shared DevSecOps guardrails generic enough to apply broadly, but make the review scope, checklist, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns specific to the skill.
5. Add domain-term assertions when the skill has non-negotiable vocabulary or risk areas.
6. Confirm the central compatibility matrix contains concrete minimum versions for every supported platform.
7. Run the scaffold and validator tests before release.

## Tests

Run the full suite before publishing changes:

```sh
go test ./...
```

The scaffold tests enforce registration, frontmatter completeness, required section presence, minimum section counts, uniqueness, missing placeholders, required domain terms for key skills, and the universal skill creator anti-generic rule. Validator tests enforce the required readiness headings for rendered skill files.
