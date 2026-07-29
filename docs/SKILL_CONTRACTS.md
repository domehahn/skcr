# Skill Artifact Specification

A current `skcr` skill is four coordinated artifacts:

```text
Skill Descriptor  +  Behavioral Contract  +  Eval Specifications  +  Instructions
skill.yaml           contract.yaml           evals/*.yaml             SKILL.md
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
| `skill.yaml` | Descriptor v2 | `schemas/skill-v2.schema.json` |
| `contract.yaml` | Contract v1 | `schemas/contract-v1.schema.json` |
| `evals/*.yaml` | Eval v1 | `schemas/eval-v1.schema.json` |

A future Contract or Eval version does not require changing the other schema
versions. Unknown current-schema fields and unsupported versions fail closed.

## Descriptor v2

`skill.yaml` owns identity, lifecycle/platform metadata, the Goal, and safe
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
```

Goal criteria use stable IDs so eval assertions and external evidence can refer
to them without interpreting prose. Contract and eval references must remain
inside the skill root. Absolute paths, traversal, and symlink escapes are
invalid.

## Contract v1 semantics

`contract.yaml` is the authoritative behavioral/security boundary. Required
capabilities express what the skill needs. Allowed capabilities express the
maximum permitted boundary. A future runtime may compare:

```text
required ⊆ available ⊆ allowed
```

`skcr` validates `required ⊆ allowed`; it does not discover `available`.

Security-relevant value semantics are normative:

- **Missing required field:** invalid.
- **`null`:** explicitly unspecified, and accepted only for limit values.
- **`[]` in an allow/scope list:** allow nothing.
- **`[]` in a deny list:** explicitly deny no named tools; default deny still
  applies to tools absent from `allow`.
- **Explicit value:** only that normalized value or scope is declared.
- **Wildcard-containing value:** invalid in security-sensitive allow/deny lists
  in Contract v1; patterns must not contain `*`.
- **Boolean `false`:** capability is not required/permitted.
- **Boolean `true`:** capability is required/permitted, subject to subset rules.

Tool precedence is:

```text
explicit deny > explicit allow > default deny
```

Tool identifiers are opaque runtime identities. Capabilities are vendor-neutral
semantic permissions. Provider-specific mappings such as a concrete update tool
implying `repository.write` belong to downstream verification systems.

Limits have `scope: invocation`, meaning one execution of one skill for one
requested task. `null` means no numeric bound is declared. Zero means zero uses
are permitted. Positive integers are inclusive upper bounds. Negative and
malformed values are invalid.

Data boundaries declare classifications, egress destinations, and explicit
source-to-sink flows. They are declarations for future verification, not taint
tracking.

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

Legacy descriptors without `schema_version` retain legacy parsing behavior and
do not acquire implicit permissions. The earlier descriptor-v2 form that
embedded `contract` in `skill.yaml` remains readable and is not automatically
migrated by bake, sync, doctor, status, or version commands.

No automatic migration is performed. `skcr migrate skill <name>` is the explicit
path: it preserves `SKILL.md`, creates split artifacts, and starts from default
deny rather than inferring permissions from natural-language instructions.

## Downstream interoperability

Independent systems such as `skil` consume the artifact files and public JSON
schemas; they do not need to import `skcr`.

```text
skcr → skill.yaml + contract.yaml + evals/*.yaml
                                      │
                                      ▼
                                    skil
                          scan / observe / verify /
                          eval / policy / enforce
```
