# Release Hardening & Pipeline Guide

## Release Process

The `skcr` release pipeline is fully automated and hardened via GitHub Actions (`.github/workflows/release.yml`) and GoReleaser (`.goreleaser.yaml`).

### Canonical Release Verification Sequence

```text
git tag vX.Y.Z
 ↓
source verification (go mod verify)
 ↓
unit & compiler tests (go test)
 ↓
race detector gate (make test-race)
 ↓
security vulnerability check (govulncheck)
 ↓
interop verification (skcr compile x skil validate)
 ↓
multi-platform compilation (linux/darwin/windows amd64/arm64)
 ↓
Software Bill of Materials (SBOM) generation
 ↓
SHA256SUMS checksum generation
 ↓
GitHub Release asset publication
```

## Release Artifacts

Every published release includes:
- Cross-platform binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`
- SHA256 checksums file (`skcr_<version>_checksums.txt`)
- Software Bill of Materials in SPDX JSON format (`skcr_<version>_<os>_<arch>.sbom.json`)
- GitHub Artifact Attestations for supply-chain integrity verification

