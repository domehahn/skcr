# Security Policy

## Canonical Responsibility Boundaries

`skcr` is strictly the **Build / Authoring / Compilation Plane** for AI Agent Skills.

Its sole responsibility is:
- Scaffolding portable skill source structures
- Authoring skill contracts, descriptors, and eval definitions
- Validating canonical skill metadata schemas
- Performing deterministic source-to-target compilation
- Emitting deterministic build provenance (`build-manifest.json`) and SHA256 checksums

`skcr` is explicitly **NOT**:
- A security scanner (delegated to `skil`)
- A package manager / dependency resolver (delegated to `skpm`)
- An admission controller or policy engine (delegated to `skgate`)
- A skill registry (delegated to `SkillForge`)
- A runtime execution sandbox (delegated to `skrun`)

## Reporting a Vulnerability

If you discover a security vulnerability in `skcr`, please report it privately to the platform engineering security team via GitHub Security Advisories or direct email.

Do not create public issues for unreleased security vulnerabilities.

## Security Controls & Guarantees

1. **Path Traversal & Filesystem Safety**: `skcr` strictly enforces relative boundary constraints (`SanitizeRelativePath`), rejecting absolute path escapes, `..` traversal attacks, and symlink overwrites.
2. **Deterministic Compilation**: Identical source inputs and compilation parameters produce 100% byte-identical compiled artifacts and manifests.
3. **Build Provenance**: All compiled assurance artifacts include deterministic build provenance metadata (`schema_version: "1.0.0"`) with cryptographic `source_digest` and `compiled_digest`.
4. **Supply Chain Security**: All CI/CD workflows pin GitHub Actions to immutable commit SHAs, run mandatory Go race detector gates, automated `govulncheck` vulnerability scans, and produce SLSA provenance attestations upon release.

