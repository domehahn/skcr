# GitLab Duo Agent Platform

Generated GitLab files:

```text
AGENTS.md
skills/<skill-name>/SKILL.md
.gitlab/duo/chat-rules.md
.gitlab/duo/flows/*.yaml
.gitlab/duo/flows/README.md
```

## Skills

Project-level skills are generated under:

```text
skills/<skill-name>/SKILL.md
```

Each skill includes:

```yaml
metadata:
  slash-command: enabled
```

## Custom Rules

Project custom rules are generated as:

```text
.gitlab/duo/chat-rules.md
```

## Flows

Flow files under `.gitlab/duo/flows/*.yaml` are templates/source-of-truth. They are not automatically activated by GitLab simply because they exist in the repo.

To use them, create or update Custom Flows in GitLab and paste the generated YAML.

Each flow passes:

```yaml
inputs:
  - from: "context:inputs.user_rule"
    as: "agents_dot_md"
    optional: true
  - from: "context:inputs.workspace_agent_skills"
    as: "workspace_agent_skills"
    optional: true
```
