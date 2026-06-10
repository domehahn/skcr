# GitLab Duo Flow Templates

These YAML files are source-of-truth templates for GitLab Duo Custom Flows.

They are not automatically activated by committing them to the repository.

To activate a flow, create or update it in GitLab through **AI > Flows** or the **AI Catalog** and paste the corresponding YAML content.

Each flow template passes:

- `context:inputs.user_rule` as `agents_dot_md`
- `context:inputs.workspace_agent_skills` as `workspace_agent_skills`

This is required so `AGENTS.md` and project-level skills are available to the flow agent.
