# Skill Artifact Specification

A current `skcr` skill is a coordinated set of source artifacts:

```text
Descriptor + Contract + Evals + Integrations + Execution Closure + Instructions
```

They answer different questions:

- **Goal:** What outcome should the skill achieve?
- **Instructions:** How should the model perform the work?
- **Contract:** What behavior is permitted while pursuing the Goal?
- **Eval:** Does observed behavior conform to the Goal and Contract?
- **Enforcement:** Can forbidden behavior actually be prevented at runtime?

`skcr` declares, scaffolds, validates, synchronizes, compares, and versions
artifacts. It does not observe runtime behavior, execute evals, apply policy, or
enforce permissions.

## Independently versioned specifications

| Artifact | Current schema | Public schema |
|---|---:|---|
| `descriptor.yaml` | Descriptor v2 | `schemas/descriptor-v2.schema.json` |
| `contract.yaml` | Contract v2 | `schemas/contract-v2.schema.json` |
| `evals/*.yaml` | Eval v2 | `schemas/eval-v2.schema.json` |
| `integrations/mcp.yaml` | MCP v1 | `schemas/mcp-v1.schema.json` |
| `integrations/a2a.yaml` | A2A v1 | `schemas/a2a-v1.schema.json` |
| `dependencies.yaml` | Dependencies v1 | `schemas/dependencies-v1.schema.json` |
| `assurance.yaml` | Assurance Requirements v1 | `schemas/assurance-v1.schema.json` |
| compiled `build-manifest.json` | Build Manifest v1 | `schemas/build-manifest-v1.schema.json` |

A future Contract or Eval version does not require changing the other schema
versions. Unknown current-schema fields and unsupported versions fail closed.

## Descriptor v2

`descriptor.yaml` owns identity, lifecycle/platform metadata, the Goal, and safe
relative references:

```yaml
schema_version: "2"
name: example
version: 1.0.0
description: Example skill.
entrypoint: SKILL.md
compatible_with: []
goal:
  objective: Produce an example result.
  success_criteria:
    - id: result-is-evidenced
      description: The result contains evidence.
  failure_conditions: []
contract:
  file: contract.yaml
evals:
  directory: evals
integrations:
  mcp: {file: integrations/mcp.yaml}
  a2a: {file: integrations/a2a.yaml}
dependencies: {file: dependencies.yaml}
assurance: {file: assurance.yaml}
```

Goal criteria use stable IDs so eval assertions and external evidence can refer
to them without interpreting prose. Contract and eval references must remain
inside the skill root. Absolute paths, traversal, and symlink escapes are
invalid.

### Deprecated descriptor security hints

Descriptor v2 still accepts the historical optional `security` object so
existing descriptors can be read and explicitly migrated without losing
metadata. It is deprecated compatibility information, not a permission model:

- independent consumers MUST ignore it when deciding whether behavior is
  permitted;
- `skcr` never derives Contract capabilities from it;
- newly scaffolded descriptors omit it;
- `skcr doctor` warns when it is present; and
- only `contract.yaml` defines required and allowed behavior.

A contradiction such as `security.requires_network: false` alongside a
Contract network allowlist does not narrow that Contract. The deprecated hint
has no authorization precedence at all. Authors should remove it after
reviewing the explicit default-deny Contract produced by migration.

## Contract v2 semantics

`contract.yaml` is the authoritative behavioral/security boundary. Contract v2
separates semantic requirements from concrete runtime authority:

```yaml
capabilities:
  semantic:
    required:
      repository.read: [source-files]
  runtime:
    required: {}
    allowed: {}
```

The runtime `required` set expresses what execution needs. Runtime `allowed`
expresses the maximum permitted boundary. A runtime may compare:

```text
required ⊆ available ⊆ allowed
```

`skcr` validates `runtime.required ⊆ runtime.allowed`; it does not discover
`available` and never expands a semantic capability such as `repository.read`
into filesystem or network authority.

Runtime v2 covers filesystem read/write/delete, inbound/outbound network hosts,
structured command rules, explicit secret IDs and exposure, environment reads,
tools, MCP servers/tools, persistence, autonomous actions, external effects,
confirmation requirements, targets, and resource budgets. Effects classify
tools as `pure`, `read`, `reversible_write`, `irreversible_write`,
`external_side_effect`, or `destructive`; destructive and irreversible effects
require an explicit approval rule.

Identity requirements constrain accepted principals and credentials by maximum
TTL, audience binding, token passthrough, and required scopes. Delegation is
default-denied and, when enabled, requires a positive bounded depth,
authenticated origin, and child-capability subset enforcement. These are
requirements for downstream verification/enforcement, never claims that the
runtime already satisfies them.

Security-relevant value semantics are normative:

- **Missing required field:** invalid.
- **`null`:** explicitly unspecified, and accepted only for limit values.
- **`[]` in an allow/scope list:** allow nothing.
- **`[]` in a deny list:** explicitly deny no named tools; default deny still
  applies to tools absent from `allow`.
- **Explicit value:** only that normalized value or scope is declared.
- **Wildcard-containing value:** interpreted only by the downstream runtime for
  fields whose target schema supports patterns; `skcr` never broadens it.
- **Boolean `false`:** capability is not required/permitted.
- **Boolean `true`:** capability is required/permitted, subject to subset rules.

Tool precedence is:

```text
explicit deny > explicit allow > default deny
```

Tool identifiers are opaque runtime identities. Capabilities are vendor-neutral
semantic permissions. Provider-specific mappings such as a concrete update tool
implying `repository.write` belong to downstream verification systems.

Resources have `scope: invocation`, meaning one execution of one skill for one
requested task. `null` means no numeric bound is declared. Zero means zero uses
are permitted. Positive integers are inclusive upper bounds. Negative and
malformed values are invalid.

Data boundaries declare classifications, purposes, sensitivity, retention,
egress destinations, and explicit source-to-sink flows. Every Contract-v2
classification requires a policy. They are declarations for future
verification, not taint tracking.

## Agentic integrations and reviewed execution closure

`integrations/mcp.yaml` binds MCP server identity, HTTPS endpoint,
authentication posture, audience binding, token-passthrough policy, and tool
allow/deny sets. Contract runtime MCP authority must resolve to these declared
servers and fully-qualified tools.

`integrations/a2a.yaml` declares incoming and outgoing agent relationships,
bounded delegation, authenticated origin, capability-subset enforcement, and
the rule that incoming outputs remain untrusted until explicitly promoted.
A2A declarations may narrow but never expand Contract delegation authority.

`dependencies.yaml` describes the reviewed execution closure. Package versions
must be exact; remote resources require HTTPS and SHA-256 digests; container
images must be digest-pinned; MCP identities must reference a declared server
and carry an identity digest. `latest`, ranges, and unbound remote resources
fail validation.

## ASPS and assurance requirements

`assurance.yaml` declares requirements, never results. It can request pinned
ASPS v1.0 profiles/properties, a minimum requested assurance level, behavioral
and adversarial evaluation, containment, native isolation, and provenance.
Unknown profiles and properties fail closed. A strict parser rejects result
claims such as `verified` or an achieved `assurance_level`.

```yaml
schema_version: "1"
asps:
  specification: ASPS
  version: "1.0"
  required_profiles: [asps-core@1.0]
  required_properties: [ASP-08.03]
assurance:
  minimum_requested_level: A1
  requirements:
    behavioral_eval: true
    adversarial_eval: true
    containment: true
    native_isolation: true
    provenance: true
security_review:
  expansion_approvals: []
```

The authoring commands are:

```bash
skcr asps list
skcr asps show ASP-08.03
skcr asps validate <skill>
skcr asps requirements <skill>
skcr asps coverage <skill>
```

Coverage reports only which source declarations route to requested properties.
It is not a security verdict. PASS/FAIL evidence and attained assurance levels
remain exclusively owned by `skil`.

## Security expansion review and versioning

`skcr version recommend <skill>` combines material source changes with the
Contract security-impact diff. Capability expansion, weaker credential or
delegation constraints, and longer retention recommend a major bump. Goal,
Eval, integration, dependency, and assurance-requirement changes recommend a
minor bump; ordinary documentation changes recommend a patch.

```bash
skcr version bump <skill> --auto --change "Describe the change"
```

A detected security expansion is blocked unless explicitly reviewed:

```bash
skcr version bump <skill> --auto \
  --approve-security-expansion \
  --approved-by security-team \
  --change "Allow reviewed deployment API access"
```

The approval binds reviewer, date, justification, and the canonical Contract
digest in `assurance.yaml`; it does not claim that verification passed.

Human approval rules declare an action and one of `per_action`,
`per_invocation`, or `per_session`. They do not implement an approval runtime.

## Normalization, digest, and comparison

Normalization sorts and deduplicates set-like lists, normalizes filesystem
separators, and orders identified rules deterministically. Ordinary authoring
commands do not rewrite user formatting.

```bash
skcr contract digest path/to/skill
skcr contract diff old/contract.yaml new/contract.yaml
skcr contract diff old new --json
```

Diff classification is deterministic and deliberately conservative:

- `NARROWING`
- `NO_SECURITY_IMPACT`
- `EXPANSION`
- `MIXED`

It covers capability scopes, tools, network destinations, limits, invariants,
and approval requirements. Required-capability changes are reported but do not
by themselves expand the allowed security boundary. The classification is not a
substitute for semantic security review.

## Compatibility

Contract v1 remains readable, validatable, diffable, and compilable. New
scaffolds use Contract v2. Legacy descriptors without `schema_version` retain legacy parsing behavior and
do not acquire implicit permissions. The legacy Descriptor v2 filename and the
earlier form that embedded `contract` in `skill.yaml` remain readable and are
not automatically
migrated by bake, sync, doctor, status, or version commands.

No automatic migration is performed. `skcr migrate skill <name>` is the explicit
path: it preserves `SKILL.md`, creates split artifacts, and starts from default
deny rather than inferring permissions from natural-language instructions.

## Downstream interoperability

`skil` consumes compiled target artifacts, never the richer source descriptor
as if it were a native runtime contract:

```text
descriptor + contract + evals + integrations + dependencies + SKILL.md
                              │
                              ▼
             skcr compile --target skil
                              │
                              ▼
       .skcr/build/skil/<name>/skill.yaml + evals/*.yaml
                              │
                              ▼
                            skil
               scan / verify / eval / assure /
                    policy / enforce
```

Compilation fails when a security-relevant source declaration has no safe,
unambiguous target mapping. Goal/invariant references, identity, delegation,
data policies, integrations, reviewed-closure, and assurance requirements that `skil`
cannot yet execute are digest-bound in `build-manifest.json` as
`verification_only`.
`--require-lossless` rejects semantics that can neither be mapped nor preserved.
The manifest also records the compiler identity, aggregate source-artifact
digest, and mapping digest so source-to-target provenance is reproducible.
