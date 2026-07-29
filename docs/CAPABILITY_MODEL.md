# Capability Model

Contract v1 uses stable, vendor-neutral capability identifiers. Scopes narrow a
capability; an empty scope list permits nothing.

## Normative vocabulary

### `repository.read`

Obtain repository content or metadata without changing repository state.
Scopes are opaque semantic resource classes such as `source_files` or
`pull_requests`. Includes reading files, diffs, commits, and review metadata.
Does not include commit, update, merge, or delete.

### `repository.write`

Change repository-controlled content or metadata. Scopes identify permitted
resource classes. Includes creating or updating files, branches, commits, or
review metadata. Contract v1 does not separately model delete or merge; those
future narrower capabilities must not be inferred as allowed merely from prose.

### `filesystem.read`

Read filesystem content or metadata. Scopes are normalized paths; Contract v1
does not define wildcard path matching.
Repository reads performed through a filesystem use both the runtime-observed
filesystem semantics and any repository semantics determined by the verifier.

### `filesystem.write`

Create or modify filesystem content or metadata. Scopes are normalized paths.
Contract v1 does not separately model filesystem deletion.

### `network.connect`

Initiate an outbound network connection. Scopes are normalized destination
identifiers, normally hostnames. `allow: []` permits no destinations.

### `process.execute`

Start or invoke a process or command. This capability is boolean in Contract v1
and has no scopes.

### `secrets.read`

Obtain secret or credential material. This capability is boolean and does not
authorize egress; data-flow restrictions apply separately.

### `tool.invoke`

Invoke a concrete runtime tool. Concrete identities live in `tools.allow` and
`tools.deny`; `tool.invoke` is the semantic vocabulary used by conditions,
approval rules, observations, and policy results.

## Required, allowed, and available

- **Required:** necessary to achieve the Goal.
- **Allowed:** maximum declared boundary.
- **Available:** supplied by a runtime; not discovered by `skcr`.

Contract v1 requires required scopes and booleans to be subsets of allowed
scopes and booleans. Observed use outside allowed is a contract violation.

Capabilities not listed above are unknown in Contract v1. Provider-specific tool
names remain opaque and do not extend the normative capability vocabulary.
