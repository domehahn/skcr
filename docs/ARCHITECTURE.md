# Architecture

Skill Creator (`skcr`) has five core responsibilities:

1. Parse `agentic.bake.yaml`
2. Resolve targets and inheritance
3. Render platform-specific templates
4. Validate source skill artifacts
5. Compile source artifacts into target-native assurance artifacts

The compiler boundary is explicit: `descriptor.yaml`, `contract.yaml`, source
evals, MCP/A2A integrations, `dependencies.yaml`, `assurance.yaml`, and `SKILL.md` are authoring inputs. A `skil` build emits a distinct
native `skill.yaml`; `skcr` does not implement scanning, verification, runtime
policy, or enforcement itself.

Contract v2 deliberately keeps semantic intent separate from runtime authority.
Only `capabilities.runtime.allowed` is compiled into the native `skil`
capability boundary; `capabilities.semantic.required` remains authoring and
verification metadata.

Agentic-security declarations that skil Contract v1 cannot execute—identity,
credential, delegation, data-governance, integration, and reviewed-closure
requirements—are validated against one another and committed into the build
manifest with source digests. They are marked `verification_only`; skcr never
turns requirements into assurance claims.

The ASPS catalog is pinned to v1.0 snapshot `2026-07-31`. `skcr asps
coverage` reports declaration routing only. Security verdicts, attained levels,
attestations, and runtime evidence are downstream `skil` outputs. Security-aware
versioning requires a digest-bound human review before a detected Contract
expansion can be auto-bumped.

## Responsibility boundary

| Function | skcr | skil | skpm |
|---|---:|---:|---:|
| Scaffold and source authoring | owns | — | — |
| Contract/Eval requirements | owns | consumes/executes | packages |
| Static scan, verification, policy, enforcement | — | owns | — |
| Assurance evidence and attained levels | — | owns | transports |
| Package, publish, install, update, revocation distribution | — | admission input | owns |

An `assurance.yaml` requirement or coverage route therefore cannot be promoted
to an assurance result by skcr. Only skil evidence can cross that boundary.

Platform adapters are implemented in the renderer by mapping one platform to its native output paths.

## Output path model

| Platform | Main instruction file | Capability files |
|---|---|---|
| Shared | `AGENTS.md` | Platform index for active targets |
| Codex | `.agentic/codex/AGENTS.md` | `.agents/skills/<skill>/SKILL.md` |
| GitLab Duo | `.agentic/gitlab-duo/AGENTS.md` | `skills/<skill>/SKILL.md`, `.gitlab/duo/chat-rules.md`, `.gitlab/duo/flows/*.yaml` |
| Claude Code | `.agentic/claude/AGENTS.md` | `.claude/skills/<skill>/SKILL.md`, `.claude/agents/*.md` |
| GitHub Copilot | `.github/copilot-instructions.md` | `.github/prompts/*.prompt.md` |
| OpenHands | `.agentic/openhands/AGENTS.md`, `.openhands/instructions.md` | capability registry in instructions |
| OpenCode | `.agentic/opencode/AGENTS.md`, `.opencode/instructions.md` | capability registry in instructions |
| Ollama | `.agentic/ollama/AGENTS.md`, `.ollama/Modelfile` | `.ollama/README.md` |
