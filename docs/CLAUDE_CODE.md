# Claude Code Integration

The Claude Code adapter generates both skills and subagents.

## Generated structure

```text
CLAUDE.md
.claude/
├── skills/
│   └── <skill-name>/
│       └── SKILL.md
└── agents/
    └── <subagent-name>.md
```

## Skills vs. Subagents

Skills are reusable instruction bundles.

```text
.claude/skills/<skill-name>/SKILL.md
```

Subagents are specialized Claude Code workers with their own context, tool access, model, permissions, and system prompt.

```text
.claude/agents/<subagent-name>.md
```

## Generated subagents

The DevSecOps SDLC edition generates these Claude Code subagents when the `claude` platform is enabled:

- `requirements-analyst`
- `architecture-reviewer`
- `devsecops-reviewer`
- `security-reviewer`
- `ci-cd-reviewer`
- `iac-gitops-reviewer`
- `test-runner`
- `release-readiness-reviewer`
- `incident-postmortem-assistant`

## Invocation

Claude Code can delegate automatically when a task matches a subagent description.

You can also ask explicitly:

```text
Use the security-reviewer subagent to review this change.
```

or:

```text
Use the test-runner subagent to run relevant tests and summarize failures.
```

## Design rules

- Keep review subagents mostly read-only.
- Use explicit tools.
- Use `skills:` to preload relevant skill instructions into subagent context.
- Prefer bounded, specialized subagents over generic workers.
- Do not use subagents as a replacement for human review on security-sensitive changes.

## Generated files

When the `claude` platform is enabled, the generator creates:

```text
CLAUDE.md
.claude/skills/<skill-name>/SKILL.md
.claude/agents/<subagent-name>.md
```

## Relationship to the DevSecOps SDLC skill set

The Claude Code adapter uses the same SDLC skill catalog as the other platform adapters.

Skills provide reusable task instructions.

Subagents provide bounded specialist workers that can preload one or more skills.

Example:

```yaml
skills:
  - security-reviewer
  - secrets-reviewer
  - dependency-supply-chain-reviewer
```

This means the subagent receives those skill instructions as part of its working context.

## Recommended usage

Use skills for direct reusable guidance.

Use subagents for bounded side tasks such as:

- reviewing a large diff
- analyzing test failures
- checking CI/CD security
- reviewing Terraform or Kubernetes changes
- summarizing logs
- preparing incident analysis
- performing release readiness review

## Safety expectations

Subagents should not be treated as autonomous production actors.

They should:

- stay within the delegated task
- avoid secrets and sensitive files
- avoid destructive commands
- avoid broad rewrites unless explicitly requested
- return concise findings to the main Claude Code session