# skcr Public Compatibility & Contract Specification

This document defines the stable public API contracts, schema versioning policies, CLI exit codes, and cross-toolchain identity rules for `skcr` (the **Build / Authoring / Compilation Plane** for AI Agent Skills).

---

## 1. Schema Specifications & Versioning

`skcr` strictly enforces schema versioning and semantic versioning rules across all source, compiled, and metadata artifacts.

### 1.1 `descriptor.yaml` (Source Descriptor)
- **Schema Version**: `2.0.0`
- **Required Fields**: `schema_version`, `name`, `version`, `description`, `contract.file`
- **Rules**: Must be a split descriptor v2 artifact. Unrecognized root fields are strictly rejected.

### 1.2 `contract.yaml` (Source Contract)
- **Schema Version**: `2.0.0`
- **Capabilities Scope**: Runtime (`allowed` / `required`), Resources, Delegation, Retention, Identity
- **Rules**: Explicit deny-by-default runtime model. Unrecognized fields trigger fail-closed validation.

### 1.3 Generated `skill.yaml` (Compiled Target Contract)
- **Target Schema Version**: `1` (for `skil` Target)
- **Structure**: Denormalized runtime view mapping capabilities (`filesystem`, `network`, `commands`, `secrets`, `environment`, `tools`, `mcp`, `agent`).

### 1.4 Evals (`evals/*.yaml`)
- **Schema Version**: `2.0.0`
- **Types**: `behavioral`, `adversarial`
- **Containment Rules**: Assertions on permitted/forbidden capabilities and tool access.

---

## 2. Build Identity Contract & Provenance

Every compilation pass produces a `build-manifest.json` adhering to `schema_version: "1.0.0"`:

```json
{
  "schema_version": "1.0.0",
  "tool": "skcr",
  "tool_version": "2.0.0",
  "source_digest": "sha256:...",
  "compiled_digest": "sha256:...",
  "target": "skil",
  "target_name": "skil",
  "contract_version": "2.0.0",
  "build_parameters": {
    "require_lossless": false
  },
  "build_parameters_digest": "sha256:..."
}
```

### Digest Determination Rules
- **`source_digest`**: Cryptographic hash (`sha256`) of canonical byte representations of `descriptor.yaml`, `contract.yaml`, sorted `evals/*.yaml`, `SKILL.md`, integrations, dependencies, and assurance files.
- **`compiled_digest`**: Hash of sorted compiled files (`SKILL.md`, `skill.yaml`, `VERSION`, `CHANGELOG.md`, `evals/*.yaml`).
- **`build_parameters_digest`**: Hash of canonical JSON representation of compilation options.

---

## 3. CLI Exit Codes & Flag Contracts

| Code | Name | Trigger |
| :---: | :--- | :--- |
| **0** | `SUCCESS` | Operation completed successfully |
| **1** | `ERROR_GENERAL` | Internal or runtime execution error |
| **2** | `ERROR_USAGE` | Invalid command arguments or flags |
| **3** | `ERROR_SCHEMA` | Source descriptor, contract, or eval validation failure |
| **4** | `ERROR_LOSSLESS` | Lossy transformation detected when `--require-lossless` is set |

---

## 4. SemVer & Schema Evolution Policy

1. **Patch Releases (`v2.0.x`)**: Bug fixes, performance enhancements; zero schema or digest calculation changes.
2. **Minor Releases (`v2.x.0`)**: Backward-compatible schema additions; new optional fields supported.
3. **Major Releases (`v3.0.0`)**: Breaking contract changes; new `schema_version` requirements.
